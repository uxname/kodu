package runbook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldAndReadConfig(t *testing.T) {
	root := t.TempDir()
	s := New()
	if err := s.Scaffold("proj", []string{"local", "dev"}, "dev", root); err != nil {
		t.Fatal(err)
	}
	if !s.Exists(root) {
		t.Fatal("config.json должен существовать")
	}
	cfg, err := s.ReadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Project != "proj" || cfg.ActiveStand != "dev" || len(cfg.Stands) != 2 {
		t.Fatalf("config = %+v", cfg)
	}
	md, err := s.ReadRunbook(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "# Runbook: proj") || !strings.Contains(md, "## Стенд: dev") {
		t.Fatalf("runbook.md без ожидаемых секций:\n%s", md)
	}
}

func TestScaffoldKeepsExistingRunbook(t *testing.T) {
	root := t.TempDir()
	s := New()
	_ = os.MkdirAll(s.DirPath(root), 0o755)
	_ = os.WriteFile(s.RunbookPath(root), []byte("CUSTOM"), 0o644)
	if err := s.Scaffold("proj", []string{"local"}, "local", root); err != nil {
		t.Fatal(err)
	}
	md, _ := s.ReadRunbook(root)
	if md != "CUSTOM" {
		t.Fatalf("существующий runbook.md перезаписан: %q", md)
	}
}

func TestEnsureGitignore(t *testing.T) {
	// no-git
	root := t.TempDir()
	s := New()
	if r, _ := s.EnsureGitignore(root); r != GitignoreNoGit {
		t.Fatalf("без .git ожидал no-git, got %s", r)
	}

	// created
	_ = os.MkdirAll(filepath.Join(root, ".git"), 0o755)
	if r, _ := s.EnsureGitignore(root); r != GitignoreCreated {
		t.Fatalf("ожидал created, got %s", r)
	}
	// present (повторно)
	if r, _ := s.EnsureGitignore(root); r != GitignorePresent {
		t.Fatalf("ожидал present, got %s", r)
	}

	// added (к существующему .gitignore)
	root2 := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root2, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(root2, ".gitignore"), []byte("node_modules\n"), 0o644)
	if r, _ := s.EnsureGitignore(root2); r != GitignoreAdded {
		t.Fatalf("ожидал added, got %s", r)
	}
	content, _ := os.ReadFile(filepath.Join(root2, ".gitignore"))
	if !strings.Contains(string(content), "/.runbook/") {
		t.Fatalf(".gitignore без записи runbook: %s", content)
	}
}

func TestDetectStack(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "Dockerfile"), []byte("x"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "docker-compose.yaml"), []byte("x"), 0o644)
	d := New().DetectStack(root)
	if !d.Dockerfile || !d.Compose || d.EnvExample {
		t.Fatalf("detect = %+v", d)
	}
}
