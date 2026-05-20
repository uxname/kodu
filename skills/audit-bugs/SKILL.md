---
name: audit-bugs
description: >
  Аудит прямых багов и логических ошибок: неверные условия, off-by-one, null dereference,
  забытый await, unreachable code, type coercion, мутация аргументов. Запускай при /audit-bugs.
---

## Правило применимости (Relevance Rule)

Применим к любому коду с бизнес-логикой. Пропускай только автогенерированные файлы (миграции, protobuf-generated, build output) и чисто декларативные конфиги (JSON, YAML без логики).

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
| BUG-01 | Преобразования типов безопасны (NaN, radix, coercion) |
| BUG-02 | async/await используется корректно (нет await в forEach, нет if(asyncFn())) |
| BUG-03 | Null-safety соблюдается — обращения к свойствам защищены от undefined/null |
| BUG-04 | Функции не мутируют входные аргументы (sort, splice, object spread) |
| BUG-05 | Exhaustive handling — все enum/union-ветки обработаны |
| BUG-06 | Математические guard-условия (деление на ноль, граничные значения) |
| BUG-07 | Off-by-one: границы диапазонов корректны (`<` vs `<=`, индексы массивов, slice) |
| BUG-08 | Float comparison не использует `===` для проверки равенства (используется epsilon или toFixed) |
| BUG-09 | Дата/время хранятся и обрабатываются в UTC, не локальном времени |
| BUG-10 | RegExp с user input не содержит катастрофического backtracking (ReDoS) |

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

**BUG-01 — Преобразования типов безопасны:**
- `parseInt` без явного основания системы счисления (radix)
- `Number()` / `parseInt()` без проверки результата на NaN
- Неявное приведение типов: конкатенация числа со строкой вместо сложения
- Сравнение с `==` вместо `===` приводит к неожиданному coercion

**BUG-02 — async/await используется корректно:**
- `await` внутри `forEach` — forEach не ждёт промисов, итерация не последовательна
- Условие `if (asyncFn())` вместо `if (await asyncFn())` — всегда truthy (Promise объект)
- Async-функция вызвана без `await` — возвращается Promise вместо значения
- `Promise` без `.catch()` или `try/await` без `catch`

**BUG-03 — Null-safety соблюдается:**
- Обращение к свойству без проверки на null/undefined
- Опциональная цепочка пропущена (`obj.a.b` там где `obj.a` может быть undefined)
- Деструктуризация без дефолтного значения при возможном undefined

**BUG-04 — Функции не мутируют входные аргументы:**
- `Array.sort()` на переданном массиве без предварительного `.slice()`
- `Array.splice()` изменяет оригинальный массив-аргумент
- Прямое присваивание свойств объекта-аргумента вместо создания копии через spread

**BUG-05 — Exhaustive handling:**
- `switch` без `default` на enum-значении — новое значение enum пройдёт незамеченным
- Union type без обработки всех вариантов в if/else цепочке
- Отсутствие never-проверки для исчерпывающего TypeScript switch

**BUG-06 — Математические guard-условия:**
- Деление на ноль без guard-проверки знаменателя
- `Number()` / `parseInt()` без проверки результата на NaN перед использованием
- Граничные значения не проверяются (отрицательный индекс, пустой массив)

**BUG-07 — Off-by-one:**
- `for (i = 0; i <= arr.length; i++)` — выход за границы массива
- `slice(0, n)` / `slice(0, n-1)` — потеря или дублирование последнего элемента
- Сравнение `index < arr.length - 1` там где нужно `index < arr.length`
- Пагинация: `offset * limit` vs `(offset - 1) * limit` при 1-based страницах

**BUG-08 — Float comparison:**
- `a === b` где a и b — результаты float-арифметики (0.1 + 0.2 !== 0.3)
- Сравнение денежных сумм через float вместо integer cents / BigDecimal
- Условие `if (parseFloat(x) === 1.0)` вместо `Math.abs(x - 1.0) < epsilon`

**BUG-09 — Дата/время в UTC:**
- `new Date()` без явной UTC-обработки — результат зависит от timezone сервера
- Сравнение дат через `toLocaleDateString()` — зависит от локали
- Хранение timestamp как local datetime строки вместо ISO 8601 с `Z`
- `new Date('2024-01-01')` парсится как UTC, `new Date('2024-01-01 00:00')` — как local

**BUG-10 — ReDoS:**
- `new RegExp(userInput)` — пользователь контролирует регулярное выражение
- Вложенные quantifiers на перекрывающихся классах: `(a+)+`, `(a*)*`, `([a-zA-Z]+\s?)+`
- Проверка: `'a'.repeat(50000) + 'x'` на подозрительном regex вешает event loop
- Особо опасно в middleware (auth, routing), где regex применяется к каждому запросу

## Граница с другими аудитами

- **Обработка ошибок** (timeout, retry, circuit breaker) → `audit-errors`
- **Race conditions, транзакции** → `audit-concurrency`
- **Валидация входных данных** → `audit-validation`
- **Безопасность (injection, XSS)** → `audit-owasp`

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение | Исправлено |
|----------|----------|--------|-------------|----------------|---------|------------|
| BUG-01 | Преобразования типов безопасны (NaN, radix, coercion) | ✅ PASS | High | `src/` — parseInt всегда с radix | — | — |
| BUG-03 | Null-safety соблюдается — обращения к свойствам защищены от undefined/null | ❌ FAIL 🟠 | High | `services/order.ts:55` | **1. Добавить опциональную цепочку: `user?.address?.city`** \\ 2. Добавить guard-проверку перед обращением \\ 3. Использовать nullish coalescing с дефолтом | Нет |
| BUG-05 | Exhaustive handling — все enum/union-ветки обработаны | ⏸ ACCEPTED | Medium | `handlers/event.ts:23` | В baseline: exhaustive check через TypeScript never | — |

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

Если все PASS — выведи: `✅ Явных логических ошибок не обнаружено.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-bugs.md`

```
# Audit Report: Bugs & Logic Errors — <YYYY-MM-DD HH:MM>
<таблица>
```
