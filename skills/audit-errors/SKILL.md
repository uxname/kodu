---
name: audit-errors
description: >
  Error handling and resiliency audit: exceptions, timeouts, retry policies,
  circuit breakers, graceful degradation. Run on /audit-errors.
---

## Relevance Rule

Applies to code with external calls (HTTP, DB, queues, file system), asynchronous code, and event handlers. For synchronous utilities without I/O, return an empty response.

## Runtime Detection & Stack Profile

This audit is stack-agnostic: the checks are framed neutrally, and the specifics (tools, idioms, anti-patterns, examples) come from the stack profile.

1. **Profile passed in context?** If the `/audit` orchestrator passed
   `runtime=<id>` and/or the profile contents, use it and skip steps 2–3.

2. **Otherwise, detect EXACTLY ONE runtime** for this directory:
   ```bash
   if   [ -f package.json ]; then echo "runtime=node"
   elif [ -f go.mod ]; then echo "runtime=go"
   elif [ -f pyproject.toml ] || [ -f requirements.txt ] || [ -f setup.py ]; then echo "runtime=python"
   elif [ -f Cargo.toml ]; then echo "runtime=rust"
   elif [ -f pom.xml ] || ls build.gradle* settings.gradle* >/dev/null 2>&1; then echo "runtime=java"
   else echo "runtime=generic"; fi
   ```
   One run = one runtime; do not mix backend and frontend. If several markers are found (monorepo), pick the one matching the current scope / files under analysis and record the choice in the Audit Coverage section.

3. **Load the profile** via Read: `./skills/audit/stacks/<runtime>.md`
   (fallback `./skills/audit/stacks/_generic.md` if the file is not found).

Then use the profile:
- **Tools** — from the profile's "Tooling by category" section (the
  "Tooling Support" section below references categories, not commands).
- **PASS expectations** — from "Idioms"; **FAIL wording** — from "Anti-patterns".
- **Targeted hints** — from "Check-ID hints" by the prefix `ERR-`.
- If the profile is `tier: general` or `runtime=generic`, mark stack-specific
  findings without unambiguous evidence as `🔍 UNVERIFIED` rather than `❌ FAIL`.
  Mark checks whose mechanism is absent in the runtime as `N/A`.

## Severity Guide

| Severity | Assignment criterion |
|----------|---------------------|
| 🔴 Critical | RCE, auth bypass, data corruption, irreversible financial risk |
| 🟠 High | production outage, privilege escalation, data leak |
| 🟡 Medium | performance or maintainability degradation without an immediate outage |
| 🟢 Low | style, readability, minor convention violation |

Rule: severity = impact × exploitability × blast radius. The same pattern → the same severity across audits.

## Checklist

| Check ID | Check |
|----------|----------|
| ERR-01 | Errors are not swallowed — a returned error is not ignored (including via `_`), catch blocks handle or rethrow |
| ERR-02 | Internal details (stack trace, paths, versions) do not leak into responses |
| ERR-03 | Failures inside handlers (panics/exceptions) are caught and do not crash the process/connection |
| ERR-04 | Uncaught top-level failures (panics in background tasks) are logged and do not lead to a silent crash |
| ERR-05 | External calls (HTTP client, DB) have explicit timeouts |
| ERR-06 | Graceful shutdown is implemented — SIGTERM is handled |
| ERR-07 | Error responses are consistent in structure across the whole application |
| ERR-08 | Retry strategies use exponential backoff with jitter |
| ERR-09 | A cancellation context is propagated into external calls and aborts them [⚡ dynamic] |

## Verification Rules

1. **Checklist only**: evaluate ONLY the checks above. Do not add new ones.
2. **Explicit verification = PASS**: assign `✅ PASS` only if you explicitly verified the mechanism (found the schema, config, guard) and confirmed there is no violation — state exactly what was checked.
3. **No evidence = UNVERIFIED**: if you cannot point to a `file:line` for either a violation or a confirmation, assign `🔍 UNVERIFIED`.
   - Checks marked `[⚡ dynamic]` cannot be confirmed statically — only `🔍 UNVERIFIED` or `❌ FAIL` (with explicit evidence), never `✅ PASS`
4. **Baseline takes priority**: if the check_id is in `docs/audit-baseline.yml` → `⏸ ACCEPTED`.
5. **Only 🔴/🟠 FAILs require a solution**: 🟡/🟢 — a solution is optional.

## Evidence Quality Rules

Every `❌ FAIL` must include:
- An exact `file:line`
- A minimal code snippet (1–3 lines)
- Causal chain: why this specific violation → what risk it creates

Not allowed:
- Assuming runtime behavior without evidence in the code
- Inferring the prod configuration from the dev configuration
- Assuming middleware is absent without checking the entire router chain
- If a conclusion rests on an assumption — only `🔍 UNVERIFIED`

## Language Rule

Audit results must be written in plain, clear language. Avoid complex terms, jargon, and abstract concepts unless necessary. Common technical terms (Docker, HTTP, API, JSON, URL) are fine. Describe problems so they are understandable to a developer of any level, not only a narrow specialist in the area.

## Baseline

Before analysis:
```bash
if [ ! -f ./docs/audit-baseline.yml ]; then
  mkdir -p ./docs
  cp ./skills/audit/audit-baseline-template.yml ./docs/audit-baseline.yml 2>/dev/null || \
    printf "accepted: []\n" > ./docs/audit-baseline.yml
fi
cat ./docs/audit-baseline.yml
```

## Analysis Context

> The examples below are illustrative (Node/TS). Take the specifics of the current runtime from
> the loaded profile (`stacks/<runtime>.md`, the Idioms/Anti-patterns/Check-ID hints sections).

**ERR-01 — Errors are not swallowed:**
- A returned error is ignored (including via `_`) and not handled
- Empty catch blocks (`catch(e) {}`)
- `catch` with only a log, without recovery or re-throw
- A Promise without `.catch()`, or `try/await` without `catch`
- Unhandled promise rejections with no handling
- In Go: `_ = f()` / `v, _ := f()` swallows the error; caught by `errcheck`/`golangci-lint`

**ERR-02 — Internal details not in responses:**
- Stack trace in production API responses
- Internal file system paths in error messages
- Dependency/framework versions in headers or responses
- DB-specific error messages (SQL syntax error) in API responses

**ERR-03 — Failures inside handlers are caught:**
- A panic/exception inside a handler crashes the process or connection instead of being caught locally
- Express async handlers without an asyncHandler wrapper or Express 5
- A Promise rejection in middleware is not propagated to the error middleware
- Unhandled exceptions in setTimeout/setInterval callbacks
- In Go: no recover middleware (chi `middleware.Recoverer`), no `RecoverFunc` in gqlgen, no recover in the task wrapper (asynq) — a panic crashes the connection/process

**ERR-04 — Uncaught top-level failures:**
- An uncaught top-level failure is not logged and leads to a silent crash
- No `process.on('unhandledRejection')`
- No `process.on('uncaughtException')`
- No logging and clean exit on critical process errors
- In Go: a panic in a background goroutine/task without `defer recover()` crashes the whole process — a `defer recover()` is needed in every goroutine

**ERR-05 — Explicit timeouts for external calls:**
- HTTP client / DB call without an explicit timeout
- DB queries without a query timeout / statement timeout
- No timeout for message queues and external gRPC calls
- Infinite retries without exponential backoff and max attempts
- In Go: `http.Client{Timeout: ...}`, `context.WithTimeout` for requests to the DB/external services

**ERR-06 — Graceful shutdown:**
- No handling of the termination signal (SIGTERM)
- DB pool and HTTP server not closed on shutdown
- In-flight requests are not awaited to completion on shutdown
- In Go: `signal.NotifyContext` + `server.Shutdown(ctx)` + closing pools (pgx), workers (asynq), redis

**ERR-07 — Consistent error response structure:**
- Different error formats across endpoints (no single error shape)
- No machine-readable error code (only a human-readable message)
- HTTP status 200 on error (should be 4xx/5xx)

**Retry & Cancellation (ERR-08 / ERR-09):**
- HTTP retry without a delay or with a fixed delay (no exponential backoff)
- No jitter — all retries synchronize during a mass failure
- An external call without a cancellation context — hanging requests after the client disconnects (Node: fetch/axios without AbortSignal)
- The cancellation context is not propagated deep into the call chain and does not abort it
- In Go: `context.Context` is the first-class cancellation mechanism, propagated into all external calls (pgx/HTTP/asynq)

## Boundary With Other Audits

- **Stack trace in responses** — this skill is primary (ERR-02). `audit-owasp` and `audit-api-contracts` refer here.
- **Handler idempotency** — primary: `audit-concurrency` (CON-05). Do not duplicate here.
- **HTTP client timeouts** — ERR-05 is primary. `audit-performance` does not duplicate it.

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| ERR-01 | Errors are not swallowed — a returned error is not ignored (including via `_`), catch blocks handle or rethrow | ✅ PASS | High | `src/` — all catch blocks log or rethrow | — | — |
| ERR-05 | External calls (HTTP client, DB) have explicit timeouts | ❌ FAIL 🟠 | High | `services/api.ts:18` | **1. Add a timeout to axios: `{ timeout: 5000 }`** \\ 2. Use an AbortController with setTimeout \\ 3. Set a global default timeout | No |
| ERR-06 | Graceful shutdown is implemented — SIGTERM is handled | ⏸ ACCEPTED | Medium | `server.ts:5` | In baseline: managed by the orchestrator | — |

Statuses: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED` / `🔍 UNVERIFIED`

Confidence: `High` — checked several key files, the pattern is obvious / `Medium` — checked selectively, the pattern is likely / `Low` — limited context, full certainty is impossible

For `❌ FAIL`: exactly 3 solution options, separated by `\\`, with option 1 in bold.

`Fixed`: FAIL → `No` (the developer changes it to `✅ Yes` manually after the fix). PASS / ACCEPTED / UNVERIFIED → `—`.

Solution requirements:
- Mutually exclusive (not rephrasings of the same thing)
- Match the project's current stack (do not propose switching frameworks)
- Do not require rewriting the whole system — realistic migration cost
- Option 3 may be "keep it, document the reason" if there is justification

At the end of the report, add a coverage section:
```
## Audit Coverage
Checked: src/module1/**, src/module2/**
Skipped: scripts/**, migrations/**, tests/**
Files checked: N | Skipped: N
```

If everything is PASS, output: `✅ Error handling is implemented correctly.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-errors.md`

```
# Audit Report: Error Handling & Resiliency — <YYYY-MM-DD HH:MM>
<table>
```
