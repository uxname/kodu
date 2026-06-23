---
name: audit
description: >
  Master orchestrator for a comprehensive codebase audit. Runs all 18 specialized
  checks and groups the results by system component. Invoke on /audit or on a
  request for a "full audit" or "comprehensive codebase audit".
---

## Task

You are the lead quality and security engineer running a full audit of the codebase.

## Step 0 — Runtime resolution (stack detection)

The audits are stack-agnostic: the specifics come from the stack profile. Detect the runtime
ONCE here and pass it to all sub-skills so they don't re-detect.

```bash
if   [ -f package.json ]; then echo "runtime=node"
elif [ -f go.mod ]; then echo "runtime=go"
elif [ -f pyproject.toml ] || [ -f requirements.txt ] || [ -f setup.py ]; then echo "runtime=python"
elif [ -f Cargo.toml ]; then echo "runtime=rust"
elif [ -f pom.xml ] || ls build.gradle* settings.gradle* >/dev/null 2>&1; then echo "runtime=java"
else echo "runtime=generic"; fi
```

One `/audit` run = ONE runtime. **Polyglot repositories** (e.g. a Go backend
+ a Node frontend in submodules) should be audited per subproject: run `/audit` separately
inside each subdirectory (`backend/`, `frontend/`). If the current directory has
multiple markers, pick one based on scope and tell the user about it.

Read the profile once via Read: `./skills/audit/stacks/<runtime>.md`
(fallback `./skills/audit/stacks/_generic.md`). The canonical detection logic is in `./skills/audit/runtime-detect.md`.

Pass `runtime=<id>` + the profile contents as context to every sub-skill
(through the same channel as the baseline). The sub-skills will see this and skip their own
detection; when run standalone they detect it themselves.

## Step 1 — Session setup

Run via Bash:

```bash
# Remove old sessions, keep the 2 most recent
ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | tail -n +3 | xargs rm -rf 2>/dev/null
# Create a folder for the new session
mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")
```

For an incremental audit (changed files only) — determine the scope:
```bash
git diff --name-only HEAD~1 2>/dev/null | grep -E '\.(ts|js|py|go|rs)$' | head -30
```
If the list has < 20 files, start the audit with them, then move on to the critical paths.

## Step 2 — Baseline

```bash
cat ./docs/audit-baseline.yml 2>/dev/null
```

If the file does not exist, create it:

```bash
cp ./skills/audit/audit-baseline-template.yml ./docs/audit-baseline.yml 2>/dev/null || \
  printf "accepted: []\n" > ./docs/audit-baseline.yml
```

Report: `📋 docs/audit-baseline.yml created. Fill it in to suppress accepted risks.`

Pass the baseline contents as context to every sub-skill (along with `runtime=<id>` and the profile from Step 0).

## Step 3 — System decomposition

List the logical components (Authentication, API, Background jobs, etc.) before running the audits. Use the folder structure as a guide.

## Step 3.5 — Critical paths (Risk-Based Prioritization)

```bash
grep -rl "auth\|login\|payment\|billing\|webhook\|cron\|migration" ./src 2>/dev/null | head -20
```

List the critical paths — the files/modules with the largest blast radius:
```
Critical paths (exhaustive depth):
- Authentication: src/auth/**
- Payments: src/payments/**
- Webhooks/Workers: src/workers/**

Standard paths: src/api/**, src/services/**
Low priority (naming/style): src/utils/**
```

Pass the critical-paths list to every sub-skill — they must start with these files.

## Step 4 — Analysis across 18 dimensions

**REQUIRED:** For each dimension, invoke the specialized skill via `Skill`. Direct analysis without the skill is not allowed.

**Skip rule:** dimension is irrelevant → skip it without mention.

The skills are grouped for parallel execution. Groups run sequentially — skills within a group can be run in parallel via Agent calls.

**Group A — Security:**
| # | Dimension | Skill |
|---|-------------|-------|
| 1 | Secrets Leak | `audit-secrets` |
| 2 | OWASP Security | `audit-owasp` |
| 3 | Boundary Validation | `audit-validation` |

**Group B — Logic:**
| # | Dimension | Skill |
|---|-------------|-------|
| 4 | Bugs & Logic | `audit-bugs` |
| 5 | Error Handling | `audit-errors` |
| 6 | Concurrency | `audit-concurrency` |

**Group C — Quality:**
| # | Dimension | Skill |
|---|-------------|-------|
| 7 | Architecture | `audit-architecture` |
| 8 | Naming | `audit-naming` |
| 9 | YAGNI | `audit-yagni` |
| 10 | Reinventing the Wheel | `audit-reinvention` |
| 11 | Documentation | `audit-docs` |

**Group D — Operations:**
| # | Dimension | Skill |
|---|-------------|-------|
| 12 | Tests & Linters | `audit-tests` |
| 13 | Logging | `audit-logging` |
| 14 | Performance | `audit-performance` |
| 15 | Deployment | `audit-deployment` |
| 16 | API Contracts | `audit-api-contracts` |
| 17 | Meta-control | `audit-meta` |

**Group E — System level:**
| # | Dimension | Skill |
|---|-------------|-------|
| 18 | Interaction matrix | `audit-matrix` |

**Each skill MUST save a file to the session folder: `./docs/audits/<SESSION>/audit-<name>.md`**

## Step 5 — Summary report by component

After all skills finish, collect only the `❌ FAIL` and `⏸ ACCEPTED` rows:

```
## Component: [Name]

| Check ID | Check | Status | Evidence | Resolution | Fixed |
|----------|----------|--------|----------------|---------|------------|
| OWA-02 | Auth on protected routes | ❌ FAIL 🔴 | `routes/admin.ts:14` | **1. Add authMiddleware** \\ 2. ... \\ 3. ... | No |
| OWA-06 | Rate limiting | ⏸ ACCEPTED | `src/app.ts` | In baseline: nginx rate limit | — |
```

If everything is PASS, output: `✅ No issues found.`

## Step 6 — Summary table

```
## Summary

| Component | ❌ FAIL 🔴 | ❌ FAIL 🟠 | ❌ FAIL 🟡🟢 | ⏸ ACCEPTED | Total FAIL |
|-----------|-----------|-----------|------------|-----------|------------|
| [Component] | N | N | N | N | N |
| **TOTAL** | **N** | **N** | **N** | **N** | **N** |
```

## Step 7 — FAIL 🔴 breakdown

For each `❌ FAIL 🔴`:

```
### 🔴 [Check ID] — [Component]
**File:** `path/file.ts:line`
**Check:** [name]
**Evidence:** [specific code]
**Resolution:** [first option from the table]
```

## Step 8 — Final verification

Invoke: `Skill("audit-verify")`

Then invoke: `Skill("audit-meta")`

## Step 9 — Save

Save the full report via Write: `./docs/audits/<SESSION>/audit-report.md`

```
# Full Audit Report — <YYYY-MM-DD HH:MM>
## System components
## Component: [Name]
## Summary
## Critical risks
```

Report the path to the session folder and the FAIL count by severity.

## Language Rule

Audit results must be written in plain, easy-to-understand language. Avoid complex terms, jargon, and abstract concepts unless necessary. Common technical terms (Docker, HTTP, API, JSON, URL) are acceptable. Describe issues so they are clear to a developer of any level, not just a narrow specialist in the given area.
