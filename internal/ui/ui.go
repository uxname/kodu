// Package ui handles output to the user: status logs and the spinner.
//
// Stream separation principle:
//   - all status messages (success/warn/error/info) and the spinner go to STDERR;
//   - data (file lists, paths, cleaned code) is written to STDOUT by the commands themselves.
//
// This way `kodu ops path x`, `kodu pack -l`, `kodu clean --stdin` produce clean
// stdout, suitable for pipes and command substitution.
package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/theckman/yacspin"
)

// UI encapsulates the output streams and color settings.
type UI struct {
	out   io.Writer // stdout: data
	err   io.Writer // stderr: status/logs/spinner
	color bool      // whether color is enabled for status logs
	tty   bool      // whether stderr is an interactive terminal
}

// Options controls output behavior.
type Options struct {
	// NoColor forcibly disables color (the --no-color flag).
	NoColor bool
}

// New creates a UI over os.Stdout/os.Stderr with auto-detection of TTY and color.
//
// Color is enabled only if stderr is a terminal, NO_COLOR is not set
// (https://no-color.org), and --no-color was not passed.
func New(opts Options) *UI {
	return newWith(os.Stdout, os.Stderr, opts)
}

func newWith(out, errw io.Writer, opts Options) *UI {
	tty := isWriterTTY(errw)
	_, noColorEnv := os.LookupEnv("NO_COLOR")
	useColor := tty && !noColorEnv && !opts.NoColor
	return &UI{out: out, err: errw, color: useColor, tty: tty}
}

// Out returns the data stream (stdout).
func (u *UI) Out() io.Writer { return u.out }

// IsTTY reports whether the terminal (stderr) is interactive.
func (u *UI) IsTTY() bool { return u.tty }

// Print writes data to stdout without prefixes or a trailing newline.
func (u *UI) Print(s string) { _, _ = fmt.Fprint(u.out, s) }

// Println writes a line of data to stdout with a trailing newline.
func (u *UI) Println(s string) { _, _ = fmt.Fprintln(u.out, s) }

func (u *UI) status(c *color.Color, symbol, msg string) {
	if u.color {
		_, _ = fmt.Fprintf(u.err, "%s %s\n", c.Sprint(symbol), msg)
		return
	}
	_, _ = fmt.Fprintf(u.err, "%s %s\n", symbol, msg)
}

// Success — a green check mark on stderr.
func (u *UI) Success(msg string) { u.status(color.New(color.FgGreen), "✔", msg) }

// Warn — a yellow warning on stderr.
func (u *UI) Warn(msg string) { u.status(color.New(color.FgYellow), "⚠", msg) }

// Error — a red error on stderr.
func (u *UI) Error(msg string) { u.status(color.New(color.FgRed), "✖", msg) }

// Info — a cyan informational message on stderr.
func (u *UI) Info(msg string) { u.status(color.New(color.FgCyan), "ℹ", msg) }

// Spinner is a progress spinner. On a non-TTY it becomes a no-op,
// so it doesn't clutter non-interactive output.
type Spinner struct {
	s  *yacspin.Spinner
	ui *UI
}

// NewSpinner creates a spinner with initial text. Returns a no-op spinner
// if stderr is not a terminal.
func (u *UI) NewSpinner(text string) *Spinner {
	if !u.tty {
		return &Spinner{ui: u}
	}
	cfg := yacspin.Config{
		Frequency:         100_000_000, // 100ms
		Writer:            u.err,
		CharSet:           yacspin.CharSets[14],
		Suffix:            " ",
		Message:           text,
		StopCharacter:     "✔",
		StopFailMessage:   text,
		StopFailCharacter: "✖",
	}
	if u.color {
		cfg.Colors = []string{"fgCyan"}
		cfg.StopColors = []string{"fgGreen"}
		cfg.StopFailColors = []string{"fgRed"}
	}
	s, err := yacspin.New(cfg)
	if err != nil {
		return &Spinner{ui: u}
	}
	return &Spinner{s: s, ui: u}
}

// Start begins the animation (no-op without a TTY).
func (sp *Spinner) Start() *Spinner {
	if sp.s != nil {
		_ = sp.s.Start()
	}
	return sp
}

// SetText updates the spinner's message.
func (sp *Spinner) SetText(text string) {
	if sp.s != nil {
		sp.s.Message(text)
	}
}

// Success stops the spinner with a check mark and a message.
func (sp *Spinner) Success(msg string) {
	if sp.s != nil {
		sp.s.StopMessage(msg)
		_ = sp.s.Stop()
		return
	}
	sp.ui.Success(msg)
}

// Fail stops the spinner with a cross and a message.
func (sp *Spinner) Fail(msg string) {
	if sp.s != nil {
		sp.s.StopFailMessage(msg)
		_ = sp.s.StopFail()
		return
	}
	sp.ui.Error(msg)
}

// Stop silently stops the spinner (without a final symbol).
func (sp *Spinner) Stop(msg string) {
	if sp.s != nil {
		sp.s.StopMessage(msg)
		_ = sp.s.Stop()
	}
}

func isWriterTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
