// Package cleaner удаляет комментарии из исходного кода (паритет cleaner.service.ts).
//
// Для ts/tsx/js/jsx и прочих C-подобных файлов используется tree-sitter
// (грамматики typescript/tsx), что безопасно для строк, regex- и
// template-литералов. Для .html/.htm комментарии `<!-- -->` вырезаются
// регуляркой.
//
// Отличия от TS (намеренные, в пользу полноты и безопасности):
//   - tree-sitter находит ВСЕ узлы-комментарии, тогда как ts-morph собирал их
//     через leading/trailing-ranges узлов и пропускал некоторые позиции
//     (комментарий после висячей запятой, одинокий комментарий в пустом блоке).
//     Поэтому очистка полнее. Код остаётся валидным; whitelist/JSDoc уважаются.
//   - для .html НЕ парсим содержимое как TypeScript (в оригинале это побочно
//     ловило `//` внутри URL и могло портить разметку); удаляем только
//     настоящие HTML-комментарии.
package cleaner

import (
	"context"
	"regexp"
	"sort"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// systemWhitelist — системные директивы, которые никогда не удаляются
// (cleaner.service.ts:31). Хранятся в нижнем регистре.
var systemWhitelist = []string{
	"@ts-ignore",
	"@ts-expect-error",
	"eslint-disable",
	"prettier-ignore",
	"biome-ignore",
	"todo",
	"fixme",
}

var htmlCommentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

// Result — итог очистки одного файла.
type Result struct {
	Content  string
	Removed  int
	Previews []string
}

// Cleaner хранит whitelist и настройку JSDoc.
type Cleaner struct {
	whitelist []string // нижний регистр, system + пользовательские
}

// New создаёт Cleaner. userWhitelist дополняет системный список.
func New(userWhitelist []string) *Cleaner {
	wl := make([]string, 0, len(systemWhitelist)+len(userWhitelist))
	wl = append(wl, systemWhitelist...)
	for _, u := range userWhitelist {
		wl = append(wl, strings.ToLower(u))
	}
	return &Cleaner{whitelist: wl}
}

type rangeKind int

const (
	kindComment rangeKind = iota
	kindJSX
)

type removal struct {
	start, end int
	text       string
	kind       rangeKind
}

// Clean удаляет комментарии из content. keepJSDoc сохраняет блоки `/** */`.
func (c *Cleaner) Clean(filename, content string, keepJSDoc bool) Result {
	src := []byte(content)
	var ranges []removal

	if isHTML(filename) {
		ranges = collectHTMLRanges(src)
	} else {
		ranges = collectCodeRanges(src, grammarFor(filename))
	}

	candidates := ranges[:0:0]
	for _, r := range ranges {
		if c.shouldRemove(r, keepJSDoc) {
			candidates = append(candidates, r)
		}
	}
	if len(candidates) == 0 {
		return Result{Content: content, Removed: 0}
	}

	previews := make([]string, len(candidates))
	for i, r := range candidates {
		previews[i] = normalizePreview(r.text)
	}

	// Удаляем с конца, чтобы индексы не съезжали.
	sorted := make([]removal, len(candidates))
	copy(sorted, candidates)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].start > sorted[j].start })

	out := append([]byte(nil), src...)
	for _, r := range sorted {
		replacement := getReplacement(src, r)
		out = append(out[:r.start], append([]byte(replacement), out[r.end:]...)...)
	}

	return Result{Content: string(out), Removed: len(candidates), Previews: previews}
}

func (c *Cleaner) shouldRemove(r removal, keepJSDoc bool) bool {
	trimmed := strings.TrimLeft(r.text, " \t\r\n")
	if keepJSDoc && strings.HasPrefix(trimmed, "/**") {
		return false
	}
	lower := strings.ToLower(r.text)
	for _, token := range c.whitelist {
		if strings.Contains(lower, token) {
			return false
		}
	}
	return true
}

// collectCodeRanges обходит AST и собирает комментарии и пустые JSX-выражения.
func collectCodeRanges(src []byte, lang *sitter.Language) []removal {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, src)
	if err != nil || tree == nil {
		return nil
	}
	defer tree.Close()

	var ranges []removal
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		switch n.Type() {
		case "jsx_expression":
			if isEmptyJSXComment(n) {
				ranges = append(ranges, removal{
					start: int(n.StartByte()), end: int(n.EndByte()),
					text: string(src[n.StartByte():n.EndByte()]), kind: kindJSX,
				})
				return // не спускаемся — внутренний комментарий уже покрыт
			}
		case "comment":
			ranges = append(ranges, removal{
				start: int(n.StartByte()), end: int(n.EndByte()),
				text: string(src[n.StartByte():n.EndByte()]), kind: kindComment,
			})
			return
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(tree.RootNode())
	return ranges
}

// isEmptyJSXComment истинно, если jsx_expression содержит только комментарии
// (нет выражения) — тогда удаляется весь блок `{/* */}`.
func isEmptyJSXComment(n *sitter.Node) bool {
	named := int(n.NamedChildCount())
	if named == 0 {
		return false
	}
	hasComment := false
	for i := 0; i < named; i++ {
		if n.NamedChild(i).Type() != "comment" {
			return false
		}
		hasComment = true
	}
	return hasComment
}

func collectHTMLRanges(src []byte) []removal {
	var ranges []removal
	for _, m := range htmlCommentRe.FindAllIndex(src, -1) {
		ranges = append(ranges, removal{
			start: m[0], end: m[1], text: string(src[m[0]:m[1]]), kind: kindComment,
		})
	}
	return ranges
}

// getReplacement возвращает строку-замену удаляемого диапазона.
// Для JSX — пусто. Для прочих: пробел, если комментарий зажат между двумя
// идентификаторными символами (иначе склеятся токены), иначе пусто.
func getReplacement(src []byte, r removal) string {
	if r.kind == kindJSX {
		return ""
	}
	var before, after byte
	if r.start > 0 {
		before = src[r.start-1]
	}
	if r.end < len(src) {
		after = src[r.end]
	}
	if isIdentByte(before) && isIdentByte(after) {
		return " "
	}
	return ""
}

func isIdentByte(b byte) bool {
	return b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '_' || b == '$'
}

var wsRe = regexp.MustCompile(`\s+`)

func normalizePreview(text string) string {
	single := strings.TrimSpace(wsRe.ReplaceAllString(text, " "))
	if len(single) <= 60 {
		return single
	}
	return single[:57] + "..."
}

func grammarFor(filename string) *sitter.Language {
	switch ext(filename) {
	case ".tsx", ".jsx":
		return tsx.GetLanguage()
	default:
		// .ts/.js/.mjs/.cjs и любые прочие — как TS (паритет ts-morph).
		return typescript.GetLanguage()
	}
}

func isHTML(filename string) bool {
	e := ext(filename)
	return e == ".html" || e == ".htm"
}

func ext(filename string) string {
	i := strings.LastIndexByte(filename, '.')
	if i < 0 {
		return ""
	}
	return strings.ToLower(filename[i:])
}
