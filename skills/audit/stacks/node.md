# Stack Profile: Node / TypeScript   (id: node)
Tier: first-class

Покрывает и бэкенд (Node), и фронтенд (браузер/Vite). Это профиль рантайма, а не
фреймворка — конкретные фреймворки (Express/Fastify/NestJS, React/Vite) упоминаются
лишь как примеры идиом.

## 1. Detection signals
- `package.json` (основной маркер)
- доп.: `tsconfig.json`, `package-lock.json` / `pnpm-lock.yaml` / `yarn.lock`

## 2. Tooling by category
| Категория | Команда | Как читать вывод |
|-----------|---------|------------------|
| unused-code | `npx knip --reporter json 2>/dev/null \| head -200 \|\| true` | неиспользуемые экспорты/зависимости/файлы → YAGNI-02, NAM-06 |
| clone-detection | `npx jscpd --min-lines 8 --min-tokens 50 --reporters json --silent --output ./.jscpd ./src 2>/dev/null \|\| true` | дубли блоков → REINV-03 |
| dep-audit | `npm audit --json 2>/dev/null \| head -100 \|\| pnpm audit --json 2>/dev/null \|\| true` | CVE в зависимостях (вне чеклиста — справочно) |
| env-extraction | `grep -rEoh 'process\.env\.[A-Z0-9_]+\|import\.meta\.env\.[A-Z0-9_]+' ./src 2>/dev/null \| sed -E 's/.*env\.//' \| sort -u` | имена env-переменных из кода → DOC-02 |
| arch-lint | `npx dependency-cruiser --validate 2>/dev/null \|\| true`; FSD: `npx steiger ./src 2>/dev/null \|\| true` | circular deps → ARC-03; нарушения слоёв FSD |
| lint/format | `npx biome check 2>/dev/null \|\| npx eslint . 2>/dev/null \|\| true` | — |
| type-check | `npx tsc --noEmit 2>/dev/null \|\| true` | — |
| test-run | `npm test 2>/dev/null \|\| true` | — |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| trufflehog filesystem . 2>/dev/null \|\| true` | стек-нейтрально |

Всегда верифицируй вывод инструмента вручную (`file:line`) перед `❌ FAIL`.

## 3. Idioms (как выглядит «правильно» → PASS)
- **Error handling:** `try/catch` вокруг `await`; типизированные ошибки; промисы либо `await`-ятся, либо явно `.catch`. Express 5 / `asyncHandler` пробрасывает rejection в error-middleware.
- **Concurrency:** `Promise.all`/`allSettled` для независимых задач; `AbortController`/`AbortSignal` для отмены и таймаутов; `process.on('SIGTERM')` для graceful shutdown.
- **Env/config:** `process.env` (бэкенд) / `import.meta.env` (Vite-фронтенд), валидируется при старте (zod/envalid) и изолируется в config-модуле.
- **Logging:** структурный логгер (pino/winston) с request/correlation ID; в production нет `console.*`.
- **Null-safety:** optional chaining `?.`, nullish coalescing `??`; строгие проверки `undefined`/`null`.
- **Type coercion:** `parseInt(x, 10)` с radix; строгое `===`; явные `Number()`/`String()`.
- **DI / абстракции:** зависимости инжектируются (конструктор/параметры), не создаются внутри функций; интерфейс оправдан >1 реализацией или тестами.
- **Deps / reinvention:** предпочитать stdlib (`structuredClone`, `crypto.randomUUID`, `[...new Set()]`, `Array.flat`) и уже установленные библиотеки самописным аналогам.
- **Build/deploy:** multi-stage Dockerfile; `npm ci` по `package-lock.json`; `NODE_ENV=production`; `USER node`/nonroot; `.dockerignore` исключает `node_modules`/`.git`/`.env`.

## 4. Anti-patterns (как выглядит FAIL)
- **Errors:** пустой `catch (e) {}`; промис без `await`/`.catch`; `if (asyncFn())` (всегда truthy).
- **Concurrency:** `async` callback в `Array.forEach`/`map` (потерянные промисы); `await` в последовательном цикле для независимых задач; module-level mutable singleton без инвалидации.
- **Logging:** `console.log`/`console.error` в production-коде; `[object Object]` в логах.
- **Type coercion:** `parseInt(x)` без radix; `==` вместо `===`; сравнение float через `===`.
- **Build/deploy:** `npm install` вместо `npm ci`; нет `NODE_ENV=production`; `node_modules` в build-контексте Docker.

## 5. Check-ID hints
- `LOG-01` → `console.log`/`console.error`/`console.*` в production.
- `BUG-01` → `parseInt` без radix, `NaN`-проверки, `==`-coercion.
- `BUG-02` / `CON-01` → `async` в `forEach`/`map`, потерянные промисы.
- `BUG-10` → катастрофический backtracking в `RegExp` на user input (ReDoS реален).
- `DEP-04` → `node_modules`/`.git`/`.env` в build-контексте.
- `DEP-09` → `NODE_ENV` не выставлен в `production`.
- `DEP-10` → `npm install` вместо `npm ci`; нет `package-lock.json`.
- `ERR-04` → нет `process.on('unhandledRejection')`/`uncaughtException`.
- `ERR-06` → нет `process.on('SIGTERM')` для graceful shutdown.
- `ERR-09` → `AbortSignal` не пробрасывается во внешние вызовы.
- `ARC-05` → `process.env.X` разбросан вместо config-модуля.
- `TST-01` → `tsconfig.json` без `strict: true`.
