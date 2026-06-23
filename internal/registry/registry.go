// Package registry хранит глобальный реестр проектов (паритет registry.service.ts
// + registry.schema.ts) в ~/.config/kodu/registry.json (учитывает XDG_CONFIG_HOME).
package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultStands — стандартный набор стендов (registry.schema.ts).
func DefaultStands() []string { return []string{"local", "dev", "stage", "prod"} }

// ProjectEntry — один проект в реестре.
type ProjectEntry struct {
	Path   string   `json:"path"`
	Repo   string   `json:"repo,omitempty"`
	Stands []string `json:"stands"`
}

// Registry — весь реестр.
type Registry struct {
	Schema   string                  `json:"$schema,omitempty"`
	Projects map[string]ProjectEntry `json:"projects"`
}

// Service читает и пишет реестр.
type Service struct {
	dir  string
	file string
}

// New создаёт сервис, вычисляя путь по XDG.
func New() *Service {
	base := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	dir := filepath.Join(base, "kodu")
	return &Service{dir: dir, file: filepath.Join(dir, "registry.json")}
}

// FilePath возвращает путь к файлу реестра.
func (s *Service) FilePath() string { return s.file }

// Load загружает реестр; если файла нет — пустой реестр.
func (s *Service) Load() (Registry, error) {
	reg := Registry{Projects: map[string]ProjectEntry{}}
	raw, err := os.ReadFile(s.file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return reg, nil
		}
		return reg, fmt.Errorf("Не удалось прочитать реестр %s: %w", s.file, err)
	}
	if err := json.Unmarshal(raw, &reg); err != nil {
		return reg, fmt.Errorf("Реестр %s невалиден:\n- %s", s.file, err.Error())
	}
	if reg.Projects == nil {
		reg.Projects = map[string]ProjectEntry{}
	}
	for name, entry := range reg.Projects {
		if entry.Path == "" {
			return reg, fmt.Errorf("Реестр %s невалиден:\n- projects.%s.path: обязательное поле", s.file, name)
		}
		if entry.Stands == nil {
			entry.Stands = DefaultStands()
			reg.Projects[name] = entry
		}
	}
	return reg, nil
}

// Save сохраняет реестр атомарно (временный файл + rename).
func (s *Service) Save(reg Registry) error {
	if reg.Projects == nil {
		reg.Projects = map[string]ProjectEntry{}
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	return writeFileAtomic(s.file, append(data, '\n'))
}

// List возвращает карту проектов.
func (s *Service) List() (map[string]ProjectEntry, error) {
	reg, err := s.Load()
	if err != nil {
		return nil, err
	}
	return reg.Projects, nil
}

// Get возвращает проект по имени (ok=false, если нет).
func (s *Service) Get(name string) (ProjectEntry, bool, error) {
	reg, err := s.Load()
	if err != nil {
		return ProjectEntry{}, false, err
	}
	e, ok := reg.Projects[name]
	return e, ok, nil
}

// Add добавляет проект. По умолчанию запрещает перезапись.
func (s *Service) Add(name string, entry ProjectEntry, overwrite bool) error {
	reg, err := s.Load()
	if err != nil {
		return err
	}
	if _, exists := reg.Projects[name]; exists && !overwrite {
		return fmt.Errorf("Проект с именем %q уже есть в реестре. Выбери другое имя или обнови существующий проект.", name)
	}
	reg.Projects[name] = entry
	return s.Save(reg)
}

// Remove удаляет проект.
func (s *Service) Remove(name string) error {
	reg, err := s.Load()
	if err != nil {
		return err
	}
	if _, ok := reg.Projects[name]; !ok {
		return fmt.Errorf("Проект %q не найден в реестре.", name)
	}
	delete(reg.Projects, name)
	return s.Save(reg)
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".registry-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
