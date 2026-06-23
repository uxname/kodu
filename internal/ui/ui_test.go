package ui

import (
	"bytes"
	"strings"
	"testing"
)

// Logs and data must go to separate streams: status to stderr, data to stdout.
func TestStreamSeparation(t *testing.T) {
	var out, errw bytes.Buffer
	u := newWith(&out, &errw, Options{NoColor: true})

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
	u := newWith(&out, &errw, Options{})
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
