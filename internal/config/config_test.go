package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultWhenNoFile(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Cleaner.KeepJSDoc || !cfg.Cleaner.UseGitignore {
		t.Fatal("expected cleaner defaults to be true")
	}
	if len(cfg.Cleaner.Whitelist) != 1 || cfg.Cleaner.Whitelist[0] != "//!" {
		t.Fatalf("whitelist = %v", cfg.Cleaner.Whitelist)
	}
	if len(cfg.Packer.Ignore) != len(DefaultPackerIgnore()) {
		t.Fatalf("packer.ignore = %v", cfg.Packer.Ignore)
	}
}

// Partial config: specified fields override, the rest keep their defaults.
func TestPartialConfigKeepsPerFieldDefaults(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"$schema":"x","cleaner":{"keepJSDoc":false}}`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Cleaner.KeepJSDoc {
		t.Fatal("keepJSDoc should become false")
	}
	if !cfg.Cleaner.UseGitignore {
		t.Fatal("useGitignore should remain the default true")
	}
	if cfg.Cleaner.Whitelist[0] != "//!" {
		t.Fatal("whitelist should remain the default")
	}
	if cfg.Schema != "x" {
		t.Fatal("$schema should be read")
	}
}

// Unknown keys are ignored (non-strict parsing, like zod without .strict()).
func TestUnknownKeysIgnored(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"totallyUnknown":123,"packer":{"useGitignore":false}}`)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unknown keys should not be an error: %v", err)
	}
	if cfg.Packer.UseGitignore {
		t.Fatal("packer.useGitignore should become false")
	}
}

// An explicit empty array overrides the default list.
func TestExplicitEmptyArrayOverrides(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"packer":{"ignore":[]}}`)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Packer.Ignore) != 0 {
		t.Fatalf("packer.ignore should be empty, got %v", cfg.Packer.Ignore)
	}
}

func TestInvalidTypeIsError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"cleaner":{"whitelist":"not-an-array"}}`)
	if _, err := Load(dir); err == nil {
		t.Fatal("expected an error on a wrong field type")
	}
}

func write(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
