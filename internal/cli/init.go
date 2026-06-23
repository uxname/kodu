package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const gitignoreContextEntry = ".kodu/context.txt"

type initCleaner struct {
	Whitelist    []string `json:"whitelist"`
	KeepJSDoc    bool     `json:"keepJSDoc"`
	UseGitignore bool     `json:"useGitignore"`
	Ignore       []string `json:"ignore"`
}

type initPacker struct {
	Ignore                      []string `json:"ignore"`
	UseGitignore                bool     `json:"useGitignore"`
	ContentBasedBinaryDetection bool     `json:"contentBasedBinaryDetection"`
}

type initConfig struct {
	Schema  string      `json:"$schema"`
	Cleaner initCleaner `json:"cleaner"`
	Packer  initPacker  `json:"packer"`
}

// defaultKoduJSON — the contents of a new kodu.json (parity with init.command.ts).
// The schema URL was fixed to point at the current uxname/kodu repository (the
// original used the stale anomalyco) — otherwise autocomplete wouldn't work for users.
func defaultKoduJSON() initConfig {
	return initConfig{
		Schema: "https://raw.githubusercontent.com/uxname/kodu/refs/heads/master/kodu.schema.json",
		Cleaner: initCleaner{
			Whitelist:    []string{"//!"},
			KeepJSDoc:    true,
			UseGitignore: true,
			Ignore:       []string{},
		},
		Packer: initPacker{
			Ignore: []string{
				"package-lock.json", "yarn.lock", "pnpm-lock.yaml",
				".git", ".kodu", "node_modules", "dist", "coverage",
			},
			UseGitignore:                true,
			ContentBasedBinaryDetection: false,
		},
	}
}

func newInitCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize kodu configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := ensureKoduJSON(app, cwd); err != nil {
				return err
			}
			if err := updateGitignore(app, cwd); err != nil {
				return err
			}
			app.UI.Success("Done.")
			return nil
		},
	}
}

func ensureKoduJSON(app *App, cwd string) error {
	path := filepath.Join(cwd, "kodu.json")
	if _, err := os.Stat(path); err == nil {
		app.UI.Info("kodu.json already exists")
		return nil
	}
	data, err := json.MarshalIndent(defaultKoduJSON(), "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return err
	}
	app.UI.Success("Created kodu.json")
	return nil
}

func updateGitignore(app *App, cwd string) error {
	path := filepath.Join(cwd, ".gitignore")
	raw, err := os.ReadFile(path)
	if err != nil {
		app.UI.Warn(".gitignore not found, skipping.")
		return nil //nolint:nilerr // a missing .gitignore is not an error
	}
	content := string(raw)
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == gitignoreContextEntry {
			app.UI.Info(gitignoreContextEntry + " already in .gitignore")
			return nil
		}
	}
	trimmed := strings.TrimRight(content, "\n\r \t")
	next := gitignoreContextEntry
	if trimmed != "" {
		next = trimmed + "\n" + gitignoreContextEntry
	}
	if err := os.WriteFile(path, []byte(next+"\n"), 0o644); err != nil {
		return err
	}
	app.UI.Success("Added " + gitignoreContextEntry + " to .gitignore")
	return nil
}
