---
name: kodu-ops
description: Use when a user asks in plain language to deploy, update environment variables, add/remove domains, or inspect server/app status on dev/prod, and these intents should be translated into `kodu ops` commands.
---

# Kodu Ops

## Overview

Справочник для `kodu ops`: как переводить пользовательские запросы в корректные команды `sysinfo`, `env`, `routes`, `service`.

Ключевой принцип: `kodu ops` должен возвращать JSON-ответы без интерактивщины; агент выбирает подкоманды и флаги точно по контракту.

## When To Use

Используйте этот скилл, когда нужно:

- Перевести бытовой/джуновский запрос на инфраструктуру в `kodu ops` (без ожидания, что пользователь знает CLI).
- Получить диагностику удаленного сервера (`sysinfo`).
- Прочитать или изменить `.env` проекта (`env`).
- Посмотреть или изменить маршруты Caddy (`routes`).
- Управлять lifecycle сервиса через Docker Compose (`service`).
- Выполнить несколько ops-действий в одном запросе в строгой последовательности.

Не используйте, если задача не про `kodu ops` (например, локальная сборка/линт/git без удаленных операций).

## Core Rules

- Всегда используйте именованные флаги; positional arguments не поддерживаются.
- Не ожидайте, что пользователь назовет `kodu ops` или `--action`; извлекайте intent из обычной речи.
- Для каждого шага проверяйте обязательные флаги до запуска следующего.
- В цепочке "сначала ... потом ..." выполняйте команды последовательно.
- В ошибках ориентируйтесь на JSON-поля `status`, `code` и `error`/`stderr`.
- Если пользователь дал неполные данные, запрашивайте только отсутствующие обязательные параметры.

## Quick Reference

| Подкоманда | Назначение | Действия | Обязательные флаги |
| --- | --- | --- | --- |
| `sysinfo` | Диагностика сервера | - | `--server` |
| `env` | Работа с `.env` проекта | `get`, `set`, `unset` | `--server --action --project` (+ `--key` для `set/unset`, + `--val` для `set`) |
| `routes` | Работа с Caddy маршрутами | `list`, `add`, `remove`, `update` | `--server --action` (+ `--domain` для `add/remove/update`, + `--upstream` для `add/update`) |
| `service` | Управление сервисом | `clone`, `pull`, `up`, `down`, `logs`, `status` | `--server --action --project` (+ `--repo` для `clone`) |

## Command Details

### `sysinfo`

Назначение: собрать `uptime`, загрузку диска `/` и свободную память.

Флаги:

- `-s, --server <name>` - алиас сервера из `kodu.json`.

Пример:

```bash
kodu ops sysinfo --server dev
```

### `env`

Назначение: читать и изменять `.env` файла проекта в `ops.servers.<alias>.paths.apps/<project>/.env`.

Флаги:

- `-s, --server <name>`
- `-a, --action <get|set|unset>`
- `-p, --project <name>`
- `--key <key>` - обязателен для `set` и `unset`
- `--val <value>` - обязателен только для `set`

Примеры:

```bash
kodu ops env --server dev --action get --project my-app
kodu ops env --server dev --action set --project my-app --key PORT --val 3123
kodu ops env --server dev --action unset --project my-app --key PORT
```

Ограничения:

- Ключ проходит валидацию: `[A-Za-z_][A-Za-z0-9_]*`.

### `routes`

Назначение: читать и редактировать блоки доменов в `Caddyfile` (reverse proxy).

Флаги:

- `-s, --server <name>`
- `-a, --action <list|add|remove|update>`
- `--domain <domain>` - обязателен для `add`, `remove`, `update`
- `--upstream <host:port>` - обязателен для `add`, `update`

Примеры:

```bash
kodu ops routes --server dev --action list
kodu ops routes --server dev --action add --domain api.example.com --upstream 127.0.0.1:3000
kodu ops routes --server dev --action update --domain api.example.com --upstream 127.0.0.1:4000
kodu ops routes --server dev --action remove --domain api.example.com
```

Поведение:

- После `add/remove/update` автоматически применяется `./caddy.sh`.
- Если маршрут не найден при `remove/update`, ожидайте ошибку с кодом `NOT_FOUND`.

### `service`

Назначение: управлять проектом в директории apps через `git` и `docker compose`.

Флаги:

- `-s, --server <name>`
- `-a, --action <clone|pull|up|down|logs|status>`
- `-p, --project <name>`
- `--repo <url>` - обязателен только для `clone`

Примеры:

```bash
kodu ops service --server dev --action clone --project temp --repo https://github.com/org/repo.git
kodu ops service --server dev --action up --project temp
kodu ops service --server dev --action status --project temp
kodu ops service --server dev --action logs --project temp
kodu ops service --server dev --action pull --project temp
kodu ops service --server dev --action down --project temp
```

Фактические действия:

- `clone` -> `git clone <repo> <projectPath>` (ошибка, если директория уже существует)
- `pull` -> `git pull`
- `up` -> `docker compose up -d`
- `down` -> `docker compose down`
- `logs` -> `docker compose logs --no-color --tail=200`
- `status` -> `docker compose ps --format json` (fallback на обычный `docker compose ps`)

## JSON Output Contract

Успех:

- Всегда есть `status: "ok"`.
- Полезная нагрузка приходит в `data` или `message` в зависимости от подкоманды.

Ошибка:

- Всегда есть `status: "error"`.
- Поле `code` может быть строкой (например, `VALIDATION_ERROR`, `NOT_FOUND`) или числовым exit-code SSH.
- Текст ошибки приходит как `error` (CLI-валидация) или `stderr` (SSH/remote ошибки).

## Multi-Step Intent Mapping

Если пользователь описывает цепочку действий, раскладывайте ее на команды в указанном порядке.

Частые разговорные формулировки:

- "залей на прод/дев" -> обычно `service pull` + `service up` (или `service clone` + `service up`, если проекта еще нет).
- "добавь домен" -> `routes --action add --domain ... --upstream ...`.
- "поменяй домен на новый порт" -> `routes --action update --domain ... --upstream ...`.
- "посмотри что с сервисом" -> `service --action status` (+ при необходимости `service --action logs`).
- "поставь переменную" -> `env --action set --key ... --val ...`.

Пример:

```text
"Покажи sysinfo на dev, поставь PORT=3123 в temp и добавь маршрут api.example.com -> 127.0.0.1:3000"
```

Порядок команд:

1. `kodu ops sysinfo --server dev`
2. `kodu ops env --server dev --action set --project temp --key PORT --val 3123`
3. `kodu ops routes --server dev --action add --domain api.example.com --upstream 127.0.0.1:3000`

## Common Mistakes

- Пропуск `--project` для `service`/`env` (он обязателен даже для `clone`).
- Использование `pull` как обновления docker-образов: здесь `pull` делает `git pull` в проекте.
- Ожидание единого формата ошибки `error` во всех случаях: для SSH-ошибок обычно поле `stderr`.
- Передача key в неверном формате (например, `my.key`), что ломает валидацию.
- Попытка передать positional args вместо `--flags`.
