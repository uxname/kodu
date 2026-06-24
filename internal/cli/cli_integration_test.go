package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/uxname/kodu/internal/registry"
	"github.com/uxname/kodu/internal/runbook"
)

// silence suppresses cobra's Usage/error template so error-path tests stay quiet.
func silence(c *cobra.Command) *cobra.Command {
	c.SilenceUsage = true
	c.SilenceErrors = true
	c.SetOut(io.Discard)
	c.SetErr(io.Discard)
	return c
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// --- init ---

func TestInitCommand(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, filepath.Join(dir, ".gitignore"), "node_modules\n")

	app, _, _ := newTestApp()
	if err := newInitCommand(app).Execute(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "kodu.json")); err != nil {
		t.Fatal("kodu.json was not created")
	}
	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if !strings.Contains(string(gi), ".kodu/context.txt") {
		t.Fatalf(".gitignore missing entry:\n%s", gi)
	}

	// Idempotent: a second run does not error and keeps the file.
	if err := newInitCommand(app).Execute(); err != nil {
		t.Fatal(err)
	}
}

// --- pack ---

func setupPackProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.ts"), "// a comment\nexport const x = 1;\n")
	writeFile(t, filepath.Join(dir, "util.ts"), "export const y = 2;\n")
	t.Chdir(dir)
	return dir
}

func TestPackXMLToFile(t *testing.T) {
	dir := setupPackProject(t)
	out := filepath.Join(dir, "ctx.txt")

	app, _, _ := newTestApp()
	cmd := newPackCommand(app)
	cmd.SetArgs([]string{"--out", out})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	if !strings.Contains(got, "<files>") || !strings.Contains(got, "main.ts") || !strings.Contains(got, "util.ts") {
		t.Fatalf("unexpected XML output:\n%s", got)
	}
}

func TestPackTextFormat(t *testing.T) {
	dir := setupPackProject(t)
	out := filepath.Join(dir, "ctx.txt")

	app, _, _ := newTestApp()
	cmd := newPackCommand(app)
	cmd.SetArgs([]string{"--format", "text", "--out", out})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(out)
	if strings.Contains(string(body), "<files>") {
		t.Fatalf("text format must not contain XML tags:\n%s", body)
	}
	if !strings.Contains(string(body), "main.ts") {
		t.Fatalf("text output missing file name:\n%s", body)
	}
}

func TestPackList(t *testing.T) {
	setupPackProject(t)

	app, out, _ := newTestApp()
	cmd := newPackCommand(app)
	cmd.SetArgs([]string{"--list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	stdout := out.String()
	if !strings.Contains(stdout, "main.ts") || !strings.Contains(stdout, "util.ts") {
		t.Fatalf("--list stdout missing file names:\n%s", stdout)
	}
	// --list prints names only, not file content.
	if strings.Contains(stdout, "export const") {
		t.Fatalf("--list leaked file content:\n%s", stdout)
	}
}

func TestPackClean(t *testing.T) {
	dir := setupPackProject(t)
	out := filepath.Join(dir, "ctx.txt")

	app, _, _ := newTestApp()
	cmd := newPackCommand(app)
	cmd.SetArgs([]string{"--clean", "--out", out})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(out)
	if strings.Contains(string(body), "a comment") {
		t.Fatalf("--clean did not strip the comment:\n%s", body)
	}
	if !strings.Contains(string(body), "export const x") {
		t.Fatalf("--clean removed real code:\n%s", body)
	}
}

func TestPackDepsRequiresEntry(t *testing.T) {
	setupPackProject(t)

	app, _, _ := newTestApp()
	cmd := silence(newPackCommand(app))
	cmd.SetArgs([]string{"--deps"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("--deps without an entry file should return an error")
	}
}

// --- clean ---

func TestCleanDryRunDoesNotModify(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.ts")
	original := "// drop me\nexport const z = 3;\n"
	writeFile(t, src, original)
	t.Chdir(dir)

	app, _, _ := newTestApp()
	cmd := newCleanCommand(app)
	cmd.SetArgs([]string{"--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(src)
	if string(after) != original {
		t.Fatalf("--dry-run modified the file:\n%s", after)
	}
}

func TestCleanModifiesAndBacksUp(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.ts")
	writeFile(t, src, "// drop me\nexport const z = 3;\n")
	t.Chdir(dir)

	app, _, _ := newTestApp()
	cmd := newCleanCommand(app)
	cmd.SetArgs([]string{"--backup"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(src)
	if strings.Contains(string(after), "drop me") {
		t.Fatalf("clean did not strip the comment:\n%s", after)
	}
	// Original is preserved under .kodu/backup/.
	backup, err := os.ReadFile(filepath.Join(dir, ".kodu", "backup", "a.ts"))
	if err != nil {
		t.Fatalf("backup not created: %v", err)
	}
	if !strings.Contains(string(backup), "drop me") {
		t.Fatalf("backup should hold the original:\n%s", backup)
	}
}

func TestCleanStdin(t *testing.T) {
	t.Chdir(t.TempDir())

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, _ = w.WriteString("// gone\nexport const q = 9;\n")
		_ = w.Close()
	}()
	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = oldStdin })

	app, out, _ := newTestApp()
	cmd := newCleanCommand(app)
	cmd.SetArgs([]string{"--stdin"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	stdout := out.String()
	if strings.Contains(stdout, "gone") {
		t.Fatalf("--stdin did not strip the comment:\n%s", stdout)
	}
	if !strings.Contains(stdout, "export const q") {
		t.Fatalf("--stdin dropped real code:\n%s", stdout)
	}
}

// --- ops ---

// opsEnv isolates both the global registry (XDG) and the working directory.
func opsEnv(t *testing.T) string {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir := t.TempDir()
	t.Chdir(dir)
	return dir
}

func TestOpsAddAndList(t *testing.T) {
	opsEnv(t)
	app, _, _ := newTestApp()

	add := newOpsAddCommand(app)
	add.SetArgs([]string{"myproj", "--path", "/repo/myproj"})
	if err := add.Execute(); err != nil {
		t.Fatal(err)
	}

	entry, ok, err := registry.New().Get("myproj")
	if err != nil || !ok {
		t.Fatalf("Get after add: ok=%v err=%v", ok, err)
	}
	if entry.Path != "/repo/myproj" {
		t.Fatalf("entry path = %q", entry.Path)
	}

	app2, out, _ := newTestApp()
	list := newOpsListCommand(app2)
	if err := list.Execute(); err != nil {
		t.Fatal(err)
	}
	_ = out // list writes to stderr (status stream); presence in registry already verified

	// Missing name is an error.
	bad := silence(newOpsAddCommand(app))
	bad.SetArgs(nil)
	if err := bad.Execute(); err == nil {
		t.Fatal("ops add without a name should error")
	}
}

func TestOpsInitScaffoldsRunbook(t *testing.T) {
	dir := opsEnv(t)
	app, _, _ := newTestApp()

	cmd := newOpsInitCommand(app)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	rb := runbook.New()
	if !rb.Exists(dir) {
		t.Fatal(".runbook/ was not scaffolded")
	}
	cfg, err := rb.ReadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Project != filepath.Base(dir) {
		t.Fatalf("project name = %q, want folder name %q", cfg.Project, filepath.Base(dir))
	}
	// The project is also registered globally.
	if _, ok, _ := registry.New().Get(filepath.Base(dir)); !ok {
		t.Fatal("ops init did not register the project")
	}
}

func TestOpsUseSwitchesStand(t *testing.T) {
	dir := opsEnv(t)
	app, _, _ := newTestApp()
	if err := newOpsInitCommand(app).Execute(); err != nil {
		t.Fatal(err)
	}

	use := newOpsUseCommand(app)
	use.SetArgs([]string{"stage"})
	if err := use.Execute(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := runbook.New().ReadConfig(dir)
	if cfg.ActiveStand != "stage" {
		t.Fatalf("active stand = %q, want stage", cfg.ActiveStand)
	}

	// A brand-new stand is appended to the list.
	use2 := newOpsUseCommand(app)
	use2.SetArgs([]string{"qa"})
	if err := use2.Execute(); err != nil {
		t.Fatal(err)
	}
	cfg, _ = runbook.New().ReadConfig(dir)
	if !contains(cfg.Stands, "qa") || cfg.ActiveStand != "qa" {
		t.Fatalf("stands = %v, active = %q", cfg.Stands, cfg.ActiveStand)
	}
}

func TestOpsPath(t *testing.T) {
	opsEnv(t)
	app, _, _ := newTestApp()
	add := newOpsAddCommand(app)
	add.SetArgs([]string{"proj", "--path", "/some/where"})
	if err := add.Execute(); err != nil {
		t.Fatal(err)
	}

	app2, out, _ := newTestApp()
	pathCmd := newOpsPathCommand(app2)
	pathCmd.SetArgs([]string{"proj"})
	if err := pathCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != "/some/where" {
		t.Fatalf("ops path stdout = %q, want /some/where", out.String())
	}

	// Unknown project is an error.
	missing := silence(newOpsPathCommand(app2))
	missing.SetArgs([]string{"ghost"})
	if err := missing.Execute(); err == nil {
		t.Fatal("ops path for an unknown project should error")
	}
}

func TestOpsStatus(t *testing.T) {
	dir := opsEnv(t)
	app, _, _ := newTestApp()
	if err := newOpsInitCommand(app).Execute(); err != nil {
		t.Fatal(err)
	}

	status := newOpsStatusCommand(app)
	status.SetArgs(nil)
	if err := status.Execute(); err != nil {
		t.Fatal(err)
	}

	// Status in a directory without .runbook/ errors.
	t.Chdir(t.TempDir())
	bad := silence(newOpsStatusCommand(app))
	bad.SetArgs(nil)
	if err := bad.Execute(); err == nil {
		t.Fatal("ops status without .runbook/ should error")
	}
	_ = dir
}

func TestOpsRunbook(t *testing.T) {
	dir := opsEnv(t)
	app, _, _ := newTestApp()
	if err := newOpsInitCommand(app).Execute(); err != nil {
		t.Fatal(err)
	}
	name := filepath.Base(dir)

	app2, out, _ := newTestApp()
	rbCmd := newOpsRunbookCommand(app2)
	rbCmd.SetArgs([]string{name})
	if err := rbCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "# Runbook:") {
		t.Fatalf("runbook stdout missing header:\n%s", out.String())
	}

	// A specific stand section can be extracted.
	app3, out3, _ := newTestApp()
	standCmd := newOpsRunbookCommand(app3)
	standCmd.SetArgs([]string{name, "dev"})
	if err := standCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out3.String(), "## Stand: dev") {
		t.Fatalf("stand section missing:\n%s", out3.String())
	}
}
