---
name: audit-secrets
description: >
  Аудит утечки секретов: поиск захардкоженных ключей, паролей, токенов, credentials в коде.
  Запускай когда пользователь просит проверить код на наличие секретов, утечки credentials,
  hardcoded паролей или при инвоке /audit-secrets.
---

## Правило применимости (Relevance Rule)

Перед анализом оцени: содержит ли код конфигурации, строки подключения, токены, ключи шифрования, credentials или работу с внешними API? Если анализируемый файл/модуль не содержит ни одного из перечисленных паттернов — верни пустой ответ без таблицы.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| SEC-01 | Нет hardcoded паролей/токенов/API-ключей в строковых литералах |
| SEC-02 | .env файлы в .gitignore |
| SEC-03 | Нет secrets в URL (query params, Basic Auth) |
| SEC-04 | .env.example не содержит реальных credentials |
| SEC-05 | Нет secrets в Dockerfile ENV директивах |
| SEC-06 | Нет credentials в комментариях |

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

**Что искать:**
- Пароли, токены, API-ключи в строковых литералах
- Строки подключения к БД с credentials
- Private keys, certificates, JWT secrets в коде
- Секреты в URL (query params, Basic Auth в URL)
- Credentials в комментариях
- Тестовые/дев credentials, которые могут попасть в прод
- Паттерны: `password = "..."`, `token = "..."`, `key = "..."`, `secret = "..."`
- Base64-encoded credentials
- Секреты в `.env`-файлах, закоммиченных в репозиторий

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| SEC-01 | Нет hardcoded паролей/токенов/API-ключей | ✅ PASS | — | — |
| SEC-02 | .env файлы в .gitignore | ❌ FAIL 🔴 | `.gitignore:1` | **1. Добавить .env в .gitignore** \\ 2. Использовать git-crypt \\ 3. Удалить .env из истории через git-filter-repo |
| SEC-03 | Нет secrets в URL | ⏸ ACCEPTED | `config.ts:9` | В baseline: legacy-интеграция, планируется замена |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

Если все PASS — выведи: `✅ Утечек секретов не обнаружено.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-secrets.md`

```
# Audit Report: Secrets Leak — <YYYY-MM-DD HH:MM>
<таблица>
```
