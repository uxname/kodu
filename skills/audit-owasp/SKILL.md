---
name: audit-owasp
description: >
  Application security audit against the OWASP Top 10: injections, broken auth, IDOR, XSS,
  CSRF, SSRF, logic vulnerabilities. Run on /audit-owasp.
---

## Relevance Rule

Applies to server-side code with HTTP routing, authentication, database access, or file system access. For purely frontend components without fetch/API calls, apply only the XSS/CSRF sections. For CLI tools without network interaction, return an empty response.

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
- **Targeted hints** — from "Check-ID hints" by the prefix `OWA-`.
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
| OWA-01 | A03: All queries to the DB/OS/LDAP are parameterized, no injection |
| OWA-02 | A01: All protected routes have auth middleware |
| OWA-03 | A01: Resource ownership is verified, no IDOR [⚡ dynamic] |
| OWA-04 | A02: Passwords are stored securely (bcrypt/argon2/scrypt) |
| OWA-05 | A05: Secure server configuration (CORS, security headers, body limits) |
| OWA-06 | A07: Brute-force protection (rate limiting on auth and sensitive endpoints) |
| OWA-07 | A09: Technical information does not leak into responses (stack trace, internal paths) |
| OWA-08 | A10: A URL from user input is not passed to an HTTP client without a whitelist (SSRF) |
| OWA-09 | A05: CSRF protection is implemented (SameSite cookies or CSRF tokens on state-changing requests) |

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

**OWA-01 — Parameterized queries, no injection:**
- SQL strings assembled via concatenation/template literals
- NoSQL injection through unescaped operators (`$where`, `$regex`)
- Command injection: user input reaches shell commands without escaping
- LDAP/XPath queries with unescaped user input

**OWA-02 — Auth middleware on protected routes:**
- Protected routes without auth middleware
- Privilege escalation: an ordinary user invokes an admin action
- Directory traversal in file operations
- Missing authorization on individual router routes

**OWA-03 — Resource ownership is verified:**
- IDOR: a resource is requested by ID without checking the current user's ownership
- Bulk operations modify other users' resources
- Indirect object reference through related entities without an access check

**OWA-04 — Secure password storage:**
- Weak hashing algorithms (MD5, SHA1 for passwords)
- Passwords not hashed with bcrypt/argon2/scrypt
- Symmetric encryption with a hardcoded key
- HTTP instead of HTTPS for transmitting credentials

**OWA-05 — Secure server configuration:**
- CORS must not be a wildcard (`*`) in production; origins by whitelist only
- Security headers not set (X-Frame-Options, CSP, HSTS, etc.)
- Request body size not limited (DoS via a huge request body)
- Open error pages with technical information
- Specifics from the profile (Go: `http.MaxBytesReader`, chi `cors`/`secure`; Node: `express.json({ limit })`, Helmet)

**OWA-06 — Brute-force protection:**
- No rate limiting on login/register/reset-password endpoints
- No rate limiting on sensitive operations (password change, OTP verification)
- Session tokens not invalidated on logout
- JWT verification without an explicit algorithm whitelist: must reject `alg:none` and the RS256↔HS256 confusion (an attacker sends `alg:none`, or RS256 with a public key passed off as the HS256 secret); weak algorithms / short key
- Specifics from the profile (Go: go-oidc `Verifier` / golang-jwt `WithValidMethods`; Node: `jwt.verify(token, secret, { algorithms: [...] })`)

**OWA-07 — Technical information does not leak:**
- Stack trace in production API responses
- Internal file system paths in error messages
- Dependency/framework versions in headers or responses
- SQL errors or DB-specific messages in API responses

**OWA-08 — SSRF: URL from user input without a whitelist:**
- A URL from user input is passed to an HTTP client without a whitelist
- Fetch to internal addresses (169.254.x.x, 10.x.x.x, localhost, metadata endpoints)
- Redirects to internal resources without validating the destination

**OWA-09 — CSRF protection:**
- State-changing requests (POST/PUT/PATCH/DELETE) accepted without a CSRF token and without checking `Origin`/`Referer`
- Cookies without proper attributes (`SameSite`, `Secure`, `HttpOnly`) — the browser will send the cookie on a cross-site request, or it is accessible to scripts / over HTTP
- Mutations (including GraphQL) reachable via GET instead of POST — bypasses CSRF protection
- `SameSite=None` without explicit justification (needed only for cross-site iframe/embed scenarios)

## Boundary With Other Audits

- **Stack trace in responses** (OWA-07) — primary: `audit-errors` (ERR-02). If found here, add a cross-ref "*see ERR-02*" in the evidence, do not create a duplicate `❌ FAIL`.
- **Validation of fields, types, ranges** — primary: `audit-validation`. Do not duplicate here.
- **Secrets in code** — primary: `audit-secrets`. Do not duplicate here.
- **API contracts** — primary: `audit-api-contracts`. Do not duplicate here.

## Tooling Support

Before analysis, use the **dep-audit** category tool from the stack profile
("Tooling by category" section): it surfaces vulnerable dependencies (known CVEs).
This is a separate area, not part of the current checklist, but critical CVEs are
worth listing in the report's notes section. If the category cell is empty
(`tier: general`/`generic`), skip this step.

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| OWA-01 | A03: All queries to the DB/OS/LDAP are parameterized, no injection | ✅ PASS | High | `db/queries.ts` checked — all queries are parameterized | — | — |
| OWA-02 | A01: All protected routes have auth middleware | ❌ FAIL 🔴 | High | `routes/admin.ts:14` | **1. Add authMiddleware to all /admin routes** \\ 2. Use router-level middleware \\ 3. Add a check in each handler | No |
| OWA-05 | A05: Secure server configuration (CORS, security headers, body limits) | ⏸ ACCEPTED | Medium | `app.ts:9` | In baseline: internal service | — |

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

If everything is PASS, output: `✅ No critical OWASP vulnerabilities found.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-owasp.md`

```
# Audit Report: OWASP Application Security — <YYYY-MM-DD HH:MM>
<table>
```
