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
| OWA-01 | A03: Нет конкатенации строк в DB/shell запросах |
| OWA-02 | A01: Все protected routes имеют auth middleware |
| OWA-03 | A01: Нет IDOR — resource ownership проверяется |
| OWA-04 | A02: Пароли хешируются через bcrypt/argon2/scrypt |
| OWA-05 | A05: CORS не wildcard в production |
| OWA-06 | A07: Rate limiting присутствует на API |
| OWA-07 | A05: Helmet/security headers установлены |
| OWA-08 | A05: express.json с limit или аналог |
| OWA-09 | A05: Stack trace не попадает в ответы |
| OWA-10 | A10: URL из user input не в HTTP-клиент без whitelist |

## Правила верификации

1. **Только чеклист**: оценивай ТОЛЬКО проверки выше. Не добавляй новые.
2. **Нет доказательства = ✅ PASS**: не можешь указать `файл:строка` — ставь PASS.
3. **Baseline приоритетен**: check_id есть в `docs/audit-baseline.yml` → `⏸ ACCEPTED`.
4. **Только 🔴/🟠 FAIL требуют решения**: 🟡/🟢 — решение необязательно.

## Baseline

До анализа:
```bash
cat ./docs/audit-baseline.yml 2>/dev/null
```

## Контекст анализа

**A01 — Broken Access Control:**
- IDOR: ресурс запрашивается по ID без проверки ownership
- Privilege escalation: обычный пользователь может вызвать admin-действие
- Отсутствие авторизации на отдельных маршрутах
- Directory traversal в file operations

**A02 — Cryptographic Failures:**
- Слабые алгоритмы хеширования (MD5, SHA1 для паролей)
- Пароли не хешируются через bcrypt/argon2/scrypt
- Симметричное шифрование с hardcoded key
- HTTP вместо HTTPS для передачи credentials

**A03 — Injection:**
- SQL строки, собранные через конкатенацию/template literals
- NoSQL injection через неэкранированные операторы (`$where`, `$regex`)
- Command injection: user input в shell команды

**A07 — Auth & Session Failures:**
- Отсутствие rate limiting на login endpoint
- Слабые JWT алгоритмы (alg:none, HS256 с коротким ключом)
- Session tokens не инвалидируются при logout

**A10 — SSRF:**
- URL из user input передаётся в HTTP-клиент без whitelist
- Fetch к внутренним адресам (169.254.x.x, 10.x.x.x, localhost)

## Граница с другими аудитами

- **Проверка полей, типов, диапазонов, обязательных полей** — зона `audit-validation`, здесь не дублируй
- **Консистентность API-контрактов** — зона `audit-api-contracts`
- **Прямые логические баги** — зона `audit-bugs`

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| OWA-01 | A03: Нет конкатенации строк в DB/shell запросах | ✅ PASS | — | — |
| OWA-02 | A01: Все protected routes имеют auth middleware | ❌ FAIL 🔴 | `routes/admin.ts:14` | **1. Добавить authMiddleware на все /admin routes** \\ 2. Использовать router-level middleware \\ 3. Добавить проверку в каждый handler |
| OWA-05 | A05: CORS не wildcard в production | ⏸ ACCEPTED | `app.ts:9` | В baseline: внутренний сервис |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

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
