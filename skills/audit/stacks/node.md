# Stack Profile: Node / TypeScript   (id: node)
Tier: first-class

Covers both the backend (Node) and the frontend (browser/Vite). This is a runtime profile, not
a framework one — specific frameworks (Express/Fastify/NestJS, React/Vite) are mentioned
only as examples of idioms.

## 1. Detection signals
- `package.json` (primary marker)
- additional: `tsconfig.json`, `package-lock.json` / `pnpm-lock.yaml` / `yarn.lock`

## 2. Tooling by category
| Category | Command | How to read the output |
|-----------|---------|------------------|
| unused-code | `npx knip --reporter json 2>/dev/null \| head -200 \|\| true` | unused exports/dependencies/files → YAGNI-02, NAM-06 |
| clone-detection | `npx jscpd --min-lines 8 --min-tokens 50 --reporters json --silent --output ./.jscpd ./src 2>/dev/null \|\| true` | duplicate blocks → REINV-03 |
| dep-audit | `npm audit --json 2>/dev/null \| head -100 \|\| pnpm audit --json 2>/dev/null \|\| true` | CVEs in dependencies (outside the checklist — for reference) |
| env-extraction | `grep -rEoh 'process\.env\.[A-Z0-9_]+\|import\.meta\.env\.[A-Z0-9_]+' ./src 2>/dev/null \| sed -E 's/.*env\.//' \| sort -u` | env variable names from code → DOC-02 |
| arch-lint | `npx dependency-cruiser --validate 2>/dev/null \|\| true`; FSD: `npx steiger ./src 2>/dev/null \|\| true` | circular deps → ARC-03; FSD layer violations |
| lint/format | `npx biome check 2>/dev/null \|\| npx eslint . 2>/dev/null \|\| true` | — |
| type-check | `npx tsc --noEmit 2>/dev/null \|\| true` | — |
| test-run | `npm test 2>/dev/null \|\| true` | — |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| trufflehog filesystem . 2>/dev/null \|\| true` | stack-neutral |

Always verify the tool's output manually (`file:line`) before `❌ FAIL`.

## 3. Idioms (what "correct" looks like → PASS)
- **Error handling:** `try/catch` around `await`; typed errors; promises are either `await`-ed or explicitly `.catch`-ed. Express 5 / `asyncHandler` propagates rejection to the error middleware.
- **Concurrency:** `Promise.all`/`allSettled` for independent tasks; `AbortController`/`AbortSignal` for cancellation and timeouts; `process.on('SIGTERM')` for graceful shutdown.
- **Env/config:** `process.env` (backend) / `import.meta.env` (Vite frontend), validated at startup (zod/envalid) and isolated in a config module.
- **Logging:** a structured logger (pino/winston) with a request/correlation ID; no `console.*` in production.
- **Null-safety:** optional chaining `?.`, nullish coalescing `??`; strict `undefined`/`null` checks.
- **Type coercion:** `parseInt(x, 10)` with radix; strict `===`; explicit `Number()`/`String()`.
- **DI / abstractions:** dependencies are injected (constructor/parameters), not created inside functions; an interface is justified by >1 implementation or by tests.
- **Deps / reinvention:** prefer the stdlib (`structuredClone`, `crypto.randomUUID`, `[...new Set()]`, `Array.flat`) and already-installed libraries over hand-written equivalents.
- **Build/deploy:** multi-stage Dockerfile; `npm ci` against `package-lock.json`; `NODE_ENV=production`; `USER node`/nonroot; `.dockerignore` excludes `node_modules`/`.git`/`.env`.

## 4. Anti-patterns (what FAIL looks like)
- **Errors:** empty `catch (e) {}`; a promise without `await`/`.catch`; `if (asyncFn())` (always truthy).
- **Concurrency:** `async` callback in `Array.forEach`/`map` (lost promises); `await` in a sequential loop for independent tasks; module-level mutable singleton without invalidation.
- **Logging:** `console.log`/`console.error` in production code; `[object Object]` in logs.
- **Type coercion:** `parseInt(x)` without radix; `==` instead of `===`; float comparison via `===`.
- **Build/deploy:** `npm install` instead of `npm ci`; no `NODE_ENV=production`; `node_modules` in the Docker build context.

## 5. Check-ID hints
- `LOG-01` → `console.log`/`console.error`/`console.*` in production.
- `BUG-01` → `parseInt` without radix, `NaN` checks, `==`-coercion.
- `BUG-02` / `CON-01` → `async` in `forEach`/`map`, lost promises.
- `BUG-10` → catastrophic backtracking in `RegExp` on user input (ReDoS is real).
- `DEP-04` → `node_modules`/`.git`/`.env` in the build context.
- `DEP-09` → `NODE_ENV` not set to `production`.
- `DEP-10` → `npm install` instead of `npm ci`; no `package-lock.json`.
- `ERR-04` → no `process.on('unhandledRejection')`/`uncaughtException`.
- `ERR-06` → no `process.on('SIGTERM')` for graceful shutdown.
- `ERR-09` → `AbortSignal` not propagated to external calls.
- `ARC-05` → `process.env.X` scattered instead of a config module.
- `TST-01` → `tsconfig.json` without `strict: true`.
