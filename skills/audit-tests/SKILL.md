---
name: audit-tests
description: >
  Аудит тестов и линтеров: конфигурации тестов, линтеров, TypeScript, покрытие критических путей,
  false-positive тесты. Запускай при /audit-tests или запросе проверить тесты/конфигурацию.
---

## Правило применимости (Relevance Rule)

Применим при наличии файлов тестов (`*.test.*`, `*.spec.*`), конфигов (`jest.config.*`, `eslint*`, `tsconfig*`, `.eslintrc`, `vitest.config.*`). Для кода без тестов и конфигов — верни пустой ответ.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| TST-01 | TypeScript strict mode включён |
| TST-02 | Coverage thresholds настроены и применяются в CI |
| TST-03 | Pre-commit/pre-push хуки запускают проверки (tests, lint, typecheck) |
| TST-04 | Критические пути покрыты тестами (auth, validation, error handling) |
| TST-05 | Тесты изолированы — нет shared mutable state между тестами |
| TST-06 | Нет пропущенных или зафиксированных тестов (.only/.skip без обоснования) |
| TST-07 | Тесты проверяют поведение, а не детали реализации |

## Правила верификации

1. **Только чеклист**: оценивай ТОЛЬКО проверки выше. Не добавляй новые.
2. **Нет доказательства = ✅ PASS**: не можешь указать `файл:строка` — ставь PASS.
3. **Baseline приоритетен**: check_id есть в `docs/audit-baseline.yml` → `⏸ ACCEPTED`.
4. **Только 🔴/🟠 FAIL требуют решения**: 🟡/🟢 — решение необязательно.

## Baseline

До анализа:
```bash
if [ ! -f ./docs/audit-baseline.yml ]; then
  mkdir -p ./docs
  cp ./skills/audit/audit-baseline-template.yml ./docs/audit-baseline.yml 2>/dev/null || \
    printf "accepted: []\n" > ./docs/audit-baseline.yml
fi
cat ./docs/audit-baseline.yml
```

## Контекст анализа

**TST-01 — TypeScript strict mode:**
- `strict: false` или отключенные важные флаги (`noImplicitAny`, `strictNullChecks`)
- `ts-ignore` / `@ts-expect-error` без объяснений
- `any` типы в публичных API
- Отключённые важные lint-правила без обоснования

**TST-02 — Coverage thresholds в CI:**
- `jest.config` / `vitest.config` без coverage thresholds
- Thresholds настроены но не применяются в CI pipeline
- Пороги установлены слишком низко (0% или не заданы)

**TST-03 — Pre-commit/pre-push хуки:**
- Отсутствие git hooks (lefthook, husky или аналог)
- Хуки не запускают typecheck / lint / тесты
- Хуки настроены но отключены или пропускаются через `--no-verify`

**TST-04 — Критические пути покрыты:**
- Auth пути (login, logout, token refresh) без тестов
- Validation logic без тестов для невалидных входных данных
- Error handling пути (что происходит при сбое DB, внешнего API) не протестированы
- Edge cases (пустой список, максимальное значение, null) не покрыты

**TST-05 — Тесты изолированы:**
- Глобальные моки, влияющие на изоляцию других тестов
- Shared mutable state между тест-кейсами в одном suite
- Тесты, зависящие от порядка выполнения
- Отсутствие setup/teardown для интеграционных тестов

**TST-06 — Нет пропущенных/зафиксированных тестов:**
- `.only` в тестах — остальные тесты не запускаются
- `.skip` без обоснования — тест пропускается в CI
- Закомментированные тесты без объяснения

**TST-07 — Тесты проверяют поведение:**
- `expect(true).toBe(true)` и другие tautology-тесты
- Тесты без assertions (пустые expect, всегда зелёные)
- Тесты, проверяющие implementation details (внутренние переменные, private методы) вместо behavior
- Один огромный тест вместо нескольких изолированных по сценарию

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| TST-01 | TypeScript strict mode включён | ✅ PASS | — | — |
| TST-02 | Coverage thresholds настроены и применяются в CI | ❌ FAIL 🟠 | `jest.config.ts:1` | **1. Добавить `coverageThreshold: { global: { lines: 80 } }`** \\ 2. Настроить thresholds только для критических модулей \\ 3. Добавить coverage check в CI без блокировки |
| TST-06 | Нет пропущенных или зафиксированных тестов (.only/.skip без обоснования) | ⏸ ACCEPTED | `tests/auth.test.ts:45` | В baseline: временно для дебага, будет убрано |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

Если все PASS — выведи: `✅ Конфигурации тестов и линтеров в порядке.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-tests.md`

```
# Audit Report: Test & Linter Integrity — <YYYY-MM-DD HH:MM>
<таблица>
```
