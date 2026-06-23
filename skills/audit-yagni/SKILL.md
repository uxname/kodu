---
name: audit-yagni
description: >
  Audit of over-engineering and YAGNI: unnecessary abstractions, premature optimization,
  unused code. Run on /audit-yagni or a request to find over-engineering.
---

## Relevance Rule

Applicable to any production code with classes, design patterns, or abstractions. For simple utility scripts without architecture, apply only to clear over-engineering patterns.

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
- **Targeted hints** — from "Check-ID hints" by the `YAGNI-` prefix.
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
| YAGNI-01 | No commented-out code |
| YAGNI-02 | No dead code — unused exports, functions, variables |
| YAGNI-03 | Abstractions are justified: an interface/factory has >1 implementation or is required by tests |
| YAGNI-04 | Feature flags are not pinned to a single value |
| YAGNI-05 | Technical debt is current — no abandoned TODO/FIXME without a date or progress |

## Verification Rules

1. **Checklist only**: evaluate ONLY the checks above. Do not add new ones.
2. **Explicit verification = PASS**: assign `✅ PASS` only if you explicitly verified the mechanism (found the schema, config, guard) and confirmed there is no violation — state exactly what was checked.
3. **No evidence = UNVERIFIED**: if you cannot point to a `file:line` for either a violation or a confirmation, assign `🔍 UNVERIFIED`.
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

> The examples are illustrative (Node); runtime anti-patterns are in the profile (Anti-patterns).

**YAGNI-01 — No commented-out code:**
- Blocks of code are commented out instead of deleted (git history preserves the history)
- Commented-out alternative implementations without explanation

**YAGNI-02 — No dead code:**
- Exported functions/classes/constants without a single import
- Variables declared but unused
- Functions defined but never called
- Unused imports

**YAGNI-03 — Abstractions are justified:**
- An interface/abstract class with a single implementation (without tests where a mock is needed)
- Factory/Builder/Strategy for an object with 1–2 creation variants
- Generic types with a single concrete use
- An event bus/pub-sub for direct calls between 2 modules
- A repository pattern over an ORM without a need for the abstraction
- A service layer that merely proxies calls without logic
- A DTO for objects identical to the entity

**YAGNI-04 — Feature flags are not pinned:**
- A feature flag is always `true` or always `false` in the code (not read from config/env)
- Configurability of parameters that never change
- Optional parameters always passed with a single value

**YAGNI-05 — Technical debt is current:**
- TODO/FIXME without a creation date or author name
- TODO/FIXME older than 6 months with no sign of progress
- TODO without a link to an issue/ticket (no tracking)

## Tooling Support

For YAGNI-02 use the **unused-code** category tool from the stack profile
(the "Tooling by category" section): it finds unused exports, functions,
dependencies, and files. Use its output as a hint for YAGNI-02 and verify
each finding manually (`file:line`) before recording it as a FAIL. If the cell is empty
(tier general/generic), check manually and mark findings as `🔍 UNVERIFIED`.

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| YAGNI-01 | No commented-out code | ✅ PASS | High | `src/` — no commented-out code found | — | — |
| YAGNI-02 | No dead code — unused exports, functions, variables | ❌ FAIL 🟡 | High | `lib/formatters.ts:89` | **1. Remove the unused exports** \\ 2. Add ts-prune to CI for automatic detection \\ 3. Mark @deprecated and remove in the next release | No |
| YAGNI-05 | Technical debt is current — no abandoned TODO/FIXME without a date or progress | ⏸ ACCEPTED | Medium | `src/auth.ts:34` | In baseline: known tech debt, tracked in Jira PROJ-123 | — |

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

If everything is PASS, output: `✅ No over-engineering found.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-yagni.md`

```
# Audit Report: Over-engineering & YAGNI — <YYYY-MM-DD HH:MM>
<table>
```
