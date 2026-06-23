// Package buildinfo holds build metadata injected via -ldflags -X.
package buildinfo

// Default values for a "from source" build. In release binaries they are
// overridden by the linker: -X github.com/uxname/kodu/internal/buildinfo.Version=...
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
