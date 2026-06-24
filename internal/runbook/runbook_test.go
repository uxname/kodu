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
		t.Fatal("config.json should exist")
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
	if !strings.Contains(md, "# Runbook: proj") || !strings.Contains(md, "## Stand: dev") {
		t.Fatalf("runbook.md is missing the expected sections:\n%s", md)
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
		t.Fatalf("existing runbook.md was overwritten: %q", md)
	}
}

func TestEnsureGitignore(t *testing.T) {
	// no-git
	root := t.TempDir()
	s := New()
	if r, _ := s.EnsureGitignore(root); r != GitignoreNoGit {
		t.Fatalf("without .git expected no-git, got %s", r)
	}

	// created
	_ = os.MkdirAll(filepath.Join(root, ".git"), 0o755)
	if r, _ := s.EnsureGitignore(root); r != GitignoreCreated {
		t.Fatalf("expected created, got %s", r)
	}
	// present (again)
	if r, _ := s.EnsureGitignore(root); r != GitignorePresent {
		t.Fatalf("expected present, got %s", r)
	}

	// added (to an existing .gitignore)
	root2 := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root2, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(root2, ".gitignore"), []byte("node_modules\n"), 0o644)
	if r, _ := s.EnsureGitignore(root2); r != GitignoreAdded {
		t.Fatalf("expected added, got %s", r)
	}
	content, _ := os.ReadFile(filepath.Join(root2, ".gitignore"))
	if !strings.Contains(string(content), "/.runbook/") {
		t.Fatalf(".gitignore is missing the runbook entry: %s", content)
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

func TestDetectStackEnvAndComposeYml(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "docker-compose.yml"), []byte("x"), 0o644)
	_ = os.WriteFile(filepath.Join(root, ".env.example"), []byte("X="), 0o644)
	d := New().DetectStack(root)
	if !d.Compose || !d.EnvExample || d.Dockerfile {
		t.Fatalf("detect = %+v", d)
	}
}

func TestDetectStackEmpty(t *testing.T) {
	d := New().DetectStack(t.TempDir())
	if d.Compose || d.Dockerfile || d.EnvExample {
		t.Fatalf("expected nothing detected, got %+v", d)
	}
}

func TestPaths(t *testing.T) {
	s := New()
	root := "/repo"
	if got := s.DirPath(root); got != filepath.Join(root, ".runbook") {
		t.Fatalf("DirPath = %q", got)
	}
	if got := s.ConfigPath(root); got != filepath.Join(root, ".runbook", "config.json") {
		t.Fatalf("ConfigPath = %q", got)
	}
	if got := s.RunbookPath(root); got != filepath.Join(root, ".runbook", "runbook.md") {
		t.Fatalf("RunbookPath = %q", got)
	}
}

func TestExistsFalse(t *testing.T) {
	if New().Exists(t.TempDir()) {
		t.Fatal("Exists should be false for an uninitialized project")
	}
}

// Missing config.json and runbook.md are surfaced as read errors.
func TestReadMissing(t *testing.T) {
	s := New()
	root := t.TempDir()
	if _, err := s.ReadConfig(root); err == nil {
		t.Fatal("ReadConfig should error when config.json is missing")
	}
	if _, err := s.ReadRunbook(root); err == nil {
		t.Fatal("ReadRunbook should error when runbook.md is missing")
	}
}

func TestReadConfigInvalidJSON(t *testing.T) {
	s := New()
	root := t.TempDir()
	_ = os.MkdirAll(s.DirPath(root), 0o755)
	_ = os.WriteFile(s.ConfigPath(root), []byte("{bad"), 0o644)
	if _, err := s.ReadConfig(root); err == nil {
		t.Fatal("ReadConfig should error on invalid JSON")
	}
}

// WriteConfig normalizes empty ActiveStand/Stands; ReadConfig round-trips them.
func TestWriteConfigDefaults(t *testing.T) {
	s := New()
	root := t.TempDir()
	if err := s.WriteConfig(ProjectConfig{Project: "p"}, root); err != nil {
		t.Fatal(err)
	}
	cfg, err := s.ReadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveStand != "local" {
		t.Fatalf("ActiveStand = %q, want local", cfg.ActiveStand)
	}
	if len(cfg.Stands) != 4 {
		t.Fatalf("Stands = %v, want the 4 defaults", cfg.Stands)
	}
}

// ReadConfig fills defaults for an empty config object on disk.
func TestReadConfigNormalizes(t *testing.T) {
	s := New()
	root := t.TempDir()
	_ = os.MkdirAll(s.DirPath(root), 0o755)
	_ = os.WriteFile(s.ConfigPath(root), []byte(`{"project":"p"}`), 0o644)
	cfg, err := s.ReadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveStand != "local" || len(cfg.Stands) != 4 {
		t.Fatalf("normalization failed: %+v", cfg)
	}
}

func TestRenderRunbookWithCompose(t *testing.T) {
	md := RenderRunbook("svc", []string{"local", "prod", "custom"}, DetectedStack{Compose: true, EnvExample: true})
	checks := []string{
		"# Runbook: svc",
		"## Stand: local (local development)",
		"## Stand: prod (production (be careful!))",
		"## Stand: custom (custom)", // unknown stand falls back to its own name
		"docker compose up -d",      // Compose hint applied
		"Copy `.env.example`",       // EnvExample hint applied
	}
	for _, want := range checks {
		if !strings.Contains(md, want) {
			t.Fatalf("runbook missing %q:\n%s", want, md)
		}
	}
}

// Without a detected stack the placeholders stay generic.
func TestRenderRunbookPlaceholders(t *testing.T) {
	md := RenderRunbook("svc", []string{"dev"}, DetectedStack{})
	if !strings.Contains(md, "<start command") || !strings.Contains(md, "<where to get environment") {
		t.Fatalf("expected generic placeholders:\n%s", md)
	}
}
