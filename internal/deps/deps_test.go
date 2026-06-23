package deps

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func mk(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCollectGraph(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "tsconfig.json", `{
		// комментарий допустим
		"compilerOptions": { "baseUrl": ".", "paths": { "@app/*": ["src/*"] } },
	}`)
	mk(t, root, "entry.ts", `import './a';
import type {T} from './t';
import x from '@app/aliased';
import 'react';`)
	mk(t, root, "a.ts", `import './lib';`)
	mk(t, root, "lib/index.ts", `export const x = 1;`)
	mk(t, root, "t.ts", `export type T = number;`)
	mk(t, root, "src/aliased.ts", `export const y = 2;`)

	res := Collect([]string{"entry.ts"}, root, Options{IncludeTypes: true})

	wantFiles := []string{"entry.ts", "a.ts", "lib/index.ts", "t.ts", "src/aliased.ts"}
	if !reflect.DeepEqual(res.Files, wantFiles) {
		t.Fatalf("files = %v\n want %v", res.Files, wantFiles)
	}

	wantExplain := map[string]string{
		"entry.ts":       "entry point",
		"a.ts":           "import from entry.ts",
		"lib/index.ts":   "import from a.ts",
		"t.ts":           "type import from entry.ts",
		"src/aliased.ts": "import from entry.ts",
	}
	for _, rel := range wantFiles {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if got := res.Explain[abs]; got != wantExplain[rel] {
			t.Fatalf("explain[%s] = %q, want %q", rel, got, wantExplain[rel])
		}
	}
}

func TestReExportAndDedup(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "entry.ts", `export * from './shared';
import './a';`)
	mk(t, root, "shared.ts", `export const s = 1;`)
	mk(t, root, "a.ts", `import './shared';`) // повтор shared — не дублируется

	res := Collect([]string{"entry.ts"}, root, Options{IncludeTypes: true})
	want := []string{"entry.ts", "shared.ts", "a.ts"}
	if !reflect.DeepEqual(res.Files, want) {
		t.Fatalf("files = %v, want %v", res.Files, want)
	}
	abs := filepath.Join(root, "shared.ts")
	if res.Explain[abs] != "re-export from entry.ts" {
		t.Fatalf("explain shared = %q", res.Explain[abs])
	}
}

func TestMaxDepth(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "entry.ts", `import './a';`)
	mk(t, root, "a.ts", `import './b';`)
	mk(t, root, "b.ts", `export const b = 1;`)

	// depth 1: entry + прямые импорты (a), без b.
	res := Collect([]string{"entry.ts"}, root, Options{IncludeTypes: true, MaxDepth: 1})
	want := []string{"entry.ts", "a.ts"}
	if !reflect.DeepEqual(res.Files, want) {
		t.Fatalf("files = %v, want %v", res.Files, want)
	}
}

func TestTypeImportExcludedWhenDisabled(t *testing.T) {
	root := t.TempDir()
	mk(t, root, "entry.ts", `import type {T} from './t';
import './a';`)
	mk(t, root, "t.ts", `export type T = number;`)
	mk(t, root, "a.ts", `export const a = 1;`)

	res := Collect([]string{"entry.ts"}, root, Options{IncludeTypes: false})
	want := []string{"entry.ts", "a.ts"}
	if !reflect.DeepEqual(res.Files, want) {
		t.Fatalf("files = %v, want %v", res.Files, want)
	}
}
