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
| PER-01 | Нет DB запросов внутри циклов (N+1) |
| PER-02 | SELECT запросы имеют LIMIT/пагинацию |
| PER-03 | Нет blocking I/O (fs.readFileSync) в request handlers |
| PER-04 | CPU-intensive операции не в main thread |
| PER-05 | Независимые async операции параллельны (Promise.all) |
| PER-06 | Кэши имеют TTL и size limit |
| PER-07 | Event listeners удаляются при cleanup |

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

**N+1 запросы:**
- Запрос к БД внутри цикла по результатам другого запроса
- ORM relations загружаются lazy внутри loop
- Отсутствие `include`/`join` где это возможно

**Тяжёлые запросы:**
- SELECT без LIMIT/пагинации (возможный full table scan)
- Запросы без WHERE на индексированных полях
- Агрегации на больших таблицах без materialized view / кэша

**Blocking I/O:**
- Синхронные файловые операции в async контексте (`fs.readFileSync`)
- `sleep`/`Thread.sleep` в request handler
- CPU-intensive операции (crypto, image processing) в main thread без worker

**Memory:**
- Накопление данных в memory без flush (buffer без drain)
- Event listeners без removeEventListener (leak в long-lived процессах)
- Кэш без TTL и без size limit (unbounded growth)

**Сетевые запросы:**
- Последовательные независимые HTTP-запросы вместо параллельных
- Повторные запросы к одному URL без мемоизации в рамках запроса

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| PER-01 | Нет DB запросов внутри циклов (N+1) | ✅ PASS | — | — |
| PER-02 | SELECT запросы имеют LIMIT/пагинацию | ❌ FAIL 🟠 | `repos/user.ts:45` | **1. Добавить .take(limit).skip(offset) в запрос** \\ 2. Добавить cursor-based пагинацию \\ 3. Установить максимальный лимит через конфиг |
| PER-06 | Кэши имеют TTL и size limit | ⏸ ACCEPTED | `cache/store.ts:12` | В baseline: кэш управляется Redis с TTL |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

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
