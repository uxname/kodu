---
name: audit-api-contracts
description: >
  API contract audit: response shape consistency, HTTP codes, versioning,
  documentation vs implementation, GraphQL/REST conventions. Run on /audit-api-contracts.
---

## Relevance Rule

Applies to code with HTTP routing, REST/GraphQL APIs, controllers, request/response schemas, OpenAPI/Swagger specs. For CLI tools and internal-only functions without an HTTP interface, return an empty response.

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
- **Targeted hints** — from "Check-ID hints" by the prefix `API-`.
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
| API-01 | Response shape is consistent — a single envelope, or its absence, across the whole API |
| API-02 | HTTP status codes are semantically correct (201 on creation, 4xx for client errors) |
| API-03 | Error responses are machine-readable and consistent in structure |
| API-04 | Field naming is consistent (camelCase or snake_case, not mixed) |
| API-05 | Stack traces and internal details do not leak into error responses |
| API-06 | Pagination includes metadata (total/hasNext) where applicable |
| API-07 | The public API has a versioning strategy |

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

> The examples in this section and in the output table (Express-style middleware, `HttpException`)
> are illustrative. The contract may be REST (Node: Express/Fastify; Go: chi routes)
> or GraphQL (Node: Apollo/Yoga; Go: gqlgen resolvers). The Check IDs and their meaning
> are the same for any transport — take the specifics from the loaded profile.

**API-01 — Consistent response shape:**
- The same entity is returned with different field sets across endpoints
- The envelope pattern is applied inconsistently (`{ data: ... }` in some, a bare object in others)
- Different success-response structures for similar operations

**API-02 — Semantically correct HTTP codes:**
- 200 returned on error (the body contains `{ error: "..." }` but the status is 200)
- 500 returned for user errors (should be 4xx)
- 201 is not used when creating a resource (POST → 200 instead of 201)
- 204 returned with a response body (No Content must not have a body)
- A GET request with side effects (state changes)

**API-03 — Machine-readable and consistent error responses:**
- Different error formats across endpoints
- Missing machine-readable error code (only a human-readable message)
- Different error structures for validation vs auth vs server errors

**API-04 — Consistent field naming:**
- Different names for the same field (`userId` vs `user_id` vs `id`)
- Mixing camelCase and snake_case in one API
- Inconsistent abbreviations (`orgId` vs `organisationId`)

**API-05 — No technical details in responses:**
- Stack trace in the response in production
- Internal filesystem paths in error messages
- SQL errors or DB-specific messages in API responses
- Library/framework versions in response headers or bodies

**API-06 — Pagination with metadata:**
- Different pagination patterns in one API (offset and cursor mixed)
- Missing pagination metadata (total, hasNext) when pagination is present
- No maximum page limit (the client can request everything)

**API-07 — Versioning strategy:**
- A field is removed or renamed without API versioning
- A field type is changed (string → number) without versioning
- No versioning strategy despite a public API (URL versioning, header versioning)

## Boundary With Other Audits

- **Stack traces in responses** (API-05) — primary owner: `audit-errors` (ERR-02). If detected here, add a cross-ref "*see ERR-02*" to the evidence; do not create a duplicate `❌ FAIL`.
- **Input validation** (missing field checks) → `audit-validation`
- **OWASP vulnerabilities** (injection, auth) → `audit-owasp`
- **Performance** (N+1 without DataLoader) → `audit-performance`

## Output Format

| Check ID | Check | Status | Confidence | Evidence | Solution | Fixed |
|----------|----------|--------|-------------|----------------|---------|------------|
| API-01 | Response shape is consistent — a single envelope, or its absence, across the whole API | ✅ PASS | High | `routes/` — all responses use a single `{ data, error }` envelope | — | — |
| API-02 | HTTP status codes are semantically correct (201 on creation, 4xx for client errors) | ❌ FAIL 🟠 | High | `routes/auth.ts:78` | **1. Replace res.status(200) with res.status(400) for validation errors** \\ 2. Add centralized error middleware \\ 3. Use HTTP Exception classes (HttpException) | No |
| API-07 | The public API has a versioning strategy | ⏸ ACCEPTED | Medium | — | In baseline: internal API with no external clients | — |

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

If everything is PASS, output: `✅ API contracts are consistent.`

## Saving Results

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-api-contracts.md`

```
# Audit Report: API Contracts — <YYYY-MM-DD HH:MM>
<table>
```
