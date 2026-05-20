---
name: audit-logging
description: >
  Аудит качества логирования: избыточность, безопасность логов, отсутствие чувствительных данных,
  соответствие best practices. Запускай при /audit-logging или запросе проверить логи/логирование.
---

## Правило применимости (Relevance Rule)

Перед анализом оцени: содержит ли код вызовы logger, console.log/error/warn, запись в файлы логов, middleware для логирования или аудит-треки? Если анализируемый файл не содержит никакого логирования — верни пустой ответ без таблицы.

## Runtime Detection

До анализа определи runtime проекта:
```bash
cat package.json 2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print('Node.js:', list(d.get('dependencies',{}).keys())[:8])" 2>/dev/null || \
ls go.mod requirements.txt pyproject.toml Cargo.toml 2>/dev/null | head -3
```

⚠️ Этот чеклист оптимизирован для **Node.js/TypeScript**. При обнаружении другого runtime:
- Go → `context.Context` вместо `AbortSignal`, `SIGTERM handler` вместо `process.on`
- Python → `asyncio cancellation`, `signal.SIGTERM`
- Java/Spring → `@Transactional`, `ApplicationContext lifecycle`
- Для неизвестного runtime — JS-специфичные проверки помечай `🔍 UNVERIFIED`

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
| LOG-01 | Production-код не использует console.log/console.error напрямую |
| LOG-02 | PII не логируется (email, телефон, имена, адреса, финансовые данные) |
| LOG-03 | Секреты и токены не попадают в логи |
| LOG-04 | Запросы трассируются (Request ID или correlation ID сквозной) |
| LOG-05 | Формат логов структурирован (JSON) в production |
| LOG-06 | Критические операции логируются (auth, create, update, delete) |
| LOG-07 | User input санитизируется перед логированием (защита от log injection) |
| LOG-08 | Критические события безопасности логируются: успешный/неудачный вход, выход, изменение прав, массовая выгрузка данных [⚡ dynamic] |

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

**LOG-01 — Нет прямого console в production:**
- `console.log`, `console.error`, `console.warn` в production-коде вместо структурированного логгера
- Логирование каждой итерации цикла без level-контроля
- Verbose-логи без флага условия (должны быть за `if (logger.isDebugEnabled())`)

**LOG-02 — PII не логируется:**
- Логирование email, телефонов, имён, адресов, дат рождения
- Логирование полных тел запросов/ответов с персональными данными
- Финансовые данные (суммы транзакций с привязкой к персоне) в логах

**LOG-03 — Секреты не в логах:**
- Пароли, API-ключи, JWT токены в log messages
- Полные строки подключения с credentials
- Секреты в stack traces при логировании ошибок

**LOG-04 — Сквозная трассируемость запросов:**
- Логи без request ID / correlation ID — невозможно отследить цепочку
- Request ID не пробрасывается в дочерние сервисы
- Логи без user ID или session context для аутентифицированных операций

**LOG-05 — Структурированный формат в production:**
- Plain text логи вместо JSON в production-среде
- Разный формат логов в разных частях приложения
- Вложенные объекты сериализуются как `[object Object]`

**LOG-06 — Критические операции логируются:**
- Операции аутентификации (login, logout, password change) без логов
- Создание/обновление/удаление критических сущностей без audit trail
- Отсутствие логов ошибок в catch-блоках критических операций

**LOG-07 — Защита от log injection:**
- User input передаётся в лог напрямую без санитизации
- Newline characters (`\n`, `\r`) в user input могут создавать поддельные log entries
- ANSI escape codes из user input могут повредить log formatters

**LOG-08 — Security audit trail:**
- Успешная и неудачная аутентификация не логируется → невозможен forensics после инцидента
- Изменение прав пользователя (role change, permission grant/revoke) без audit записи
- Mass data export (выгрузка > N записей, bulk delete) без лога кто/когда/что
- Password change / email change без записи в аудит-лог
- Критические административные операции выполняются без trace

## Формат вывода

| Check ID | Проверка | Статус | Уверенность | Доказательство | Решение | Исправлено |
|----------|----------|--------|-------------|----------------|---------|------------|
| LOG-01 | Production-код не использует console.log/console.error напрямую | ✅ PASS | High | `src/` grep — console.* не найдено | — | — |
| LOG-02 | PII не логируется (email, телефон, имена, адреса, финансовые данные) | ❌ FAIL 🔴 | High | `auth/login.ts:34` | **1. Удалить email из лога, логировать только userId** \\ 2. Маскировать PII через log sanitizer \\ 3. Заменить на структурированный лог без PII | Нет |
| LOG-04 | Запросы трассируются (Request ID или correlation ID сквозной) | ⏸ ACCEPTED | Medium | — | В baseline: трейсинг через внешний сервис | — |

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

Если все PASS — выведи: `✅ Логирование соответствует best practices.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-logging.md`

```
# Audit Report: Logging Best Practices — <YYYY-MM-DD HH:MM>
<таблица>
```
