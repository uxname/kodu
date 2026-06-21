---
name: audit-deployment
description: >
  Аудит сборки и деплоя: оптимизация Dockerfile, переменные окружения, CI/CD конфигурации,
  secrets в конфигах, non-root пользователи. Запускай при /audit-deployment.
---

## Правило применимости (Relevance Rule)

Применим при наличии Dockerfile, docker-compose.yml, CI/CD конфигов (`.github/workflows`, `gitlab-ci.yml`, `Jenkinsfile`), `.env` файлов, Kubernetes манифестов. Для проектов без deployment конфигурации — верни пустой ответ.

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
- **Точечные подсказки** — из «Check-ID hints» по префиксу `DEP-`.
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
| DEP-01 | Docker images используют pinned versions (нет :latest) |
| DEP-02 | Контейнеры запускаются от непривилегированного пользователя (USER nonroot) |
| DEP-03 | Multi-stage build: финальный образ без инструментов сборки и dev-артефактов |
| DEP-04 | .dockerignore исключает артефакты сборки/зависимости, VCS-каталог и секреты |
| DEP-05 | HEALTHCHECK определён в Dockerfile |
| DEP-06 | Секреты не hardcoded в Dockerfile (нет в ENV) |
| DEP-07 | .env исключён из VCS |
| DEP-08 | .env.example документирует все переменные окружения |
| DEP-09 | Окружение переключено в production-режим: dev-артефакты выключены в проде |
| DEP-10 | Детерминированная установка по локфайлу с проверкой целостности |
| DEP-11 | Ограничения ресурсов контейнера определены (CPU limits, Memory limits) |
| DEP-12 | Возможность запуска с read-only root filesystem проверена |

## Правила верификации

1. **Только чеклист**: оценивай ТОЛЬКО проверки выше. Не добавляй новые.
2. **Явная верификация = PASS**: ставь `✅ PASS` только если явно проверил механизм (нашёл схему, конфиг, guard) и подтвердил отсутствие нарушения — укажи что именно проверено.
3. **Нет доказательства = UNVERIFIED**: не можешь указать `файл:строка` ни для нарушения, ни для подтверждения — ставь `🔍 UNVERIFIED`.
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

> Примеры ниже — иллюстративные. Конкретные инструменты, идиомы и анти-паттерны
> сборки/деплоя для текущего рантайма бери из загруженного профиля
> (`stacks/<runtime>.md`, секции Idioms/Anti-patterns/Check-ID hints по префиксу
> `DEP-`). Node: `npm ci`/`package-lock.json`, `NODE_ENV`, `node_modules`.
> Go: builder→distroless, `go.mod`+`go.sum`, `CGO_ENABLED=0`, нет `net/http/pprof`.

**DEP-01 — Pinned versions:**
- Базовый образ `:latest` без pinning версии (непредсказуемые обновления)
- Тег `:alpine` без конкретной версии
- Digest-based pinning отсутствует для критичных образов

**DEP-02 — Непривилегированный пользователь:**
- Запуск контейнера как root без переключения на non-root (Node: `USER node`; Go: distroless `nonroot`)
- Отсутствие создания non-root пользователя перед переключением (или образа со встроенным nonroot)

**DEP-03 — Multi-stage build: финальный образ без инструментов сборки и dev-артефактов:**
- Отсутствие multi-stage build (инструменты сборки/dev-зависимости в финальном образе)
- Node: `devDependencies` устанавливаются в production stage
- Go: компилятор/тесты в финальном образе вместо `builder → distroless`
- Build artifacts не копируются из builder stage

**DEP-04 — .dockerignore:**
- Нет `.dockerignore` или он не исключает зависимости/артефакты сборки (Node: `node_modules`; Go: локальный `vendor/`, build-кэш) и VCS-каталог `.git`
- `.env` и прочие секреты не исключены из Docker build context
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

**DEP-09 — Окружение переключено в production-режим:**
- Dev-артефакты не выключены в проде: debug-эндпоинты, verbose-логи, dev-зависимости
- Node: `NODE_ENV` не выставлен или `development` в prod образе (ведёт к загрузке devDependencies в runtime)
- Go: нет признаков prod-режима (нет `APP_ENV`/собственного флага), `net/http/pprof` и debug-роуты доступны в проде, нет `-ldflags "-s -w"`, нет `CGO_ENABLED=0`

**DEP-10 — Детерминированная установка по локфайлу с проверкой целостности:**
- Node: `npm install` вместо `npm ci` (не детерминированная, медленнее); отсутствие `package-lock.json`
- Go: `go.mod`/`go.sum` не закоммичены; в Dockerfile `go get` вместо `go mod download` по локфайлу; нет `GOFLAGS=-mod=readonly` / `go mod verify`

**Ограничения ресурсов:**
- `docker-compose.yml` без `deploy.resources.limits.memory` и `cpus`
- Kubernetes Deployment без `resources.limits` в контейнере
- Нет ограничений → один контейнер с memory leak роняет весь хост
- Go-доп.: при memory-лимите контейнера выставить `GOMEMLIMIT`/`GOMAXPROCS` под cgroup-квоты (иначе GC видит память/CPU хоста)

**Read-only filesystem:**
- Контейнер без `read_only: true` (docker-compose) или `readOnlyRootFilesystem: true` (K8s)
- Если приложение пишет во временные файлы — проверить наличие tmpfs mount для `/tmp`

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение | Исправлено |
|----------|----------|--------|-------------|----------------|---------|------------|
| DEP-01 | Docker images используют pinned versions (нет :latest) | ✅ PASS | High | `Dockerfile:1` — pinned version указана | — | — |
| DEP-02 | Контейнеры запускаются от непривилегированного пользователя (USER nonroot) | ❌ FAIL 🟠 | High | `Dockerfile:12` | **1. Добавить `RUN addgroup -S app && adduser -S app -G app` и `USER app`** \\ 2. Использовать образ со встроенным nonroot (Node: node:alpine; Go: distroless nonroot) \\ 3. Добавить `USER node` если базовый образ node | Нет |
| DEP-05 | HEALTHCHECK определён в Dockerfile | ⏸ ACCEPTED | Medium | — | В baseline: health check управляется Kubernetes liveness probe | — |

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
