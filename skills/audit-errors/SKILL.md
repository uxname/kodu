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
| ERR-01 | Нет пустых catch блоков |
| ERR-02 | Stack trace не в production response |
| ERR-03 | Async handlers в asyncHandler или Express 5 |
| ERR-04 | Есть process.on('unhandledRejection') |
| ERR-05 | HTTP клиенты имеют явный timeout |
| ERR-06 | Graceful shutdown реализован (SIGTERM) |
| ERR-07 | Error responses имеют консистентную структуру |
| ERR-08 | DB запросы имеют timeout |

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

**Проглоченные ошибки:**
- Пустые catch-блоки (`catch(e) {}`)
- `catch` только с логом без восстановления и re-throw
- Promise без `.catch()` или `try/await` без `catch`
- Unhandled promise rejections

**Таймауты:**
- HTTP-клиенты без явного timeout
- БД-запросы без query timeout
- Отсутствие timeout для очередей сообщений
- Бесконечные retry без exponential backoff и max attempts

**Retry policies:**
- Retry без различия transient vs permanent ошибок
- Retry на non-idempotent операции (POST без идемпотентного ключа)
- Отсутствие jitter в retry (thundering herd проблема)

**Graceful degradation:**
- Падение всего сервиса при недоступности некритичного зависимости
- Отсутствие circuit breaker для внешних HTTP-зависимостей
- Нет fallback для кэша при недоступности Redis

## Граница с другими аудитами

- **null/undefined dereference, array bounds, деление на ноль** — логические баги, зона `audit-bugs`
- **Race conditions, транзакции БД** — зона `audit-concurrency`

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| ERR-01 | Нет пустых catch блоков | ✅ PASS | — | — |
| ERR-05 | HTTP клиенты имеют явный timeout | ❌ FAIL 🟠 | `services/api.ts:18` | **1. Добавить timeout в axios: `{ timeout: 5000 }`** \\ 2. Использовать AbortController с setTimeout \\ 3. Установить глобальный default timeout |
| ERR-06 | Graceful shutdown реализован (SIGTERM) | ⏸ ACCEPTED | `server.ts:5` | В baseline: управляется оркестратором |

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
