# Stack Profile: Python   (id: python)
Tier: general

A general-tier profile: provides neutral idioms and tooling pointers, but
without guarantees they are present. Mark stack-specific findings without unambiguous evidence
as `🔍 UNVERIFIED`.

## 1. Detection signals
- `pyproject.toml` / `requirements.txt` / `setup.py`
- additional: `Pipfile`, `poetry.lock`, `tox.ini`

## 2. Tooling by category
| Category | Command | How to read the output |
|-----------|---------|------------------|
| unused-code | `vulture . 2>/dev/null \|\| true` | dead code → YAGNI-02 |
| clone-detection | `pylint --disable=all --enable=duplicate-code . 2>/dev/null \|\| true` | duplicates → REINV-03 |
| dep-audit | `pip-audit 2>/dev/null \|\| safety check 2>/dev/null \|\| true` | CVEs in dependencies |
| env-extraction | `grep -rEoh 'os\.environ(\.get)?\(?\["'\'']?[A-Z0-9_]+' . 2>/dev/null \| sort -u` | env from code → DOC-02 |
| arch-lint | `import-linter 2>/dev/null \|\| true` | layers/cycles |
| lint/format | `ruff check . 2>/dev/null \|\| flake8 2>/dev/null \|\| true` | — |
| type-check | `mypy . 2>/dev/null \|\| pyright 2>/dev/null \|\| true` | — |
| test-run | `pytest 2>/dev/null \|\| true` | — |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| true` | stack-neutral |

## 3. Idioms
- **Errors:** specific exceptions, not a bare `except:`; `finally`/context managers (`with`) for resources.
- **Concurrency:** `asyncio` with `async/await`; cancellation via `asyncio.CancelledError`/`asyncio.timeout`; `concurrent.futures` for CPU.
- **Env/config:** `os.environ`/pydantic-settings, validation at startup; isolation in a config module.
- **Logging:** the `logging` module with structure; no `print()` in production.
- **Null-safety:** explicit `is None` checks; `Optional[...]`.
- **Type coercion:** explicit `int()`/`str()` with `ValueError` handling.
- **Deps:** stdlib (`itertools`, `functools`, `collections`, `datetime`, `secrets`); `pydantic` for validation.

## 4. Anti-patterns
- Bare `except:` / `except Exception: pass`.
- `print()` for logging in production.
- Mutable default arguments (`def f(x=[])`).
- Scattered `os.environ[...]` without centralization.

## 5. Check-ID hints
- `LOG-01` → `print()` instead of `logging`.
- `ERR-01` → `except: pass` (swallowing).
- `BUG-04` → mutable default argument.
- `ARC-05` → scattered `os.environ`.
- Other stack-specific → when evidence is lacking, `🔍 UNVERIFIED`.
