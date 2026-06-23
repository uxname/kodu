// Package ui отвечает за вывод пользователю: статус-логи и спиннер.
//
// Принцип разделения потоков:
//   - все статус-сообщения (success/warn/error/info) и спиннер идут в STDERR;
//   - данные (списки файлов, пути, очищенный код) команды пишут в STDOUT сами.
//
// Так `kodu ops path x`, `kodu pack -l`, `kodu clean --stdin` дают чистый
// stdout, пригодный для пайпов и подстановки команд.
package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/theckman/yacspin"
)

// UI инкапсулирует поток вывода и настройки цвета.
type UI struct {
	out   io.Writer // stdout: данные
	err   io.Writer // stderr: статус/логи/спиннер
	color bool      // включён ли цвет для статус-логов
	tty   bool      // является ли stderr интерактивным терминалом
}

// Options управляет поведением вывода.
type Options struct {
	// NoColor принудительно отключает цвет (флаг --no-color).
	NoColor bool
}

// New создаёт UI поверх os.Stdout/os.Stderr с авто-детектом TTY и цвета.
//
// Цвет включается, только если stderr — это терминал, не задан NO_COLOR
// (https://no-color.org) и не передан --no-color.
func New(opts Options) *UI {
	return newWith(os.Stdout, os.Stderr, opts)
}

func newWith(out, errw io.Writer, opts Options) *UI {
	tty := isWriterTTY(errw)
	_, noColorEnv := os.LookupEnv("NO_COLOR")
	useColor := tty && !noColorEnv && !opts.NoColor
	return &UI{out: out, err: errw, color: useColor, tty: tty}
}

// Out возвращает поток для данных (stdout).
func (u *UI) Out() io.Writer { return u.out }

// IsTTY сообщает, интерактивен ли терминал (stderr).
func (u *UI) IsTTY() bool { return u.tty }

// Print пишет данные в stdout без префиксов и перевода строки.
func (u *UI) Print(s string) { _, _ = fmt.Fprint(u.out, s) }

// Println пишет строку данных в stdout с переводом строки.
func (u *UI) Println(s string) { _, _ = fmt.Fprintln(u.out, s) }

func (u *UI) status(c *color.Color, symbol, msg string) {
	if u.color {
		_, _ = fmt.Fprintf(u.err, "%s %s\n", c.Sprint(symbol), msg)
		return
	}
	_, _ = fmt.Fprintf(u.err, "%s %s\n", symbol, msg)
}

// Success — зелёная галочка в stderr.
func (u *UI) Success(msg string) { u.status(color.New(color.FgGreen), "✔", msg) }

// Warn — жёлтое предупреждение в stderr.
func (u *UI) Warn(msg string) { u.status(color.New(color.FgYellow), "⚠", msg) }

// Error — красная ошибка в stderr.
func (u *UI) Error(msg string) { u.status(color.New(color.FgRed), "✖", msg) }

// Info — голубое информационное сообщение в stderr.
func (u *UI) Info(msg string) { u.status(color.New(color.FgCyan), "ℹ", msg) }

// Spinner — крутилка прогресса. На не-TTY превращается в no-op,
// чтобы не засорять неинтерактивный вывод.
type Spinner struct {
	s  *yacspin.Spinner
	ui *UI
}

// NewSpinner создаёт спиннер с начальным текстом. Возвращает no-op спиннер,
// если stderr не является терминалом.
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

// Start запускает анимацию (no-op без TTY).
func (sp *Spinner) Start() *Spinner {
	if sp.s != nil {
		_ = sp.s.Start()
	}
	return sp
}

// SetText обновляет сообщение спиннера.
func (sp *Spinner) SetText(text string) {
	if sp.s != nil {
		sp.s.Message(text)
	}
}

// Success останавливает спиннер с галочкой и сообщением.
func (sp *Spinner) Success(msg string) {
	if sp.s != nil {
		sp.s.StopMessage(msg)
		_ = sp.s.Stop()
		return
	}
	sp.ui.Success(msg)
}

// Fail останавливает спиннер с крестиком и сообщением.
func (sp *Spinner) Fail(msg string) {
	if sp.s != nil {
		sp.s.StopFailMessage(msg)
		_ = sp.s.StopFail()
		return
	}
	sp.ui.Error(msg)
}

// Stop тихо останавливает спиннер (без финального символа).
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
