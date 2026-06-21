---
name: audit-validation
description: >
  Аудит валидации граничных данных: проверка входящих данных на границах системы,
  отсутствие sanitization, trust boundary нарушения. Запускай при /audit-validation.
---

## Правило применимости (Relevance Rule)

Применим к коду, принимающему внешние данные: HTTP handlers, WebSocket, CLI args, file parsers, event consumers, gRPC endpoints. Для чисто внутреннего кода без внешних входов — верни пустой ответ.

## Runtime Detection & Stack Profile

Этот аудит стек-агностичен: проверки сформулированы нейтрально, а конкретика
(инструменты, идиомы, анти-паттерны, примеры) берётся из профиля стека.

1. **Профиль передан контекстом?** Если оркестратор `/audit` передал
   `runtime=<id>` и/или содержимое профиля — используй его, шаги 2–3 пропусти.

2. **Иначе определи РОВНО ОДИН рантайм** этого каталога:
   ```bash
   if   [ -f package.json ]; then echo "runtime=node"
   elif [ -f go.mod ]; then echo "runtime=go"
   elif [ -f pyproject.toml ] || [ -f requirements.txt ] || [ -f setup.py ]; then echo "runtime=python"
   elif [ -f Cargo.toml ]; then echo "runtime=rust"
   elif [ -f pom.xml ] || ls build.gradle* settings.gradle* >/dev/null 2>&1; then echo "runtime=java"
   else echo "runtime=generic"; fi
   ```
   Один запуск = один рантайм; не миксуй backend и frontend. Если найдено
   несколько маркеров (монорепо) — выбери соответствующий текущему scope/анализируемым
   файлам и зафиксируй выбор в разделе Audit Coverage.

3. **Загрузи профиль** через Read: `./skills/audit/stacks/<runtime>.md`
   (fallback `./skills/audit/stacks/_generic.md`, если файл не найден).

Дальше используй профиль:
- **Инструменты** — из секции «Tooling by category» профиля (раздел
  «Инструментальная поддержка» ниже ссылается на категории, а не на команды).
- **Ожидания PASS** — из «Idioms»; **формулировки FAIL** — из «Anti-patterns».
- **Точечные подсказки** — из «Check-ID hints» по префиксу `VAL-`.
- Если профиль `tier: general` или `runtime=generic` → стек-специфичные находки
  без однозначного evidence помечай `🔍 UNVERIFIED`, а не `❌ FAIL`. Проверки,
  чей механизм в рантайме отсутствует, помечай `N/A`.

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
| VAL-01 | Все входящие данные (body, params, query) проходят schema-валидацию |
| VAL-02 | Строки имеют maxLength, числа — диапазон, enum-значения — whitelist |
| VAL-03 | Парсинг недоверенного ввода обрабатывает ошибки и валидирует структуру результата |
| VAL-04 | Identity данные берутся из аутентифицированного контекста (не из user input) |
| VAL-05 | Вложенные структуры и массивы ограничены (глубина, minItems/maxItems) |
| VAL-06 | Валидатор не выполняет неявный coercion (строка "false" → boolean true) [⚡ dynamic] |
| VAL-07 | Prototype pollution: merge/assign с user input фильтрует `__proto__`, `constructor`, `prototype` |
| VAL-08 | Загрузка файлов: MIME тип проверяется по содержимому, имя файла санитизировано, размер ограничен |

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

## Language Rule

Результаты аудита должны быть написаны простым и понятным языком. Избегай сложных терминов, жаргона и абстрактных понятий без необходимости. Общепринятые технические термины (Docker, HTTP, API, JSON, URL) допустимы. Описывай проблемы так, чтобы они были понятны разработчику любого уровня, а не только узкому специалисту в данной области.

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
- HTTP request body используется напрямую без schema-валидации (Node: zod/joi/yup/etc; Go: проверка в resolver/handler — go-playground/validator или ручная проверка; gqlgen типизирует вход на уровне схемы, но прикладные инварианты всё равно проверяй)
- Query params / path params используются без типизации и проверки
- Отсутствие проверки обязательных полей
- Нет валидации типов (строка может прийти вместо числа)
- WebSocket, CLI args, event payloads без валидации входных данных

**VAL-02 — Строки, числа, enum ограничены:**
- Отсутствие maxLength для строк (DoS через огромную строку)
- Нет проверки диапазонов чисел (отрицательные ID, огромные offset, NaN)
- Отсутствие whitelist для enum-полей (принимается любое строковое значение)
- Нет проверки формата (email, UUID, дата) там где она применима

**VAL-03 — Парсинг недоверенного ввода с защитой:**
- Node: `JSON.parse` без try/catch — выбросит SyntaxError при невалидном input
- Go: `json.Unmarshal` / парсинг без проверки возвращаемой ошибки (err проигнорирован)
- Парсинг без последующей валидации структуры (тип полей не проверен)
- Доверие структуре распарсенного результата без schema-проверки

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

**VAL-07 — Prototype pollution:**
- Преимущественно JS-специфично (`__proto__`/`constructor`/`prototype`). В Go prototype pollution неприменим (нет прототипной модели объектов) → для Go-рантайма помечай VAL-07 как `N/A`. Концепт ниже актуален для Node:
- `Object.assign(target, userInput)` без проверки — ключ `__proto__` загрязняет Object.prototype
- `_.merge(obj, userInput)` в lodash < 4.17.21 — уязвим к prototype pollution
- Deep merge из user input без sanitize ключей (`constructor`, `prototype`, `__proto__`)
- Последствие: `({}).isAdmin === true` для всех объектов после атаки

**VAL-08 — Безопасная загрузка файлов:**
- Проверка типа только через `file.mimetype` — значение подставлено клиентом, не верифицировано
- `path.join(uploadDir, file.originalname)` — `originalname` может содержать path traversal (`../../../etc/passwd`)
- Нет ограничения `maxFileSize` — DoS через огромный файл
- Разрешённые расширения не ограничены — возможна загрузка исполняемых файлов (.sh, .exe, .php)

## Граница с другими аудитами

- **Validation** — этот скилл первичный. `audit-owasp` и `audit-bugs` ссылаются сюда при находках типа "missing input check".
- **User ID из auth контекста** — VAL-04 первичный. `audit-owasp` (IDOR) — вторичный.

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение | Исправлено |
|----------|----------|--------|-------------|----------------|---------|------------|
| VAL-01 | Все входящие данные (body, params, query) проходят schema-валидацию | ✅ PASS | High | `handlers/` — все routes используют zod-схемы | — | — |
| VAL-02 | Строки имеют maxLength, числа — диапазон, enum-значения — whitelist | ❌ FAIL 🟠 | High | `handlers/user.ts:22` | **1. Добавить maxLength в zod-схему** \\ 2. Ручная проверка длины в handler \\ 3. Ограничение на уровне БД | Нет |
| VAL-04 | Identity данные берутся из аутентифицированного контекста (не из user input) | ⏸ ACCEPTED | Medium | `routes/order.ts:9` | В baseline: legacy endpoint, запланирован рефактор | — |

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
