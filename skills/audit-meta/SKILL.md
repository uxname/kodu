---
name: audit-meta
description: >
  Audit quality meta-control: checks codebase coverage, baseline currency,
  and evidence quality in the reports. Call AFTER all audits and audit-verify — /audit-meta
  or automatically at the end of /audit.
---

## Task

You are the coordinator of the audit process. You check not the code, but the quality of the audit itself: is everything covered, are there any stale exceptions, is there enough evidence.

## Step 1 — Data collection

```bash
# Latest session folder
ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
```

Read all `*.md` files from the session folder and `docs/audit-baseline.yml`.

## Step 2 — Scope Verification

1. Get the list of all source files in the project:
   ```bash
   find ./src -type f \( -name "*.ts" -o -name "*.js" -o -name "*.py" \) 2>/dev/null | head -100
   ```
2. Cross-check against the evidence in the reports: every directory/module is mentioned in at least one audit file.
3. Output the list of directories not mentioned in any report:

```
### Uncovered modules
- `src/module-name/` — no audit contains references to files in this directory
```

If all modules are covered → `✅ Scope: all modules checked.`

## Step 3 — Baseline Expiry

For each entry in `docs/audit-baseline.yml` with an `expires` field:
- Compare the date against the current one (`date +%Y-%m-%d`)
- If it has expired, output a warning:

```
### ⚠️ Expired exceptions in baseline
| check_id | expires | accepted_by | reason |
|----------|---------|-------------|--------|
| OWA-06 | 2025-12-31 | username | nginx rate limit |
```

If none have expired → `✅ Baseline: all exceptions are current.`

## Step 4 — Quality of Evidence

Check every row with `❌ FAIL` in the reports:
- Is there a `file:line` in the "Evidence" column? If not → low evidence quality.
- Is there a specific piece of code/value, or only a file name?

Output:
```
### Findings with insufficient evidence
| Audit file | Check ID | Problem |
|------------|----------|----------|
| audit-owasp | OWA-02 | No specific line, only a file name |
```

If all evidence is sufficient → `✅ Evidence: all FAILs are backed by evidence.`

## Step 4.5 — False Positive Suppression Audit

Check the types of entries in `docs/audit-baseline.yml`. A correct baseline distinguishes:
- `accepted-risk` — the risk is known and intentionally accepted
- `false-positive` — the tool was wrong, there is no violation
- `intentional-design` — an architectural decision, not a bug

For entries without a `type` field, output:
```
### Baseline entries without a type (need clarification)
| check_id | reason | Recommended type |
|----------|--------|-------------------|
| OWA-06 | nginx rate limit | accepted-risk |
```

Without `type`, the baseline turns into a dumping ground — clarification is mandatory.

## Step 5 — Report

```
# Audit Meta Report — <YYYY-MM-DD HH:MM>

## Scope Coverage
<result of Step 2>

## Baseline Expiry
<result of Step 3>

## Evidence Quality
<result of Step 4>

## Summary
| Check | Status |
|----------|--------|
| Scope Coverage | ✅ / ⚠️ N modules uncovered |
| Baseline Expiry | ✅ / ⚠️ N expired |
| Evidence Quality | ✅ / ⚠️ N weak evidence |
| Baseline Types | ✅ / ⚠️ N entries without type |
```

## Language Rule

Audit results must be written in plain, clear language. Avoid complex terms, jargon, and abstract concepts unless necessary. Common technical terms (Docker, HTTP, API, JSON, URL) are fine. Describe problems so they are understandable to a developer of any level, not only a narrow specialist in the area.

## Saving

1. Find the session folder:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If empty, create it: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Save via Write: `<AUDIT_DIR>/audit-meta.md`
