package cli

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/spf13/cobra"

	"github.com/uxname/kodu/internal/config"
	"github.com/uxname/kodu/internal/packer"
	"github.com/uxname/kodu/internal/prompt"
	"github.com/uxname/kodu/internal/ui"
)

// newTestApp returns an App whose UI writes into captured buffers (non-TTY,
// so spinners are no-ops and there are no animation goroutines / races).
func newTestApp() (*App, *bytes.Buffer, *bytes.Buffer) {
	var out, errw bytes.Buffer
	return &App{UI: ui.NewWith(&out, &errw, ui.Options{NoColor: true})}, &out, &errw
}

func TestExplainReason(t *testing.T) {
	explain := map[string]string{
		"/cwd/a.ts": "entry point",
		"b.ts":      "imported by a.ts",
	}
	if got := explainReason(explain, "/cwd", "a.ts"); got != "entry point" {
		t.Errorf("abs lookup = %q, want %q", got, "entry point")
	}
	if got := explainReason(explain, "/cwd", "b.ts"); got != "imported by a.ts" {
		t.Errorf("rel fallback = %q, want %q", got, "imported by a.ts")
	}
	if got := explainReason(explain, "/cwd", "missing.ts"); got != "" {
		t.Errorf("unknown file = %q, want empty", got)
	}
}

func TestCleanSpinnerText(t *testing.T) {
	cases := []struct {
		opts cleanOptions
		want string
	}{
		{cleanOptions{staged: true}, "Cleaning staged files..."},
		{cleanOptions{changed: true}, "Cleaning changed files..."},
		{cleanOptions{dryRun: true}, "Analysing..."},
		{cleanOptions{}, "Cleaning..."},
	}
	for _, c := range cases {
		opts := c.opts
		if got := cleanSpinnerText(&opts); got != c.want {
			t.Errorf("cleanSpinnerText(%+v) = %q, want %q", c.opts, got, c.want)
		}
	}
}

func TestNoCleanFilesMessage(t *testing.T) {
	cases := []struct {
		opts cleanOptions
		want string
	}{
		{cleanOptions{staged: true}, "No staged files to clean."},
		{cleanOptions{changed: true}, "No changed files to clean."},
		{cleanOptions{}, "No files to clean."},
	}
	for _, c := range cases {
		opts := c.opts
		if got := noCleanFilesMessage(&opts); got != c.want {
			t.Errorf("noCleanFilesMessage(%+v) = %q, want %q", c.opts, got, c.want)
		}
	}
}

func TestFilterByGitSet(t *testing.T) {
	supported := []string{"a.ts", "b.ts", "c.ts"}

	got, err := filterByGitSet(supported, []string{"b.ts", "x.ts"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"b.ts"}) {
		t.Fatalf("filterByGitSet = %v, want [b.ts]", got)
	}

	// An upstream git error is propagated.
	if _, err := filterByGitSet(supported, nil, errSentinel); err == nil {
		t.Fatal("expected propagated git error")
	}
}

var errSentinel = &gitErr{}

type gitErr struct{}

func (*gitErr) Error() string { return "git failed" }

func TestCollectCleanTargets(t *testing.T) {
	all := []string{"src/a.ts", "src/b.go", "src/c.tsx", "readme.md"}

	// No inputs, no git filters: only supported extensions survive.
	got, err := collectCleanTargets("/cwd", all, nil, &cleanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"src/a.ts", "src/c.tsx"}) {
		t.Fatalf("supported filter = %v", got)
	}

	// Path-prefix input narrows to a subtree.
	got, err = collectCleanTargets("/cwd", all, []string{"src/a.ts"}, &cleanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"src/a.ts"}) {
		t.Fatalf("exact input = %v", got)
	}
}

func TestDefaultKoduJSON(t *testing.T) {
	cfg := defaultKoduJSON()
	if cfg.Schema == "" {
		t.Error("schema URL must be set")
	}
	if !cfg.Cleaner.KeepJSDoc || !cfg.Cleaner.UseGitignore {
		t.Error("cleaner defaults wrong")
	}
	if len(cfg.Cleaner.Whitelist) == 0 || cfg.Cleaner.Whitelist[0] != "//!" {
		t.Errorf("whitelist = %v", cfg.Cleaner.Whitelist)
	}
	if len(cfg.Packer.Ignore) == 0 {
		t.Error("packer ignore list must not be empty")
	}
}

func TestExtractStand(t *testing.T) {
	md := "# Runbook\n\n## Stand: dev (dev stand)\n- a\n- b\n\n## Stand: prod (prod)\n- c\n"

	dev := extractStand(md, "dev")
	if dev == "" || !contains2(dev, "## Stand: dev") || !contains2(dev, "- b") {
		t.Fatalf("dev section = %q", dev)
	}
	if contains2(dev, "## Stand: prod") {
		t.Fatalf("dev section leaked into prod: %q", dev)
	}
	if got := extractStand(md, "missing"); got != "" {
		t.Fatalf("missing stand = %q, want empty", got)
	}
}

func contains2(s, sub string) bool { return bytes.Contains([]byte(s), []byte(sub)) }

func TestContains(t *testing.T) {
	xs := []string{"local", "dev", "prod"}
	if !contains(xs, "dev") {
		t.Error("should contain dev")
	}
	if contains(xs, "stage") {
		t.Error("should not contain stage")
	}
	if contains(nil, "x") {
		t.Error("nil slice contains nothing")
	}
}

func TestCompleteFirstArgProjects(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// With an existing first arg, no completion is offered.
	names, directive := completeFirstArgProjects(&cobra.Command{}, []string{"already"}, "")
	if names != nil || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("expected no completion for second arg, got %v", names)
	}

	// With no args it falls through to project-name completion (empty registry → no names).
	names, directive = completeFirstArgProjects(&cobra.Command{}, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("unexpected directive %v", directive)
	}
	if len(names) != 0 {
		t.Fatalf("empty registry should yield no names, got %v", names)
	}
}

func TestApplyConfiguredPrompt(t *testing.T) {
	app, _, errw := newTestApp()
	ps := prompt.New(t.TempDir())
	tctx := packer.TemplateContext{Context: "RAW CONTEXT"}

	// No prompts configured → raw context is returned unchanged.
	if got := applyConfiguredPrompt(app, ps, config.Config{}, tctx); got != "RAW CONTEXT" {
		t.Fatalf("no-prompt = %q, want raw context", got)
	}

	// A configured-but-missing prompt file warns and falls back to raw context.
	cfg := config.Config{Prompts: &config.Prompts{Pack: "missing.md"}}
	if got := applyConfiguredPrompt(app, ps, cfg, tctx); got != "RAW CONTEXT" {
		t.Fatalf("missing-prompt fallback = %q", got)
	}
	if errw.Len() == 0 {
		t.Fatal("expected a warning for the missing prompt file")
	}
}
