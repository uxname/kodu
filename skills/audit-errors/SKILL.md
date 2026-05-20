---
name: audit-errors
description: >
  Аудит обработки ошибок и отказоустойчивости: исключения, таймауты, retry policies,
  circuit breakers, graceful degradation. Запускай при /audit-errors.
---

## Правило применимости (Relevance Rule)

Применим к коду с внешними вызовами (HTTP, БД, очереди, файловая система), асинхронному коду, обработчикам событий. Для синхронных утилит без I/O — верни пустой ответ.

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

## Граница с другими аудитами

- **null/undefined dereference, array bounds, деление на ноль** — логические баги, зона `audit-bugs`
- **Race conditions, транзакции БД** — зона `audit-concurrency`

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| ERR-01 | Ошибки не проглатываются — catch-блоки обрабатывают или пробрасывают | ✅ PASS | — | — |
| ERR-05 | Внешние вызовы (HTTP-клиенты, DB) имеют явные таймауты | ❌ FAIL 🟠 | `services/api.ts:18` | **1. Добавить timeout в axios: `{ timeout: 5000 }`** \\ 2. Использовать AbortController с setTimeout \\ 3. Установить глобальный default timeout |
| ERR-06 | Graceful shutdown реализован — SIGTERM обрабатывается | ⏸ ACCEPTED | `server.ts:5` | В baseline: управляется оркестратором |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

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
