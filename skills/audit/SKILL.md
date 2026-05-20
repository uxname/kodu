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

Для incremental-аудита (только изменённые файлы) — определи scope:
```bash
git diff --name-only HEAD~1 2>/dev/null | grep -E '\.(ts|js|py|go|rs)$' | head -30
```
Если список < 20 файлов — начинай аудит с них, затем критические пути.

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

## Шаг 3.5 — Критические пути (Risk-Based Prioritization)

```bash
grep -rl "auth\|login\|payment\|billing\|webhook\|cron\|migration" ./src 2>/dev/null | head -20
```

Перечисли critical paths — файлы/модули с наибольшим blast radius:
```
Critical paths (exhaustive depth):
- Authentication: src/auth/**
- Payments: src/payments/**
- Webhooks/Workers: src/workers/**

Standard paths: src/api/**, src/services/**
Low priority (naming/style): src/utils/**
```

Передавай список critical paths каждому sub-скиллу — они должны начинать с этих файлов.

## Шаг 4 — Анализ по 15 направлениям

**ОБЯЗАТЕЛЬНО:** Для каждого направления вызови специализированный скилл через `Skill`. Прямой анализ без скилла недопустим.

**Правило пропуска:** направление нерелевантно → пропусти без упоминания.

Скиллы сгруппированы для параллельного запуска. Группы выполняются последовательно — скиллы внутри группы можно запускать параллельно через Agent-вызовы.

**Группа А — Безопасность:**
| # | Направление | Скилл |
|---|-------------|-------|
| 1 | Secrets Leak | `audit-secrets` |
| 2 | OWASP Security | `audit-owasp` |
| 3 | Boundary Validation | `audit-validation` |

**Группа Б — Логика:**
| # | Направление | Скилл |
|---|-------------|-------|
| 4 | Bugs & Logic | `audit-bugs` |
| 5 | Error Handling | `audit-errors` |
| 6 | Concurrency | `audit-concurrency` |

**Группа В — Качество:**
| # | Направление | Скилл |
|---|-------------|-------|
| 7 | Architecture | `audit-architecture` |
| 8 | Naming | `audit-naming` |
| 9 | YAGNI | `audit-yagni` |

**Группа Г — Операции:**
| # | Направление | Скилл |
|---|-------------|-------|
| 10 | Tests & Linters | `audit-tests` |
| 11 | Logging | `audit-logging` |
| 12 | Performance | `audit-performance` |
| 13 | Deployment | `audit-deployment` |
| 14 | API Contracts | `audit-api-contracts` |
| 15 | Meta-контроль | `audit-meta` |

**Каждый скилл ОБЯЗАН сохранить файл в папку сессии: `./docs/audits/<SESSION>/audit-<name>.md`**

## Шаг 5 — Сводный отчёт по компонентам

После всех скиллов собери только строки `❌ FAIL` и `⏸ ACCEPTED`:

```
## Компонент: [Название]

| Check ID | Проверка | Статус | Доказательство | Решение | Исправлено |
|----------|----------|--------|----------------|---------|------------|
| OWA-02 | Auth на protected routes | ❌ FAIL 🔴 | `routes/admin.ts:14` | **1. Добавить authMiddleware** \\ 2. ... \\ 3. ... | Нет |
| OWA-06 | Rate limiting | ⏸ ACCEPTED | `src/app.ts` | В baseline: nginx rate limit | — |
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

Затем вызови: `Skill("audit-meta")`

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

---

> **Матрица взаимодействий** (межкомпонентные риски и сценарии сбоя) генерируется отдельно: `/audit-matrix`
> Запускай после /audit для полного анализа системы.
