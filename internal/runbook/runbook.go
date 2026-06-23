// Package runbook управляет per-project директорией .runbook/ (паритет
// runbook.service.ts): скаффолд config.json + runbook.md, детекция стека,
// идемпотентная настройка .gitignore.
package runbook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	runbookDir     = ".runbook"
	gitignoreEntry = "/.runbook/"
)

// gitignoreEquivalents — варианты записи `.runbook/`, считающиеся дублем.
var gitignoreEquivalents = map[string]bool{
	"/.runbook/": true, "/.runbook": true, ".runbook/": true, ".runbook": true,
}

// GitignoreResult — итог настройки .gitignore.
type GitignoreResult string

// Возможные исходы EnsureGitignore.
const (
	GitignoreCreated GitignoreResult = "created"
	GitignoreAdded   GitignoreResult = "added"
	GitignorePresent GitignoreResult = "present"
	GitignoreNoGit   GitignoreResult = "no-git"
)

// ProjectConfig — содержимое .runbook/config.json.
type ProjectConfig struct {
	Schema      string   `json:"$schema,omitempty"`
	Project     string   `json:"project"`
	ActiveStand string   `json:"activeStand"`
	Stands      []string `json:"stands"`
}

// Service инкапсулирует работу с .runbook/.
type Service struct{}

// New создаёт сервис.
func New() *Service { return &Service{} }

// DirPath возвращает путь к .runbook/.
func (s *Service) DirPath(root string) string { return filepath.Join(root, runbookDir) }

// ConfigPath возвращает путь к .runbook/config.json.
func (s *Service) ConfigPath(root string) string {
	return filepath.Join(root, runbookDir, "config.json")
}

// RunbookPath возвращает путь к .runbook/runbook.md.
func (s *Service) RunbookPath(root string) string {
	return filepath.Join(root, runbookDir, "runbook.md")
}

// Exists сообщает, инициализирован ли проект (есть config.json).
func (s *Service) Exists(root string) bool {
	_, err := os.Stat(s.ConfigPath(root))
	return err == nil
}

// ReadConfig читает и нормализует config.json.
func (s *Service) ReadConfig(root string) (ProjectConfig, error) {
	var cfg ProjectConfig
	raw, err := os.ReadFile(s.ConfigPath(root))
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, err
	}
	if cfg.ActiveStand == "" {
		cfg.ActiveStand = "local"
	}
	if cfg.Stands == nil {
		cfg.Stands = defaultStands()
	}
	return cfg, nil
}

// WriteConfig пишет config.json.
func (s *Service) WriteConfig(cfg ProjectConfig, root string) error {
	if cfg.ActiveStand == "" {
		cfg.ActiveStand = "local"
	}
	if cfg.Stands == nil {
		cfg.Stands = defaultStands()
	}
	if err := os.MkdirAll(s.DirPath(root), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.ConfigPath(root), append(data, '\n'), 0o644)
}

// ReadRunbook читает runbook.md.
func (s *Service) ReadRunbook(root string) (string, error) {
	b, err := os.ReadFile(s.RunbookPath(root))
	return string(b), err
}

// Scaffold создаёт .runbook/ (config.json + runbook.md). Существующий
// runbook.md не перезаписывается.
func (s *Service) Scaffold(project string, stands []string, activeStand, root string) error {
	if err := os.MkdirAll(s.DirPath(root), 0o755); err != nil {
		return err
	}
	if err := s.WriteConfig(ProjectConfig{Project: project, ActiveStand: activeStand, Stands: stands}, root); err != nil {
		return err
	}
	if _, err := os.Stat(s.RunbookPath(root)); os.IsNotExist(err) {
		detected := s.DetectStack(root)
		md := RenderRunbook(project, stands, detected)
		if err := os.WriteFile(s.RunbookPath(root), []byte(md), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// DetectStack определяет наличие docker-compose / Dockerfile / .env.example.
func (s *Service) DetectStack(root string) DetectedStack {
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(root, name))
		return err == nil
	}
	return DetectedStack{
		Compose:    exists("docker-compose.yml") || exists("docker-compose.yaml"),
		Dockerfile: exists("Dockerfile"),
		EnvExample: exists(".env.example"),
	}
}

// EnsureGitignore идемпотентно добавляет `.runbook/` в .gitignore.
func (s *Service) EnsureGitignore(root string) (GitignoreResult, error) {
	if _, err := os.Stat(filepath.Join(root, ".git")); os.IsNotExist(err) {
		return GitignoreNoGit, nil
	}

	gitignorePath := filepath.Join(root, ".gitignore")
	content := ""
	fileExists := true
	if b, err := os.ReadFile(gitignorePath); err == nil {
		content = string(b)
	} else {
		fileExists = false
	}

	for _, line := range strings.Split(content, "\n") {
		if gitignoreEquivalents[strings.TrimSpace(line)] {
			return GitignorePresent, nil
		}
	}

	trimmed := strings.TrimRight(content, "\n\r \t")
	next := gitignoreEntry + "\n"
	if trimmed != "" {
		next = trimmed + "\n" + gitignoreEntry + "\n"
	}
	if err := os.WriteFile(gitignorePath, []byte(next), 0o644); err != nil {
		return "", err
	}
	if fileExists {
		return GitignoreAdded, nil
	}
	return GitignoreCreated, nil
}

func defaultStands() []string { return []string{"local", "dev", "stage", "prod"} }
