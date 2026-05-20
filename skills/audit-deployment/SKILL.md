---
name: audit-deployment
description: >
  Аудит сборки и деплоя: оптимизация Dockerfile, переменные окружения, CI/CD конфигурации,
  secrets в конфигах, non-root пользователи. Запускай при /audit-deployment.
---

## Правило применимости (Relevance Rule)

Применим при наличии Dockerfile, docker-compose.yml, CI/CD конфигов (`.github/workflows`, `gitlab-ci.yml`, `Jenkinsfile`), `.env` файлов, Kubernetes манифестов. Для проектов без deployment конфигурации — верни пустой ответ.

## Задача

Ты — DevSecOps эксперт, проводящий аудит конфигурации сборки и деплоя. Найди проблемы безопасности, оптимизации и надёжности.

## Что анализировать

**Dockerfile:**
- Базовый образ `:latest` без pinning версии
- Запуск как root без `USER nonroot`
- Secrets в ENV директиве (видны в docker inspect и layers)
- Отсутствие multi-stage build (dev dependencies в prod образе)
- Неправильный порядок слоёв (COPY package.json до npm install нарушен)
- Нет `.dockerignore` или `.dockerignore` не включает node_modules, .git
- Отсутствие health check
- Установка пакетов без `--no-cache` / `rm -rf /var/lib/apt/lists/*`

**Переменные окружения:**
- Секреты в docker-compose.yml в открытом виде
- `.env` файл с реальными credentials закоммичен в репозиторий
- Отсутствие `.env.example` для документирования required vars
- Production secrets в CI/CD конфиге в открытом виде (не masked)
- Нет разделения между dev/staging/prod конфигурациями

**CI/CD:**
- Отсутствие проверки secrets scanning в pipeline
- Нет dependency vulnerability scan (npm audit, Snyk)
- Артефакты сборки с возможными секретами публикуются в логах
- Непривилегированные шаги используют GITHUB_TOKEN с избыточными правами
- Нет timeout для CI jobs (бесконечный run при зависании)

**Kubernetes / Compose:**
- Containers без resource limits (CPU/Memory)
- Отсутствие readinessProbe / livenessProbe
- Privileged: true без необходимости
- hostNetwork: true / hostPID: true

## Формат вывода

| Сценарий | Что происходит | Риск | Текущее поведение / Меры защиты | Варианты решений | Статус |
|----------|---------------|------|--------------------------------|------------------|--------|
| [файл:строка + директива] | [риск в prod] | 🔴/🟠/🟡/🟢 | [текущая конфигурация] | **1. [Secure конфигурация]** \\ 2. [Альтернативный подход] \\ 3. [Минимальный фикс] | [ ] |

## Требования к вариантам решений

Ровно 3 варианта. Вариант 1 жирным с конкретной конфигурацией. Разделитель `\\`. Без `<br>`.

Если проблем не обнаружено — выведи: `✅ Конфигурация сборки и деплоя в порядке.`

## Сохранение результатов

После завершения анализа выполни следующие шаги через инструменты:

1. Найди папку текущей сессии через Bash:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если вывод пустой — создай новую: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")` и используй её путь.
2. Сохрани отчёт через Write в файл: `<AUDIT_DIR>/audit-deployment.md`

Структура файла:
```
# Audit Report: Build & Deployment Configuration — <YYYY-MM-DD HH:MM>

<таблица с результатами или строка об отсутствии находок>
```

Сообщи пользователю путь к созданному файлу.
