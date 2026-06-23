package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uxname/kodu/internal/buildinfo"
	"github.com/uxname/kodu/internal/ui"
)

// Execute builds the root command with all its subcommands and runs it.
func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	app := &App{}
	var noColor bool

	root := &cobra.Command{
		Use:   "kodu",
		Short: "High-performance CLI to prepare a codebase for LLMs",
		Long: "Kodu collects project context into a single file (pack), strips comments\n" +
			"from code (clean), and manages a registry of projects and stands (ops).",
		Version:       buildinfo.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		// Color depends on the --no-color flag, which is only visible after parsing,
		// so the UI is created here, before the subcommand's RunE runs.
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			app.UI = ui.New(ui.Options{NoColor: noColor})
		},
	}

	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	root.SetVersionTemplate(fmt.Sprintf(
		"kodu %s (commit %s, built %s)\n",
		buildinfo.Version, buildinfo.Commit, buildinfo.Date,
	))

	// Subcommands are registered as they are implemented (init, pack, clean, ops).
	registerCommands(root, app)

	return root
}

// registerCommands adds all subcommands to the root.
func registerCommands(root *cobra.Command, app *App) {
	root.AddCommand(newInitCommand(app))
	root.AddCommand(newPackCommand(app))
	root.AddCommand(newCleanCommand(app))
	root.AddCommand(newOpsCommand(app))
}
