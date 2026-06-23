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
		t.Fatal("ожидал дефолты cleaner true")
	}
	if len(cfg.Cleaner.Whitelist) != 1 || cfg.Cleaner.Whitelist[0] != "//!" {
		t.Fatalf("whitelist = %v", cfg.Cleaner.Whitelist)
	}
	if len(cfg.Packer.Ignore) != len(DefaultPackerIgnore()) {
		t.Fatalf("packer.ignore = %v", cfg.Packer.Ignore)
	}
}

// Частичный конфиг: заданные поля переопределяют, остальные сохраняют дефолт.
func TestPartialConfigKeepsPerFieldDefaults(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"$schema":"x","cleaner":{"keepJSDoc":false}}`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Cleaner.KeepJSDoc {
		t.Fatal("keepJSDoc должен стать false")
	}
	if !cfg.Cleaner.UseGitignore {
		t.Fatal("useGitignore должен остаться дефолтным true")
	}
	if cfg.Cleaner.Whitelist[0] != "//!" {
		t.Fatal("whitelist должен остаться дефолтным")
	}
	if cfg.Schema != "x" {
		t.Fatal("$schema должен читаться")
	}
}

// Неизвестные ключи игнорируются (нестрогий разбор, как zod без .strict()).
func TestUnknownKeysIgnored(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"totallyUnknown":123,"packer":{"useGitignore":false}}`)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("неизвестные ключи не должны быть ошибкой: %v", err)
	}
	if cfg.Packer.UseGitignore {
		t.Fatal("packer.useGitignore должен стать false")
	}
}

// Явный пустой массив переопределяет дефолтный список.
func TestExplicitEmptyArrayOverrides(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"packer":{"ignore":[]}}`)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Packer.Ignore) != 0 {
		t.Fatalf("packer.ignore должен быть пустым, got %v", cfg.Packer.Ignore)
	}
}

func TestInvalidTypeIsError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"cleaner":{"whitelist":"not-an-array"}}`)
	if _, err := Load(dir); err == nil {
		t.Fatal("ожидал ошибку на неверном типе поля")
	}
}

func write(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
