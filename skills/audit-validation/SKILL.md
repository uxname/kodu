---
name: audit-validation
description: >
  Audit of boundary data validation: checking incoming data at the system boundaries,
  missing sanitization, trust boundary violations. Run on /audit-validation.
---

## Relevance Rule

Applicable to code that accepts external data: HTTP handlers, WebSocket, CLI args, file parsers, event consumers, gRPC endpoints. For purely internal code with no external inputs, return an empty response.

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
- **Targeted hints** — from "Check-ID hints" by the `VAL-` prefix.
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
| VAL-01 | All incoming data (body, params, query) passes schema validation |
| VAL-02 | Strings have a maxLength, numbers have a range, enum values use a whitelist |
| VAL-03 | Parsing untrusted input handles errors and validates the structure of the result |
| VAL-04 | Identity data is taken from the authenticated context (not from user input) |
| VAL-05 | Nested structures and arrays are bounded (depth, minItems/maxItems) |
| VAL-06 | The validator does not perform implicit coercion (string "false" → boolean true) [⚡ dynamic] |
| VAL-07 | Prototype pollution: merge/assign with user input filters out `__proto__`, `constructor`, `prototype` |
| VAL-08 | File uploads: the MIME type is checked by content, the filename is sanitized, the size is limited |

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

**VAL-01 — All incoming data passes schema validation:**
- The HTTP request body is used directly without schema validation (Node: zod/joi/yup/etc; Go: validation in the resolver/handler — go-playground/validator or manual checks; gqlgen types the input at the schema level, but application invariants must still be checked)
- Query params / path params are used without typing and validation
- No check for required fields
- No type validation (a string may arrive instead of a number)
- WebSocket, CLI args, event payloads without input validation

**VAL-02 — Strings, numbers, enums are bounded:**
- No maxLength for strings (DoS via a huge string)
- No range checks for numbers (negative IDs, huge offsets, NaN)
- No whitelist for enum fields (any string value is accepted)
- No format check (email, UUID, date) where applicable

**VAL-03 — Parsing untrusted input with protection:**
- Node: `JSON.parse` without try/catch — throws a SyntaxError on invalid input
- Go: `json.Unmarshal` / parsing without checking the returned error (err ignored)
- Parsing without subsequent structure validation (field types not checked)
- Trusting the structure of the parsed result without a schema check

**VAL-04 — Identity from the authenticated context:**
- JWT claims used without signature verification
- The user ID is taken from the request body instead of `req.user` / the authenticated context
- Role/permission data is taken from user-controlled input
- Mass assignment: an object from the body is saved directly to the DB without a field whitelist

**Nesting and collections:**
- Recursive / deeply nested schemas without a depth limit → ReDoS / stack overflow
- Arrays without maxItems → unbounded payload growth
- Nested objects without maxProperties

**Coercion:**
- Zod: `.coerce.boolean()` accepts the string "false" as true
- Joi: without `.options({ convert: false })` it implicitly casts types
- express-validator: without explicit type checks it accepts "1" as the number 1

**VAL-07 — Prototype pollution:**
- Mostly JS-specific (`__proto__`/`constructor`/`prototype`). In Go prototype pollution does not apply (there is no prototype-based object model) → for the Go runtime mark VAL-07 as `N/A`. The concept below is relevant for Node:
- `Object.assign(target, userInput)` without a check — the `__proto__` key pollutes Object.prototype
- `_.merge(obj, userInput)` in lodash < 4.17.21 — vulnerable to prototype pollution
- Deep merge from user input without sanitizing keys (`constructor`, `prototype`, `__proto__`)
- Consequence: `({}).isAdmin === true` for all objects after the attack

**VAL-08 — Secure file uploads:**
- Type check only via `file.mimetype` — the value is supplied by the client, not verified
- `path.join(uploadDir, file.originalname)` — `originalname` may contain path traversal (`../../../etc/passwd`)
- No `maxFileSize` limit — DoS via a huge file
- Allowed extensions are not restricted — executable files (.sh, .exe, .php) can be uploaded

## Boundary With Other Audits

- **Validation** — this skill is primary. `audit-owasp` and `audit-bugs` defer here for findings like "missing input check".
- **User ID from the auth context** — VAL-04 is primary. `audit-owasp` (IDOR) is secondary.

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| VAL-01 | All incoming data (body, params, query) passes schema validation | ✅ PASS | High | `handlers/` — all routes use zod schemas | — | — |
| VAL-02 | Strings have a maxLength, numbers have a range, enum values use a whitelist | ❌ FAIL 🟠 | High | `handlers/user.ts:22` | **1. Add maxLength to the zod schema** \\ 2. Manual length check in the handler \\ 3. Constraint at the DB level | No |
| VAL-04 | Identity data is taken from the authenticated context (not from user input) | ⏸ ACCEPTED | Medium | `routes/order.ts:9` | In baseline: legacy endpoint, refactor planned | — |

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

If everything is PASS, output: `✅ Boundary data validation is implemented correctly.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-validation.md`

```
# Audit Report: Boundary Data Validation — <YYYY-MM-DD HH:MM>
<table>
```
