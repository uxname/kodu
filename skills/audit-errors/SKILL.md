---
name: audit-errors
description: >
  Аудит обработки ошибок и отказоустойчивости: исключения, таймауты, retry policies,
  circuit breakers, graceful degradation. Запускай при /audit-errors.
---

## Правило применимости (Relevance Rule)

Применим к коду с внешними вызовами (HTTP, БД, очереди, файловая система), асинхронному коду, обработчикам событий. Для синхронных утилит без I/O — верни пустой ответ.

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
| ERR-01 | Ошибки не проглатываются — catch-блоки обрабатывают или пробрасывают |
| ERR-02 | Внутренние детали (stack trace, пути, версии) не попадают в ответы |
| ERR-03 | Async handlers корректно пробрасывают исключения в error middleware |
| ERR-04 | Unhandled rejections и uncaught exceptions имеют process-level обработчики |
| ERR-05 | Внешние вызовы (HTTP-клиенты, DB) имеют явные таймауты |
| ERR-06 | Graceful shutdown реализован — SIGTERM обрабатывается |
| ERR-07 | Error responses консистентны по структуре во всём приложении |
| ERR-08 | Retry-стратегии используют exponential backoff с jitter |
| ERR-09 | AbortSignal/CancellationToken пробрасывается во внешние вызовы [⚡ dynamic] |

## Правила верификации

1. **Только чеклист**: оценивай ТОЛЬКО проверки выше. Не добавляй новые.
2. **Явная верификация = PASS**: ставь `✅ PASS` только если явно проверил механизм (нашёл схему, конфиг, guard) и подтвердил отсутствие нарушения — укажи что именно проверено.
3. **Нет доказательства = UNVERIFIED**: не можешь указать `файл:строка` ни для нарушения, ни для подтверждения — ставь `🔍 UNVERIFIED`.
   - Проверки с `[⚡ dynamic]` нельзя статически подтвердить — только `🔍 UNVERIFIED` или `❌ FAIL` (при явном evidence), но не `✅ PASS`
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

**ERR-01 — Ошибки не проглатываются:**
- Пустые catch-блоки (`catch(e) {}`)
- `catch` только с логом без восстановления и re-throw
- Promise без `.catch()` или `try/await` без `catch`
- Unhandled promise rejections без обработки

**ERR-02 — Внутренние детали не в ответах:**
- Stack trace в production API responses
- Внутренние пути файловой системы в error messages
- Версии зависимостей/фреймворка в заголовках или ответах
- DB-специфичные сообщения об ошибках (SQL syntax error) в API responses

**ERR-03 — Async handlers пробрасывают исключения:**
- Express async handlers без asyncHandler wrapper или Express 5
- Promise rejection в middleware не пробрасывается в error middleware
- Необработанные исключения в setTimeout/setInterval колбэках

**ERR-04 — Process-level обработчики:**
- Отсутствие `process.on('unhandledRejection')` 
- Отсутствие `process.on('uncaughtException')`
- Нет логирования и корректного выхода при критических ошибках процесса

**ERR-05 — Явные таймауты для внешних вызовов:**
- HTTP-клиенты (axios, fetch, got) без явного timeout
- БД-запросы без query timeout / statement timeout
- Отсутствие timeout для очередей сообщений и внешних gRPC вызовов
- Бесконечные retry без exponential backoff и max attempts

**ERR-06 — Graceful shutdown:**
- Отсутствие `process.on('SIGTERM')` обработчика
- Нет закрытия DB-пула и HTTP-сервера при shutdown
- Незавершённые запросы не дожидаются окончания при shutdown

**ERR-07 — Консистентная структура error responses:**
- Разный формат ошибок в разных endpoint'ах (нет единого error shape)
- Отсутствие machine-readable error code (только human-readable message)
- HTTP статус 200 при ошибке (должен быть 4xx/5xx)

**Retry & Cancellation:**
- HTTP-retry без задержки или с фиксированной задержкой (нет exponential backoff)
- Отсутствие jitter — все ретраи синхронизируются при массовом сбое
- fetch/axios без AbortSignal — висящие запросы после disconnect клиента
- Цепочка обработчиков не передаёт AbortController вглубь

## Граница с другими аудитами

- **Stack trace в ответах** — этот скилл первичный (ERR-02). `audit-owasp` и `audit-api-contracts` ссылаются сюда.
- **Идемпотентность handlers** — первичный: `audit-concurrency` (CON-05). Здесь не дублируй.
- **Таймауты HTTP-клиентов** — ERR-05 первичный. `audit-performance` не дублирует.

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение | Исправлено |
|----------|----------|--------|-------------|----------------|---------|------------|
| ERR-01 | Ошибки не проглатываются — catch-блоки обрабатывают или пробрасывают | ✅ PASS | High | `src/` — все catch-блоки логируют или пробрасывают | — | — |
| ERR-05 | Внешние вызовы (HTTP-клиенты, DB) имеют явные таймауты | ❌ FAIL 🟠 | High | `services/api.ts:18` | **1. Добавить timeout в axios: `{ timeout: 5000 }`** \\ 2. Использовать AbortController с setTimeout \\ 3. Установить глобальный default timeout | Нет |
| ERR-06 | Graceful shutdown реализован — SIGTERM обрабатывается | ⏸ ACCEPTED | Medium | `server.ts:5` | В baseline: управляется оркестратором | — |

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

Если все PASS — выведи: `✅ Обработка ошибок реализована корректно.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-errors.md`

```
# Audit Report: Error Handling & Resiliency — <YYYY-MM-DD HH:MM>
<таблица>
```
