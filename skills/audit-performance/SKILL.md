---
name: audit-performance
description: >
  Аудит ресурсов и производительности: блокирующие вызовы, N+1 запросы, тяжёлые запросы
  без лимитов, memory leaks, отсутствие пагинации. Запускай при /audit-performance.
---

## Правило применимости (Relevance Rule)

Применим к коду с I/O операциями (БД, HTTP, файлы), обработкой коллекций, кэшированием. Для stateless математических утилит без I/O — верни пустой ответ.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| PER-01 | Нет N+1: DB-запросы не выполняются внутри циклов |
| PER-02 | Выборки из БД ограничены (LIMIT, пагинация) |
| PER-03 | Обработчики запросов не содержат блокирующего I/O |
| PER-04 | CPU-интенсивные операции вынесены из main thread |
| PER-05 | Независимые async-операции выполняются параллельно |
| PER-06 | Кэши ограничены по размеру и времени жизни (TTL + size limit) |
| PER-07 | Event listeners и subscriptions очищаются при завершении |

## Правила верификации

1. **Только чеклист**: оценивай ТОЛЬКО проверки выше. Не добавляй новые.
2. **Явная верификация = PASS**: ставь `✅ PASS` только если явно проверил механизм (нашёл схему, конфиг, guard) и подтвердил отсутствие нарушения — укажи что именно проверено.
3. **Нет доказательства = UNVERIFIED**: не можешь указать `файл:строка` ни для нарушения, ни для подтверждения — ставь `🔍 UNVERIFIED`.
4. **Baseline приоритетен**: check_id есть в `docs/audit-baseline.yml` → `⏸ ACCEPTED`.
5. **Только 🔴/🟠 FAIL требуют решения**: 🟡/🟢 — решение необязательно.

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

**PER-01 — Нет N+1:**
- Запрос к БД внутри цикла по результатам другого запроса
- ORM relations загружаются lazy внутри loop
- Отсутствие `include`/`join` там где возможна одна выборка
- DataLoader / batch loading не используется при множественных точечных запросах

**PER-02 — Выборки ограничены:**
- SELECT без LIMIT/пагинации (возможный full table scan)
- Запросы без WHERE на индексированных полях
- Агрегации на больших таблицах без materialized view / кэша
- Нет cursor-based пагинации при больших наборах данных

**PER-03 — Нет блокирующего I/O в handlers:**
- Синхронные файловые операции в async контексте (`fs.readFileSync`)
- `sleep`/busy-wait в request handler
- Синхронные операции с большими буферами блокирующие event loop

**PER-04 — CPU-операции вне main thread:**
- CPU-intensive операции (crypto, image processing, compression) в main thread без worker
- Тяжёлые вычисления (сортировка больших массивов, regex на длинных строках) в request path

**PER-05 — Параллельные независимые операции:**
- Последовательные независимые HTTP-запросы вместо `Promise.all`
- Sequential await там где возможен parallel fetch
- Повторные запросы к одному URL без мемоизации в рамках одного запроса

**PER-06 — Кэши ограничены:**
- Кэш без TTL (unbounded growth, stale data навсегда)
- Кэш без size limit (memory leak в long-lived процессах)
- In-memory кэш без механизма инвалидации при обновлении данных

**PER-07 — Event listeners очищаются:**
- Event listeners без `removeEventListener` / `off` (leak в long-lived процессах)
- RxJS subscriptions без `unsubscribe` в destroy/cleanup
- WebSocket / SSE connections без cleanup при завершении request lifecycle
- Накопление данных в memory без flush (buffer без drain)

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение |
|----------|----------|--------|-------------|----------------|---------|
| PER-01 | Нет N+1: DB-запросы не выполняются внутри циклов | ✅ PASS | High | `repos/` проверены — запросы вне циклов | — |
| PER-02 | Выборки из БД ограничены (LIMIT, пагинация) | ❌ FAIL 🟠 | High | `repos/user.ts:45` | **1. Добавить .take(limit).skip(offset) в запрос** \\ 2. Добавить cursor-based пагинацию \\ 3. Установить максимальный лимит через конфиг |
| PER-06 | Кэши ограничены по размеру и времени жизни (TTL + size limit) | ⏸ ACCEPTED | Medium | `cache/store.ts:12` | В baseline: кэш управляется Redis с TTL |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED` / `🔍 UNVERIFIED`

Уверенность: `High` — проверил несколько ключевых файлов, паттерн очевиден / `Medium` — проверил выборочно, паттерн вероятен / `Low` — ограниченный контекст, полная уверенность невозможна

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

Если все PASS — выведи: `✅ Производительных антипаттернов не обнаружено.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-performance.md`

```
# Audit Report: Resource & Performance — <YYYY-MM-DD HH:MM>
<таблица>
```
