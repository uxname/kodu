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

// loadTsconfig читает tsconfig.json/tsconfig.base.json и возвращает базовую
// директорию для разрешения путей и карту paths. tsconfig может содержать
// комментарии/висячие запятые — нормализуем через jsonc.
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
