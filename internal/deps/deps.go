// Package deps строит граф импортов от входных файлов (паритет deps.service.ts).
//
// Реализация best-effort: модульные спецификаторы извлекаются через tree-sitter,
// а пути разрешаются вручную — относительные импорты, расширения (.ts/.tsx/.d.ts
// /.js/.jsx), index-файлы и алиасы tsconfig (paths/baseUrl). Не покрывается
// экзотика (conditional exports, package.json "exports"); такие импорты, как и
// node_modules, пропускаются.
package deps

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Options управляют обходом графа.
type Options struct {
	MaxDepth     int  // <= 0 — без ограничения
	IncludeTypes bool // включать ли import type
}

// Result — итог обхода.
type Result struct {
	Files   []string          // относительные POSIX-пути в порядке обхода
	Explain map[string]string // abs-путь → причина включения
}

type collector struct {
	root         string
	maxDepth     int
	includeTypes bool
	tsBaseDir    string
	tsPaths      map[string][]string
	visited      map[string]bool
	order        []string
	explain      map[string]string
}

// Collect обходит граф импортов от entryFiles.
func Collect(entryFiles []string, projectRoot string, opts Options) Result {
	c := &collector{
		root:         projectRoot,
		maxDepth:     opts.MaxDepth,
		includeTypes: opts.IncludeTypes,
		visited:      map[string]bool{},
		explain:      map[string]string{},
	}
	c.tsBaseDir, c.tsPaths = loadTsconfig(projectRoot)

	for _, f := range entryFiles {
		abs := f
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(projectRoot, f)
		}
		abs = filepath.Clean(abs)
		c.explain[abs] = "entry point"
		c.collect(abs, 0)
	}

	files := make([]string, 0, len(c.order))
	for _, abs := range c.order {
		files = append(files, toRel(projectRoot, abs))
	}
	return Result{Files: files, Explain: c.explain}
}

func (c *collector) collect(abs string, depth int) {
	if c.visited[abs] {
		return
	}
	c.visited[abs] = true
	c.order = append(c.order, abs)

	if c.maxDepth > 0 && depth >= c.maxDepth {
		return
	}

	content, err := os.ReadFile(abs)
	if err != nil {
		return
	}

	relFrom := toRel(c.root, abs)
	for _, imp := range extractImports(content, abs) {
		if imp.typeOnly && !c.includeTypes {
			continue
		}
		resolved := c.resolve(imp.spec, abs)
		if resolved == "" || strings.Contains(resolved, "node_modules") {
			continue
		}
		if _, ok := c.explain[resolved]; !ok {
			c.explain[resolved] = importReason(imp, relFrom)
		}
		c.collect(resolved, depth+1)
	}
}

type importRef struct {
	spec     string
	typeOnly bool
	reExport bool
}

func importReason(imp importRef, relFrom string) string {
	switch {
	case imp.reExport:
		return "re-export from " + relFrom
	case imp.typeOnly:
		return "type import from " + relFrom
	default:
		return "import from " + relFrom
	}
}

var typeOnlyRe = regexp.MustCompile(`^(?:import|export)\s+type\b`)

// extractImports достаёт спецификаторы из import/export ... from.
func extractImports(src []byte, filename string) []importRef {
	parser := sitter.NewParser()
	parser.SetLanguage(grammarFor(filename))
	tree, err := parser.ParseCtx(context.Background(), nil, src)
	if err != nil || tree == nil {
		return nil
	}
	defer tree.Close()

	var refs []importRef
	root := tree.RootNode()
	for i := 0; i < int(root.ChildCount()); i++ {
		n := root.Child(i)
		t := n.Type()
		if t != "import_statement" && t != "export_statement" {
			continue
		}
		srcNode := n.ChildByFieldName("source")
		if srcNode == nil {
			continue // export const ... — без источника
		}
		spec := strings.Trim(string(src[srcNode.StartByte():srcNode.EndByte()]), "\"'`")
		if spec == "" {
			continue
		}
		stmt := string(src[n.StartByte():n.EndByte()])
		refs = append(refs, importRef{
			spec:     spec,
			typeOnly: typeOnlyRe.MatchString(stmt),
			reExport: t == "export_statement",
		})
	}
	return refs
}

// resolve превращает спецификатор в абсолютный путь к файлу или "".
func (c *collector) resolve(spec, fromFile string) string {
	if strings.HasPrefix(spec, ".") {
		base := filepath.Join(filepath.Dir(fromFile), spec)
		return firstExisting(resolveCandidates(base))
	}
	// Алиасы tsconfig paths.
	if c.tsBaseDir != "" {
		if p := c.resolveTsPaths(spec); p != "" {
			return p
		}
		// baseUrl: import "src/foo" → <baseDir>/src/foo
		base := filepath.Join(c.tsBaseDir, spec)
		if r := firstExisting(resolveCandidates(base)); r != "" {
			return r
		}
	}
	return ""
}

func (c *collector) resolveTsPaths(spec string) string {
	for pattern, targets := range c.tsPaths {
		star := strings.IndexByte(pattern, '*')
		if star < 0 {
			if pattern != spec {
				continue
			}
			for _, tgt := range targets {
				if r := firstExisting(resolveCandidates(filepath.Join(c.tsBaseDir, tgt))); r != "" {
					return r
				}
			}
			continue
		}
		prefix, suffix := pattern[:star], pattern[star+1:]
		if !strings.HasPrefix(spec, prefix) || !strings.HasSuffix(spec, suffix) {
			continue
		}
		matched := spec[len(prefix) : len(spec)-len(suffix)]
		for _, tgt := range targets {
			resolvedTgt := strings.Replace(tgt, "*", matched, 1)
			if r := firstExisting(resolveCandidates(filepath.Join(c.tsBaseDir, resolvedTgt))); r != "" {
				return r
			}
		}
	}
	return ""
}

var resolveExts = []string{".ts", ".tsx", ".d.ts", ".js", ".jsx"}

// sourceExts — расширения, которые считаем исходным кодом при точном совпадении
// пути. .json и прочее намеренно исключены: ts-morph без resolveJsonModule их
// не подхватывает, и для упаковки кода они не нужны.
var sourceExts = map[string]bool{
	".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".mjs": true, ".cjs": true,
}

func hasSourceExt(p string) bool {
	if strings.HasSuffix(p, ".d.ts") {
		return true
	}
	return sourceExts[strings.ToLower(filepath.Ext(p))]
}

// resolveCandidates перечисляет возможные файлы для базового пути.
func resolveCandidates(p string) []string {
	var cands []string
	// Точный путь принимаем только если это исходный файл (не .json и т.п.).
	if hasSourceExt(p) {
		cands = append(cands, p)
	}
	for _, e := range resolveExts {
		cands = append(cands, p+e)
	}
	for _, e := range resolveExts {
		cands = append(cands, filepath.Join(p, "index"+e))
	}
	// .js → .ts/.tsx сиблинг (NodeNext-стиль).
	if strings.HasSuffix(p, ".js") {
		base := strings.TrimSuffix(p, ".js")
		cands = append(cands, base+".ts", base+".tsx")
	}
	if strings.HasSuffix(p, ".jsx") {
		cands = append(cands, strings.TrimSuffix(p, ".jsx")+".tsx")
	}
	return cands
}

func firstExisting(cands []string) string {
	for _, c := range cands {
		info, err := os.Stat(c)
		if err == nil && info.Mode().IsRegular() {
			return filepath.Clean(c)
		}
	}
	return ""
}

func toRel(root, abs string) string {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return filepath.ToSlash(abs)
	}
	return filepath.ToSlash(rel)
}

func grammarFor(filename string) *sitter.Language {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".tsx", ".jsx":
		return tsx.GetLanguage()
	default:
		return typescript.GetLanguage()
	}
}
