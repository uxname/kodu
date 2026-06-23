// Package prompt loads prompt templates (parity with prompt.service.ts).
package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Service reads templates from files or accepts inline text, with a cache.
type Service struct {
	root  string
	mu    sync.Mutex
	cache map[string]string
}

// New creates a service rooted at root (usually the cwd).
func New(root string) *Service {
	return &Service{root: root, cache: map[string]string{}}
}

// Load treats source as a file path or as an inline prompt.
func (s *Service) Load(source string) (string, error) {
	resolved := source
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(s.root, source)
	}
	if s.exists(resolved) {
		return s.readAndCache(resolved)
	}
	if looksLikeInline(source) {
		return source, nil
	}
	rel, _ := filepath.Rel(s.root, resolved)
	return "", fmt.Errorf("prompt file not found: %s", rel)
}

// LoadFromPromptsDir looks for a template in .kodu/prompts (name.md, name.txt, or name with an extension).
func (s *Service) LoadFromPromptsDir(name string) (string, error) {
	var names []string
	if filepath.Ext(name) != "" {
		names = []string{name}
	} else {
		names = []string{name + ".md", name + ".txt"}
	}
	root := filepath.Join(s.root, ".kodu", "prompts")
	candidates := make([]string, len(names))
	for i, n := range names {
		candidates[i] = filepath.Join(root, n)
	}
	for _, c := range candidates {
		if s.exists(c) {
			return s.readAndCache(c)
		}
	}
	rels := make([]string, len(candidates))
	for i, c := range candidates {
		rels[i], _ = filepath.Rel(s.root, c)
	}
	return "", fmt.Errorf("template %s not found. Expected files: %s", name, strings.Join(rels, ", "))
}

func (s *Service) readAndCache(target string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.cache[target]; ok {
		return v, nil
	}
	b, err := os.ReadFile(target)
	if err != nil {
		return "", err
	}
	s.cache[target] = string(b)
	return string(b), nil
}

func (s *Service) exists(target string) bool {
	_, err := os.Stat(target)
	return err == nil
}

// looksLikeInline mirrors the prompt.service.ts heuristic: multiline text,
// or a short string with no slashes and no extension — that's the prompt itself, not a path.
func looksLikeInline(value string) bool {
	if strings.Contains(value, "\n") {
		return true
	}
	hasPathSegments := strings.ContainsAny(value, "/\\")
	hasExtension := filepath.Ext(value) != ""
	return strings.TrimSpace(value) != "" && !hasPathSegments && !hasExtension
}
