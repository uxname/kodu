// Package fsutil обходит проект и фильтрует файлы для упаковки/очистки.
//
// Паритет с src/core/file-system/fs.service.ts. Отличие реализации: вместо
// связки tinyglobby + npm-пакета ignore используется один авторитетный
// gitignore-матчер go-git (он покрывает и спецификацию gitignore, и
// разворачивание паттернов, которое в TS делал buildGlobIgnorePatterns).
package fsutil

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/uxname/kodu/internal/config"
	"github.com/uxname/kodu/internal/sortx"
)

// Finder ищет файлы проекта относительно корневой директории.
type Finder struct {
	Root string        // корень (обычно cwd)
	Cfg  config.Config // конфиг для дефолтов packer
	Warn func(string)  // опционально: предупреждения (крупные файлы)
}

// FindOptions переопределяют поведение поиска. nil-указатели означают
// «использовать значение из конфига» (паритет оператора ?? в TS).
type FindOptions struct {
	Ignore                      []string
	UseGitignore                *bool
	ExcludeBinary               *bool
	ContentBasedBinaryDetection *bool
	MaxFileSizeBytes            int64 // 0 → MaxFileSizeBytes по умолчанию
	RootPaths                   []string
}

// New создаёт Finder.
func New(root string, cfg config.Config, warn func(string)) *Finder {
	return &Finder{Root: root, Cfg: cfg, Warn: warn}
}

// Find возвращает отсортированный список относительных POSIX-путей к текстовым файлам.
func (f *Finder) Find(opts FindOptions) ([]string, error) {
	useGitignore := f.Cfg.Packer.UseGitignore
	if opts.UseGitignore != nil {
		useGitignore = *opts.UseGitignore
	}

	var gitignorePatterns []string
	if useGitignore {
		gitignorePatterns = f.readIgnoreFile(".gitignore")
	}
	koduignorePatterns := f.readIgnoreFile(".koduignore")

	baseIgnore := opts.Ignore
	if baseIgnore == nil {
		baseIgnore = f.Cfg.Packer.Ignore
	}

	combined := make([]string, 0, len(baseIgnore)+len(gitignorePatterns)+len(koduignorePatterns))
	combined = append(combined, normalizeIgnorePatterns(baseIgnore)...)
	combined = append(combined, gitignorePatterns...)
	combined = append(combined, koduignorePatterns...)

	patterns := make([]gitignore.Pattern, 0, len(combined))
	for _, p := range combined {
		p = strings.ReplaceAll(p, "\\", "/")
		patterns = append(patterns, gitignore.ParsePattern(p, nil))
	}
	matcher := gitignore.NewMatcher(patterns)

	excludeBinary := true
	if opts.ExcludeBinary != nil {
		excludeBinary = *opts.ExcludeBinary
	}
	useContent := f.Cfg.Packer.ContentBasedBinaryDetection
	if opts.ContentBasedBinaryDetection != nil {
		useContent = *opts.ContentBasedBinaryDetection
	}
	maxSize := opts.MaxFileSizeBytes
	if maxSize == 0 {
		maxSize = MaxFileSizeBytes
	}

	rootPrefixes := normalizeRootPaths(opts.RootPaths)

	var result []string
	walkErr := filepath.WalkDir(f.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Недоступную запись пропускаем, обход продолжаем.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if path == f.Root {
			return nil
		}

		rel, relErr := filepath.Rel(f.Root, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")

		if d.IsDir() {
			// .git всегда игнорируется (GLOB_IGNORE = ['.git/**']).
			if d.Name() == ".git" || matcher.Match(parts, true) {
				return fs.SkipDir
			}
			return nil
		}

		if matcher.Match(parts, false) {
			return nil
		}
		if !underRootPaths(rel, rootPrefixes) {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}
		if info.Size() > maxSize {
			if f.Warn != nil {
				f.Warn(fmt.Sprintf("Skipping large file: %s (>%dMB)", rel, maxSize/(1024*1024)))
			}
			return nil
		}

		if excludeBinary && f.shouldExcludeBinary(rel, path, useContent) {
			return nil
		}

		result = append(result, rel)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sortx.LocaleStrings(result)
	return result, nil
}

// ReadFileRelative читает файл по относительному пути от корня.
func (f *Finder) ReadFileRelative(rel string) (string, error) {
	abs := filepath.Join(f.Root, filepath.FromSlash(rel))
	b, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// normalizeRootPaths приводит rootPaths к POSIX-префиксам (паритет `${p}/**`).
func normalizeRootPaths(roots []string) []string {
	if len(roots) == 0 {
		return nil
	}
	out := make([]string, 0, len(roots))
	for _, r := range roots {
		r = filepath.ToSlash(strings.TrimSuffix(r, "/"))
		if r == "" || r == "." {
			return nil // эквивалент отсутствия ограничения
		}
		out = append(out, r)
	}
	return out
}

// underRootPaths повторяет семантику паттерна `${p}/**`: файл должен лежать
// строго внутри одного из rootPaths.
func underRootPaths(rel string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}
	for _, p := range prefixes {
		if strings.HasPrefix(rel, p+"/") {
			return true
		}
	}
	return false
}

func normalizeIgnorePatterns(patterns []string) []string {
	out := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (f *Finder) readIgnoreFile(name string) []string {
	file, err := os.Open(filepath.Join(f.Root, name))
	if err != nil {
		return nil
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func (f *Finder) shouldExcludeBinary(rel, abs string, detectByContent bool) bool {
	if isKnownTextFile(rel) {
		return false
	}
	if isBinaryExtension(rel) {
		return true
	}
	if !detectByContent {
		return false
	}
	return hasNullByte(abs)
}

func isBinaryExtension(rel string) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	if ext == "" {
		return false
	}
	_, ok := binaryExtensions[ext]
	return ok
}

func isKnownTextFile(rel string) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	if ext != "" {
		if _, ok := knownTextExtensions[ext]; ok {
			return true
		}
	}
	base := strings.ToLower(filepath.Base(rel))
	_, ok := knownTextExtensions[base]
	return ok
}

// hasNullByte читает первые binaryProbeSize байт и ищет 0x00.
// При ошибке чтения считает файл бинарным (как catch → true в TS).
func hasNullByte(abs string) bool {
	file, err := os.Open(abs)
	if err != nil {
		return true
	}
	defer func() { _ = file.Close() }()

	buf := make([]byte, binaryProbeSize)
	n, err := file.Read(buf)
	if err != nil && n == 0 {
		// Пустой файл (EOF) — не бинарный; иная ошибка чтения — бинарный.
		if errors.Is(err, io.EOF) {
			return false
		}
		return true
	}
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}
