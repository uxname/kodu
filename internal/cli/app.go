// Package cli wires up Kodu's cobra commands and their dependencies.
package cli

import "github.com/uxname/kodu/internal/ui"

// App holds the shared dependencies available to all commands.
// Dependencies are wired up manually via constructors (no reflection/DI container).
type App struct {
	UI *ui.UI
}
