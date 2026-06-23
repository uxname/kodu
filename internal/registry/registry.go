// Package registry stores the global project registry (parity with registry.service.ts
// + registry.schema.ts) in ~/.config/kodu/registry.json (respects XDG_CONFIG_HOME).
package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultStands is the standard set of stands (registry.schema.ts).
func DefaultStands() []string { return []string{"local", "dev", "stage", "prod"} }

// ProjectEntry is a single project in the registry.
type ProjectEntry struct {
	Path   string   `json:"path"`
	Repo   string   `json:"repo,omitempty"`
	Stands []string `json:"stands"`
}

// Registry is the entire registry.
type Registry struct {
	Schema   string                  `json:"$schema,omitempty"`
	Projects map[string]ProjectEntry `json:"projects"`
}

// Service reads and writes the registry.
type Service struct {
	dir  string
	file string
}

// New creates the service, computing the path via XDG.
func New() *Service {
	base := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	dir := filepath.Join(base, "kodu")
	return &Service{dir: dir, file: filepath.Join(dir, "registry.json")}
}

// FilePath returns the path to the registry file.
func (s *Service) FilePath() string { return s.file }

// Load loads the registry; if the file does not exist, returns an empty registry.
func (s *Service) Load() (Registry, error) {
	reg := Registry{Projects: map[string]ProjectEntry{}}
	raw, err := os.ReadFile(s.file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return reg, nil
		}
		return reg, fmt.Errorf("failed to read registry %s: %w", s.file, err)
	}
	if err := json.Unmarshal(raw, &reg); err != nil {
		return reg, fmt.Errorf("registry %s is invalid:\n- %s", s.file, err.Error())
	}
	if reg.Projects == nil {
		reg.Projects = map[string]ProjectEntry{}
	}
	for name, entry := range reg.Projects {
		if entry.Path == "" {
			return reg, fmt.Errorf("registry %s is invalid:\n- projects.%s.path: required field", s.file, name)
		}
		if entry.Stands == nil {
			entry.Stands = DefaultStands()
			reg.Projects[name] = entry
		}
	}
	return reg, nil
}

// Save persists the registry atomically (temp file + rename).
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

// List returns the map of projects.
func (s *Service) List() (map[string]ProjectEntry, error) {
	reg, err := s.Load()
	if err != nil {
		return nil, err
	}
	return reg.Projects, nil
}

// Get returns a project by name (ok=false if not found).
func (s *Service) Get(name string) (ProjectEntry, bool, error) {
	reg, err := s.Load()
	if err != nil {
		return ProjectEntry{}, false, err
	}
	e, ok := reg.Projects[name]
	return e, ok, nil
}

// Add adds a project. By default it forbids overwriting.
func (s *Service) Add(name string, entry ProjectEntry, overwrite bool) error {
	reg, err := s.Load()
	if err != nil {
		return err
	}
	if _, exists := reg.Projects[name]; exists && !overwrite {
		return fmt.Errorf("A project named %q already exists in the registry. Choose a different name or update the existing project.", name)
	}
	reg.Projects[name] = entry
	return s.Save(reg)
}

// Remove deletes a project.
func (s *Service) Remove(name string) error {
	reg, err := s.Load()
	if err != nil {
		return err
	}
	if _, ok := reg.Projects[name]; !ok {
		return fmt.Errorf("Project %q not found in the registry.", name)
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
