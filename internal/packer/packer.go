// Package packer форматирует собранный контекст (паритет pack.command.ts).
package packer

import (
	"strconv"
	"strings"
)

// Format — формат вывода контекста.
type Format string

const (
	FormatXML  Format = "xml"
	FormatText Format = "text"
)

// File — путь и содержимое одного файла.
type File struct {
	Path    string
	Content string
}

// BuildContext собирает контекст в выбранном формате.
//
//	xml:  <files>\n<file path="P">\n{content}\n</file>\n\n...\n</files>
//	text: // file: P\n{content}\n\n...
func BuildContext(files []File, format Format) string {
	chunks := make([]string, len(files))
	for i, f := range files {
		if format == FormatXML {
			chunks[i] = "<file path=\"" + f.Path + "\">\n" + f.Content + "\n</file>"
		} else {
			chunks[i] = "// file: " + f.Path + "\n" + f.Content
		}
	}
	joined := strings.Join(chunks, "\n\n")
	if format == FormatXML {
		return "<files>\n" + joined + "\n</files>"
	}
	return joined
}

// TemplateContext — данные для подстановки в шаблон промпта.
type TemplateContext struct {
	Context     string
	FileList    string
	TokenCount  int
	USDEstimate float64
}

// FillTemplate подставляет плейсхолдеры. Если в шаблоне нет {{context}},
// контекст дописывается в конец (паритет pack.command.ts:341).
func FillTemplate(tmpl string, ctx TemplateContext) string {
	r := strings.NewReplacer(
		"{{context}}", ctx.Context,
		"{{fileList}}", ctx.FileList,
		"{{tokenCount}}", strconv.Itoa(ctx.TokenCount),
		"{{usdEstimate}}", strconv.FormatFloat(ctx.USDEstimate, 'f', 4, 64),
	)
	filled := r.Replace(tmpl)
	if !strings.Contains(tmpl, "{{context}}") {
		return filled + "\n\n" + ctx.Context
	}
	return filled
}
