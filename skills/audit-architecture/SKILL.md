---
name: audit-architecture
description: >
  Аудит архитектуры и файловой структуры: корректность связей между слоями, нарушения
  dependency rules, структура папок, circular dependencies. Запускай при /audit-architecture.
---

## Правило применимости (Relevance Rule)

Применим к проектам с выраженной слоистой архитектурой (MVC, Clean Architecture, DDD, Hexagonal). Для single-file скриптов или утилит без архитектурного деления — верни пустой ответ.

## Runtime Detection

До анализа определи runtime проекта:
```bash
cat package.json 2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print('Node.js:', list(d.get('dependencies',{}).keys())[:8])" 2>/dev/null || \
ls go.mod requirements.txt pyproject.toml Cargo.toml 2>/dev/null | head -3
```

⚠️ Этот чеклист оптимизирован для **Node.js/TypeScript**. При обнаружении другого runtime:
- Go → `context.Context` вместо `AbortSignal`, `SIGTERM handler` вместо `process.on`
- Python → `asyncio cancellation`, `signal.SIGTERM`
- Java/Spring → `@Transactional`, `ApplicationContext lifecycle`
- Для неизвестного runtime — JS-специфичные проверки помечай `🔍 UNVERIFIED`

## Severity Guide

| Severity | Критерий назначения |
|----------|---------------------|
| 🔴 Critical | RCE, auth bypass, data corruption, необратимый финансовый риск |
| 🟠 High | Падение production, privilege escalation, утечка данных |
| 🟡 Medium | Деградация производительности или поддерживаемости без immediate outage |
| 🟢 Low | Стиль, читаемость, слабое нарушение конвенции |

Правило: severity = impact × exploitability × blast radius. Одинаковый паттерн → одинаковый severity между аудитами.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| ARC-01 | Бизнес-логика вынесена из route handlers в service/domain слой |
| ARC-02 | Presentation layer не взаимодействует с БД напрямую |
| ARC-03 | Нет circular dependencies между модулями |
| ARC-04 | Нет god-объектов: файлы и классы имеют единственную ответственность |
| ARC-05 | Конфигурация и env-переменные изолированы в config-модуле |
| ARC-06 | Внешние зависимости инжектируются (DI), не импортируются напрямую |
| ARC-07 | Доменный слой не импортирует инфраструктурные модули |

## Правила верификации

1. **Только чеклист**: оценивай ТОЛЬКО проверки выше. Не добавляй новые.
2. **Явная верификация = PASS**: ставь `✅ PASS` только если явно проверил механизм (нашёл схему, конфиг, guard) и подтвердил отсутствие нарушения — укажи что именно проверено.
3. **Нет доказательства = UNVERIFIED**: не можешь указать `файл:строка` ни для нарушения, ни для подтверждения — ставь `🔍 UNVERIFIED`.
4. **Baseline приоритетен**: check_id есть в `docs/audit-baseline.yml` → `⏸ ACCEPTED`.
5. **Только 🔴/🟠 FAIL требуют решения**: 🟡/🟢 — решение необязательно.

## Evidence Quality Rules

Любой `❌ FAIL` обязан содержать:
- Точный `file:line`
- Минимальный код-фрагмент (1–3 строки)
- Causal chain: почему именно это нарушение → какой риск возникает

Запрещено:
- Предполагать runtime behavior без evidence в коде
- Предполагать prod-конфигурацию по dev-конфигу
- Предполагать отсутствие middleware без проверки всей router chain
- Если вывод основан на предположении — только `🔍 UNVERIFIED`

## Language Rule

Результаты аудита должны быть написаны простым и понятным языком. Избегай сложных терминов, жаргона и абстрактных понятий без необходимости. Общепринятые технические термины (Docker, HTTP, API, JSON, URL) допустимы. Описывай проблемы так, чтобы они были понятны разработчику любого уровня, а не только узкому специалисту в данной области.

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

**ARC-01 — Бизнес-логика в service/domain слое:**
- Бизнес-правила и вычисления прямо в route handler / controller
- Сложные условия и трансформации данных в middleware вместо сервиса
- Операции с несколькими сущностями выполняются в handler без сервиса

**ARC-02 — Presentation layer без прямых DB-вызовов:**
- Прямые обращения к ORM/query builder из роутеров/контроллеров
- Импорт репозиториев или DB-клиента непосредственно в слой представления
- SQL / Prisma / Mongoose вызовы в middleware

**ARC-03 — Нет circular dependencies:**
- Модуль A импортирует модуль B, который импортирует модуль A
- Circular deps через несколько уровней (A→B→C→A)
- Прямые импорты между feature-модулями (должны идти через shared)

**ARC-04 — Нет god-объектов:**
- Файлы >500 строк с несвязанной логикой (несколько разных ответственностей)
- Классы с методами из разных доменных областей
- Модуль делает несвязанные вещи (low cohesion)

**ARC-05 — Конфигурация изолирована:**
- `process.env.XYZ` используется напрямую вне config-модуля
- Магические строки с именами env-переменных разбросаны по коду
- Нет единой точки валидации конфигурации при старте приложения

**ARC-06 — Внешние зависимости инжектируются:**
- HTTP-клиенты, email-сервисы, БД-клиенты создаются внутри функций (не инжектируются)
- Зависимость от конкретных реализаций вместо интерфейсов/абстракций
- Отсутствие dependency injection делает код нетестируемым (нельзя подменить в тестах)

**Направление зависимостей (Dependency Rule):**
- Domain/service слой импортирует типы Prisma напрямую (должен через repository interface)
- Бизнес-логика зависит от Express Request/Response типов
- Domain entity содержит ORM-декораторы (TypeORM @Entity в domain классе)
- Нарушение: `domain/ → infrastructure/` (правильно: `infrastructure/ → domain/`)

## Инструментальная поддержка

Перед анализом выполни (если инструменты установлены):
```bash
npx steiger ./src 2>/dev/null || true   # FSD layer violations
npx dependency-cruiser --validate .dependency-cruiser.js src 2>/dev/null || true
```
Используй вывод как подсказку, верифицируй находки вручную.

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение | Исправлено |
|----------|----------|--------|-------------|----------------|---------|------------|
| ARC-01 | Бизнес-логика вынесена из route handlers в service/domain слой | ✅ PASS | High | `routes/` — handlers делегируют в services | — | — |
| ARC-03 | Нет circular dependencies между модулями | ❌ FAIL 🟠 | High | `modules/order/index.ts:3` | **1. Вынести общий код в shared модуль** \\ 2. Применить dependency inversion \\ 3. Разбить модуль на независимые части | Нет |
| ARC-06 | Внешние зависимости инжектируются (DI), не импортируются напрямую | ⏸ ACCEPTED | Medium | `services/payment.ts:1` | В baseline: refactor запланирован в Q3 | — |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED` / `🔍 UNVERIFIED`

Уверенность: `High` — проверил несколько ключевых файлов, паттерн очевиден / `Medium` — проверил выборочно, паттерн вероятен / `Low` — ограниченный контекст, полная уверенность невозможна

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

`Исправлено`: FAIL → `Нет` (разработчик меняет на `✅ Да` вручную после фикса). PASS / ACCEPTED / UNVERIFIED → `—`.

Требования к решениям:
- Взаимно исключающие (не перефразировки одного и того же)
- Соответствуют текущему стеку проекта (не предлагать смену фреймворка)
- Не требуют переписать всю систему — realistic migration cost
- Вариант 3 может быть «оставить, задокументировать причину» при наличии обоснования

В конце отчёта добавь раздел покрытия:
```
## Audit Coverage
Проверено: src/module1/**, src/module2/**
Пропущено: scripts/**, migrations/**, tests/**
Файлов проверено: N | Пропущено: N
```

Если все PASS — выведи: `✅ Архитектурные принципы соблюдены.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-architecture.md`

```
# Audit Report: Architecture & File Structure — <YYYY-MM-DD HH:MM>
<таблица>
```
