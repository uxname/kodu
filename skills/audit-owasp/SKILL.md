---
name: audit-owasp
description: >
  Аудит безопасности приложения по OWASP Top 10: инъекции, broken auth, IDOR, XSS,
  CSRF, SSRF, логические уязвимости. Запускай при /audit-owasp.
---

## Правило применимости (Relevance Rule)

Применим к серверному коду с HTTP-роутингом, аутентификацией, работой с БД или файловой системой. Для чисто фронтендовых компонентов без fetch/API calls — применяй только XSS/CSRF секции. Для CLI-инструментов без сетевого взаимодействия — верни пустой ответ.

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
- **Точечные подсказки** — из «Check-ID hints» по префиксу `OWA-`.
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
| OWA-01 | A03: Все запросы к БД/OS/LDAP параметризованы, нет injection |
| OWA-02 | A01: Все защищённые маршруты имеют auth-middleware |
| OWA-03 | A01: Resource ownership проверяется, нет IDOR [⚡ dynamic] |
| OWA-04 | A02: Пароли хранятся безопасно (bcrypt/argon2/scrypt) |
| OWA-05 | A05: Безопасная конфигурация сервера (CORS, security headers, body limits) |
| OWA-06 | A07: Защита от перебора (rate limiting на auth и чувствительных эндпоинтах) |
| OWA-07 | A09: Техническая информация не утекает в ответы (stack trace, внутренние пути) |
| OWA-08 | A10: URL из user input не передаётся в HTTP-клиент без whitelist (SSRF) |
| OWA-09 | A05: CSRF-защита реализована (SameSite cookies или CSRF-токены на state-changing запросах) |

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

**OWA-01 — Параметризованные запросы, нет injection:**
- SQL строки, собранные через конкатенацию/template literals
- NoSQL injection через неэкранированные операторы (`$where`, `$regex`)
- Command injection: user input попадает в shell-команды без экранирования
- LDAP/XPath запросы с неэкранированным user input

**OWA-02 — Auth middleware на защищённых маршрутах:**
- Protected routes без auth middleware
- Privilege escalation: обычный пользователь вызывает admin-действие
- Directory traversal в file operations
- Отсутствие авторизации на отдельных маршрутах роутера

**OWA-03 — Resource ownership проверяется:**
- IDOR: ресурс запрашивается по ID без проверки ownership текущего пользователя
- Bulk operations изменяют ресурсы других пользователей
- Indirect object reference через связанные сущности без проверки доступа

**OWA-04 — Безопасное хранение паролей:**
- Слабые алгоритмы хеширования (MD5, SHA1 для паролей)
- Пароли не хешируются через bcrypt/argon2/scrypt
- Симметричное шифрование с hardcoded key
- HTTP вместо HTTPS для передачи credentials

**OWA-05 — Безопасная конфигурация сервера:**
- CORS не должен быть wildcard (`*`) в production; origins только по whitelist
- Не выставлены security-заголовки (X-Frame-Options, CSP, HSTS и т.п.)
- Не ограничен размер тела запроса (DoS через огромное тело запроса)
- Открытые error pages с технической информацией
- Конкретика из профиля (Go: `http.MaxBytesReader`, chi `cors`/`secure`; Node: `express.json({ limit })`, Helmet)

**OWA-06 — Защита от перебора:**
- Отсутствие rate limiting на login/register/reset-password эндпоинтах
- Нет rate limiting на чувствительных операциях (смена пароля, OTP-проверка)
- Session tokens не инвалидируются при logout
- Проверка JWT без явного whitelist алгоритмов: нужно отклонять `alg:none` и подмену RS256↔HS256 (атакующий передаёт `alg:none` либо RS256 с публичным ключом, выданным за HS256-секрет); слабые алгоритмы / короткий ключ
- Конкретика из профиля (Go: go-oidc `Verifier` / golang-jwt `WithValidMethods`; Node: `jwt.verify(token, secret, { algorithms: [...] })`)

**OWA-07 — Техническая информация не утекает:**
- Stack trace в ответах production API
- Внутренние пути файловой системы в error messages
- Версии зависимостей/фреймворка в заголовках или ответах
- SQL-ошибки или DB-специфичные сообщения в API responses

**OWA-08 — SSRF: URL из user input без whitelist:**
- URL из user input передаётся в HTTP-клиент без whitelist
- Fetch к внутренним адресам (169.254.x.x, 10.x.x.x, localhost, metadata endpoints)
- Редиректы на внутренние ресурсы без валидации destination

**OWA-09 — CSRF-защита:**
- State-changing запросы (POST/PUT/PATCH/DELETE) принимаются без CSRF-токена и без проверки `Origin`/`Referer`
- Cookies без корректных атрибутов (`SameSite`, `Secure`, `HttpOnly`) — браузер отправит cookie в cross-site запросе либо она доступна скрипту/по HTTP
- Мутации (включая GraphQL) доступны через GET вместо POST — обходит CSRF-защиту
- `SameSite=None` без явного обоснования (нужно только для cross-site iframe/embed сценариев)

## Граница с другими аудитами

- **Stack trace в ответах** (OWA-07) — первичный: `audit-errors` (ERR-02). Если обнаружено здесь — добавь cross-ref «*см. ERR-02*» в доказательство, не создавай дублирующий `❌ FAIL`.
- **Валидация полей, типов, диапазонов** — первичный: `audit-validation`. Здесь не дублируй.
- **Secrets в коде** — первичный: `audit-secrets`. Здесь не дублируй.
- **API-контракты** — первичный: `audit-api-contracts`. Здесь не дублируй.

## Инструментальная поддержка

Перед анализом используй инструмент категории **dep-audit** из профиля стека
(секция «Tooling by category»): он выявляет уязвимые зависимости (известные CVE).
Это самостоятельная зона, не входит в текущий чеклист, но критические CVE стоит
вынести в раздел замечаний отчёта. Если ячейка категории пустая
(`tier: general`/`generic`) — пропусти этот шаг.

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение | Исправлено |
|----------|----------|--------|-------------|----------------|---------|------------|
| OWA-01 | A03: Все запросы к БД/OS/LDAP параметризованы, нет injection | ✅ PASS | High | `db/queries.ts` проверен — все запросы параметризованы | — | — |
| OWA-02 | A01: Все защищённые маршруты имеют auth-middleware | ❌ FAIL 🔴 | High | `routes/admin.ts:14` | **1. Добавить authMiddleware на все /admin routes** \\ 2. Использовать router-level middleware \\ 3. Добавить проверку в каждый handler | Нет |
| OWA-05 | A05: Безопасная конфигурация сервера (CORS, security headers, body limits) | ⏸ ACCEPTED | Medium | `app.ts:9` | В baseline: внутренний сервис | — |

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

Если все PASS — выведи: `✅ Критических OWASP-уязвимостей не обнаружено.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-owasp.md`

```
# Audit Report: OWASP Application Security — <YYYY-MM-DD HH:MM>
<таблица>
```
