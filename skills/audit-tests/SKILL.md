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
| TST-02 | Coverage thresholds настроены |
| TST-03 | Pre-commit/pre-push хуки запускают тесты |
| TST-04 | Критические пути покрыты (auth, validation, errors) |
| TST-05 | Тесты изолированы (нет shared mutable state) |
| TST-06 | Нет .only/.skip в тестах |
| TST-07 | Нет tautology-тестов |

## Правила верификации

1. **Только чеклист**: оценивай ТОЛЬКО проверки выше. Не добавляй новые.
2. **Нет доказательства = ✅ PASS**: не можешь указать `файл:строка` — ставь PASS.
3. **Baseline приоритетен**: check_id есть в `docs/audit-baseline.yml` → `⏸ ACCEPTED`.
4. **Только 🔴/🟠 FAIL требуют решения**: 🟡/🟢 — решение необязательно.

## Baseline

До анализа:
```bash
cat ./docs/audit-baseline.yml 2>/dev/null
```

## Контекст анализа

**Конфигурация тестов:**
- `jest.config` / `vitest.config` без coverage thresholds
- Моки, перекрывающие реальное поведение (mock DB вместо real DB)
- Глобальные моки, влияющие на изоляцию тестов
- Отсутствие setup/teardown для интеграционных тестов
- Тесты, зависящие от порядка выполнения

**Качество тестов:**
- Тесты без assertions (пустые expect)
- `expect(true).toBe(true)` и другие tautology-тесты
- Тесты, проверяющие implementation details вместо behavior
- Отсутствие тестов для error paths и edge cases
- Один огромный тест вместо нескольких изолированных

**TypeScript конфигурация:**
- `strict: false` или отключенные важные флаги (`noImplicitAny`, `strictNullChecks`)
- `ts-ignore` / `@ts-expect-error` без объяснений
- `any` типы в публичных API

**Линтер конфигурация:**
- Отключённые важные правила без обоснования (no-any, no-explicit-any)
- `eslint-disable` без комментария
- Несовместимые правила между lint и formatter

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| TST-01 | TypeScript strict mode включён | ✅ PASS | — | — |
| TST-02 | Coverage thresholds настроены | ❌ FAIL 🟠 | `jest.config.ts:1` | **1. Добавить `coverageThreshold: { global: { lines: 80 } }`** \\ 2. Настроить thresholds только для критических модулей \\ 3. Добавить coverage check в CI без блокировки |
| TST-06 | Нет .only/.skip в тестах | ⏸ ACCEPTED | `tests/auth.test.ts:45` | В baseline: временно для дебага, будет убрано |

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
