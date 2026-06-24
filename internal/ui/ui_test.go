package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
)

// Logs and data must go to separate streams: status to stderr, data to stdout.
func TestStreamSeparation(t *testing.T) {
	var out, errw bytes.Buffer
	u := NewWith(&out, &errw, Options{NoColor: true})

	u.Println("data-line")
	u.Info("status-line")
	u.Success("done")

	if got := out.String(); got != "data-line\n" {
		t.Fatalf("stdout = %q, wanted only data", got)
	}
	if errStr := errw.String(); !strings.Contains(errStr, "status-line") || !strings.Contains(errStr, "done") {
		t.Fatalf("stderr = %q, wanted status messages", errStr)
	}
	if strings.Contains(out.String(), "status-line") {
		t.Fatalf("status message leaked into stdout")
	}
}

// On a non-TTY, color is disabled and the spinner becomes a no-op.
func TestNoColorAndNoSpinnerOffTTY(t *testing.T) {
	var out, errw bytes.Buffer
	u := NewWith(&out, &errw, Options{})
	if u.color {
		t.Fatal("color must be disabled on a non-TTY writer")
	}
	sp := u.NewSpinner("work").Start()
	if sp.s != nil {
		t.Fatal("spinner must be a no-op on a non-TTY")
	}
	sp.Success("ok") // should go to stderr as a regular status
	if !strings.Contains(errw.String(), "ok") {
		t.Fatalf("expected spinner fallback in stderr, got %q", errw.String())
	}
}

func TestGetters(t *testing.T) {
	var out, errw bytes.Buffer
	u := NewWith(&out, &errw, Options{NoColor: true})
	if u.Out() != &out {
		t.Fatal("Out() should return the configured stdout writer")
	}
	if u.IsTTY() {
		t.Fatal("IsTTY() should be false for a buffer")
	}
}

func TestPrintNoNewline(t *testing.T) {
	var out, errw bytes.Buffer
	u := NewWith(&out, &errw, Options{NoColor: true})
	u.Print("raw")
	if out.String() != "raw" {
		t.Fatalf("Print = %q, want %q (no trailing newline)", out.String(), "raw")
	}
}

// All four status helpers write their symbol and message to stderr, not stdout.
func TestStatusHelpers(t *testing.T) {
	cases := []struct {
		name   string
		call   func(u *UI, msg string)
		symbol string
	}{
		{"Success", (*UI).Success, "✔"},
		{"Warn", (*UI).Warn, "⚠"},
		{"Error", (*UI).Error, "✖"},
		{"Info", (*UI).Info, "ℹ"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errw bytes.Buffer
			u := NewWith(&out, &errw, Options{NoColor: true})
			c.call(u, "the message")
			got := errw.String()
			if !strings.Contains(got, c.symbol) || !strings.Contains(got, "the message") {
				t.Fatalf("%s stderr = %q, want symbol %q + message", c.name, got, c.symbol)
			}
			if out.Len() != 0 {
				t.Fatalf("%s leaked to stdout: %q", c.name, out.String())
			}
		})
	}
}

// With color forced on, status output is wrapped in ANSI escape codes.
func TestStatusColor(t *testing.T) {
	// fatih/color suppresses ANSI globally when stdout isn't a TTY; force it on.
	old := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = old }()

	var out, errw bytes.Buffer
	u := NewWith(&out, &errw, Options{})
	u.color = true // simulate a color-enabled terminal without a real TTY
	u.Error("boom")
	got := errw.String()
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("expected ANSI color codes, got %q", got)
	}
	if !strings.Contains(got, "boom") {
		t.Fatalf("message missing: %q", got)
	}
}

// On a non-TTY the spinner methods are no-ops or fall back to plain status lines.
func TestSpinnerNonTTY(t *testing.T) {
	var out, errw bytes.Buffer
	u := NewWith(&out, &errw, Options{NoColor: true})

	sp := u.NewSpinner("start").Start()
	sp.SetText("update") // no-op, must not panic
	sp.Fail("failed")
	if !strings.Contains(errw.String(), "failed") {
		t.Fatalf("Fail should fall back to UI.Error, got %q", errw.String())
	}

	// Stop is silent (no symbol/text) on a non-TTY.
	errw.Reset()
	sp2 := u.NewSpinner("again").Start()
	sp2.Stop("done")
	if errw.Len() != 0 {
		t.Fatalf("Stop should be silent on non-TTY, got %q", errw.String())
	}
}

func TestNew(t *testing.T) {
	u := New(Options{NoColor: true})
	if u == nil {
		t.Fatal("New returned nil")
	}
	if u.color {
		t.Fatal("NoColor option should disable color")
	}
}

// Force the TTY branch (without starting the animation goroutine, to stay
// race-free) so the real-spinner code paths are exercised.
func TestSpinnerTTYBranch(t *testing.T) {
	var out, errw bytes.Buffer
	u := NewWith(&out, &errw, Options{NoColor: true})
	u.tty = true
	u.color = true // also exercise the colored spinner config branch

	sp := u.NewSpinner("work")
	if sp.s == nil {
		t.Fatal("expected a real spinner when tty is true")
	}
	sp.SetText("progress") // s.Message branch
	sp.Stop("stopped")     // s != nil branch of Stop

	sp2 := u.NewSpinner("work2")
	sp2.Success("ok") // s != nil branch of Success

	sp3 := u.NewSpinner("work3")
	sp3.Fail("nope") // s != nil branch of Fail
}

func TestNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var out, errw bytes.Buffer
	// Even without NoColor option, NO_COLOR env disables color (and buffers aren't TTYs anyway).
	u := NewWith(&out, &errw, Options{})
	if u.color {
		t.Fatal("color must be disabled when NO_COLOR is set")
	}
}
