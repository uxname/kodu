# Stack Profile: Generic (fallback)   (id: generic)
Tier: fallback

Used when the runtime is not detected (no known marker file) or the required
profile is not found. No tools are available; only **stack-neutral** checks work.

## 1. Detection signals
- none of the known markers (`package.json`, `go.mod`, `pyproject.toml`/`requirements.txt`, `Cargo.toml`, `pom.xml`/`build.gradle`)

## 2. Tooling by category
| Category | Command | How to read the output |
|-----------|---------|------------------|
| unused-code | — | no tool → manual scan of symbol references → `🔍 UNVERIFIED` |
| clone-detection | — | grep for repeated blocks manually → `🔍 UNVERIFIED` |
| dep-audit | — | none → mark "dependency audit unavailable" |
| env-extraction | `grep -rEoh '[A-Z][A-Z0-9_]{2,}' . 2>/dev/null \| sort -u` (rough) | env variable candidates, verify manually |
| arch-lint | — | manual layer analysis → `🔍 UNVERIFIED` |
| lint/format | — | — |
| type-check | — | — |
| test-run | — | — |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| trufflehog filesystem . 2>/dev/null \|\| true` | stack-neutral, always works |

## 3. Idioms
Stack-neutral expectations (apply to any language):
- Errors are handled or explicitly propagated, not silently suppressed.
- External calls have timeouts and a cancellation mechanism.
- Secrets are not hardcoded; configuration is isolated.
- Logs are structured, with no PII or secrets.
- Names describe purpose; there is no logic duplication.
- Documentation (README, links, env) is in sync with the code.

## 4. Anti-patterns
- Suppressing errors without handling.
- Hardcoded secrets, scattered configuration.
- Logic duplication, dead code.

## 5. Check-ID hints
For all stack-specific Check IDs without unambiguous evidence, mark `🔍 UNVERIFIED`.
Real `✅ PASS`/`❌ FAIL` are allowed only for stack-neutral dimensions:
secrets (SEC-*), docs (DOC-*), naming readability (NAM-02/03/05), bug logic
(BUG-06/07/09), api-contracts (API-*) when there is explicit evidence.
