# Stack Profile: Go   (id: go)
Tier: first-class

Профиль рантайма Go. Конкретные библиотеки (chi, gqlgen, sqlc, pgx, asynq, slog,
go-oidc) упоминаются лишь как примеры идиом, а не как требования.

## 1. Detection signals
- `go.mod` (основной маркер)
- доп.: `go.sum`, `Taskfile.yml`/`Makefile`, `.golangci.yml`, `.go-arch-lint.yml`

## 2. Tooling by category
| Категория | Команда | Как читать вывод |
|-----------|---------|------------------|
| unused-code | `deadcode ./... 2>/dev/null \|\| true` (golang.org/x/tools/cmd/deadcode) | недостижимые функции → YAGNI-02; верифицируй перед FAIL |
| clone-detection | `dupl -threshold 50 ./... 2>/dev/null \|\| true` | дубли токенов → REINV-03 |
| dep-audit | `govulncheck ./... 2>/dev/null \|\| true` | CVE в модулях (вне чеклиста — справочно) |
| env-extraction | `grep -rEoh 'os\.Getenv\("[A-Z0-9_]+"\)' . 2>/dev/null \| sed -E 's/.*"(.*)".*/\1/' \| sort -u` | имена env из кода → DOC-02 (учти теги `env:"..."` для caarlos0/env) |
| arch-lint | `go vet ./... 2>/dev/null`; слои: `go-arch-lint check 2>/dev/null \|\| true` | циклы импортов запрещены компилятором → ARC-03 на уровне пакетов авто-PASS |
| lint/format | `golangci-lint run 2>/dev/null \|\| true`; `gofmt -l . 2>/dev/null` | включает `errcheck`, `gosec`, `nilness`, `staticcheck` |
| type-check | `go build ./... 2>/dev/null \|\| true` | компиляция = проверка типов |
| test-run | `go test ./... 2>/dev/null \|\| true`; гонки: `go test -race ./...` | `-race` обнаруживает data race → CON-03 |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| trufflehog filesystem . 2>/dev/null \|\| true` | стек-нейтрально |

Всегда верифицируй вывод инструмента вручную (`file:line`) перед `❌ FAIL`.

## 3. Idioms (как выглядит «правильно» → PASS)
- **Error handling:** явная проверка `if err != nil`; оборачивание `fmt.Errorf("...: %w", err)`; sentinel-ошибки + `errors.Is`/`errors.As`. Возвращаемая ошибка не игнорируется.
- **Concurrency:** отмена через `context.Context` (не сигналы); shared state под `sync.Mutex`/каналами; fan-out через `errgroup`/`sync.WaitGroup`; каждая goroutine имеет путь завершения по `ctx.Done()`.
- **Graceful shutdown:** `signal.NotifyContext(ctx, syscall.SIGTERM, os.Interrupt)`; `server.Shutdown(ctx)`; закрытие пулов (pgx), воркеров (asynq), redis.
- **Panic safety:** `defer recover()` в HTTP-хендлерах (chi `middleware.Recoverer`), в gqlgen (`RecoverFunc`), в каждой фоновой goroutine и в обёртке job (asynq).
- **Env/config:** централизованная config-структура (например caarlos0/env), загружается при старте; нет `os.Getenv` вразброс.
- **Logging:** структурный `log/slog` (JSON в prod) с request/correlation ID через `context`; нет `fmt.Print*`/`log.Print*` в request-путях.
- **Null-safety:** проверка nil-указателей/интерфейсов перед разыменованием; запись в nil-map не допускается; map-доступ через `v, ok := m[k]`.
- **Type coercion:** `strconv.Atoi`/`ParseInt` с проверкой ошибки; type assertion через `v, ok := x.(T)` (comma-ok).
- **Deps / reinvention:** stdlib `slices`/`maps`/`sync`; `database/sql`/sqlc вместо самописного query builder.
- **Build/deploy:** multi-stage (builder → distroless/`nonroot`); `go.mod`+`go.sum` закоммичены; `GOFLAGS=-mod=readonly`; `CGO_ENABLED=0` для статической линковки; нет `go get` в Dockerfile; нет `net/http/pprof`/debug-роутов в prod.

## 4. Anti-patterns (как выглядит FAIL)
- **Errors:** игнор ошибки через `_ = f()` или `v, _ := f()`; `panic` для обычного потока управления; потеря wrap-цепочки.
- **Concurrency:** goroutine без `WaitGroup`/`errgroup`/`ctx` (результат теряется / процесс не ждёт / leak); запись в map из >1 goroutine без lock (data race, ловит `go test -race`); `for ... { go f(loopVar) }` — захват переменной цикла (до Go 1.22); незакрытый канал блокирует получателей.
- **Panic:** паника в goroutine без `recover()` роняет весь процесс.
- **Logging:** `fmt.Print*`/`log.Print*`/`println` в production-путях вместо `slog`.
- **Resources:** `defer` внутри цикла (ресурс держится до конца функции, не итерации).
- **Build/deploy:** `go get` в Dockerfile; отсутствие `go.sum`; компилятор/тесты в финальном образе.

## 5. Check-ID hints
- `LOG-01` → `fmt.Print*`, `log.Print*`, `println` вне обёртки `slog`.
- `BUG-01` → **N/A**, если нет парсинга строк; иначе — type assertion `x.(T)` без comma-ok, проигнорированная ошибка `strconv`.
- `BUG-02` → **N/A** (нет синтаксиса промисов/forEach); риск переезжает в `CON-01`.
- `BUG-03` → nil-deref указателя/интерфейса/map; запись в nil-map (паника).
- `BUG-08` → сравнение float через `==`; деньги — в `int64` (центы) или `shopspring/decimal`.
- `BUG-10` → **N/A или Low**: `regexp` (RE2) не имеет backtracking, ReDoS практически невозможен.
- `CON-01` → goroutine без механизма завершения (WaitGroup/errgroup/ctx) → goroutine leak.
- `CON-03` → конкурентный доступ к map/переменной пакета без `sync.Mutex`/atomic (data race).
- `CON-06` → фоновая goroutine не слушает `ctx.Done()` → не завершается при shutdown.
- `ERR-01` → ошибка проигнорирована через `_`.
- `ERR-03` → нет recover-middleware (паника в хендлере роняет соединение/процесс).
- `ERR-04` → нет `defer recover()` в фоновых goroutine/asynq-handler.
- `ERR-06` → нет `signal.NotifyContext`/`server.Shutdown` (graceful shutdown).
- `ERR-09` → `context.Context` не пробрасывается во внешние вызовы (pgx/HTTP/asynq).
- `PER-08` → `defer` в цикле/долгой функции держит ресурсы (rows/файлы/locks).
- `ARC-03` → **N/A на уровне пакетов** (циклы импортов запрещены компилятором); проверять только логические слои через go-arch-lint.
- `ARC-05` → `os.Getenv` разбросан вместо config-структуры.
- `DEP-09` → нет переключения в production-режим (debug-роуты/pprof/verbose в prod-образе).
- `DEP-10` → `go get` в сборке вместо `go mod download` по закоммиченному `go.sum`; нет `-mod=readonly`.
- `TST-01` → `go vet`/`staticcheck`/`golangci-lint` не включены/не обязательны.
