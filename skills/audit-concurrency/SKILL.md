---
name: audit-concurrency
description: >
  Аудит управления состоянием и конкурентности: race conditions, deadlocks, shared mutable state,
  неатомарные операции. Запускай при /audit-concurrency.
---

## Правило применимости (Relevance Rule)

Применим к коду с параллельными операциями, shared state, кэшированием, транзакциями БД, очередями, WebSocket. Для однопоточных скриптов без параллелизма — верни пустой ответ.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| CON-01 | Нет await внутри forEach/map |
| CON-02 | DB read-modify-write операции в транзакциях |
| CON-03 | Нет shared mutable state на module level |
| CON-04 | Module-level кэш имеет механизм инвалидации |
| CON-05 | Webhook/event handlers идемпотентны |

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

**Race Conditions:**
- Check-then-act без атомарности (читаем → проверяем → пишем без lock)
- Двойное списание/начисление без транзакции
- Кэш-инвалидация между чтением и записью
- Конкурентные запись в один файл без координации

**Shared Mutable State:**
- Глобальные переменные, изменяемые из нескольких мест
- Синглтоны с mutable state без синхронизации
- Closure над изменяемой переменной в async callback
- Module-level кэш без TTL и без механизма инвалидации

**Транзакции БД:**
- SELECT + UPDATE без транзакции (TOCTOU)
- Транзакции с lock escalation risk
- Optimistic locking без retry при конфликте

**Async/Promise специфика (JS/TS):**
- `await` внутри `forEach` (не работает как ожидается)
- Параллельные Promise без `Promise.all`
- Shared state между параллельными async операциями

**Идемпотентность:**
- Обработчики событий/сообщений без идемпотентности
- Нет защиты от дублирующихся webhook-вызовов

## Граница с другими аудитами

- **Timeout, retry, circuit breaker, проглоченные ошибки** — зона `audit-errors`
- **Логические баги в условиях** — зона `audit-bugs`

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| CON-01 | Нет await внутри forEach/map | ✅ PASS | — | — |
| CON-02 | DB read-modify-write операции в транзакциях | ❌ FAIL 🔴 | `services/wallet.ts:67` | **1. Обернуть в db.transaction() с SELECT FOR UPDATE** \\ 2. Использовать optimistic locking с retry \\ 3. Добавить уникальное ограничение на уровне БД |
| CON-05 | Webhook/event handlers идемпотентны | ⏸ ACCEPTED | `handlers/stripe.ts:12` | В baseline: идемпотентность обеспечена через event_id |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

Если все PASS — выведи: `✅ Проблем с конкурентностью не обнаружено.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-concurrency.md`

```
# Audit Report: State & Concurrency — <YYYY-MM-DD HH:MM>
<таблица>
```
