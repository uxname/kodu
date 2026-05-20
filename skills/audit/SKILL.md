---
name: audit
description: >
  Мастер-оркестратор комплексного аудита кодовой базы. Запускает все 14 специализированных
  проверок и группирует результаты по компонентам системы. Вызывай при /audit или
  запросе "полный аудит", "комплексный аудит кодовой базы".
---

## Задача

Ты — ведущий инженер по качеству и безопасности, проводящий полный аудит кодовой базы.

## Шаг 1 — Подготовка сессии

Выполни через Bash:

```bash
# Удалить старые сессии, оставить 2 последних
ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | tail -n +3 | xargs rm -rf 2>/dev/null
# Создать папку новой сессии
mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")
```

## Шаг 2 — Baseline

```bash
cat ./docs/audit-baseline.yml 2>/dev/null
```

Если файл не существует — создай:

```bash
cp ./skills/audit/audit-baseline-template.yml ./docs/audit-baseline.yml 2>/dev/null || \
  printf "accepted: []\n" > ./docs/audit-baseline.yml
```

Сообщи: `📋 docs/audit-baseline.yml создан. Заполни для подавления принятых рисков.`

Содержимое baseline передавай как контекст каждому sub-скиллу.

## Шаг 3 — Декомпозиция системы

Перечисли логические компоненты (Аутентификация, API, Фоновые задачи и т.д.) перед запуском аудитов. Используй структуру папок как ориентир.

## Шаг 4 — Анализ по 14 направлениям

**ОБЯЗАТЕЛЬНО:** Для каждого направления вызови специализированный скилл через `Skill`. Прямой анализ без скилла недопустим.

| # | Направление | Скилл |
|---|-------------|-------|
| 1 | Secrets Leak | `audit-secrets` |
| 2 | Logging | `audit-logging` |
| 3 | Naming | `audit-naming` |
| 4 | Architecture | `audit-architecture` |
| 5 | Tests & Linters | `audit-tests` |
| 6 | YAGNI | `audit-yagni` |
| 7 | Boundary Validation | `audit-validation` |
| 8 | Error Handling | `audit-errors` |
| 9 | Concurrency | `audit-concurrency` |
| 10 | OWASP Security | `audit-owasp` |
| 11 | Performance | `audit-performance` |
| 12 | Deployment | `audit-deployment` |
| 13 | Bugs & Logic | `audit-bugs` |
| 14 | API Contracts | `audit-api-contracts` |

Порядок: последовательно. **Правило пропуска:** направление нерелевантно → пропусти без упоминания.

**Каждый скилл ОБЯЗАН сохранить файл в папку сессии: `./docs/audits/<SESSION>/audit-<name>.md`**

## Шаг 5 — Сводный отчёт по компонентам

После всех скиллов собери только строки `❌ FAIL` и `⏸ ACCEPTED`:

```
## Компонент: [Название]

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| OWA-02 | Auth на protected routes | ❌ FAIL 🔴 | `routes/admin.ts:14` | **1. Добавить authMiddleware** \\ 2. ... \\ 3. ... |
| OWA-06 | Rate limiting | ⏸ ACCEPTED | `src/app.ts` | В baseline: nginx rate limit |
```

Если все PASS — выведи: `✅ Проблем не обнаружено.`

## Шаг 6 — Итоговая таблица

```
## Сводка

| Компонент | ❌ FAIL 🔴 | ❌ FAIL 🟠 | ❌ FAIL 🟡🟢 | ⏸ ACCEPTED | Итого FAIL |
|-----------|-----------|-----------|------------|-----------|------------|
| [Компонент] | N | N | N | N | N |
| **ИТОГО** | **N** | **N** | **N** | **N** | **N** |
```

## Шаг 7 — Разбор FAIL 🔴

Для каждого `❌ FAIL 🔴`:

```
### 🔴 [Check ID] — [Компонент]
**Файл:** `path/file.ts:line`
**Проверка:** [название]
**Доказательство:** [конкретный код]
**Решение:** [первый вариант из таблицы]
```

## Шаг 8 — Финальная верификация

Вызови: `Skill("audit-verify")`

## Шаг 9 — Сохранение

Сохрани полный отчёт через Write: `./docs/audits/<SESSION>/audit-report.md`

```
# Full Audit Report — <YYYY-MM-DD HH:MM>
## Компоненты системы
## Компонент: [Название]
## Сводка
## Критические риски
```

Сообщи путь к папке сессии и число FAIL по severity.
