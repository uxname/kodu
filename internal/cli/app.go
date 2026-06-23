// Package cli собирает cobra-команды Kodu и их зависимости.
package cli

import "github.com/uxname/kodu/internal/ui"

// App держит общие зависимости, доступные всем командам.
// Зависимости создаются вручную через конструкторы (без рефлексии/DI-контейнера).
type App struct {
	UI *ui.UI
}
