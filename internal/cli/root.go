package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uxname/kodu/internal/buildinfo"
	"github.com/uxname/kodu/internal/ui"
)

// Execute строит корневую команду со всеми подкомандами и запускает её.
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
		// Цвет зависит от флага --no-color, который виден только после парсинга,
		// поэтому UI создаётся здесь, до запуска RunE подкоманды.
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			app.UI = ui.New(ui.Options{NoColor: noColor})
		},
	}

	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	root.SetVersionTemplate(fmt.Sprintf(
		"kodu %s (commit %s, built %s)\n",
		buildinfo.Version, buildinfo.Commit, buildinfo.Date,
	))

	// Подкоманды регистрируются по мере реализации (init, pack, clean, ops).
	registerCommands(root, app)

	return root
}

// registerCommands добавляет все подкоманды к корню. Наполняется в следующих шагах.
func registerCommands(_ *cobra.Command, _ *App) {}
