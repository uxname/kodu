// Package config читает и валидирует kodu.json.
//
// Паритет с src/core/config/config.schema.ts + config.service.ts:
//   - ищется только kodu.json в текущей директории;
//   - разбор нестрогий (неизвестные ключи игнорируются, как zod без .strict());
//   - каждое поле имеет значение по умолчанию (зеркалит zod .default()).
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Cleaner — настройки удаления комментариев.
type Cleaner struct {
	Whitelist    []string `json:"whitelist"`
	KeepJSDoc    bool     `json:"keepJSDoc"`
	UseGitignore bool     `json:"useGitignore"`
	Ignore       []string `json:"ignore"`
}

// Packer — настройки сбора контекста.
type Packer struct {
	Ignore                      []string `json:"ignore"`
	UseGitignore                bool     `json:"useGitignore"`
	ContentBasedBinaryDetection bool     `json:"contentBasedBinaryDetection"`
}

// Prompts — пользовательские шаблоны промптов.
type Prompts struct {
	Pack string `json:"pack,omitempty"`
}

// Config — корневая конфигурация kodu.json.
type Config struct {
	Schema  string   `json:"$schema,omitempty"`
	Cleaner Cleaner  `json:"cleaner"`
	Packer  Packer   `json:"packer"`
	Prompts *Prompts `json:"prompts,omitempty"`
}

// ConfigFileName — единственное имя файла конфига (как searchPlaces в lilconfig).
const ConfigFileName = "kodu.json"

// DefaultPackerIgnore — дефолтный список игнора packer (config.schema.ts:13).
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

// Default возвращает полностью заполненную дефолтами конфигурацию.
// Значения 1:1 с zod .default() в config.schema.ts.
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

// Load читает kodu.json из dir. Если файла нет — возвращает дефолты.
// Невалидный JSON (в т.ч. неверные типы полей) возвращается ошибкой.
//
// Дефолты реализованы через разбор поверх предзаполненной структуры:
// присутствующие в файле ключи перезаписывают дефолт, отсутствующие — нет.
// Это зеркалит поведение per-field .default() в zod.
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
