// Package config reads and validates kodu.json.
//
// Parity with src/core/config/config.schema.ts + config.service.ts:
//   - only kodu.json in the current directory is looked up;
//   - parsing is non-strict (unknown keys are ignored, like zod without .strict());
//   - every field has a default value (mirrors zod .default()).
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Cleaner holds comment-stripping settings.
type Cleaner struct {
	Whitelist    []string `json:"whitelist"`
	KeepJSDoc    bool     `json:"keepJSDoc"`
	UseGitignore bool     `json:"useGitignore"`
	Ignore       []string `json:"ignore"`
}

// Packer holds context-collection settings.
type Packer struct {
	Ignore                      []string `json:"ignore"`
	UseGitignore                bool     `json:"useGitignore"`
	ContentBasedBinaryDetection bool     `json:"contentBasedBinaryDetection"`
}

// Prompts holds user-defined prompt templates.
type Prompts struct {
	Pack string `json:"pack,omitempty"`
}

// Config is the root configuration of kodu.json.
type Config struct {
	Schema  string   `json:"$schema,omitempty"`
	Cleaner Cleaner  `json:"cleaner"`
	Packer  Packer   `json:"packer"`
	Prompts *Prompts `json:"prompts,omitempty"`
}

// ConfigFileName is the single config file name (like searchPlaces in lilconfig).
const ConfigFileName = "kodu.json"

// DefaultPackerIgnore is the default packer ignore list (config.schema.ts:13).
func DefaultPackerIgnore() []string {
	return []string{
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		".git",
		".kodu",
		"node_modules",
		"dist",
		"coverage",
	}
}

// Default returns a configuration fully populated with default values.
// Values match zod .default() in config.schema.ts 1:1.
func Default() Config {
	return Config{
		Cleaner: Cleaner{
			Whitelist:    []string{"//!"},
			KeepJSDoc:    true,
			UseGitignore: true,
			Ignore:       []string{},
		},
		Packer: Packer{
			Ignore:                      DefaultPackerIgnore(),
			UseGitignore:                true,
			ContentBasedBinaryDetection: false,
		},
	}
}

// Load reads kodu.json from dir. If the file is missing, it returns defaults.
// Invalid JSON (including wrong field types) is returned as an error.
//
// Defaults are implemented by parsing on top of a pre-populated struct:
// keys present in the file overwrite the default, absent ones do not.
// This mirrors the per-field .default() behavior in zod.
func Load(dir string) (Config, error) {
	cfg := Default()

	path := filepath.Join(dir, ConfigFileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read %s: %w", ConfigFileName, err)
	}

	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("%s is invalid: %w", ConfigFileName, err)
	}

	return cfg, nil
}
