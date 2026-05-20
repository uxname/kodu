---
name: audit-validation
description: >
  Аудит валидации граничных данных: проверка входящих данных на границах системы,
  отсутствие sanitization, trust boundary нарушения. Запускай при /audit-validation.
---

## Правило применимости (Relevance Rule)

Применим к коду, принимающему внешние данные: HTTP handlers, WebSocket, CLI args, file parsers, event consumers, gRPC endpoints. Для чисто внутреннего кода без внешних входов — верни пустой ответ.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| VAL-01 | Все входящие данные (body, params, query) проходят schema-валидацию |
| VAL-02 | Строки имеют maxLength, числа — диапазон, enum-значения — whitelist |
| VAL-03 | JSON.parse обёрнут в try/catch с последующей валидацией структуры |
| VAL-04 | Identity данные берутся из аутентифицированного контекста (не из user input) |
| VAL-05 | Вложенные структуры и массивы ограничены (глубина, minItems/maxItems) |
| VAL-06 | Валидатор не выполняет неявный coercion (строка "false" → boolean true) |

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

**VAL-01 — Все входящие данные проходят schema-валидацию:**
- HTTP request body используется напрямую без schema validation (zod/joi/yup/etc)
- Query params / path params используются без типизации и проверки
- Отсутствие проверки обязательных полей
- Нет валидации типов (строка может прийти вместо числа)
- WebSocket, CLI args, event payloads без валидации входных данных

**VAL-02 — Строки, числа, enum ограничены:**
- Отсутствие maxLength для строк (DoS через огромную строку)
- Нет проверки диапазонов чисел (отрицательные ID, огромные offset, NaN)
- Отсутствие whitelist для enum-полей (принимается любое строковое значение)
- Нет проверки формата (email, UUID, дата) там где она применима

**VAL-03 — JSON.parse с защитой:**
- JSON.parse без try/catch — выбросит SyntaxError при невалидном input
- JSON.parse без последующей валидации структуры (тип полей не проверен)
- Доверие структуре распарсенного JSON без schema-проверки

**VAL-04 — Identity из аутентифицированного контекста:**
- JWT claims используются без верификации подписи
- User ID берётся из тела запроса вместо `req.user` / аутентифицированного контекста
- Данные о ролях/правах берутся из user-controlled input
- Массовое присваивание: объект из body напрямую сохраняется в БД без whitelist полей

**Вложенность и коллекции:**
- Рекурсивные/глубоко вложенные схемы без ограничения глубины → ReDoS / stack overflow
- Массивы без maxItems → неограниченный рост payload
- Вложенные объекты без maxProperties

**Coercion:**
- Zod: `.coerce.boolean()` принимает строку "false" как true
- Joi: без `.options({ convert: false })` неявно кастует типы
- express-validator: без explicit type checks принимает "1" как число 1

## Граница с другими аудитами

- **Validation** — этот скилл первичный. `audit-owasp` и `audit-bugs` ссылаются сюда при находках типа "missing input check".
- **User ID из auth контекста** — VAL-04 первичный. `audit-owasp` (IDOR) — вторичный.

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение |
|----------|----------|--------|-------------|----------------|---------|
| VAL-01 | Все входящие данные (body, params, query) проходят schema-валидацию | ✅ PASS | High | `handlers/` — все routes используют zod-схемы | — |
| VAL-02 | Строки имеют maxLength, числа — диапазон, enum-значения — whitelist | ❌ FAIL 🟠 | High | `handlers/user.ts:22` | **1. Добавить maxLength в zod-схему** \\ 2. Ручная проверка длины в handler \\ 3. Ограничение на уровне БД |
| VAL-04 | Identity данные берутся из аутентифицированного контекста (не из user input) | ⏸ ACCEPTED | Medium | `routes/order.ts:9` | В baseline: legacy endpoint, запланирован рефактор |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED` / `🔍 UNVERIFIED`

Уверенность: `High` — проверил несколько ключевых файлов, паттерн очевиден / `Medium` — проверил выборочно, паттерн вероятен / `Low` — ограниченный контекст, полная уверенность невозможна

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

Если все PASS — выведи: `✅ Валидация граничных данных реализована корректно.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-validation.md`

```
# Audit Report: Boundary Data Validation — <YYYY-MM-DD HH:MM>
<таблица>
```
