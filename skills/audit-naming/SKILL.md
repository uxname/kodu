---
name: audit-naming
description: >
  Аудит именования: читаемость кода, стандарты именования переменных, функций, классов, файлов.
  Запускай при /audit-naming или запросе проверить code style / naming conventions.
---

## Правило применимости (Relevance Rule)

Этот аудит применим к любому коду с идентификаторами. Пропускай только автогенерированные файлы (миграции, protobuf-generated, build output). Для конфигурационных файлов без кода (JSON, YAML) — верни пустой ответ.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| NAM-01 | camelCase для переменных/функций (JS/TS), snake_case (Python) |
| NAM-02 | Нет однобуквенных переменных вне for-циклов |
| NAM-03 | Boolean переменные с префиксом is/has/can |
| NAM-04 | Нет misleading names (getX не меняет данные) |
| NAM-05 | Magic numbers заменены именованными константами |
| NAM-06 | Нет utils.ts/helpers.ts как свалки |

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

**Плохие паттерны именования:**
- Однобуквенные переменные вне циклов (`d`, `x`, `tmp`)
- Аббревиатуры без расшифровки (`mgr`, `proc`, `srv`, `usr`)
- Misleading names (функция `getUser` меняет данные)
- Слишком общие имена (`data`, `info`, `manager`, `handler`, `util`)
- Отрицательные булевы (`isNotValid`, `notDisabled`)
- Несоответствие конвенции языка/фреймворка (snake_case в JS, camelCase в Python)
- Непоследовательное именование одной сущности в разных местах (`userId` vs `user_id` vs `uid`)
- Magic numbers без именованных констант
- Классы без существительного, функции без глагола

**Файловая структура:**
- Файлы с именами не отражающими содержимое
- `index.ts` с экспортом несвязанных сущностей
- `utils.ts`, `helpers.ts` как свалка

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| NAM-01 | camelCase для переменных/функций (JS/TS) | ✅ PASS | — | — |
| NAM-05 | Magic numbers заменены именованными константами | ❌ FAIL 🟡 | `utils/date.ts:18` | **1. Вынести в `const MAX_RETRY_ATTEMPTS = 3`** \\ 2. Добавить объяснение в комментарий рядом \\ 3. Переместить в конфиг |
| NAM-06 | Нет utils.ts/helpers.ts как свалки | ⏸ ACCEPTED | `src/utils.ts` | В baseline: рефактор запланирован |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

Если все PASS — выведи: `✅ Стандарты именования соблюдены.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-naming.md`

```
# Audit Report: Naming — <YYYY-MM-DD HH:MM>
<таблица>
```
