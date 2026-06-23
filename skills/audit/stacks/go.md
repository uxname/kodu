# Stack Profile: Go   (id: go)
Tier: first-class

Go runtime profile. Specific libraries (chi, gqlgen, sqlc, pgx, asynq, slog,
go-oidc) are mentioned only as examples of idioms, not as requirements.

## 1. Detection signals
- `go.mod` (primary marker)
- additional: `go.sum`, `Taskfile.yml`/`Makefile`, `.golangci.yml`, `.go-arch-lint.yml`

## 2. Tooling by category
| Category | Command | How to read the output |
|-----------|---------|------------------|
| unused-code | `deadcode ./... 2>/dev/null \|\| true` (golang.org/x/tools/cmd/deadcode) | unreachable functions → YAGNI-02; verify before FAIL |
| clone-detection | `dupl -threshold 50 ./... 2>/dev/null \|\| true` | token duplicates → REINV-03 |
| dep-audit | `govulncheck ./... 2>/dev/null \|\| true` | CVEs in modules (outside the checklist — for reference) |
| env-extraction | `grep -rEoh 'os\.Getenv\("[A-Z0-9_]+"\)' . 2>/dev/null \| sed -E 's/.*"(.*)".*/\1/' \| sort -u` | env names from code → DOC-02 (account for `env:"..."` tags for caarlos0/env) |
| arch-lint | `go vet ./... 2>/dev/null`; layers: `go-arch-lint check 2>/dev/null \|\| true` | import cycles are forbidden by the compiler → ARC-03 at the package level is auto-PASS |
| lint/format | `golangci-lint run 2>/dev/null \|\| true`; `gofmt -l . 2>/dev/null` | includes `errcheck`, `gosec`, `nilness`, `staticcheck` |
| type-check | `go build ./... 2>/dev/null \|\| true` | compilation = type check |
| test-run | `go test ./... 2>/dev/null \|\| true`; races: `go test -race ./...` | `-race` detects data races → CON-03 |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| trufflehog filesystem . 2>/dev/null \|\| true` | stack-neutral |

Always verify the tool's output manually (`file:line`) before `❌ FAIL`.

## 3. Idioms (what "correct" looks like → PASS)
- **Error handling:** explicit `if err != nil` check; wrapping with `fmt.Errorf("...: %w", err)`; sentinel errors + `errors.Is`/`errors.As`. A returned error is not ignored.
- **Concurrency:** cancellation via `context.Context` (not signals); shared state under `sync.Mutex`/channels; fan-out via `errgroup`/`sync.WaitGroup`; each goroutine has a termination path on `ctx.Done()`.
- **Graceful shutdown:** `signal.NotifyContext(ctx, syscall.SIGTERM, os.Interrupt)`; `server.Shutdown(ctx)`; closing pools (pgx), workers (asynq), redis.
- **Panic safety:** `defer recover()` in HTTP handlers (chi `middleware.Recoverer`), in gqlgen (`RecoverFunc`), in every background goroutine and in the job wrapper (asynq).
- **Env/config:** a centralized config struct (e.g. caarlos0/env), loaded at startup; no scattered `os.Getenv`.
- **Logging:** structured `log/slog` (JSON in prod) with a request/correlation ID via `context`; no `fmt.Print*`/`log.Print*` on request paths.
- **Null-safety:** check nil pointers/interfaces before dereferencing; writing to a nil map is not allowed; map access via `v, ok := m[k]`.
- **Type coercion:** `strconv.Atoi`/`ParseInt` with error checking; type assertion via `v, ok := x.(T)` (comma-ok).
- **Deps / reinvention:** stdlib `slices`/`maps`/`sync`; `database/sql`/sqlc instead of a hand-written query builder.
- **Build/deploy:** multi-stage (builder → distroless/`nonroot`); `go.mod`+`go.sum` committed; `GOFLAGS=-mod=readonly`; `CGO_ENABLED=0` for static linking; no `go get` in the Dockerfile; no `net/http/pprof`/debug routes in prod.

## 4. Anti-patterns (what FAIL looks like)
- **Errors:** ignoring an error via `_ = f()` or `v, _ := f()`; `panic` for ordinary control flow; losing the wrap chain.
- **Concurrency:** goroutine without `WaitGroup`/`errgroup`/`ctx` (result is lost / process does not wait / leak); writing to a map from >1 goroutine without a lock (data race, caught by `go test -race`); `for ... { go f(loopVar) }` — loop variable capture (before Go 1.22); an unclosed channel blocks receivers.
- **Panic:** a panic in a goroutine without `recover()` brings down the whole process.
- **Logging:** `fmt.Print*`/`log.Print*`/`println` on production paths instead of `slog`.
- **Resources:** `defer` inside a loop (the resource is held until the end of the function, not the iteration).
- **Build/deploy:** `go get` in the Dockerfile; missing `go.sum`; compiler/tests in the final image.

## 5. Check-ID hints
- `LOG-01` → `fmt.Print*`, `log.Print*`, `println` outside a `slog` wrapper.
- `BUG-01` → **N/A** if there is no string parsing; otherwise — type assertion `x.(T)` without comma-ok, an ignored `strconv` error.
- `BUG-02` → **N/A** (no promise/forEach syntax); the risk moves to `CON-01`.
- `BUG-03` → nil-deref of a pointer/interface/map; writing to a nil map (panic).
- `BUG-08` → float comparison via `==`; money — in `int64` (cents) or `shopspring/decimal`.
- `BUG-10` → **N/A or Low**: `regexp` (RE2) has no backtracking, ReDoS is practically impossible.
- `CON-01` → goroutine without a termination mechanism (WaitGroup/errgroup/ctx) → goroutine leak.
- `CON-03` → concurrent access to a map/package variable without `sync.Mutex`/atomic (data race).
- `CON-06` → a background goroutine does not listen on `ctx.Done()` → does not terminate on shutdown.
- `ERR-01` → error ignored via `_`.
- `ERR-03` → no recover middleware (a panic in a handler drops the connection/process).
- `ERR-04` → no `defer recover()` in background goroutines/asynq handler.
- `ERR-06` → no `signal.NotifyContext`/`server.Shutdown` (graceful shutdown).
- `ERR-09` → `context.Context` is not propagated to external calls (pgx/HTTP/asynq).
- `PER-08` → `defer` in a loop/long function holds resources (rows/files/locks).
- `ARC-03` → **N/A at the package level** (import cycles are forbidden by the compiler); only check logical layers via go-arch-lint.
- `ARC-05` → `os.Getenv` scattered instead of a config struct.
- `DEP-09` → no switch to production mode (debug routes/pprof/verbose in the prod image).
- `DEP-10` → `go get` in the build instead of `go mod download` against a committed `go.sum`; no `-mod=readonly`.
- `TST-01` → `go vet`/`staticcheck`/`golangci-lint` not enabled/not required.
