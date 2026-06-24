package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/uxname/kodu/internal/registry"
	"github.com/uxname/kodu/internal/runbook"
	"github.com/uxname/kodu/internal/sortx"
)

func newOpsCommand(app *App) *cobra.Command {
	ops := &cobra.Command{
		Use:   "ops",
		Short: "Manage projects and stands (local/dev/stage/prod) from anywhere",
	}
	ops.AddCommand(
		newOpsInitCommand(app),
		newOpsListCommand(app),
		newOpsAddCommand(app),
		newOpsStatusCommand(app),
		newOpsUseCommand(app),
		newOpsPathCommand(app),
		newOpsRunbookCommand(app),
	)
	return ops
}

// completeProjectNames offers project names from the registry for shell completion.
func completeProjectNames(toComplete string) ([]string, cobra.ShellCompDirective) {
	projects, err := registry.New().List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var names []string
	for n := range projects {
		if strings.HasPrefix(n, toComplete) {
			names = append(names, n)
		}
	}
	sortx.LocaleStrings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeFirstArgProjects completes the project name only for the first argument.
func completeFirstArgProjects(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeProjectNames(toComplete)
}

// resolveProjectRoot returns the project's repository path by name.
func resolveProjectRoot(reg *registry.Service, name string) (string, error) {
	entry, ok, err := reg.Get(name)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("project %q not found in the registry. List projects: kodu ops list", name)
	}
	return entry.Path, nil
}

// --- ops init ---

func newOpsInitCommand(app *App) *cobra.Command {
	var name, active string
	var stand []string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up stands in the current project: create .runbook/, .gitignore and register the project",
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			reg, rb := registry.New(), runbook.New()

			projectName := name
			if projectName == "" {
				projectName = filepath.Base(cwd)
			}
			stands := stand
			if len(stands) == 0 {
				stands = registry.DefaultStands()
			}
			activeStand := active
			if activeStand == "" && len(stands) > 0 {
				activeStand = stands[0]
			}
			if activeStand == "" {
				activeStand = "local"
			}

			if err := rb.Scaffold(projectName, stands, activeStand, cwd); err != nil {
				return err
			}
			app.UI.Success(fmt.Sprintf("Created .runbook/ for project %q.", projectName))
			app.UI.Info("Active stand: " + activeStand)
			app.UI.Info("Fill in the deploy steps in " + rb.RunbookPath(cwd))

			res, err := rb.EnsureGitignore(cwd)
			if err != nil {
				return err
			}
			reportGitignore(app, res)

			return registerProject(app, reg, projectName, cwd, stands)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Unique project name (defaults to the folder name)")
	cmd.Flags().StringVarP(&active, "active", "a", "", "Default active stand (defaults to local)")
	cmd.Flags().StringArrayVarP(&stand, "stand", "s", nil, "Project stand (repeatable)")
	return cmd
}

func reportGitignore(app *App, res runbook.GitignoreResult) {
	switch res {
	case runbook.GitignoreCreated:
		app.UI.Success("Created .gitignore with a /.runbook/ entry")
	case runbook.GitignoreAdded:
		app.UI.Success("Added /.runbook/ to .gitignore")
	case runbook.GitignorePresent:
		app.UI.Info("/.runbook/ already in .gitignore")
	case runbook.GitignoreNoGit:
		app.UI.Warn("This is not a git repository — .gitignore was not configured. Don't commit .runbook/ manually.")
	}
}

func registerProject(app *App, reg *registry.Service, name, root string, stands []string) error {
	existing, ok, err := reg.Get(name)
	if err != nil {
		return err
	}
	if !ok {
		if err := reg.Add(name, registry.ProjectEntry{Path: root, Stands: stands}, false); err != nil {
			return err
		}
		app.UI.Success(fmt.Sprintf("Project %q added to the registry.", name))
		return nil
	}
	if existing.Path == root {
		app.UI.Info(fmt.Sprintf("Project %q is already in the registry.", name))
		return nil
	}
	app.UI.Warn(fmt.Sprintf("Name %q is already taken by a different path (%s). Re-run with a different name: kodu ops init --name <other-name>", name, existing.Path))
	return nil
}

// --- ops list ---

func newOpsListCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all projects in the registry and their active stands",
		RunE: func(_ *cobra.Command, _ []string) error {
			reg, rb := registry.New(), runbook.New()
			projects, err := reg.List()
			if err != nil {
				return err
			}
			names := make([]string, 0, len(projects))
			for n := range projects {
				names = append(names, n)
			}
			sortx.LocaleStrings(names)

			if len(names) == 0 {
				app.UI.Info("The registry is empty. Add a project: kodu ops add <name> --path <dir>")
				app.UI.Info("Registry file: " + reg.FilePath())
				return nil
			}
			for _, n := range names {
				entry := projects[n]
				label := ""
				if active := readActiveStand(rb, entry.Path); active != "" {
					label = " [active: " + active + "]"
				}
				app.UI.Info(n + label)
				app.UI.Info("  path:   " + entry.Path)
				app.UI.Info("  stands: " + strings.Join(entry.Stands, ", "))
			}
			return nil
		},
	}
}

func readActiveStand(rb *runbook.Service, root string) string {
	if !rb.Exists(root) {
		return ""
	}
	cfg, err := rb.ReadConfig(root)
	if err != nil {
		return ""
	}
	return cfg.ActiveStand
}

// --- ops add ---

func newOpsAddCommand(app *App) *cobra.Command {
	var path, repo string
	var stand []string
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Register a project in the global registry under a unique name",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a project name: kodu ops add <name> [--path <dir>]")
			}
			name := args[0]
			reg := registry.New()

			projectPath := path
			if projectPath == "" {
				projectPath, _ = os.Getwd()
			}
			projectPath, _ = filepath.Abs(projectPath)
			stands := stand
			if len(stands) == 0 {
				stands = registry.DefaultStands()
			}
			if err := reg.Add(name, registry.ProjectEntry{Path: projectPath, Repo: repo, Stands: stands}, false); err != nil {
				return err
			}
			app.UI.Success(fmt.Sprintf("Project %q added to the registry.", name))
			app.UI.Info("Path: " + projectPath)
			app.UI.Info("Stands: " + strings.Join(stands, ", "))
			return nil
		},
	}
	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to the repository (defaults to the current directory)")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Repository URL (git clone)")
	cmd.Flags().StringArrayVarP(&stand, "stand", "s", nil, "Project stand (repeatable)")
	return cmd
}

// --- ops status ---

func newOpsStatusCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "status [name]",
		Short:             "Show the active stand and the project's stands (by name or in the current folder)",
		ValidArgsFunction: completeFirstArgProjects,
		RunE: func(_ *cobra.Command, args []string) error {
			reg, rb := registry.New(), runbook.New()
			root, _ := os.Getwd()
			if len(args) > 0 {
				r, err := resolveProjectRoot(reg, args[0])
				if err != nil {
					return err
				}
				root = r
			}
			if !rb.Exists(root) {
				return fmt.Errorf("%s has no .runbook/. Initialize the project: kodu ops init", root)
			}
			cfg, err := rb.ReadConfig(root)
			if err != nil {
				return err
			}
			app.UI.Info("Project:      " + cfg.Project)
			app.UI.Info("Active stand: " + cfg.ActiveStand)
			app.UI.Info("Stands:       " + strings.Join(cfg.Stands, ", "))
			app.UI.Info("Path:         " + root)
			return nil
		},
	}
}

// --- ops use ---

func newOpsUseCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "use <args...>",
		Aliases:           []string{"switch"},
		Short:             "Switch the active stand: kodu ops use <stand> or kodu ops use <name> <stand>",
		ValidArgsFunction: completeFirstArgProjects,
		RunE: func(_ *cobra.Command, args []string) error {
			reg, rb := registry.New(), runbook.New()
			root, _ := os.Getwd()
			var stand string
			if len(args) >= 2 {
				r, err := resolveProjectRoot(reg, args[0])
				if err != nil {
					return err
				}
				root, stand = r, args[1]
			} else if len(args) == 1 {
				stand = args[0]
			}

			if stand == "" {
				return errors.New("specify a stand: kodu ops use <stand> or kodu ops use <name> <stand>")
			}
			if !rb.Exists(root) {
				return fmt.Errorf("%s has no .runbook/. Initialize the project: kodu ops init", root)
			}
			cfg, err := rb.ReadConfig(root)
			if err != nil {
				return err
			}
			if !contains(cfg.Stands, stand) {
				cfg.Stands = append(cfg.Stands, stand)
				app.UI.Info(fmt.Sprintf("Stand %q added to the project's list of stands.", stand))
			}
			cfg.ActiveStand = stand
			if err := rb.WriteConfig(cfg, root); err != nil {
				return err
			}
			app.UI.Success(fmt.Sprintf("Active stand for project %q → %s", cfg.Project, stand))
			return nil
		},
	}
}

// --- ops path ---

func newOpsPathCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "path <name>",
		Short:             "Print the project's repository path (handy for cd $(kodu ops path <name>))",
		ValidArgsFunction: completeFirstArgProjects,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a project name: kodu ops path <name>")
			}
			reg := registry.New()
			root, err := resolveProjectRoot(reg, args[0])
			if err != nil {
				return err
			}
			app.UI.Println(root) // clean stdout for command substitution
			return nil
		},
	}
}

// --- ops runbook ---

func newOpsRunbookCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "runbook <name> [stand]",
		Short:             "Print the project's runbook (or a specific stand's section)",
		ValidArgsFunction: completeFirstArgProjects,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("specify a project name: kodu ops runbook <name> [stand]")
			}
			name := args[0]
			var stand string
			if len(args) > 1 {
				stand = args[1]
			}
			reg, rb := registry.New(), runbook.New()
			root, err := resolveProjectRoot(reg, name)
			if err != nil {
				return err
			}
			if !rb.Exists(root) {
				return fmt.Errorf("project %q has no .runbook/. Run: kodu ops init", name)
			}
			md, err := rb.ReadRunbook(root)
			if err != nil {
				return err
			}
			output := md
			if stand != "" {
				output = extractStand(md, stand)
				if output == "" {
					return fmt.Errorf("no section for stand %q found in the runbook", stand)
				}
			}
			app.UI.Println(strings.TrimRight(output, "\n \t"))
			return nil
		},
	}
}

// extractStand returns the `## Stand: <stand> ...` block up to the next `## `.
func extractStand(md, stand string) string {
	lines := strings.Split(md, "\n")
	prefix := "## Stand: " + stand
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			start = i
			break
		}
	}
	if start == -1 {
		return ""
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
