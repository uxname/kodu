package packer

import "testing"

func TestBuildContextXML(t *testing.T) {
	got := BuildContext([]File{
		{Path: "a.ts", Content: "const a = 1;"},
		{Path: "b.ts", Content: "const b = 2;"},
	}, FormatXML)
	want := "<files>\n" +
		"<file path=\"a.ts\">\nconst a = 1;\n</file>\n\n" +
		"<file path=\"b.ts\">\nconst b = 2;\n</file>\n" +
		"</files>"
	if got != want {
		t.Fatalf("xml mismatch:\n got=%q\nwant=%q", got, want)
	}
}

func TestBuildContextText(t *testing.T) {
	got := BuildContext([]File{
		{Path: "a.ts", Content: "const a = 1;"},
		{Path: "b.ts", Content: "const b = 2;"},
	}, FormatText)
	want := "// file: a.ts\nconst a = 1;\n\n// file: b.ts\nconst b = 2;"
	if got != want {
		t.Fatalf("text mismatch:\n got=%q\nwant=%q", got, want)
	}
}

func TestFillTemplateWithPlaceholder(t *testing.T) {
	got := FillTemplate("Files: {{fileList}}\nTokens: {{tokenCount}}\nCost: {{usdEstimate}}\n---\n{{context}}", TemplateContext{
		Context:     "CTX",
		FileList:    "a.ts",
		TokenCount:  42,
		USDEstimate: 0.0123456,
	})
	want := "Files: a.ts\nTokens: 42\nCost: 0.0123\n---\nCTX"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFillTemplateAppendsContextWhenMissing(t *testing.T) {
	got := FillTemplate("Just a prompt", TemplateContext{Context: "CTX"})
	want := "Just a prompt\n\nCTX"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
