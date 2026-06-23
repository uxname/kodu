---
name: audit-secrets
description: >
  Secrets leak audit: searching for hardcoded keys, passwords, tokens, and credentials in code.
  Run when the user asks to check code for secrets, credential leaks,
  hardcoded passwords, or on the /audit-secrets invocation.
---

## Relevance Rule

Before analysis, assess: does the code contain configuration, connection strings, tokens, encryption keys, credentials, or work with external APIs? If the file/module under analysis contains none of the listed patterns, return an empty response without a table.

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
- **Targeted hints** — from "Check-ID hints" by the prefix `SEC-`.
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
| SEC-01 | No hardcoded credentials in code (passwords, tokens, API keys, private keys) |
| SEC-02 | Files with secrets are excluded from VCS (.env* in .gitignore) |
| SEC-03 | Secrets are not passed via URL (query params, Basic Auth in URL) |
| SEC-04 | .env.example contains only placeholder values, no real data |
| SEC-05 | Dockerfile contains no secrets in ENV directives |
| SEC-06 | Comments in code contain no credentials |
| SEC-07 | Automated secret scanning is configured (pre-commit or CI) |

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

**SEC-01 — No hardcoded credentials in code:**
- Passwords, tokens, API keys in string literals
- DB connection strings with credentials (`postgres://user:pass@host`)
- Private keys, certificates, JWT secrets in code
- Base64-encoded credentials in code
- Test/dev credentials that could end up in prod
- Patterns: `password = "..."`, `token = "..."`, `key = "..."`, `secret = "..."`

**SEC-02 — Files with secrets are excluded from VCS:**
- `.env`, `.env.local`, `.env.production` not in `.gitignore`
- Committed `.env` files with real data in the repository
- Certificate and key files (`*.pem`, `*.key`, `*.p12`) not excluded from VCS

**SEC-03 — Secrets not in URLs:**
- API keys or tokens in query params (`?api_key=...`)
- Basic Auth credentials in a URL (`https://user:pass@host`)
- Secrets in redirect_uri or callback URL parameters

**SEC-04 — .env.example without real data:**
- `.env.example` contains real values instead of placeholders (`DB_PASS=realpassword`)
- Placeholder values do not describe the expected format (`DB_URL=` without explanation)

**SEC-05 — Dockerfile without secrets in ENV:**
- Secrets in `ENV` directives of the Dockerfile (visible in docker inspect and image layers)
- Credentials in `ARG` without using build secrets
- Secrets in LABEL or COPY commands

**SEC-06 — Comments without credentials:**
- Passwords or tokens in commented-out code
- TODO comments with examples of real credentials
- Setup instructions with real values

**Automated scanning (SEC-07):**
- Take the tool from the **secret-scan** category of the stack profile (cross-stack: `gitleaks` / `trufflehog`; also detect-secrets) — it is stack-neutral.
- No such scanner in the git hooks (lefthook / pre-commit / husky) — no automatic check on commit
- No secret scanning in CI/the task runner (Taskfile/Makefile target, GitHub Actions secret scanning, GitLab SAST)
- `.gitleaks.toml` / `.secrets.baseline` not configured
- If any of these tools is present in hooks or CI → `✅ PASS`

## Boundary With Other Audits

- **Secrets in code** — this skill is primary. `audit-logging` (LOG-03) and `audit-deployment` (DEP-06) refer here.
- **Secrets in logs** — primary: `audit-logging` (LOG-03). Do not duplicate here.
- **Secrets in Dockerfile ENV** — duplicated intentionally (DEP-06 + SEC-05): the criticality justifies a double check.

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| SEC-01 | No hardcoded credentials in code (passwords, tokens, API keys, private keys) | ✅ PASS | High | `.gitignore`, `src/` checked — no patterns found | — | — |
| SEC-02 | Files with secrets are excluded from VCS (.env* in .gitignore) | ❌ FAIL 🔴 | High | `.gitignore:1` | **1. Add .env to .gitignore** \\ 2. Use git-crypt \\ 3. Remove .env from history with git-filter-repo | No |
| SEC-03 | Secrets are not passed via URL (query params, Basic Auth in URL) | ⏸ ACCEPTED | Medium | `config.ts:9` | In baseline: legacy integration, replacement planned | — |
| SEC-07 | Automated secret scanning is configured (pre-commit or CI) | 🔍 UNVERIFIED | Low | — | — | — |

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

If everything is PASS, output: `✅ No secret leaks found.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-secrets.md`

```
# Audit Report: Secrets Leak — <YYYY-MM-DD HH:MM>
<table>
```
