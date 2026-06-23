package fsutil

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/uxname/kodu/internal/config"
)

// mkfile создаёт файл с содержимым, создавая родительские директории.
func mkfile(t *testing.T, root, rel string, content []byte) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func newFinder(t *testing.T) (*Finder, string) {
	t.Helper()
	root := t.TempDir()
	cfg := config.Default()
	return New(root, cfg, nil), root
}

func TestFindBasicIgnoreAndSort(t *testing.T) {
	f, root := newFinder(t)
	mkfile(t, root, "b.ts", []byte("b"))
	mkfile(t, root, "a.ts", []byte("a"))
	mkfile(t, root, "node_modules/dep/index.js", []byte("x")) // default ignore
	mkfile(t, root, "dist/out.js", []byte("x"))               // default ignore
	mkfile(t, root, "src/util.ts", []byte("u"))

	got, err := f.Find(FindOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.ts", "b.ts", "src/util.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestGitignoreRespected(t *testing.T) {
	f, root := newFinder(t)
	mkfile(t, root, ".gitignore", []byte("secret.txt\nlogs/\n"))
	mkfile(t, root, "secret.txt", []byte("s"))
	mkfile(t, root, "logs/app.log", []byte("l"))
	mkfile(t, root, "keep.txt", []byte("k"))

	got, err := f.Find(FindOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// .gitignore сам по себе не игнорируется и попадает в список (паритет с TS dot:true).
	want := []string{".gitignore", "keep.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestKoduignoreRespected(t *testing.T) {
	f, root := newFinder(t)
	mkfile(t, root, ".koduignore", []byte("*.md\n"))
	mkfile(t, root, "readme.md", []byte("r"))
	mkfile(t, root, "code.ts", []byte("c"))

	got, err := f.Find(FindOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{".koduignore", "code.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBinaryExtensionExcludedTextForced(t *testing.T) {
	f, root := newFinder(t)
	mkfile(t, root, "img.png", []byte("binary"))
	mkfile(t, root, "code.ts", []byte("text"))
	// .go в списке knownText, расширения нет в binary — точно текст
	mkfile(t, root, "main.go", []byte("package x"))

	got, err := f.Find(FindOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"code.ts", "main.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestContentBasedBinaryDetection(t *testing.T) {
	f, root := newFinder(t)
	// Расширение неизвестно, есть нулевой байт → бинарный при включённой детекции.
	mkfile(t, root, "blob.xyz", []byte{0x41, 0x00, 0x42})
	mkfile(t, root, "plain.xyz", []byte("hello"))

	on := true
	got, err := f.Find(FindOptions{ContentBasedBinaryDetection: &on})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"plain.xyz"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestRootPathsRestrictsToSubtree(t *testing.T) {
	f, root := newFinder(t)
	mkfile(t, root, "src/a.ts", []byte("a"))
	mkfile(t, root, "src/nested/b.ts", []byte("b"))
	mkfile(t, root, "other/c.ts", []byte("c"))

	got, err := f.Find(FindOptions{RootPaths: []string{"src"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"src/a.ts", "src/nested/b.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestLargeFileSkipped(t *testing.T) {
	f, root := newFinder(t)
	var warned []string
	f.Warn = func(s string) { warned = append(warned, s) }
	mkfile(t, root, "big.ts", make([]byte, 2*1024*1024))
	mkfile(t, root, "small.ts", []byte("ok"))

	got, err := f.Find(FindOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"small.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if len(warned) != 1 {
		t.Fatalf("ожидал 1 предупреждение о крупном файле, got %v", warned)
	}
}

func TestGitDirAlwaysIgnored(t *testing.T) {
	f, root := newFinder(t)
	mkfile(t, root, ".git/config", []byte("x"))
	mkfile(t, root, "a.ts", []byte("a"))

	got, err := f.Find(FindOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.ts"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
