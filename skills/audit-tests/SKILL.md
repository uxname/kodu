---
name: audit-tests
description: >
  Audit of tests and linters: test, linter, and TypeScript configurations, critical-path coverage,
  false-positive tests. Run on /audit-tests or a request to check tests/configuration.
---

## Relevance Rule

Applicable when test files (`*.test.*`, `*.spec.*`) or configs (`jest.config.*`, `eslint*`, `tsconfig*`, `.eslintrc`, `vitest.config.*`) are present. For code without tests and configs, return an empty response.

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
- **Targeted hints** — from "Check-ID hints" by the `TST-` prefix.
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
| TST-01 | Strict static analysis is enabled and mandatory (the compiler/linter does not let unsafe constructs through) |
| TST-02 | Coverage thresholds are configured and enforced |
| TST-03 | Hooks/CI run the checks (tests, lint, types) |
| TST-04 | Critical paths are covered by tests (auth, validation, error handling) |
| TST-05 | Tests are isolated — no shared mutable state between tests |
| TST-06 | No disabled/focused tests without justification |
| TST-07 | Tests verify behavior, not implementation details |
| TST-08 | Sources of nondeterminism are mocked/injected [⚡ dynamic] |
| TST-09 | Golden/snapshot tests cover what matters, not the entire object |

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

> The examples below are illustrative. Take the concrete strict-analysis tools, idioms,
> and anti-patterns for the current runtime from the loaded profile
> (`stacks/<runtime>.md`, the Idioms/Anti-patterns/Tooling by category/Check-ID
> hints sections by the `TST-` prefix). Node: tsconfig `strict`, jest/vitest, husky/lefthook.
> Go: `go vet`/`staticcheck`/`golangci-lint`, `go test -cover`, Taskfile/lefthook.

**TST-01 — Strict static analysis is enabled and mandatory:**
- Node: `tsconfig.json` with `strict: false` or important flags disabled (`noImplicitAny`, `strictNullChecks`); `ts-ignore`/`@ts-expect-error` without explanations; `any` in public APIs
- Go: `go vet` + `staticcheck` + `golangci-lint` (including `errcheck`/`nilness`/`gosec`) are not configured or not mandatory in CI
- Important linter/analyzer rules disabled without justification

**TST-02 — Coverage thresholds are configured and enforced:**
- Node: `jest.config`/`vitest.config` without `coverageThreshold`
- Go: no `go test -cover -coverprofile`; the threshold is not set and not checked in CI/Taskfile (there is no built-in threshold flag — the check must be explicit)
- Thresholds are configured but not enforced in the pipeline; or set too low (0% / unset)

**TST-03 — Hooks/CI run the checks (tests, lint, types):**
- Missing git hooks (lefthook — cross-stack, husky — Node) or equivalent CI steps
- Hooks/targets do not run static analysis / lint / tests (npm scripts ↔ Taskfile targets)
- Hooks are configured but disabled or bypassed via `--no-verify`

**TST-04 — Critical paths are covered:**
- Auth paths (login, logout, token refresh) without tests
- Validation logic without tests for invalid input
- Error handling paths (what happens on a DB or external-API failure) are not tested
- Edge cases (empty list, maximum value, null) are not covered

**TST-05 — Tests are isolated:**
- Global mocks that affect the isolation of other tests
- Shared mutable state between test cases in the same suite
- Tests that depend on execution order
- Missing setup/teardown for integration tests
- Go: `t.Parallel()` with shared mutable state without synchronization

**TST-06 — No disabled/focused tests without justification:**
- A focused test silences the rest (Node: `.only`)
- A test is disabled without a reason (Node: `.skip`; Go: `t.Skip()` without explanation, hidden behind build tags)
- Commented-out tests without explanation

**TST-07 — Tests verify behavior:**
- Tautology tests (Node: `expect(true).toBe(true)`)
- Tests without assertions (always green)
- Tests that verify implementation details (internal variables, private methods / unexported functions) instead of behavior
- One huge test instead of several isolated by scenario

**TST-08 — Sources of nondeterminism are mocked/injected:**
- Uncontrolled randomness (Node: `Math.random()` without a mock; Go: `math/rand` without a fixed seed)
- Real time instead of injection (Node: `new Date()`/`Date.now()` without `useFakeTimers`; Go: direct `time.Now()` instead of injecting a `Clock`)
- Synchronization via delay instead of waiting for an event (Node: `setTimeout`/`sleep(N)`; Go: `time.Sleep` to synchronize goroutines instead of a channel/`WaitGroup`)
- Tests that depend on execution order (shared state)

**TST-09 — Golden/snapshot tests cover what matters:**
- A snapshot of the entire object/component (500+ lines) — changing one line breaks everything (Node: snapshot, Go: `testdata/*.golden`)
- A snapshot of objects with dynamic fields (id, createdAt) without masking
- Updating golden/snapshot without reviewing the changes

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| TST-01 | Strict static analysis is enabled and mandatory (the compiler/linter does not let unsafe constructs through) | ✅ PASS | High | `tsconfig.json:5` — strict: true | — | — |
| TST-02 | Coverage thresholds are configured and enforced | ❌ FAIL 🟠 | High | `jest.config.ts:1` | **1. Add `coverageThreshold: { global: { lines: 80 } }`** \\ 2. Configure thresholds for critical modules only \\ 3. Add a coverage check to CI without blocking | No |
| TST-06 | No disabled/focused tests without justification | ⏸ ACCEPTED | Medium | `tests/auth.test.ts:45` | In baseline: temporary for debugging, will be removed | — |

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

If everything is PASS, output: `✅ Test and linter configurations are in order.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-tests.md`

```
# Audit Report: Test & Linter Integrity — <YYYY-MM-DD HH:MM>
<table>
```
