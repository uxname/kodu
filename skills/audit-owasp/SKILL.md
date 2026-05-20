---
name: audit-owasp
description: >
  Аудит безопасности приложения по OWASP Top 10: инъекции, broken auth, IDOR, XSS,
  CSRF, SSRF, логические уязвимости. Запускай при /audit-owasp.
---

## Правило применимости (Relevance Rule)

Применим к серверному коду с HTTP-роутингом, аутентификацией, работой с БД или файловой системой. Для чисто фронтендовых компонентов без fetch/API calls — применяй только XSS/CSRF секции. Для CLI-инструментов без сетевого взаимодействия — верни пустой ответ.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| OWA-01 | A03: Все запросы к БД/OS/LDAP параметризованы, нет injection |
| OWA-02 | A01: Все защищённые маршруты имеют auth-middleware |
| OWA-03 | A01: Resource ownership проверяется, нет IDOR |
| OWA-04 | A02: Пароли хранятся безопасно (bcrypt/argon2/scrypt) |
| OWA-05 | A05: Безопасная конфигурация сервера (CORS, security headers, body limits) |
| OWA-06 | A07: Защита от перебора (rate limiting на auth и чувствительных эндпоинтах) |
| OWA-07 | A09: Техническая информация не утекает в ответы (stack trace, внутренние пути) |
| OWA-08 | A10: URL из user input не передаётся в HTTP-клиент без whitelist (SSRF) |

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
- CORS wildcard (`*`) в production или открытый origins без whitelist
- Отсутствие security headers (Helmet или аналог: X-Frame-Options, CSP, HSTS)
- express.json / bodyParser без limit (DoS через огромное тело запроса)
- Открытые error pages с технической информацией

**OWA-06 — Защита от перебора:**
- Отсутствие rate limiting на login/register/reset-password эндпоинтах
- Нет rate limiting на чувствительных операциях (смена пароля, OTP-проверка)
- Слабые JWT алгоритмы (alg:none, HS256 с коротким ключом)
- Session tokens не инвалидируются при logout

**OWA-07 — Техническая информация не утекает:**
- Stack trace в ответах production API
- Внутренние пути файловой системы в error messages
- Версии зависимостей/фреймворка в заголовках или ответах
- SQL-ошибки или DB-специфичные сообщения в API responses

**OWA-08 — SSRF: URL из user input без whitelist:**
- URL из user input передаётся в HTTP-клиент (axios, fetch, got) без whitelist
- Fetch к внутренним адресам (169.254.x.x, 10.x.x.x, localhost, metadata endpoints)
- Редиректы на внутренние ресурсы без валидации destination

## Граница с другими аудитами

- **Stack trace в ответах** — первичный: `audit-errors` (ERR-02). Если обнаружено в audit-owasp — добавь перекрёстную ссылку: *«см. audit-errors ERR-02»*, не дублируй FAIL.
- **Валидация полей, типов, диапазонов** — первичный: `audit-validation`. Здесь не дублируй.
- **Secrets в коде** — первичный: `audit-secrets`. Здесь не дублируй.
- **API-контракты** — первичный: `audit-api-contracts`. Здесь не дублируй.

## Инструментальная поддержка

Перед анализом:
```bash
npm audit --json 2>/dev/null | head -100 || pnpm audit --json 2>/dev/null | head -100 || true
```
`npm audit` выявляет уязвимые зависимости — это самостоятельная зона, не входит в текущий чеклист, но критические CVE стоит вынести в раздел замечаний отчёта.

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение |
|----------|----------|--------|-------------|----------------|---------|
| OWA-01 | A03: Все запросы к БД/OS/LDAP параметризованы, нет injection | ✅ PASS | High | `db/queries.ts` проверен — все запросы параметризованы | — |
| OWA-02 | A01: Все защищённые маршруты имеют auth-middleware | ❌ FAIL 🔴 | High | `routes/admin.ts:14` | **1. Добавить authMiddleware на все /admin routes** \\ 2. Использовать router-level middleware \\ 3. Добавить проверку в каждый handler |
| OWA-05 | A05: Безопасная конфигурация сервера (CORS, security headers, body limits) | ⏸ ACCEPTED | Medium | `app.ts:9` | В baseline: внутренний сервис |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED` / `🔍 UNVERIFIED`

Уверенность: `High` — проверил несколько ключевых файлов, паттерн очевиден / `Medium` — проверил выборочно, паттерн вероятен / `Low` — ограниченный контекст, полная уверенность невозможна

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

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
