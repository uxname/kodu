package deps

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/tidwall/jsonc"
)

type tsconfig struct {
	CompilerOptions struct {
		BaseURL string              `json:"baseUrl"`
		Paths   map[string][]string `json:"paths"`
	} `json:"compilerOptions"`
}

// loadTsconfig reads tsconfig.json/tsconfig.base.json and returns the base
// directory for path resolution and the paths map. A tsconfig may contain
// comments/trailing commas — we normalize it via jsonc.
func loadTsconfig(projectRoot string) (baseDir string, paths map[string][]string) {
	for _, name := range []string{"tsconfig.json", "tsconfig.base.json"} {
		p := filepath.Join(projectRoot, name)
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var cfg tsconfig
		if err := json.Unmarshal(jsonc.ToJSON(raw), &cfg); err != nil {
			continue
		}
		base := projectRoot
		if cfg.CompilerOptions.BaseURL != "" {
			base = filepath.Join(projectRoot, cfg.CompilerOptions.BaseURL)
		}
		return base, cfg.CompilerOptions.Paths
	}
	return "", nil
}
