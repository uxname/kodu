---
name: audit-reinvention
description: >
  Reinventing-the-wheel audit: hand-rolled implementations of things already provided by the stdlib/language,
  by installed dependencies, or duplicated within the project; proposes a mature off-the-shelf solution.
  Run on /audit-reinvention or a request to find reinvented wheels.
---

## Relevance Rule

Applies to any production code that contains application logic. Skip thin glue modules (configs, pure types, generated code). The goal is to find code that does by hand what the language, runtime, an installed library, or another part of this same project already does.

**Do not confuse with neighboring audits** (responsibility boundary, to avoid duplicating findings):
- `audit-yagni` — about the **superfluous**: unnecessary abstractions, dead code, premature optimization. Here — about the **reinvented**: needed functionality, but written by hand instead of using a ready-made solution.
- `audit-architecture` (ARC-04 god objects) — about module size/responsibility. REINV-03 — about **semantic duplication** of the same logic in different places.
- If a violation is better described by a neighboring audit, hand it off there and do not duplicate.

## Runtime Detection & Stack Profile

This audit is stack-agnostic: the checks are framed neutrally, and the specifics
(stdlib equivalents, idioms, anti-patterns, examples) come from the stack profile.

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
   One run = one runtime; do not mix backend and frontend. If there are several
   markers (monorepo), pick the one matching the current scope and record the choice in Audit Coverage.

3. **Load the profile** via Read: `./skills/audit/stacks/<runtime>.md`
   (fallback `./skills/audit/stacks/_generic.md`).

Then: stdlib equivalents and tools — from the "Idioms"/"Tooling by category"
sections of the profile; FAIL wording — from "Anti-patterns"; targeted hints — from
"Check-ID hints" by the prefix `REINV-`. For `tier: general`/`generic`, mark
recommendations where you are not sure an equivalent exists as `🔍 UNVERIFIED`.

**Important:** what counts as a "ready-made solution" is determined ONLY by the
dependency manifest/lock file and the runtime's stdlib (see the profile). Do not assume a
library is present without its entry in the dependencies.

## Severity Guide

| Severity | Assignment criterion |
|----------|---------------------|
| 🔴 Critical | A hand-rolled **security** primitive: your own password hashing, your own crypto, your own JWT parser, your own SQL escaping, your own HTML sanitizer. A manual implementation almost always contains a vulnerability. |
| 🟠 High | Reinvention that causes real bugs in prod: hand-rolled retry/timeout without jitter, hand-rolled date/timezone parser, manual concurrent queue, manual money/decimal arithmetic on float. |
| 🟡 Medium | Maintainability: something that exists in the stdlib or an installed library is hand-written but works correctly (dedup, deepClone, debounce, groupBy). |
| 🟢 Low | Minor: a trivial one-line stdlib equivalent, readability is not harmed. |

Rule: severity = impact × exploitability × blast radius. **Security primitives are always escalated** to at least 🟠, more often 🔴 — even if "it seems to work". The same pattern → the same severity across audits.

## Checklist

| Check ID | Check |
|----------|----------|
| REINV-01 | No hand-rolled implementation of something the stdlib/language/runtime provides |
| REINV-02 | No hand-rolled implementation of something an installed dependency already does |
| REINV-03 | No semantic duplication: the same logic is not rewritten in ≥2 places instead of a shared utility |
| REINV-04 | Large mechanisms (ORM, DI, scheduler, logger, job queue, validator) are not written from scratch when a mature standard solution exists |

## Verification Rules

1. **Checklist only**: evaluate ONLY the checks above. Do not add new ones.
2. **Explicit verification = PASS**: assign `✅ PASS` only if you reviewed the key modules (utils, helpers, lib, core, services) and confirmed there is no reinvention — state exactly what was checked.
3. **No evidence = UNVERIFIED**: if you cannot point to a `file:line` for a hand-rolled implementation AND confirm a ready-made equivalent exists, assign `🔍 UNVERIFIED`.
4. **An equivalent must exist**: a `❌ FAIL` for REINV-01/02/04 is valid ONLY if you named a specific ready-made equivalent (a stdlib API, a package from `package.json`, or a mature library) — otherwise it is not a reinvented wheel, just code.
5. **Baseline takes priority**: if the check_id is in `docs/audit-baseline.yml` → `⏸ ACCEPTED`.
6. **Only 🔴/🟠 FAILs require a solution**: 🟡/🟢 — a solution is optional.

## Evidence Quality Rules

Every `❌ FAIL` must include:
- An exact `file:line` of the hand-rolled implementation
- A minimal code snippet (1–3 lines)
- **The name of the ready-made equivalent**: `crypto.randomUUID()`, `structuredClone`, `zod` (in deps), `date-fns`, etc.
- Causal chain: why this is a reinvented wheel → what risk (bug/vulnerability/maintenance cost)

Not allowed:
- Flagging as a reinvented wheel code for which you did NOT name a specific ready-made equivalent
- Assuming a library is present without an entry in the dependency manifest/lock
- Treating an intentionally thin wrapper around a library as a "reinvented wheel" (that is an adapter, not reinvention)
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

> The examples below are illustrative (Node/TS). Take the specific stdlib
> equivalents and installed libraries for the current runtime from the loaded profile
> (`stacks/<runtime>.md`, the Idioms/Anti-patterns/Check-ID hints sections). For example,
> for Go: `[...new Set()]` → `slices`/`maps`, a hand-rolled query builder →
> `database/sql`/sqlc, manual cancellation → `context.Context`.

**REINV-01 — Reimplementing stdlib/language/runtime:**
- Deduplicating an array by hand instead of `[...new Set(arr)]`
- A hand-rolled deep clone instead of `structuredClone()`
- Generating UUIDs by hand instead of `crypto.randomUUID()`
- Manual base64 instead of `Buffer`/`btoa`/`atob`
- Hand-rolled `debounce`/`throttle`/`groupBy`/`flatten` when `Array.flat`, `Object.groupBy` exist, or they are already in an installed library (then it is more like REINV-02)
- A loop with an accumulator instead of `Promise.all`/`Promise.allSettled` for independent tasks
- Manual query string parsing instead of `URLSearchParams`
- Hand-rolled date comparison/sorting instead of `Intl`/an installed date library

**REINV-02 — Duplicating an installed dependency:**
- A hand-rolled HTTP retry/backoff when `axios-retry`/`p-retry`/`got` (retry built in) is present
- Manual field validation when `zod`/`yup`/`joi`/`class-validator`/`pydantic` is present
- A hand-rolled `deepEqual`/`cloneDeep`/`pick`/`omit` when `lodash`/`ramda` is present
- Manual date arithmetic when `date-fns`/`dayjs`/`luxon` is present
- Reimplementing what the framework already provides (your own body-parser/router/CORS when a built-in one exists)
- Your own in-memory cache with TTL when `lru-cache`/`node-cache` is in deps

**REINV-03 — Internal semantic duplicates:**
- A semantically identical function copied into ≥2 files instead of a shared utility
- Two helpers under different names that do the same thing
- A repeated inline block (formatting, mapping, guard) that should be a single function
- Boundary: if it is about module size, hand it to `audit-architecture` (ARC-04); if it is about dead code, to `audit-yagni`

**REINV-04 — A hand-rolled large mechanism when a mature solution exists:**
- Your own ORM/query builder, your own DI container, your own scheduler/cron, your own logger, your own job queue, your own state machine, your own event bus
- A hand-rolled mechanism for which the runtime's ecosystem has a mature standard solution that is NOT yet installed in the project
- For such a FAIL, the solution must include a **migration cost estimate** (realistic, without "rewrite everything")

## Recommendation Rules (new-dependency policy)

Recommendations are **stack-agnostic**: rely on Runtime Detection, do not tie yourself to specific project templates.

Priority of solution options, top to bottom:
1. **stdlib/language/runtime** — always preferred if an equivalent exists (zero cost).
2. **An already installed dependency** — if the dependency manifest (see the profile) has a suitable library, use it.
3. **A new dependency — with caution**: propose ONLY for mature/popular libraries (active maintenance, wide adoption) AND with a mandatory migration cost estimate. For minor things (dedup, clone) do NOT propose a new library — the stdlib is enough there.

For security primitives (REINV-01/04 with severity 🔴), "keep the hand-rolled version" **cannot** be a valid option — only a move to a proven solution.

## Tooling Support

For REINV-03, use the **clone-detection** category tool from the stack profile
("Tooling by category" section): it finds duplicated code blocks. Verify each
match manually before recording it as a FAIL (false positives happen: generated
code, blocks that look alike but are not semantically identical). If the category
cell is empty (`tier: general`/`generic`), search for duplicates manually and mark
findings `🔍 UNVERIFIED`.

For REINV-02, cross-check the hand-rolled implementations you find against the list of installed
dependencies (manifest/lock from the profile) — what is actually available in the project.

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| REINV-01 | No hand-rolled implementation of something the stdlib/language/runtime provides | ❌ FAIL 🟡 | High | `src/utils/array.ts:12` — manual dedup with a loop; `[...new Set()]` exists | **1. Replace with `[...new Set(arr)]`** \\ 2. Extract into a shared helper `unique()` \\ 3. Keep it if order and a custom comparator matter — document it | No |
| REINV-02 | No hand-rolled implementation of something an installed dependency already does | ❌ FAIL 🟠 | High | `src/http/client.ts:40` — hand-rolled retry loop; `axios-retry` is in deps | **1. Wire up `axios-retry` with exponential backoff + jitter** \\ 2. Replace with `p-retry` (already in deps) \\ 3. Keep it, adding jitter and tests — justify it | No |
| REINV-04 | Large mechanisms are not written from scratch when a mature solution exists | ✅ PASS | Medium | `src/` — no hand-rolled ORM/DI/scheduler found | — | — |
| REINV-03 | No semantic duplication of logic | ⏸ ACCEPTED | Medium | `src/a.ts`, `src/b.ts` | In baseline: duplicates in a legacy module, tracked in Jira PROJ-456 | — |

Statuses: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED` / `🔍 UNVERIFIED`

Confidence: `High` — found the hand-rolled implementation and confirmed an equivalent exists / `Medium` — the pattern is likely, an equivalent exists, but the usage context was checked selectively / `Low` — limited context, full certainty is impossible

For `❌ FAIL`: exactly 3 solution options, separated by `\\`, with option 1 in bold.

`Fixed`: FAIL → `No` (the developer changes it to `✅ Yes` manually after the fix). PASS / ACCEPTED / UNVERIFIED → `—`.

Solution requirements:
- Each option names a **specific** ready-made equivalent (a stdlib API or a package name)
- Mutually exclusive (not rephrasings of the same thing)
- Match the current runtime and the dependency policy above
- Do not require rewriting the whole system — realistic migration cost
- Option 3 may be "keep it, document the reason" — BUT it is forbidden for security primitives (🔴)

At the end of the report, add a coverage section:
```
## Audit Coverage
Checked: src/utils/**, src/services/**, src/http/**
Skipped: scripts/**, migrations/**, tests/**, generated/**
Files checked: N | Skipped: N
```

If everything is PASS, output: `✅ No reinvented wheels found.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-reinvention.md`

```
# Audit Report: Reinventing the Wheel — <YYYY-MM-DD HH:MM>
<table>
```
