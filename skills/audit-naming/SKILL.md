---
name: audit-naming
description: >
  Naming audit: code readability, naming standards for variables, functions, classes, and files.
  Run on /audit-naming or a request to check code style / naming conventions.
---

## Relevance Rule

This audit applies to any code with identifiers. Skip only auto-generated files (migrations, protobuf-generated, build output). For configuration files without code (JSON, YAML), return an empty response.

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
- **Targeted hints** — from "Check-ID hints" by the prefix `NAM-`.
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
| NAM-01 | Naming convention is followed consistently (camelCase/snake_case) |
| NAM-02 | Variable, function, and class names describe purpose, not implementation |
| NAM-03 | Boolean variables have predicate-style names (is/has/can/should) |
| NAM-04 | Reader functions (get*/find*) have no side effects |
| NAM-05 | Magic numbers and magic strings are replaced with named constants |
| NAM-06 | Utility modules are not a dumping ground for unrelated code |
| NAM-07 | Key entities are named in line with the project's domain glossary |

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

**NAM-01 — Naming convention consistency:**
- The specific convention is determined by the runtime/profile (Idioms section): in Node/JS — camelCase for variables/functions, in Go — PascalCase for exported and camelCase for private identifiers. The examples below are illustrative (Node).
- Mismatch with the language/framework convention (snake_case in JS, camelCase in Python)
- Inconsistent naming of the same entity in different places (`userId` vs `user_id` vs `uid`)
- Mixing styles within a single file or module

**NAM-02 — Names describe purpose:**
- Single-letter variables outside loops (`d`, `x`, `tmp`)
- Abbreviations without expansion (`mgr`, `proc`, `srv`, `usr`)
- Overly generic names (`data`, `info`, `manager`, `handler`, `util`)
- Classes without a noun, functions without a verb
- Names that reflect the implementation rather than the purpose (`arrayOfUsers` instead of `users`)

**NAM-03 — Booleans with predicate-style names:**
- Boolean variables without an is/has/can/should prefix (`enabled`, `valid`, `error`)
- Negative booleans (`isNotValid`, `notDisabled`) — double negation in conditions
- Names that do not make the expected value clear when true

**NAM-04 — Reader functions without side effects:**
- A `getUser` function mutates data or has side effects
- `find*`/`fetch*` methods perform writes/updates
- Violation of the principle: a reading name → a pure operation

**NAM-05 — Named constants instead of magic values:**
- Magic numbers without named constants (`timeout = 86400`, `limit = 100`)
- Magic strings without a constant (`status === 'active'` without an enum/constant)
- Repeated literal values in different places in the code

**NAM-06 — Utility modules are not a dumping ground:**
- `utils.ts`, `helpers.ts` as a dumping ground for unrelated code
- `index.ts` exporting unrelated entities
- Files whose names do not reflect their contents

**Ubiquitous Language:**
- If the project has a `GLOSSARY.md`, `SPEC.md`, or `README.md` with business terms, check for consistency
- `client` in the code while the domain uses `customer` — a ubiquitous language mismatch
- Synonyms for one entity across different layers (`User` / `Account` / `Member`) without explicit justification
- If there is no glossary, assign `🔍 UNVERIFIED`

## Tooling Support

For NAM-06, use the **unused-code** category tool from the stack profile
("Tooling by category" section): it finds unused exports and dead code in
utility modules (a dumping ground for unrelated code). Use the output as a hint
and verify each finding manually (`file:line`) before recording it as a FAIL. If
the cell is empty (tier general/generic), check manually and mark findings
`🔍 UNVERIFIED`.

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| NAM-01 | Naming convention is followed consistently (camelCase/snake_case) | ✅ PASS | High | `src/` — camelCase style is consistent | — | — |
| NAM-05 | Magic numbers and magic strings are replaced with named constants | ❌ FAIL 🟡 | High | `utils/date.ts:18` | **1. Extract into `const MAX_RETRY_ATTEMPTS = 3`** \\ 2. Add an explanation in a comment nearby \\ 3. Move to config | No |
| NAM-06 | Utility modules are not a dumping ground for unrelated code | ⏸ ACCEPTED | Medium | `src/utils.ts` | In baseline: refactor planned | — |
| NAM-07 | Key entities are named in line with the project's domain glossary | 🔍 UNVERIFIED | Low | — | — | — |

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

If everything is PASS, output: `✅ Naming standards are followed.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-naming.md`

```
# Audit Report: Naming — <YYYY-MM-DD HH:MM>
<table>
```
