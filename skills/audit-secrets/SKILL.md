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
| SEC-01 | Нет hardcoded credentials в коде (пароли, токены, API-ключи, приватные ключи) |
| SEC-02 | Файлы с секретами исключены из VCS (.env* в .gitignore) |
| SEC-03 | Секреты не передаются через URL (query params, Basic Auth в URL) |
| SEC-04 | .env.example содержит только placeholder-значения без реальных данных |
| SEC-05 | Dockerfile не содержит секретов в ENV-директивах |
| SEC-06 | Комментарии в коде не содержат credentials |

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

**SEC-01 — Нет hardcoded credentials в коде:**
- Пароли, токены, API-ключи в строковых литералах
- Строки подключения к БД с credentials (`postgres://user:pass@host`)
- Private keys, certificates, JWT secrets в коде
- Base64-encoded credentials в коде
- Тестовые/дев credentials, которые могут попасть в прод
- Паттерны: `password = "..."`, `token = "..."`, `key = "..."`, `secret = "..."`

**SEC-02 — Файлы с секретами исключены из VCS:**
- `.env`, `.env.local`, `.env.production` не в `.gitignore`
- Закоммиченные `.env`-файлы с реальными данными в репозитории
- Certificate и key файлы (`*.pem`, `*.key`, `*.p12`) не исключены из VCS

**SEC-03 — Секреты не в URL:**
- API-ключи или токены в query params (`?api_key=...`)
- Basic Auth credentials в URL (`https://user:pass@host`)
- Секреты в redirect_uri или callback URL параметрах

**SEC-04 — .env.example без реальных данных:**
- `.env.example` содержит реальные значения вместо placeholders (`DB_PASS=realpassword`)
- Placeholder-значения не описывают ожидаемый формат (`DB_URL=` без пояснения)

**SEC-05 — Dockerfile без секретов в ENV:**
- Секреты в `ENV` директивах Dockerfile (видны в docker inspect и слоях образа)
- Credentials в `ARG` без использования build secrets
- Секреты в LABEL или COPY-командах

**SEC-06 — Комментарии без credentials:**
- Пароли или токены в закомментированном коде
- TODO-комментарии с примерами реальных credentials
- Инструкции по настройке с реальными значениями

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| SEC-01 | Нет hardcoded credentials в коде (пароли, токены, API-ключи, приватные ключи) | ✅ PASS | — | — |
| SEC-02 | Файлы с секретами исключены из VCS (.env* в .gitignore) | ❌ FAIL 🔴 | `.gitignore:1` | **1. Добавить .env в .gitignore** \\ 2. Использовать git-crypt \\ 3. Удалить .env из истории через git-filter-repo |
| SEC-03 | Секреты не передаются через URL (query params, Basic Auth в URL) | ⏸ ACCEPTED | `config.ts:9` | В baseline: legacy-интеграция, планируется замена |

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
