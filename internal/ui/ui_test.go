package ui

import (
	"bytes"
	"strings"
	"testing"
)

// Логи и данные должны идти в разные потоки: статус — в stderr, данные — в stdout.
func TestStreamSeparation(t *testing.T) {
	var out, errw bytes.Buffer
	u := newWith(&out, &errw, Options{NoColor: true})

	u.Println("data-line")
	u.Info("status-line")
	u.Success("done")

	if got := out.String(); got != "data-line\n" {
		t.Fatalf("stdout = %q, хотел только данные", got)
	}
	if errStr := errw.String(); !strings.Contains(errStr, "status-line") || !strings.Contains(errStr, "done") {
		t.Fatalf("stderr = %q, хотел статус-сообщения", errStr)
	}
	if strings.Contains(out.String(), "status-line") {
		t.Fatalf("статус-сообщение протекло в stdout")
	}
}

// На не-TTY цвет отключается, а спиннер становится no-op.
func TestNoColorAndNoSpinnerOffTTY(t *testing.T) {
	var out, errw bytes.Buffer
	u := newWith(&out, &errw, Options{})
	if u.color {
		t.Fatal("цвет должен быть выключен на не-TTY writer")
	}
	sp := u.NewSpinner("work").Start()
	if sp.s != nil {
		t.Fatal("спиннер должен быть no-op на не-TTY")
	}
	sp.Success("ok") // должно уйти в stderr как обычный статус
	if !strings.Contains(errw.String(), "ok") {
		t.Fatalf("ожидал фолбэк спиннера в stderr, получил %q", errw.String())
	}
}
