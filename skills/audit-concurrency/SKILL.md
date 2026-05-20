---
name: audit-concurrency
description: >
  Аудит управления состоянием и конкурентности: race conditions, deadlocks, shared mutable state,
  неатомарные операции. Запускай при /audit-concurrency.
---

## Правило применимости (Relevance Rule)

Применим к коду с параллельными операциями, shared state, кэшированием, транзакциями БД, очередями, WebSocket. Для однопоточных скриптов без параллелизма — верни пустой ответ.

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
| CON-01 | async/await не используется в неасинхронных итераторах (forEach, map) |
| CON-02 | Read-modify-write операции выполняются в транзакциях [⚡ dynamic] |
| CON-03 | Нет shared mutable state на уровне модуля (синглтоны, кэши без locks) |
| CON-04 | Module-level кэш имеет механизм инвалидации |
| CON-05 | Обработчики событий и webhook-handlers идемпотентны [⚡ dynamic] |
| CON-06 | Background async операции имеют механизм отмены (AbortController/signal) и не блокируют graceful shutdown |

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

**CON-06 — Background операции с механизмом отмены:**
- Promise-цепочка запущена в фоне без `AbortController` / `signal` — при SIGTERM процесс не завершается чисто
- `setInterval` / `setImmediate` внутри request handler без очистки — утечка при завершении запроса
- Background job без timeout и без механизма принудительной остановки
- Graceful shutdown не ожидает завершения фоновых задач (нет `await backgroundJob`)
- `Promise.all` с неотменяемыми задачами — при ошибке одной остальные продолжают выполняться

## Граница с другими аудитами

- **Идемпотентность** — этот скилл первичный (CON-05). `audit-errors` ссылается сюда.
- **async/await в forEach** — первичный: `audit-bugs` (BUG-02). `audit-concurrency` (CON-01) фокусируется на параллелизме, а не на синтаксической ошибке.

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение | Исправлено |
|----------|----------|--------|-------------|----------------|---------|------------|
| CON-01 | async/await не используется в неасинхронных итераторах (forEach, map) | ✅ PASS | High | `src/` — async forEach не найден | — | — |
| CON-02 | Read-modify-write операции выполняются в транзакциях | ❌ FAIL 🔴 | High | `services/wallet.ts:67` | **1. Обернуть в db.transaction() с SELECT FOR UPDATE** \\ 2. Использовать optimistic locking с retry \\ 3. Добавить уникальное ограничение на уровне БД [⚡ dynamic] | Нет |
| CON-05 | Обработчики событий и webhook-handlers идемпотентны | ⏸ ACCEPTED | Medium | `handlers/stripe.ts:12` | В baseline: идемпотентность обеспечена через event_id [⚡ dynamic] | — |

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
