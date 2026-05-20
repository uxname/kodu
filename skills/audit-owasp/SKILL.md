---
name: audit-owasp
description: >
  Аудит безопасности приложения по OWASP Top 10: инъекции, broken auth, IDOR, XSS,
  CSRF, SSRF, логические уязвимости. Запускай при /audit-owasp.
---

## Правило применимости (Relevance Rule)

Применим к серверному коду с HTTP-роутингом, аутентификацией, работой с БД или файловой системой. Для чисто фронтендовых компонентов без fetch/API calls — применяй только XSS/CSRF секции. Для CLI-инструментов без сетевого взаимодействия — верни пустой ответ.

## Задача

Ты — эксперт по application security, проводящий аудит по OWASP Top 10. Найди логические уязвимости, небезопасные паттерны и нарушения security best practices.

## Что анализировать

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
- LDAP/XPath injection

**A07 — Auth & Session Failures:**
- Отсутствие rate limiting на login endpoint
- Слабые JWT алгоритмы (alg:none, HS256 с коротким ключом)
- Session tokens не инвалидируются при logout
- Refresh tokens без rotation

**A10 — SSRF:**
- URL из user input передаётся в HTTP-клиент без whitelist
- Fetch к внутренним адресам (169.254.x.x, 10.x.x.x, localhost)

**CSRF:**
- State-changing запросы без CSRF-токена или SameSite cookie

**XSS (для SSR/template engines):**
- User input рендерится в HTML без escaping

## Граница с другими аудитами

- **Проверка полей, типов, диапазонов, обязательных полей** — зона `audit-validation`, здесь не дублируй
- **Консистентность API-контрактов** — зона `audit-api-contracts`
- **Прямые логические баги** — зона `audit-bugs`

## Формат вывода

| Сценарий | Что происходит | Риск | Текущее поведение / Меры защиты | Варианты решений | Статус |
|----------|---------------|------|--------------------------------|------------------|--------|
| [файл:строка + OWASP категория] | [вектор атаки + эксплоит-сценарий] | 🔴/🟠/🟡/🟢 | [текущая защита] | **1. [Secure implementation]** \\ 2. [Mitigation контроль] \\ 3. [Минимальный патч] | [ ] |

## Требования к вариантам решений

Ровно 3 варианта. Вариант 1 жирным с конкретным secure паттерном или библиотекой. Разделитель `\\`. Без `<br>`.

Если уязвимостей не обнаружено — выведи: `✅ Критических OWASP-уязвимостей не обнаружено.`

## Сохранение результатов

После завершения анализа выполни следующие шаги через инструменты:

1. Найди папку текущей сессии через Bash:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если вывод пустой — создай новую: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")` и используй её путь.
2. Сохрани отчёт через Write в файл: `<AUDIT_DIR>/audit-owasp.md`

Структура файла:
```
# Audit Report: OWASP Application Security — <YYYY-MM-DD HH:MM>

<таблица с результатами или строка об отсутствии находок>
```

Сообщи пользователю путь к созданному файлу.
