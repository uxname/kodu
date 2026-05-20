---
name: audit-deployment
description: >
  Аудит сборки и деплоя: оптимизация Dockerfile, переменные окружения, CI/CD конфигурации,
  secrets в конфигах, non-root пользователи. Запускай при /audit-deployment.
---

## Правило применимости (Relevance Rule)

Применим при наличии Dockerfile, docker-compose.yml, CI/CD конфигов (`.github/workflows`, `gitlab-ci.yml`, `Jenkinsfile`), `.env` файлов, Kubernetes манифестов. Для проектов без deployment конфигурации — верни пустой ответ.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| DEP-01 | Dockerfile не использует :latest |
| DEP-02 | Dockerfile имеет USER nonroot |
| DEP-03 | Multi-stage build (dev deps не в prod) |
| DEP-04 | .dockerignore включает node_modules, .git |
| DEP-05 | HEALTHCHECK определён |
| DEP-06 | Нет secrets в ENV Dockerfile |
| DEP-07 | .env в .gitignore |
| DEP-08 | .env.example существует |
| DEP-09 | NODE_ENV устанавливается при деплое |
| DEP-10 | npm ci вместо npm install в Docker |

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

**Dockerfile:**
- Базовый образ `:latest` без pinning версии
- Запуск как root без `USER nonroot`
- Secrets в ENV директиве (видны в docker inspect и layers)
- Отсутствие multi-stage build (dev dependencies в prod образе)
- Нет `.dockerignore` или `.dockerignore` не включает node_modules, .git
- Отсутствие health check
- `npm install` вместо `npm ci` (не детерминированная сборка)

**Переменные окружения:**
- Секреты в docker-compose.yml в открытом виде
- `.env` файл с реальными credentials закоммичен в репозиторий
- Отсутствие `.env.example` для документирования required vars
- Production secrets в CI/CD конфиге в открытом виде (не masked)

**CI/CD:**
- Отсутствие проверки secrets scanning в pipeline
- Нет dependency vulnerability scan (npm audit, Snyk)
- Нет timeout для CI jobs (бесконечный run при зависании)

**Kubernetes / Compose:**
- Containers без resource limits (CPU/Memory)
- Отсутствие readinessProbe / livenessProbe

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| DEP-01 | Dockerfile не использует :latest | ✅ PASS | — | — |
| DEP-02 | Dockerfile имеет USER nonroot | ❌ FAIL 🟠 | `Dockerfile:12` | **1. Добавить `RUN addgroup -S app && adduser -S app -G app` и `USER app`** \\ 2. Использовать образ с встроенным nonroot пользователем (node:alpine) \\ 3. Добавить `USER node` если базовый образ node |
| DEP-05 | HEALTHCHECK определён | ⏸ ACCEPTED | — | В baseline: health check управляется Kubernetes liveness probe |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

Если все PASS — выведи: `✅ Конфигурация сборки и деплоя в порядке.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-deployment.md`

```
# Audit Report: Build & Deployment Configuration — <YYYY-MM-DD HH:MM>
<таблица>
```
