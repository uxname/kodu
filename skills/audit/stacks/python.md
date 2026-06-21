# Stack Profile: Python   (id: python)
Tier: general

Профиль уровня general: даёт нейтральные идиомы и ориентиры по инструментам, но
без гарантий их наличия. Стек-специфичные находки без однозначного evidence
помечай `🔍 UNVERIFIED`.

## 1. Detection signals
- `pyproject.toml` / `requirements.txt` / `setup.py`
- доп.: `Pipfile`, `poetry.lock`, `tox.ini`

## 2. Tooling by category
| Категория | Команда | Как читать вывод |
|-----------|---------|------------------|
| unused-code | `vulture . 2>/dev/null \|\| true` | мёртвый код → YAGNI-02 |
| clone-detection | `pylint --disable=all --enable=duplicate-code . 2>/dev/null \|\| true` | дубли → REINV-03 |
| dep-audit | `pip-audit 2>/dev/null \|\| safety check 2>/dev/null \|\| true` | CVE в зависимостях |
| env-extraction | `grep -rEoh 'os\.environ(\.get)?\(?\["'\'']?[A-Z0-9_]+' . 2>/dev/null \| sort -u` | env из кода → DOC-02 |
| arch-lint | `import-linter 2>/dev/null \|\| true` | слои/циклы |
| lint/format | `ruff check . 2>/dev/null \|\| flake8 2>/dev/null \|\| true` | — |
| type-check | `mypy . 2>/dev/null \|\| pyright 2>/dev/null \|\| true` | — |
| test-run | `pytest 2>/dev/null \|\| true` | — |
| secret-scan | `gitleaks detect --no-banner 2>/dev/null \|\| true` | стек-нейтрально |

## 3. Idioms
- **Errors:** конкретные исключения, не голый `except:`; `finally`/context managers (`with`) для ресурсов.
- **Concurrency:** `asyncio` с `async/await`; отмена через `asyncio.CancelledError`/`asyncio.timeout`; `concurrent.futures` для CPU.
- **Env/config:** `os.environ`/pydantic-settings, валидация при старте; изоляция в config-модуле.
- **Logging:** модуль `logging` со структурой; нет `print()` в production.
- **Null-safety:** явные проверки `is None`; `Optional[...]`.
- **Type coercion:** явные `int()`/`str()` с обработкой `ValueError`.
- **Deps:** stdlib (`itertools`, `functools`, `collections`, `datetime`, `secrets`); `pydantic` для валидации.

## 4. Anti-patterns
- Голый `except:` / `except Exception: pass`.
- `print()` для логирования в production.
- Изменяемые аргументы по умолчанию (`def f(x=[])`).
- Разбросанный `os.environ[...]` без централизации.

## 5. Check-ID hints
- `LOG-01` → `print()` вместо `logging`.
- `ERR-01` → `except: pass` (проглатывание).
- `BUG-04` → mutable default argument.
- `ARC-05` → `os.environ` вразброс.
- Прочие стек-специфичные → при нехватке evidence `🔍 UNVERIFIED`.
