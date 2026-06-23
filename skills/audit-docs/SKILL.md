---
name: audit-docs
description: >
  Documentation audit: mismatches between the README/instructions and the actual code, broken links,
  env-variable drift, comments and JSDoc that contradict signatures, stale
  references to removed entities. Run on /audit-docs or a request to check documentation.
---

## Relevance Rule

Applies to any project that has documentation (README, `docs/**`, JSDoc, run instructions) or code comments. For projects with no documentation at all, check only DOC-01 (the minimum for onboarding) and DOC-04 (comments).

**Focus only on verifiable discrepancies.** This audit does NOT subjectively judge whether "there is enough documentation". `❌ FAIL` is assigned only when the documentation **contradicts** the code or references something that does not exist — and this is provable with a `file:line`. "It would be nice to add a description" without a concrete contradiction is not a finding.

**Boundary With Adjacent Audits** (to avoid duplicating findings):
- `audit-api-contracts` — owns "API documentation vs implementation" (endpoints, response shapes, OpenAPI/GraphQL schema). Hand REST/GraphQL contracts off there.
- `audit-yagni` (YAGNI-05) — owns stale TODO/FIXME with no progress. We do not check TODOs here.
- `audit-naming` — owns self-documenting names. We do not judge "name clarity" here.
- `audit-deployment` (DEP-08) — owns the completeness of `.env.example`. DOC-02 focuses on the **drift** between "code ↔ documentation", not on the file's presence.

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
- **Targeted hints** — from "Check-ID hints" by the prefix `DOC-`.
- If the profile is `tier: general` or `runtime=generic`, mark stack-specific
  findings without unambiguous evidence as `🔍 UNVERIFIED` rather than `❌ FAIL`.
  Mark checks whose mechanism is absent in the runtime as `N/A`.

## Severity Guide

| Severity | Assignment criterion |
|----------|---------------------|
| 🔴 Critical | The documentation steers toward an unsafe action: instructions to disable a security mechanism, a real secret published "as an example", a command with a destructive side effect and no warning |
| 🟠 High | An instruction that breaks onboarding/deployment: an incorrect install/run command, a missing required env variable that causes prod to crash or be misconfigured |
| 🟡 Medium | Drift: README/comment/JSDoc contradicts the current code, broken internal links, references to removed entities |
| 🟢 Low | Minor gap: a public export without a doc comment, a small staleness with no risk |

Rule: severity = impact × the likelihood of misleading a developer/operator × blast radius. The same pattern → the same severity across audits.

## Checklist

| Check ID | Check |
|----------|----------|
| DOC-01 | The README/onboarding documents install, run, test, and build; the commands match the project's task manifest (package.json scripts / Makefile / Taskfile) |
| DOC-02 | Env variables are in sync: every variable used in the code is documented, and there are no documented-but-unused ones |
| DOC-03 | Internal links and paths in the Markdown documentation point to existing files and sections |
| DOC-04 | Comments and JSDoc/docstrings do not contradict the code (signature, parameter names, types, described behavior) |
| DOC-05 | The public surface (the exported API of the library/package) has a doc comment for its purpose and parameters |
| DOC-06 | Project/architecture documentation does not reference removed or renamed entities or outdated facts |

## Verification Rules

1. **Checklist only**: evaluate ONLY the checks above. Do not add new ones.
2. **Discrepancy, not absence**: `❌ FAIL` is valid only with a provable "documentation ↔ code" contradiction or a reference to something nonexistent. A subjective "not enough documentation" is not a FAIL.
3. **Two anchors per finding**: every FAIL must point to BOTH the place in the documentation (`file:line`) AND the place in the code that refutes it (`file:line`).
4. **Explicit verification = PASS**: assign `✅ PASS` only if you compared the document against the code and confirmed they match — state exactly what was compared.
5. **No evidence = UNVERIFIED**: if you cannot provide both anchors, assign `🔍 UNVERIFIED`.
6. **Baseline takes priority**: if the check_id is in `docs/audit-baseline.yml` → `⏸ ACCEPTED`.
7. **Only 🔴/🟠 FAILs require a solution**: 🟡/🟢 — a solution is optional.

## Evidence Quality Rules

Every `❌ FAIL` must include:
- A documentation anchor: `file:line` + a quote of the claim (1–2 lines)
- A code anchor: `file:line` that refutes the claim
- Causal chain: why the discrepancy → who it misleads and how (a developer during onboarding / an operator during deployment / an API reader)

Not allowed:
- Treating a missing comment as a FAIL where the code is self-evident (that is not a contradiction)
- Checking external URLs for availability (the network is unstable) — only internal links/paths and anchors within the repository
- Assuming a command/variable is "probably stale" without comparing it against the code — only `🔍 UNVERIFIED`
- Duplicating findings that belong to `audit-api-contracts`/`audit-yagni`/`audit-naming` (see the boundaries above)

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

**DOC-01 — Instructions match reality:**
- The README references a script/command that is not in the project's task manifest (package.json scripts / Makefile / Taskfile)
- Install steps are documented with the wrong package manager or runtime version (mismatch with `engines`/the lock file)
- Basic onboarding is missing (how to install / run / run tests) even though these scripts exist in the project

**DOC-02 — Env-variable synchronization:**
- A variable is used in the code but absent from `.env.example`/README — the operator will not learn about it
- A variable is documented but no longer read anywhere (a stale instruction)
- A variable's description contradicts the code (default, format, whether it is required)
- Take the syntax for extracting env variables from code from the stack profile (`process.env` / `import.meta.env` / `os.Getenv`)

**DOC-03 — Valid links and paths:**
- A relative link in Markdown points to a nonexistent file
- An anchor (`#section`) points to a missing heading
- A file/folder path in an instruction does not exist in the repository

**DOC-04 — Comments do not lie:**
- JSDoc/TSDoc `@param`/`@returns` does not match the actual signature (name, count, type of parameters)
- A comment describes behavior opposite to the code ("returns null if..." while the code throws an exception)
- A docstring/comment references a parameter/step that is no longer in the function

**DOC-05 — Public surface documentation:**
- Apply ONLY if the project is a published library/package (per the project manifest)
- An exported public function/class/type without a doc comment for its purpose
- For applications (not a published library), DOC-05 = `N/A`, not FAIL

**DOC-06 — Project documentation currency:**
- An architecture document describes a module/service/endpoint that was removed or renamed in the code
- A diagram/chart mentions a technology that is no longer in the dependencies
- The version in the documentation diverges from `package.json`/the tag (if the documentation declares a version)

## Tooling Support

Env variables — code vs documentation (DOC-02): extract env-variable names from the code
with a tool from the **env-extraction** category of the stack profile (the "Tooling by
category" section; the syntax depends on the runtime — Node: `process.env`/`import.meta.env`;
Go: `os.Getenv`). Compare the resulting list against the documented variables
(`.env.example`/README); the difference between the lists is the set of DOC-02 candidates. Verify
each one manually (some variables come from CI/infrastructure, not from `.env`).

README commands vs the task manifest (DOC-01): take the list of tasks/scripts from
the **project's task manifest** (per the stack profile: `package.json` scripts / `Makefile` /
`Taskfile`). Compare the commands mentioned in the README against this manifest — a command from
the README that is not in the manifest → a DOC-01 candidate.

Internal Markdown links (DOC-03): extract relative links `[...](./path)` and check that the paths exist via `Read`/`ls`. Do NOT check external `http(s)://` links for availability.

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| DOC-01 | README documents install/run/tests, commands match the scripts | ❌ FAIL 🟠 | High | `README.md:24` — `npm run start:dev`; `package.json` only has `dev` | **1. Fix the README to `npm run dev`** \\ 2. Add a `start:dev` alias to scripts \\ 3. Generate the commands section from scripts automatically | No |
| DOC-02 | Env variables are in sync | ❌ FAIL 🟠 | High | `src/config.ts:9` reads `REDIS_URL`; not in `.env.example` | **1. Add `REDIS_URL` to `.env.example` with a description** \\ 2. Validate env at startup (zod/envalid) \\ 3. Remove the variable read if it is not needed | No |
| DOC-04 | Comments and JSDoc do not contradict the code | ❌ FAIL 🟡 | Medium | `src/user.ts:40` JSDoc `@returns User`; the function returns `User | null` | **1. Fix `@returns` to `User \| null`** \\ 2. Remove the stale JSDoc \\ 3. Enable `eslint-plugin-jsdoc` for automatic checking | No |
| DOC-05 | The public surface is documented | ✅ PASS | Medium | per the manifest — an application, not a published library → N/A | — | — |

Statuses: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED` / `🔍 UNVERIFIED` / `N/A`

Confidence: `High` — compared the documentation claim against the code, the discrepancy is unambiguous / `Medium` — the discrepancy is likely, context checked selectively / `Low` — limited context, full certainty is impossible

For `❌ FAIL`: exactly 3 solution options, separated by `\\`, with option 1 in bold.

`Fixed`: FAIL → `No` (the developer changes it to `✅ Yes` manually after the fix). PASS / ACCEPTED / UNVERIFIED / N/A → `—`.

Solution requirements:
- Mutually exclusive (not rephrasings of the same thing)
- At least one option is "fix the document", at least one is "fix/remove the code or automate the comparison"
- Realistic cost, without "rewrite all the documentation"
- Option 3 may be "remove the stale section/comment" if it is no longer needed

At the end of the report, add a coverage section:
```
## Audit Coverage
Checked: README.md, docs/**, src/**/*.ts (JSDoc), .env.example
Skipped: CHANGELOG.md (auto-generated), node_modules/**
Files checked: N | Skipped: N
```

If everything is PASS, output: `✅ No discrepancies between documentation and code were found.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-docs.md`

```
# Audit Report: Documentation — <YYYY-MM-DD HH:MM>
<table>
```
