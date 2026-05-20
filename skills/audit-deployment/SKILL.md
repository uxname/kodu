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
| DEP-01 | Docker images используют pinned versions (нет :latest) |
| DEP-02 | Контейнеры запускаются от непривилегированного пользователя (USER nonroot) |
| DEP-03 | Multi-stage build разделяет dev и prod зависимости |
| DEP-04 | .dockerignore исключает node_modules, .git, .env |
| DEP-05 | HEALTHCHECK определён в Dockerfile |
| DEP-06 | Секреты не hardcoded в Dockerfile (нет в ENV) |
| DEP-07 | .env исключён из VCS |
| DEP-08 | .env.example документирует все переменные окружения |
| DEP-09 | NODE_ENV корректно устанавливается для production |
| DEP-10 | npm ci используется вместо npm install в Docker |
| DEP-11 | Ограничения ресурсов контейнера определены (CPU limits, Memory limits) |
| DEP-12 | Возможность запуска с read-only root filesystem проверена |

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

**DEP-01 — Pinned versions:**
- Базовый образ `:latest` без pinning версии (непредсказуемые обновления)
- Тег `:alpine` без конкретной версии
- Digest-based pinning отсутствует для критичных образов

**DEP-02 — Непривилегированный пользователь:**
- Запуск контейнера как root без `USER nonroot` / `USER node`
- Отсутствие создания non-root пользователя перед переключением

**DEP-03 — Multi-stage build:**
- Отсутствие multi-stage build (dev dependencies в prod образе)
- devDependencies устанавливаются в production stage
- Build artifacts не копируются из builder stage

**DEP-04 — .dockerignore:**
- Нет `.dockerignore` или `.dockerignore` не включает node_modules, .git
- `.env` файлы не исключены из Docker build context
- Тесты и документация попадают в production образ

**DEP-05 — HEALTHCHECK:**
- HEALTHCHECK отсутствует в Dockerfile
- Health check endpoint не существует в приложении

**DEP-06 — Секреты не в Dockerfile ENV:**
- Secrets в `ENV` директивах Dockerfile (видны в docker inspect и слоях образа)
- Credentials в `ARG` без использования build secrets (`--secret`)

**DEP-07 — .env исключён из VCS:**
- `.env` файл с реальными credentials закоммичен в репозиторий
- Отсутствие `.env*` в `.gitignore`

**DEP-08 — .env.example документирует переменные:**
- `.env.example` отсутствует
- `.env.example` содержит реальные credentials
- Не все обязательные переменные задокументированы

**DEP-09 — NODE_ENV для production:**
- `NODE_ENV` не устанавливается или устанавливается как `development` в prod образе
- Отсутствие `NODE_ENV=production` ведёт к загрузке devDependencies в runtime

**DEP-10 — npm ci в Docker:**
- `npm install` вместо `npm ci` (не детерминированная, более медленная сборка)
- Отсутствие `package-lock.json` при использовании npm

**Ограничения ресурсов:**
- `docker-compose.yml` без `deploy.resources.limits.memory` и `cpus`
- Kubernetes Deployment без `resources.limits` в контейнере
- Нет ограничений → один контейнер с memory leak роняет весь хост

**Read-only filesystem:**
- Контейнер без `read_only: true` (docker-compose) или `readOnlyRootFilesystem: true` (K8s)
- Если приложение пишет во временные файлы — проверить наличие tmpfs mount для `/tmp`

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение |
|----------|----------|--------|-------------|----------------|---------|
| DEP-01 | Docker images используют pinned versions (нет :latest) | ✅ PASS | High | `Dockerfile:1` — pinned version указана | — |
| DEP-02 | Контейнеры запускаются от непривилегированного пользователя (USER nonroot) | ❌ FAIL 🟠 | High | `Dockerfile:12` | **1. Добавить `RUN addgroup -S app && adduser -S app -G app` и `USER app`** \\ 2. Использовать образ с встроенным nonroot пользователем (node:alpine) \\ 3. Добавить `USER node` если базовый образ node |
| DEP-05 | HEALTHCHECK определён в Dockerfile | ⏸ ACCEPTED | Medium | — | В baseline: health check управляется Kubernetes liveness probe |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED` / `🔍 UNVERIFIED`

Уверенность: `High` — проверил несколько ключевых файлов, паттерн очевиден / `Medium` — проверил выборочно, паттерн вероятен / `Low` — ограниченный контекст, полная уверенность невозможна

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
