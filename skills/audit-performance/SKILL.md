---
name: audit-performance
description: >
  Resource and performance audit: blocking calls, N+1 queries, heavy queries
  without limits, memory leaks, missing pagination. Run on /audit-performance.
---

## Relevance Rule

Applies to code with I/O operations (DB, HTTP, files), collection processing, or caching. For stateless math utilities without I/O, return an empty response.

## Runtime Detection & Stack Profile

This audit is stack-agnostic: the checks are framed neutrally, and the specifics
(tools, idioms, anti-patterns, examples) come from the stack profile.

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
   One run = one runtime; do not mix backend and frontend. If several markers
   are found (monorepo), pick the one matching the current scope / files under
   analysis and record the choice in the Audit Coverage section.

3. **Load the profile** via Read: `./skills/audit/stacks/<runtime>.md`
   (fallback `./skills/audit/stacks/_generic.md` if the file is not found).

Then use the profile:
- **Tools** — from the profile's "Tooling by category" section (the
  "Tooling Support" section below references categories, not commands).
- **PASS expectations** — from "Idioms"; **FAIL wording** — from "Anti-patterns".
- **Targeted hints** — from "Check-ID hints" by the prefix `PER-`.
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
| PER-01 | No N+1: DB queries are not executed inside loops [⚡ dynamic] |
| PER-02 | DB result sets are bounded (LIMIT, pagination) |
| PER-03 | Request handlers contain no blocking I/O |
| PER-04 | CPU-intensive operations are moved off the main thread |
| PER-05 | Independent async operations run in parallel |
| PER-06 | Caches are bounded by size and lifetime (TTL + size limit) |
| PER-07 | Event listeners and subscriptions are cleaned up on teardown |
| PER-08 | No memory leaks: timers and closures do not retain large objects in a long-lived scope |

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

**PER-01 — No N+1:**
- A DB query inside a loop over the results of another query
- ORM relations loaded lazily inside a loop
- Missing `include`/`join` where a single fetch is possible
- DataLoader / batch loading not used when there are many point queries

**PER-02 — Result sets are bounded:**
- SELECT without LIMIT/pagination (possible full table scan)
- Queries without WHERE on indexed fields
- Aggregations on large tables without a materialized view / cache
- No cursor-based pagination for large data sets

**PER-03 — No blocking I/O in handlers:**
- Synchronous file operations in an async context (Node: `fs.readFileSync`)
- `sleep`/busy-wait in a request handler
- Synchronous operations on large buffers that block the event loop (Node)
- Go: blocking network/I/O calls without a timeout/`context.Context` in a handler; synchronous heavy calls outside a goroutine (less critical than blocking the event loop in Node, but they still hold the connection)

**PER-04 — CPU operations off the main thread:**
- CPU-intensive operations (crypto, image processing, compression) on the main thread without a worker
- Heavy computations (sorting large arrays, regex on long strings) in the request path

**PER-05 — Parallel independent operations:**
- Sequential independent HTTP requests instead of `Promise.all`
- Sequential await where a parallel fetch is possible
- Repeated requests to the same URL without memoization within a single request

**PER-06 — Caches are bounded:**
- Cache without a TTL (unbounded growth, stale data forever)
- Cache without a size limit (memory leak in long-lived processes)
- In-memory cache without an invalidation mechanism when data is updated

**PER-07 — Event listeners are cleaned up:**
- Event listeners without `removeEventListener` / `off` (leak in long-lived processes)
- RxJS subscriptions without `unsubscribe` in destroy/cleanup
- WebSocket / SSE connections without cleanup at the end of the request lifecycle
- Data accumulating in memory without flush (buffer without drain)
- Go: a subscriber/worker goroutine with no termination path via `ctx.Done()` (goroutine leak); an unclosed channel/`time.Ticker` without `Stop()`

**PER-08 — No memory leaks via timers and closures:**
- `setInterval` / `setTimeout` without a corresponding `clearInterval` / `clearTimeout` in cleanup (Node)
- A closure in a long-lived object captures a large array/object — GC cannot collect it
- A `global` object or module-level variable accumulates entries without a bound (unbounded growth)
- Circular reference between objects with WeakMap/WeakRef where a strong reference is needed
- Go: `defer` in a loop/long function holds resources (rows/files/locks) longer than needed — the release is deferred to the end of the function, not the iteration

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| PER-01 | No N+1: DB queries are not executed inside loops | ✅ PASS | High | `repos/` checked — queries are outside loops | — [⚡ dynamic] | — |
| PER-02 | DB result sets are bounded (LIMIT, pagination) | ❌ FAIL 🟠 | High | `repos/user.ts:45` | **1. Add .take(limit).skip(offset) to the query** \\ 2. Add cursor-based pagination \\ 3. Set a maximum limit via config | No |
| PER-06 | Caches are bounded by size and lifetime (TTL + size limit) | ⏸ ACCEPTED | Medium | `cache/store.ts:12` | In baseline: cache managed by Redis with TTL | — |

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

If everything is PASS, output: `✅ No performance anti-patterns found.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-performance.md`

```
# Audit Report: Resource & Performance — <YYYY-MM-DD HH:MM>
<table>
```
