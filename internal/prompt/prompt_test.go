package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLooksLikeInline(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"multiline", "line one\nline two", true},
		{"path with slash", "prompts/pack.md", false},
		{"path with backslash", `prompts\pack.md`, false},
		{"has extension", "pack.md", false},
		{"short inline word", "summarize", true},
		{"inline phrase", "summarize this codebase", true},
		{"empty string", "", false},
		{"only whitespace", "   ", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := looksLikeInline(c.value); got != c.want {
				t.Fatalf("looksLikeInline(%q) = %v, want %v", c.value, got, c.want)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "prompt.md"), "file content")
	s := New(root)

	got, err := s.Load("prompt.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != "file content" {
		t.Fatalf("Load = %q, want %q", got, "file content")
	}
}

func TestLoadAbsolutePath(t *testing.T) {
	root := t.TempDir()
	abs := filepath.Join(root, "abs.md")
	write(t, abs, "absolute content")
	s := New(root)

	got, err := s.Load(abs)
	if err != nil {
		t.Fatal(err)
	}
	if got != "absolute content" {
		t.Fatalf("Load(abs) = %q, want %q", got, "absolute content")
	}
}

func TestLoadInlineText(t *testing.T) {
	s := New(t.TempDir())
	got, err := s.Load("just an inline prompt")
	if err != nil {
		t.Fatal(err)
	}
	if got != "just an inline prompt" {
		t.Fatalf("Load(inline) = %q", got)
	}
}

// A path-like source that does not exist on disk is an error, not inline text.
func TestLoadMissingFileError(t *testing.T) {
	s := New(t.TempDir())
	_, err := s.Load("missing/file.md")
	if err == nil {
		t.Fatal("expected error for missing path-like source")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %v, want 'not found'", err)
	}
}

func TestLoadCaches(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "cached.md")
	write(t, p, "original")
	s := New(root)

	first, err := s.Load("cached.md")
	if err != nil || first != "original" {
		t.Fatalf("first Load = %q, err %v", first, err)
	}
	// Change the file on disk; the cached value must be returned unchanged.
	write(t, p, "modified")
	second, err := s.Load("cached.md")
	if err != nil {
		t.Fatal(err)
	}
	if second != "original" {
		t.Fatalf("expected cached %q, got %q", "original", second)
	}
}

func TestLoadFromPromptsDir(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".kodu", "prompts", "review.md"), "review template")
	s := New(root)

	got, err := s.LoadFromPromptsDir("review")
	if err != nil {
		t.Fatal(err)
	}
	if got != "review template" {
		t.Fatalf("LoadFromPromptsDir = %q", got)
	}
}

// .md is preferred over .txt when both exist.
func TestLoadFromPromptsDirPrefersMarkdown(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".kodu", "prompts", "dup.md"), "from md")
	write(t, filepath.Join(root, ".kodu", "prompts", "dup.txt"), "from txt")
	s := New(root)

	got, err := s.LoadFromPromptsDir("dup")
	if err != nil {
		t.Fatal(err)
	}
	if got != "from md" {
		t.Fatalf("expected .md preferred, got %q", got)
	}
}

func TestLoadFromPromptsDirExplicitExtension(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".kodu", "prompts", "note.txt"), "txt body")
	s := New(root)

	got, err := s.LoadFromPromptsDir("note.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got != "txt body" {
		t.Fatalf("LoadFromPromptsDir(note.txt) = %q", got)
	}
}

func TestLoadFromPromptsDirMissing(t *testing.T) {
	s := New(t.TempDir())
	_, err := s.LoadFromPromptsDir("nope")
	if err == nil {
		t.Fatal("expected error for missing template")
	}
	if !strings.Contains(err.Error(), "nope.md") || !strings.Contains(err.Error(), "nope.txt") {
		t.Fatalf("error should list expected candidates, got %v", err)
	}
}

// The cache must be safe under concurrent Load calls (run under -race).
func TestConcurrentLoad(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "shared.md"), "shared content")
	s := New(root)

	var wg sync.WaitGroup
	errs := make(chan error, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := s.Load("shared.md")
			if err != nil {
				errs <- err
				return
			}
			if got != "shared content" {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent Load failed: %v", err)
	}
}
