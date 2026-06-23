package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func gitInit(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git не установлен")
	}
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "t@t.t"},
		{"config", "user.name", "t"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, rel), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestEnsureRepo(t *testing.T) {
	g := New(gitInit(t))
	if err := g.EnsureRepo(); err != nil {
		t.Fatalf("EnsureRepo на git-репо вернул ошибку: %v", err)
	}
	if err := New(t.TempDir()).EnsureRepo(); err == nil {
		t.Fatal("EnsureRepo вне репо должен вернуть ошибку")
	}
}

func TestChangedAndStaged(t *testing.T) {
	dir := gitInit(t)
	g := New(dir)

	write(t, dir, "tracked.ts", "1")
	gitRun(t, dir, "add", "tracked.ts")
	gitRun(t, dir, "commit", "-m", "init")

	write(t, dir, "tracked.ts", "2")   // modified, unstaged
	write(t, dir, "staged.ts", "s")    // new, staged
	write(t, dir, "untracked.ts", "u") // new, untracked
	gitRun(t, dir, "add", "staged.ts")

	staged, err := g.StagedFiles()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(staged, []string{"staged.ts"}) {
		t.Fatalf("staged = %v", staged)
	}

	changed, err := g.ChangedFiles()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"staged.ts", "tracked.ts", "untracked.ts"}
	if !reflect.DeepEqual(changed, want) {
		t.Fatalf("changed = %v, want %v", changed, want)
	}
}
