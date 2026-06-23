---
name: audit-logging
description: >
  Logging quality audit: verbosity, log safety, absence of sensitive data,
  adherence to best practices. Run on /audit-logging or a request to check logs/logging.
---

## Relevance Rule

Before analysis, assess: does the code contain logger calls, console.log/error/warn, writes to log files, logging middleware, or audit trails? If the file under analysis contains no logging at all, return an empty response without a table.

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
- **Targeted hints** — from "Check-ID hints" by the prefix `LOG-`.
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
| LOG-01 | Production code does not write to stdout/stderr directly, bypassing the structured logger |
| LOG-02 | PII is not logged (email, phone, names, addresses, financial data) |
| LOG-03 | Secrets and tokens do not end up in logs |
| LOG-04 | Requests are traceable (a request ID or correlation ID end to end) |
| LOG-05 | Log format is structured (JSON) in production |
| LOG-06 | Critical operations are logged (auth, create, update, delete) |
| LOG-07 | User input is sanitized before logging (protection against log injection) |
| LOG-08 | Critical security events are logged: successful/failed login, logout, permission changes, bulk data export [⚡ dynamic] |

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

**LOG-01 — No direct writes to stdout/stderr bypassing the logger:**
- Node: `console.log`, `console.error`, `console.warn` in production code instead of the structured logger
- Go: `fmt.Print*` / `log.Print*` / `println` on production paths instead of `slog`
- Logging every loop iteration without level control
- Verbose logs without a conditional flag (should be behind the debug level, e.g. `if (logger.isDebugEnabled())`)

**LOG-02 — PII is not logged:**
- Logging email, phone numbers, names, addresses, dates of birth
- Logging full request/response bodies containing personal data
- Financial data (transaction amounts tied to a person) in logs

**LOG-03 — Secrets not in logs:**
- Passwords, API keys, JWT tokens in log messages
- Full connection strings with credentials
- Secrets in stack traces when logging errors

**LOG-04 — End-to-end request traceability:**
- Logs without a request ID / correlation ID — the chain cannot be traced
- The request ID is not propagated to downstream services
- Logs without a user ID or session context for authenticated operations
- Go: no middleware generating a request ID (chi `middleware.RequestID`), and the ID is not placed in `context` for subsequent `slog` calls

**LOG-05 — Structured format in production:**
- Plain-text logs instead of JSON in the production environment
- Different log formats across parts of the application
- Node: nested objects serialized as `[object Object]`
- Go: a text handler instead of JSON in prod (no `slog.NewJSONHandler` for the production configuration)

**LOG-06 — Critical operations are logged:**
- Authentication operations (login, logout, password change) without logs
- Creation/update/deletion of critical entities without an audit trail
- No error logs in the catch blocks of critical operations

**LOG-07 — Protection against log injection:**
- User input passed into a log directly without sanitization
- Newline characters (`\n`, `\r`) in user input can create forged log entries
- ANSI escape codes from user input can corrupt log formatters

**LOG-08 — Security audit trail:**
- Successful and failed authentication are not logged → forensics after an incident is impossible
- A change to a user's permissions (role change, permission grant/revoke) without an audit record
- Mass data export (exporting > N records, bulk delete) without a log of who/when/what
- Password change / email change without a record in the audit log
- Critical administrative operations are performed without a trace

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| LOG-01 | Production code does not write to stdout/stderr directly, bypassing the structured logger | ✅ PASS | High | `src/` grep — no console.* found | — | — |
| LOG-02 | PII is not logged (email, phone, names, addresses, financial data) | ❌ FAIL 🔴 | High | `auth/login.ts:34` | **1. Remove the email from the log, log only userId** \\ 2. Mask PII via a log sanitizer \\ 3. Replace with a structured log without PII | No |
| LOG-04 | Requests are traceable (a request ID or correlation ID end to end) | ⏸ ACCEPTED | Medium | — | In baseline: tracing via an external service | — |

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

If everything is PASS, output: `✅ Logging adheres to best practices.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-logging.md`

```
# Audit Report: Logging Best Practices — <YYYY-MM-DD HH:MM>
<table>
```
