---
name: audit-architecture
description: >
  Architecture and file structure audit: correctness of inter-layer relationships, dependency
  rule violations, folder structure, circular dependencies. Run on /audit-architecture.
---

## Relevance Rule

Applies to projects with a pronounced layered architecture (MVC, Clean Architecture, DDD, Hexagonal). For single-file scripts or utilities without architectural separation, return an empty response.

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
- **Targeted hints** — from "Check-ID hints" by the prefix `ARC-`.
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
| ARC-01 | Business logic is moved out of route handlers into the service/domain layer |
| ARC-02 | The presentation layer does not interact with the database directly |
| ARC-03 | No circular dependencies between modules |
| ARC-04 | No god objects: files and classes have a single responsibility |
| ARC-05 | Configuration and env variables are isolated in a config module |
| ARC-06 | External dependencies are injected (DI), not imported directly |
| ARC-07 | The domain layer does not import infrastructure modules |

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

**ARC-01 — Business logic in the service/domain layer:**
- Business rules and calculations directly in a route handler / controller
- Complex conditions and data transformations in middleware instead of a service
- Multi-entity operations performed in the handler without a service

**ARC-02 — Presentation layer without direct DB calls:**
- Direct access to the ORM/query builder from routers/controllers
- Importing repositories or the DB client directly into the presentation layer
- SQL / Prisma / Mongoose calls in middleware (the examples are illustrative — Node; for the specific ORM/drivers of the current runtime see the profile, Idioms/Anti-patterns sections)

**ARC-03 — No circular dependencies:**
- Module A imports module B, which imports module A
- Circular deps across several levels (A→B→C→A)
- Direct imports between feature modules (they should go through shared)

**ARC-04 — No god objects:**
- Files >500 lines with unrelated logic (several different responsibilities)
- Classes with methods from different domain areas
- A module doing unrelated things (low cohesion)

**ARC-05 — Configuration is isolated:**
- Access to configuration/env is centralized in one module; reading env scattered outside the config module is a violation (the profile specifies the mechanism: `process.env` in Node / `os.Getenv` in Go / `caarlos0/env` tags, etc.)
- Magic strings with env variable names scattered across the code
- No single point of configuration validation at application startup

**ARC-06 — External dependencies are injected:**
- HTTP clients, email services, DB clients are created inside functions (not injected)
- Dependence on concrete implementations instead of interfaces/abstractions
- The absence of dependency injection makes the code untestable (it cannot be substituted in tests)

**Dependency Rule:**
> The examples below are illustrative (Node); for the specific ORM/infrastructure types of the current runtime see the profile (Idioms/Anti-patterns/Check-ID hints sections, ARC-07).
- The domain/service layer imports Prisma types directly (it should go through a repository interface)
- Business logic depends on Express Request/Response types
- A domain entity contains ORM decorators (TypeORM @Entity in a domain class)
- Violation: `domain/ → infrastructure/` (correct: `infrastructure/ → domain/`)

## Tooling Support

For ARC-03 (circular dependencies) and layer violations, use a tool in the
**arch-lint** category from the stack profile (the "Tooling by category" section). Use
the output as a hint and verify each finding manually (`file:line`) before
recording it as a FAIL. If the cell is empty (tier general/generic), check manually and
mark findings `🔍 UNVERIFIED`.

Note: in Go, import cycles are forbidden by the compiler, so ARC-03 at the
package level is an auto-PASS / N/A; you only need to check logical layers (see the Check-ID
hints in the go profile).

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| ARC-01 | Business logic is moved out of route handlers into the service/domain layer | ✅ PASS | High | `routes/` — handlers delegate to services | — | — |
| ARC-03 | No circular dependencies between modules | ❌ FAIL 🟠 | High | `modules/order/index.ts:3` | **1. Move shared code into a shared module** \\ 2. Apply dependency inversion \\ 3. Split the module into independent parts | No |
| ARC-06 | External dependencies are injected (DI), not imported directly | ⏸ ACCEPTED | Medium | `services/payment.ts:1` | In baseline: refactor planned for Q3 | — |

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

If everything is PASS, output: `✅ Architectural principles are upheld.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-architecture.md`

```
# Audit Report: Architecture & File Structure — <YYYY-MM-DD HH:MM>
<table>
```
