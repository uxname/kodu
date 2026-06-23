---
name: audit-concurrency
description: >
  State management and concurrency audit: race conditions, deadlocks, shared mutable state,
  non-atomic operations. Run on /audit-concurrency.
---

## Relevance Rule

Applies to code with concurrent operations, shared state, caching, DB transactions, queues, WebSocket. For single-threaded scripts without concurrency, return an empty response.

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
- **Targeted hints** — from "Check-ID hints" by the prefix `CON-`.
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
| CON-01 | Concurrent operations are started and synchronized correctly; results are awaited, background tasks do not leak |
| CON-02 | Read-modify-write operations run inside transactions [⚡ dynamic] |
| CON-03 | Shared mutable state is protected by synchronization (singletons, caches, module-level variables) |
| CON-04 | A module-level cache has an invalidation mechanism |
| CON-05 | Event handlers and webhook handlers are idempotent [⚡ dynamic] |
| CON-06 | Background operations have a cancellation mechanism and do not block graceful shutdown |

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

**CON-01 — Concurrent operations are started and synchronized correctly:**
- A concurrent operation is started, but its result is not awaited (it is lost)
- A background task leaks — there is no path to complete it
- Shared state between concurrent operations without synchronization
- In Node: `await` inside `forEach` (the iteration does not wait for the promises); `async` in `Array.map` without `Promise.all` (the promises are not awaited)
- In Go: a goroutine without `WaitGroup`/`errgroup`/`ctx` → lost result or goroutine leak; capturing the loop variable in `for { go f(loopVar) }` (before Go 1.22); an unclosed channel blocks receivers

**CON-02 — Read-modify-write in transactions:**
- SELECT + UPDATE without a transaction (TOCTOU — time-of-check to time-of-use)
- Double debit/credit without a transaction with locking
- Check-then-act without atomicity (read → check → write without a lock)
- Optimistic locking without retry on a version conflict
- Cache invalidation between read and write

**CON-03 — Shared mutable state is protected by synchronization:**
- Global variables mutated from several places without synchronization
- Singletons with mutable state without synchronization
- A closure over a mutable variable in an async callback
- Concurrent writes to the same file/resource without coordination
- In Go: a package-level map/variable without `sync.Mutex`/atomic under concurrent access → data race; caught by `go test -race`

**CON-04 — A module-level cache is invalidated:**
- A module-level cache without an invalidation mechanism when the data is updated
- A cache without TTL (stale data is never refreshed)
- No strategy for refreshing the cache when the source data changes

**CON-05 — Handler idempotency:**
- Event/message handlers without idempotency (redelivery is not safe)
- No protection against duplicate webhook calls (no event_id check)
- Financial or critical operations without an idempotency key

**CON-06 — Background operations with a cancellation mechanism:**
- A background operation is started without a cancellation mechanism — on shutdown the process does not terminate cleanly
- A background job without a timeout and without a forced-stop mechanism
- Graceful shutdown does not wait for background tasks to finish
- In Node: a background promise chain without `AbortController`/`signal`; `setInterval`/`setImmediate` in a request handler without cleanup; `Promise.all` with non-cancelable tasks
- In Go: a background goroutine does not listen on `ctx.Done()` → it does not terminate on shutdown; a `context.Context` with a cancellation check is needed

## Boundary With Other Audits

- **Idempotency** — this skill is the primary owner (CON-05). `audit-errors` references it here.
- **async/await in forEach** — primary owner: `audit-bugs` (BUG-02). `audit-concurrency` (CON-01) focuses on concurrency, not on the syntactic error.

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| CON-01 | Concurrent operations are started and synchronized correctly; results are awaited, background tasks do not leak | ✅ PASS | High | `src/` — no async forEach found | — | — |
| CON-02 | Read-modify-write operations run inside transactions | ❌ FAIL 🔴 | High | `services/wallet.ts:67` | **1. Wrap in db.transaction() with SELECT FOR UPDATE** \\ 2. Use optimistic locking with retry \\ 3. Add a unique constraint at the DB level [⚡ dynamic] | No |
| CON-05 | Event handlers and webhook handlers are idempotent | ⏸ ACCEPTED | Medium | `handlers/stripe.ts:12` | In baseline: idempotency ensured via event_id [⚡ dynamic] | — |

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

If everything is PASS, output: `✅ No concurrency issues found.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-concurrency.md`

```
# Audit Report: State & Concurrency — <YYYY-MM-DD HH:MM>
<table>
```
