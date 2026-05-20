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
| CON-01 | async/await не используется в неасинхронных итераторах (forEach, map) |
| CON-02 | Read-modify-write операции выполняются в транзакциях |
| CON-03 | Нет shared mutable state на уровне модуля (синглтоны, кэши без locks) |
| CON-04 | Module-level кэш имеет механизм инвалидации |
| CON-05 | Обработчики событий и webhook-handlers идемпотентны |

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

**CON-01 — Корректное использование async в итераторах:**
- `await` внутри `forEach` — forEach не ждёт промисов, итерация выполняется некорректно
- `async` функция в `Array.map` без `Promise.all` — промисы создаются но не ожидаются
- Shared state между параллельными async операциями без синхронизации

**CON-02 — Read-modify-write в транзакциях:**
- SELECT + UPDATE без транзакции (TOCTOU — time-of-check to time-of-use)
- Двойное списание/начисление без транзакции с блокировкой
- Check-then-act без атомарности (читаем → проверяем → пишем без lock)
- Optimistic locking без retry при конфликте версий
- Кэш-инвалидация между чтением и записью

**CON-03 — Нет shared mutable state на уровне модуля:**
- Глобальные переменные, изменяемые из нескольких мест
- Синглтоны с mutable state без синхронизации
- Closure над изменяемой переменной в async callback
- Конкурентная запись в один файл/ресурс без координации

**CON-04 — Module-level кэш инвалидируется:**
- Module-level кэш без механизма инвалидации при обновлении данных
- Кэш без TTL (stale data не обновляется никогда)
- Нет стратегии обновления кэша при изменении исходных данных

**CON-05 — Идемпотентность обработчиков:**
- Обработчики событий/сообщений без идемпотентности (повторная доставка не безопасна)
- Нет защиты от дублирующихся webhook-вызовов (нет проверки event_id)
- Финансовые или критические операции без идемпотентного ключа

## Граница с другими аудитами

- **Timeout, retry, circuit breaker, проглоченные ошибки** — зона `audit-errors`
- **Логические баги в условиях** — зона `audit-bugs`

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| CON-01 | async/await не используется в неасинхронных итераторах (forEach, map) | ✅ PASS | — | — |
| CON-02 | Read-modify-write операции выполняются в транзакциях | ❌ FAIL 🔴 | `services/wallet.ts:67` | **1. Обернуть в db.transaction() с SELECT FOR UPDATE** \\ 2. Использовать optimistic locking с retry \\ 3. Добавить уникальное ограничение на уровне БД |
| CON-05 | Обработчики событий и webhook-handlers идемпотентны | ⏸ ACCEPTED | `handlers/stripe.ts:12` | В baseline: идемпотентность обеспечена через event_id |

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
