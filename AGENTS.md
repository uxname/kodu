# AGENTS.md

Guidelines for AI assistants working on the Kodu project.

## 1. Project Overview

**Kodu** is a high-performance CLI that bridges local development with LLMs:
it bundles project context (`pack`), strips comments deterministically (`clean`),
initializes config (`init`), and manages a registry of projects/stands (`ops`).

- **Key goals:** fast startup, deterministic file ops (no AI for critical paths), good DX.
- **Commands:** `init`, `pack`, `clean`, `ops` (+ subcommands).

## 2. Technology Stack

Single native Go binary. CGO is required (tree-sitter C grammars).

| Category | Library |
|----------|---------|
| CLI engine | `spf13/cobra` |
| Config | `encoding/json` (defaults via `config.Default()`, non-strict) |
| File walk / ignore | `filepath.WalkDir` + `go-git/.../gitignore` matcher |
| AST / comment removal | `smacker/go-tree-sitter` (typescript/tsx grammars) |
| Import graph (`--deps`) | tree-sitter + manual resolver (best-effort) |
| Tokens | `pkoukk/tiktoken-go` + offline loader (o200k_base) |
| Sorting (file order) | `x/text/collate` (matches JS `localeCompare`) |
| Clipboard | `atotto/clipboard` |
| Spinner | `theckman/yacspin` (TTY-gated) |
| Color | `fatih/color` (respects `NO_COLOR` / `--no-color` / non-TTY) |
| Git | `os/exec` wrapper |

## 3. Architecture

```
cmd/kodu/main.go          # entry point
internal/
  cli/                    # cobra commands: root, init, pack, clean, ops
  config/                 # kodu.json loader + defaults
  fsutil/                 # project file finder (walk + gitignore + binary detect)
  cleaner/                # tree-sitter comment removal
  deps/                   # best-effort import graph
  tokenizer/              # tiktoken-go
  packer/                 # xml/text context formatting + templates
  prompt/                 # .kodu/prompts template loading
  git/                    # git wrapper
  registry/               # ~/.config/kodu/registry.json (XDG, atomic write)
  runbook/                # .runbook/ scaffold + templates
  ui/                     # logger (stderr) + spinner; data goes to stdout
  sortx/                  # locale-aware sorting
  buildinfo/              # version vars (set via -ldflags -X)
scripts/parity-check.sh   # parity sweep vs the legacy Node build
```

Dependencies are wired by hand via constructors in `internal/cli` (no DI container).

## 4. Conventions

- **Streams:** status/logs/spinner → **stderr** (`ui.UI.Success/Warn/Error/Info`);
  machine-readable data (file lists, paths, cleaned code) → **stdout** (`ui.UI.Println`).
  Keep `pack -l`, `ops path`, `clean --stdin` free of logs/ANSI.
- **Determinism:** file order uses `sortx.LocaleStrings` (collate), not byte sort.
- **CGO:** build/test/vet with `CGO_ENABLED=1`. tree-sitter parsers do not cross-compile
  trivially — releases build natively per platform (see `.github/workflows/release.yml`).
- **Errors shown to users** may be capitalized/localized (Russian for `ops`/registry);
  `ST1005`/revive `error-strings` are disabled for this reason in `.golangci.yml`.

## 5. Workflow

```
make build    # CGO_ENABLED=1 go build -ldflags=... → dist/kodu
make test     # go test ./...
make lint     # gofmt + go vet + golangci-lint
```

Local git hooks (lefthook): `gofmt` + `go vet` on pre-commit, `go test` on pre-push.

## 6. Parity with the former TypeScript version

Kodu was migrated from NestJS/TypeScript to Go. Behavior is byte-for-byte compatible
for `pack` (xml/text/clean/deps), `clean` (stdin/in-place), `init`, and `ops`. Two
intentional improvements: the comment cleaner is more thorough (tree-sitter finds
comments ts-morph's range-walk missed), and HTML is not parsed as TypeScript (no URL
corruption). See `internal/cleaner/cleaner.go` for the documented differences.
