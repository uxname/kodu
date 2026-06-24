package cli

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/uxname/kodu/internal/cleaner"
	"github.com/uxname/kodu/internal/config"
	"github.com/uxname/kodu/internal/fsutil"
	"github.com/uxname/kodu/internal/git"
)

var supportedCleanExt = map[string]bool{
	".ts": true, ".tsx": true, ".js": true, ".jsx": true,
	".mjs": true, ".cjs": true, ".html": true, ".htm": true,
}

type cleanOptions struct {
	dryRun  bool
	changed bool
	staged  bool
	backup  bool
	noJsdoc bool
	verbose bool
	stdin   bool
}

func newCleanCommand(app *App) *cobra.Command {
	opts := &cleanOptions{}
	cmd := &cobra.Command{
		Use:   "clean [files...]",
		Short: "Remove comments from code",
		Example: strings.Join([]string{
			"  kodu clean --dry-run",
			"  kodu clean --changed",
			"  kodu clean src/ --backup",
			"  cat file.ts | kodu clean --stdin",
		}, "\n"),
		RunE: func(_ *cobra.Command, args []string) error {
			return runClean(app, args, opts)
		},
	}
	f := cmd.Flags()
	f.BoolVarP(&opts.dryRun, "dry-run", "d", false, "Show what will be removed")
	f.BoolVarP(&opts.changed, "changed", "c", false, "Clean only git-changed files (staged + unstaged + untracked)")
	f.BoolVarP(&opts.staged, "staged", "s", false, "Clean only git-staged files")
	f.BoolVarP(&opts.backup, "backup", "b", false, "Save originals to .kodu/backup/ before modifying")
	f.BoolVarP(&opts.noJsdoc, "no-jsdoc", "n", false, "Remove JSDoc comments (overrides config keepJSDoc)")
	f.BoolVarP(&opts.verbose, "verbose", "v", false, "Show all removed comments in dry-run (not just first 3)")
	f.BoolVar(&opts.stdin, "stdin", false, "Read from stdin, write cleaned result to stdout")
	return cmd
}

func runClean(app *App, args []string, opts *cleanOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg, err := config.Load(cwd)
	if err != nil {
		return err
	}
	cl := cleaner.New(cfg.Cleaner.Whitelist)
	keepJSDoc := cfg.Cleaner.KeepJSDoc
	if opts.noJsdoc {
		keepJSDoc = false
	}

	if opts.stdin {
		input, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			return rerr
		}
		res := cl.Clean("stdin.ts", string(input), keepJSDoc)
		app.UI.Print(res.Content)
		return nil
	}

	sp := app.UI.NewSpinner(cleanSpinnerText(opts)).Start()

	finder := newFinder(app, cwd, cfg)
	ignore := append(append([]string{}, cfg.Packer.Ignore...), cfg.Cleaner.Ignore...)
	ug := cfg.Cleaner.UseGitignore
	allFiles, err := finder.Find(fsutil.FindOptions{UseGitignore: &ug, Ignore: ignore})
	if err != nil {
		sp.Fail("Error during cleaning")
		return err
	}

	targets, err := collectCleanTargets(cwd, allFiles, args, opts)
	if err != nil {
		sp.Fail("Error during cleaning")
		return err
	}

	if len(targets) == 0 {
		msg := noCleanFilesMessage(opts)
		sp.Stop(msg)
		app.UI.Warn(msg)
		return nil
	}

	summary := cleanFiles(finder, cl, targets, keepJSDoc, opts, func(cur, total int) {
		sp.SetText(fmt.Sprintf("%s (%d/%d)", cleanSpinnerText(opts), cur, total))
	})

	if opts.dryRun {
		sp.Success("Analysis complete")
	} else {
		sp.Success("Cleaning complete")
	}

	bytesSaved := summary.bytesBefore - summary.bytesAfter
	tokensSaved := int(math.Round(float64(bytesSaved) / 4))

	if opts.dryRun {
		app.UI.Info(fmt.Sprintf("Files affected: %d/%d, comments: %d",
			summary.filesChanged, summary.filesProcessed, summary.commentsRemoved))
		app.UI.Info(fmt.Sprintf("Bytes saved: %d (~%d tokens)", bytesSaved, tokensSaved))
		for _, r := range summary.reports {
			if r.removed == 0 {
				continue
			}
			previews := r.previews
			more := ""
			if !opts.verbose && len(previews) > 3 {
				more = fmt.Sprintf(" +%d more", len(previews)-3)
				previews = previews[:3]
			}
			quoted := make([]string, len(previews))
			for i, p := range previews {
				quoted[i] = "\"" + p + "\""
			}
			app.UI.Info(fmt.Sprintf("  %s (%d): %s%s", r.file, r.removed, strings.Join(quoted, ", "), more))
		}
		return nil
	}

	app.UI.Success(fmt.Sprintf("Files cleaned: %d, comments removed: %d",
		summary.filesChanged, summary.commentsRemoved))
	app.UI.Info(fmt.Sprintf("Bytes saved: %d (~%d tokens)", bytesSaved, tokensSaved))
	if opts.backup && summary.filesChanged > 0 {
		app.UI.Info("Originals backed up to .kodu/backup/")
	}
	return nil
}

type fileReport struct {
	file     string
	removed  int
	previews []string
}

type cleanSummary struct {
	filesProcessed  int
	filesChanged    int
	commentsRemoved int
	bytesBefore     int
	bytesAfter      int
	reports         []fileReport
}

func cleanFiles(finder *fsutil.Finder, cl *cleaner.Cleaner, files []string, keepJSDoc bool, opts *cleanOptions, onProgress func(cur, total int)) cleanSummary {
	var s cleanSummary
	s.filesProcessed = len(files)
	for i, file := range files {
		onProgress(i+1, len(files))
		original, err := finder.ReadFileRelative(file)
		if err != nil {
			continue
		}
		s.bytesBefore += len(original)
		res := cl.Clean(file, original, keepJSDoc)
		s.bytesAfter += len(res.Content)
		if res.Removed > 0 {
			s.filesChanged++
			s.commentsRemoved += res.Removed
			if !opts.dryRun {
				if opts.backup {
					_ = backupFile(finder.Root, file, original)
				}
				_ = os.WriteFile(filepath.Join(finder.Root, filepath.FromSlash(file)), []byte(res.Content), 0o644)
			}
		}
		s.reports = append(s.reports, fileReport{file: file, removed: res.Removed, previews: res.Previews})
	}
	return s
}

func backupFile(root, file, content string) error {
	target := filepath.Join(root, ".kodu", "backup", filepath.FromSlash(file))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, []byte(content), 0o644)
}

func collectCleanTargets(cwd string, allFiles, inputs []string, opts *cleanOptions) ([]string, error) {
	supported := make([]string, 0, len(allFiles))
	for _, f := range allFiles {
		if supportedCleanExt[strings.ToLower(filepath.Ext(f))] {
			supported = append(supported, f)
		}
	}

	if len(inputs) > 0 {
		var out []string
		for _, f := range supported {
			for _, in := range inputs {
				if f == in || strings.HasPrefix(f, strings.TrimSuffix(in, "/")+"/") {
					out = append(out, f)
					break
				}
			}
		}
		return out, nil
	}

	if opts.staged {
		list, err := git.New(cwd).StagedFiles()
		return filterByGitSet(supported, list, err)
	}
	if opts.changed {
		list, err := git.New(cwd).ChangedFiles()
		return filterByGitSet(supported, list, err)
	}
	return supported, nil
}

func filterByGitSet(supported []string, list []string, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(list))
	for _, f := range list {
		set[f] = true
	}
	var out []string
	for _, f := range supported {
		if set[f] {
			out = append(out, f)
		}
	}
	return out, nil
}

func cleanSpinnerText(opts *cleanOptions) string {
	switch {
	case opts.staged:
		return "Cleaning staged files..."
	case opts.changed:
		return "Cleaning changed files..."
	case opts.dryRun:
		return "Analysing..."
	default:
		return "Cleaning..."
	}
}

func noCleanFilesMessage(opts *cleanOptions) string {
	switch {
	case opts.staged:
		return "No staged files to clean."
	case opts.changed:
		return "No changed files to clean."
	default:
		return "No files to clean."
	}
}
