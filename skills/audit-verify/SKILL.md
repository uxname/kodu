---
name: audit-verify
description: >
  Final verification of all completed audits: checks that findings match the actual code,
  removes false positives, adds missing critical risks, fixes the audit documents.
  Call LAST after all audits — /audit-verify or at the end of /audit.
---

## Task

You are a senior security and quality engineer performing the final verification of audit results. Your goal: make sure every finding actually exists in the code, rather than being a hallucination or a stale artifact.

## Step 1 — Collecting the audit documents

1. Find the latest session folder via Bash:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
2. Read all `*.md` files from this folder via Read.

If the folder is not found, tell the user: `⚠️ No audit documents found. Run an audit first.` and stop.

## Step 1.5 — Checking the audit's freshness

```bash
SESSION_DIR=$(ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||')
SESSION_DATE=$(basename "$SESSION_DIR" | cut -c1-10)
DAYS_OLD=$(( ($(date +%s) - $(date -d "$SESSION_DATE" +%s 2>/dev/null || date -j -f "%Y-%m-%d" "$SESSION_DATE" +%s 2>/dev/null || echo 0)) / 86400 ))
echo "Audit age: $DAYS_OLD days (stale after 30)"
```

If `$DAYS_OLD > 30`, add to the final report:
```
⚠️ STALE AUDIT — the report was created N days ago. The code may have changed. A re-audit is recommended.
```

## Step 2 — Verifying each finding

For each row in each audit table, perform the following check:

### 2.1 Checking that the file and line exist

- Extract `file:line` from the "Scenario" column.
- Open the file via Read (specifying offset and limit around the line).
- Confirm that the indicated code actually exists.

### 2.2 Classifying the finding

| Status | Condition |
|--------|---------|
| ✅ **Confirmed** | The code exists, the risk is current |
| ❌ **False Positive** | The file/line does not exist, or the code does not match the description |
| ⚠️ **Stale** | The code exists, but the risk has already been mitigated (protection added) |
| 🔍 **Missed** | During verification, a critical risk was found that was not in the original audit |

### 2.3 Checking critical risks (🔴)

For each 🔴 Critical risk, additionally:
- Read the entire context of the function/class (±30 lines).
- Make sure there are no existing protections in the nearby code that the audit missed.
- Check whether a similar pattern exists in neighboring files (grep over the directory).

### 2.4 Checking the baseline expiry

For each `⏸ ACCEPTED` entry in the audit files:
1. Find the corresponding entry in `docs/audit-baseline.yml`.
2. If the `expires` field is set and the date has passed, the status automatically changes to `❌ FAIL 🟠` with the note `[baseline expired: <date>]`.
3. If the `expires` field is absent, keep `⏸ ACCEPTED` (an indefinite exception).

## Step 3 — Fixing the audit documents

For each file with detected problems:

1. **Remove** rows with `❌ False Positive` — they clutter the report.
2. **Mark** rows with `⚠️ Stale` — add the value `[✓ stale]` to the "Status" column.
3. **Add** rows with `🔍 Missed` to the appropriate audit file.
4. Rewrite the file via Write with the corrected content.
5. **Expired ACCEPTED** rows — change the status from `⏸ ACCEPTED` to `❌ FAIL 🟠 [baseline expired: YYYY-MM-DD]` and remove the baseline reference from the "Solution" column.
6. **Stale references** — if the file or line from a FAIL/ACCEPTED entry no longer exists (refactoring), remove the entry from the report with a note in the log: `[removed: <file> not found]`.

If a file does not need changes, do not rewrite it.

## Step 4 — Verification report

Output the final verification report:

```
## Verification Results

| Audit file | ✅ Confirmed | ❌ False Positive | ⚠️ Stale | 🔍 Missed |
|------------|---------------|-----------------|------------|-------------|
| audit-secrets | N | N | N | N |
| audit-owasp   | N | N | N | N |
| ...           | N | N | N | N |
| **TOTAL**     | **N** | **N** | **N** | **N** |

### Fixed documents
- `<SESSION>/audit-<name>.md` — N false positives removed, N missed risks added
- ...

### Missed critical risks
[List of new 🔴 risks with file:line, if any were found]
```

## Step 5 — Saving the verification report

1. Find the current session folder via Bash:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   If the output is empty, create a new one: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")` and use its path.
2. Save the report via Write to the file: `<AUDIT_DIR>/audit-verify.md`

File structure:
```
# Audit Verification Report — <YYYY-MM-DD HH:MM>

## Verification Results
<table from Step 4>

### Fixed documents
<list>

### Missed critical risks
<list or "None found">
```

Tell the user the path to the session folder and a brief summary: how many false positives were removed, how many risks were added.

## Rules

- Do not invent new risks — only confirm or refute existing ones, plus obvious omissions in the code you reviewed.
- Do not change the risk levels (🔴/🟠/🟡/🟢) of confirmed findings.
- Do not edit the "Solution options" column of confirmed findings.
- If an audit file contains only `✅ ... none found`, no verification is needed for it — skip it.

## Language Rule

Audit results must be written in plain, clear language. Avoid complex terms, jargon, and abstract concepts unless necessary. Common technical terms (Docker, HTTP, API, JSON, URL) are fine. Describe problems so they are understandable to a developer of any level, not only a narrow specialist in the area.
