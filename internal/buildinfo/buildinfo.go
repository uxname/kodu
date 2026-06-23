// Package buildinfo хранит метаданные сборки, проставляемые через -ldflags -X.
package buildinfo

// Значения по умолчанию для сборки «из исходников». В релизных бинарях
// перекрываются линковщиком: -X github.com/uxname/kodu/internal/buildinfo.Version=...
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
