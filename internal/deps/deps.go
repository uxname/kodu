// Package deps builds an import graph starting from the entry files (parity with deps.service.ts).
//
// The implementation is best-effort: module specifiers are extracted via tree-sitter,
// and paths are resolved by hand — relative imports, extensions (.ts/.tsx/.d.ts
// /.js/.jsx), index files, and tsconfig aliases (paths/baseUrl). Exotic cases
// (conditional exports, package.json "exports") are not covered; such imports, like
// node_modules, are skipped.
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

// Options control the graph traversal.
type Options struct {
	MaxDepth     int  // <= 0 — unlimited
	IncludeTypes bool // whether to include import type
}

// Result — the outcome of the traversal.
type Result struct {
	Files   []string          // relative POSIX paths in traversal order
	Explain map[string]string // abs path → reason for inclusion
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

// Collect traverses the import graph starting from entryFiles.
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

// extractImports pulls specifiers out of import/export ... from statements.
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
			continue // export const ... — no source
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

// resolve turns a specifier into an absolute file path, or "".
func (c *collector) resolve(spec, fromFile string) string {
	if strings.HasPrefix(spec, ".") {
		base := filepath.Join(filepath.Dir(fromFile), spec)
		return firstExisting(resolveCandidates(base))
	}
	// tsconfig paths aliases.
	if c.tsBaseDir != "" {
		if p := c.resolveTSPaths(spec); p != "" {
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

func (c *collector) resolveTSPaths(spec string) string {
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

// sourceExts — extensions we treat as source code on an exact path match.
// .json and the like are intentionally excluded: ts-morph without resolveJsonModule
// doesn't pick them up, and they aren't needed for packing code.
var sourceExts = map[string]bool{
	".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".mjs": true, ".cjs": true,
}

func hasSourceExt(p string) bool {
	if strings.HasSuffix(p, ".d.ts") {
		return true
	}
	return sourceExts[strings.ToLower(filepath.Ext(p))]
}

// resolveCandidates lists the possible files for a base path.
func resolveCandidates(p string) []string {
	var cands []string
	// Accept the exact path only if it's a source file (not .json, etc.).
	if hasSourceExt(p) {
		cands = append(cands, p)
	}
	for _, e := range resolveExts {
		cands = append(cands, p+e)
	}
	for _, e := range resolveExts {
		cands = append(cands, filepath.Join(p, "index"+e))
	}
	// .js → .ts/.tsx sibling (NodeNext style).
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
