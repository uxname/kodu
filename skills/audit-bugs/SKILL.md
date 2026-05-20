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
| BUG-01 | Number()/parseInt() проверяются на NaN |
| BUG-02 | Нет await внутри forEach |
| BUG-03 | Нет обращений к свойствам без null-проверки |
| BUG-04 | Array.sort() не мутирует входной аргумент |
| BUG-05 | parseInt всегда с radix 10 |
| BUG-06 | Нет if(asyncFn()) без await |
| BUG-07 | switch на enum имеет default |
| BUG-08 | Нет деления на ноль без guard |

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

**Неверные условия:**
- Инвертированные флаги (`!isValid` там где нужен `isValid`)
- Неправильные операторы сравнения (`>=` вместо `>`)
- Неверная булева логика (`&&` вместо `||` и наоборот)

**Null / undefined dereference:**
- Обращение к свойству без проверки на null/undefined
- Опциональная цепочка пропущена (`obj.a.b` там где `obj.a` может быть undefined)
- Деструктуризация без дефолтного значения при возможном undefined

**Забытый await:**
- Async-функция вызвана без `await` — возвращается Promise вместо значения
- `await` внутри `forEach` (forEach не ждёт промисов)
- Условие `if (asyncFn())` вместо `if (await asyncFn())`

**Type coercion и сравнение типов:**
- `parseInt` без явного основания системы счисления
- Конкатенация числа со строкой вместо сложения

**Мутация аргументов:**
- Функция изменяет переданный объект/массив вместо создания копии
- `Array.sort()` на переданном массиве без предварительного `.slice()`

**Деление и математика:**
- Деление на ноль без guard-проверки
- `Number()` / `parseInt()` без проверки на NaN

**Switch / enum exhaustiveness:**
- `switch` без `default` на enum-значении (новое значение enum пройдёт незамеченным)

## Граница с другими аудитами

- **Обработка ошибок** (timeout, retry, circuit breaker) → `audit-errors`
- **Race conditions, транзакции** → `audit-concurrency`
- **Валидация входных данных** → `audit-validation`
- **Безопасность (injection, XSS)** → `audit-owasp`

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| BUG-01 | Number()/parseInt() проверяются на NaN | ✅ PASS | — | — |
| BUG-03 | Нет обращений к свойствам без null-проверки | ❌ FAIL 🟠 | `services/order.ts:55` | **1. Добавить опциональную цепочку: `user?.address?.city`** \\ 2. Добавить guard-проверку перед обращением \\ 3. Использовать nullish coalescing с дефолтом |
| BUG-07 | switch на enum имеет default | ⏸ ACCEPTED | `handlers/event.ts:23` | В baseline: exhaustive check через TypeScript never |

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
