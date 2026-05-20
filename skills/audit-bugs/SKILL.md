---
name: audit-bugs
description: >
  Аудит прямых багов и логических ошибок: неверные условия, off-by-one, null dereference,
  забытый await, unreachable code, type coercion, мутация аргументов. Запускай при /audit-bugs.
---

## Правило применимости (Relevance Rule)

Применим к любому коду с бизнес-логикой. Пропускай только автогенерированные файлы (миграции, protobuf-generated, build output) и чисто декларативные конфиги (JSON, YAML без логики).

## Чеклист

| Check ID | Проверка |
|----------|----------|
| BUG-01 | Преобразования типов безопасны (NaN, radix, coercion) |
| BUG-02 | async/await используется корректно (нет await в forEach, нет if(asyncFn())) |
| BUG-03 | Null-safety соблюдается — обращения к свойствам защищены от undefined/null |
| BUG-04 | Функции не мутируют входные аргументы (sort, splice, object spread) |
| BUG-05 | Exhaustive handling — все enum/union-ветки обработаны |
| BUG-06 | Математические guard-условия (деление на ноль, граничные значения) |

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

## Граница с другими аудитами

- **Обработка ошибок** (timeout, retry, circuit breaker) → `audit-errors`
- **Race conditions, транзакции** → `audit-concurrency`
- **Валидация входных данных** → `audit-validation`
- **Безопасность (injection, XSS)** → `audit-owasp`

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| BUG-01 | Преобразования типов безопасны (NaN, radix, coercion) | ✅ PASS | — | — |
| BUG-03 | Null-safety соблюдается — обращения к свойствам защищены от undefined/null | ❌ FAIL 🟠 | `services/order.ts:55` | **1. Добавить опциональную цепочку: `user?.address?.city`** \\ 2. Добавить guard-проверку перед обращением \\ 3. Использовать nullish coalescing с дефолтом |
| BUG-05 | Exhaustive handling — все enum/union-ветки обработаны | ⏸ ACCEPTED | `handlers/event.ts:23` | В baseline: exhaustive check через TypeScript never |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

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
