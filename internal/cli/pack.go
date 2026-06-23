package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/uxname/kodu/internal/cleaner"
	"github.com/uxname/kodu/internal/config"
	"github.com/uxname/kodu/internal/deps"
	"github.com/uxname/kodu/internal/fsutil"
	"github.com/uxname/kodu/internal/packer"
	"github.com/uxname/kodu/internal/prompt"
	"github.com/uxname/kodu/internal/tokenizer"
)

type packOptions struct {
	copy      bool
	template  string
	out       string
	paths     []string
	exclude   []string
	list      bool
	clean     bool
	deps      bool
	depsDepth int
	explain   bool
	format    string
}

func newPackCommand(app *App) *cobra.Command {
	opts := &packOptions{}
	cmd := &cobra.Command{
		Use:   "pack [files...]",
		Short: "Collect project context into a single file",
		Example: strings.Join([]string{
			"  kodu pack --copy",
			"  kodu pack src/ -f text -o ctx.txt",
			"  kodu pack --clean",
			"  kodu pack src/main.ts --deps --explain",
		}, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPack(cmd, app, args, opts)
		},
	}
	f := cmd.Flags()
	f.BoolVarP(&opts.copy, "copy", "c", false, "Copy result to clipboard")
	f.StringVarP(&opts.template, "template", "t", "", "Template name from .kodu/prompts")
	f.StringVarP(&opts.out, "out", "o", "", "Path to save result")
	f.StringArrayVarP(&opts.paths, "path", "p", nil, "Directory or glob to include (repeatable)")
	f.StringArrayVarP(&opts.exclude, "exclude", "e", nil, "Additional exclude pattern (repeatable)")
	f.BoolVarP(&opts.list, "list", "l", false, "Print file list only, without content")
	f.BoolVar(&opts.clean, "clean", false, "Strip comments in-memory before packing (files not modified)")
	f.BoolVar(&opts.deps, "deps", false, "Trace imports from entry point(s) and include their dependencies")
	f.IntVar(&opts.depsDepth, "deps-depth", 0, "Max import depth when using --deps (default: unlimited)")
	f.BoolVar(&opts.explain, "explain", false, "Print why each file was included (requires --deps)")
	f.StringVarP(&opts.format, "format", "f", "xml", "Output format: xml (default) or text")
	return cmd
}

func runPack(cmd *cobra.Command, app *App, args []string, opts *packOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg, err := config.Load(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sp := app.UI.NewSpinner("Collecting files...").Start()

	var files []string
	var explainMap map[string]string

	if opts.deps {
		if len(args) == 0 {
			sp.Fail("--deps requires at least one entry file as argument")
			app.UI.Error("Usage: kodu pack <entry.ts> [more.ts...] --deps")
			os.Exit(1)
		}
		sp.SetText("Resolving dependency graph...")
		maxDepth := opts.depsDepth
		if cmd.Flags().Changed("deps-depth") && opts.depsDepth < 1 {
			app.UI.Warn(fmt.Sprintf("Invalid --deps-depth %d, ignoring", opts.depsDepth))
			maxDepth = 0
		}
		res := deps.Collect(args, cwd, deps.Options{MaxDepth: maxDepth, IncludeTypes: true})
		files = res.Files
		explainMap = res.Explain
	} else {
		finder := newFinder(app, cwd, cfg)
		ignore := append(append([]string{}, cfg.Packer.Ignore...), opts.exclude...)
		roots := opts.paths
		if len(args) > 0 {
			roots = args
		}
		ug := cfg.Packer.UseGitignore
		cbd := cfg.Packer.ContentBasedBinaryDetection
		files, err = finder.Find(fsutil.FindOptions{
			Ignore:                      ignore,
			UseGitignore:                &ug,
			ContentBasedBinaryDetection: &cbd,
			RootPaths:                   roots,
		})
		if err != nil {
			sp.Fail("Error collecting context")
			app.UI.Error(err.Error())
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		sp.Stop("No files to pack.")
		app.UI.Warn("No files to pack.")
		return nil
	}

	if opts.list {
		sp.Success(fmt.Sprintf("Found %d files", len(files)))
		for _, file := range files {
			if opts.explain && explainMap != nil {
				app.UI.Println(fmt.Sprintf("%s  ← %s", file, explainReason(explainMap, cwd, file)))
			} else {
				app.UI.Println(file)
			}
		}
		return nil
	}

	format := packer.FormatXML
	if opts.format == "text" {
		format = packer.FormatText
	} else if opts.format != "xml" {
		app.UI.Warn(fmt.Sprintf("Unknown format %q, using \"xml\"", opts.format))
	}

	context, err := buildContext(app, cwd, cfg, files, format, opts.clean)
	if err != nil {
		sp.Fail("Error collecting context")
		app.UI.Error(err.Error())
		os.Exit(1)
	}

	tk := tokenizer.New()
	est, err := tk.Count(context)
	if err != nil {
		sp.Fail("Error collecting context")
		app.UI.Error(err.Error())
		os.Exit(1)
	}

	fileList := strings.Join(files, "\n")
	tctx := packer.TemplateContext{
		Context:     context,
		FileList:    fileList,
		TokenCount:  est.Tokens,
		USDEstimate: est.USDEstimate,
	}

	ps := prompt.New(cwd)
	output := applyConfiguredPrompt(app, ps, cfg, tctx)
	if opts.template != "" {
		tmpl, terr := ps.LoadFromPromptsDir(opts.template)
		if terr != nil {
			sp.Fail("Error collecting context")
			app.UI.Error(terr.Error())
			os.Exit(1)
		}
		output = packer.FillTemplate(tmpl, tctx)
	}

	outputPath, err := writeOutput(cwd, output, opts.out)
	if err != nil {
		sp.Fail("Error collecting context")
		app.UI.Error(err.Error())
		os.Exit(1)
	}

	copied := false
	if opts.copy {
		if err := clipboard.WriteAll(output); err != nil {
			app.UI.Warn("Clipboard unavailable (install xclip/xsel/wl-clipboard): " + err.Error())
		} else {
			copied = true
		}
	}

	sp.Success("Collection complete")

	if opts.explain && explainMap != nil {
		app.UI.Info("Dependency explanation:")
		for _, file := range files {
			app.UI.Info(fmt.Sprintf("  %s  ← %s", file, explainReason(explainMap, cwd, file)))
		}
	}

	app.UI.Info(fmt.Sprintf("Files: %d", len(files)))
	app.UI.Info(fmt.Sprintf("Tokens: %d", est.Tokens))
	app.UI.Info(fmt.Sprintf("Cost estimate: ~$%.4f", est.USDEstimate))
	cleanNote := ""
	if opts.clean {
		cleanNote = " (comments stripped)"
	}
	app.UI.Info(fmt.Sprintf("Format: %s%s", opts.format, cleanNote))
	app.UI.Success("Saved to " + outputPath)
	if copied {
		app.UI.Success("Result copied to clipboard")
	}
	return nil
}

func buildContext(app *App, cwd string, cfg config.Config, files []string, format packer.Format, clean bool) (string, error) {
	finder := newFinder(app, cwd, cfg)
	cl := cleaner.New(cfg.Cleaner.Whitelist)

	out := make([]packer.File, len(files))
	g := new(errgroup.Group)
	g.SetLimit(runtime.NumCPU())
	for i, file := range files {
		i, file := i, file
		g.Go(func() error {
			content, err := finder.ReadFileRelative(file)
			if err != nil {
				return err
			}
			if clean {
				content = cl.Clean(file, content, cfg.Cleaner.KeepJSDoc).Content
			}
			out[i] = packer.File{Path: file, Content: content}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return "", err
	}
	return packer.BuildContext(out, format), nil
}

func applyConfiguredPrompt(app *App, ps *prompt.Service, cfg config.Config, tctx packer.TemplateContext) string {
	if cfg.Prompts == nil || cfg.Prompts.Pack == "" {
		return tctx.Context
	}
	tmpl, err := ps.Load(cfg.Prompts.Pack)
	if err != nil {
		app.UI.Warn(fmt.Sprintf("Prompt file not found: %s, using raw context", cfg.Prompts.Pack))
		return tctx.Context
	}
	return packer.FillTemplate(tmpl, tctx)
}

func writeOutput(cwd, content, outPath string) (string, error) {
	target := outPath
	if target == "" {
		target = filepath.Join(cwd, ".kodu", "context.txt")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(target, []byte(content+"\n"), 0o644); err != nil {
		return "", err
	}
	return target, nil
}

// explainReason mirrors the lookup in pack.command.ts: by abs path, then by rel.
func explainReason(explain map[string]string, cwd, file string) string {
	abs := filepath.Join(cwd, file)
	if r, ok := explain[abs]; ok {
		return r
	}
	return explain[file]
}

func newFinder(app *App, cwd string, cfg config.Config) *fsutil.Finder {
	return fsutil.New(cwd, cfg, func(msg string) { app.UI.Warn(msg) })
}
