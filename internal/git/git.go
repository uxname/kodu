// Package git — тонкая обёртка над системным git (паритет git.service.ts).
package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// Git выполняет git-команды в заданной директории.
type Git struct {
	Dir string
}

// New создаёт обёртку для каталога dir.
func New(dir string) *Git { return &Git{Dir: dir} }

func (g *Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.Dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

// EnsureRepo проверяет, что мы внутри git-репозитория.
func (g *Git) EnsureRepo() error {
	if _, err := g.run("rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("git repository not found. Initialize git before running the command")
	}
	return nil
}

// ChangedFiles — объединение изменённых (unstaged + staged + untracked) файлов.
func (g *Git) ChangedFiles() ([]string, error) {
	if err := g.EnsureRepo(); err != nil {
		return nil, err
	}
	set := map[string]struct{}{}
	for _, args := range [][]string{
		{"diff", "--name-only"},
		{"diff", "--name-only", "--staged"},
		{"ls-files", "--others", "--exclude-standard"},
	} {
		out, err := g.run(args...)
		if err != nil {
			return nil, err
		}
		addLines(set, out)
	}
	return sortedKeys(set), nil
}

// StagedFiles — только staged-файлы.
func (g *Git) StagedFiles() ([]string, error) {
	if err := g.EnsureRepo(); err != nil {
		return nil, err
	}
	out, err := g.run("diff", "--name-only", "--staged")
	if err != nil {
		return nil, err
	}
	set := map[string]struct{}{}
	addLines(set, out)
	return sortedKeys(set), nil
}

func addLines(set map[string]struct{}, out string) {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			set[line] = struct{}{}
		}
	}
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
