---
name: audit-architecture
description: >
  Аудит архитектуры и файловой структуры: корректность связей между слоями, нарушения
  dependency rules, структура папок, circular dependencies. Запускай при /audit-architecture.
---

## Правило применимости (Relevance Rule)

Применим к проектам с выраженной слоистой архитектурой (MVC, Clean Architecture, DDD, Hexagonal). Для single-file скриптов или утилит без архитектурного деления — верни пустой ответ.

## Чеклист

| Check ID | Проверка |
|----------|----------|
| ARC-01 | Бизнес-логика в service слое, не в route handlers |
| ARC-02 | Нет прямых DB вызовов из presentation layer |
| ARC-03 | Нет circular dependencies между модулями |
| ARC-04 | Нет God-файлов (>500 строк несвязанной логики) |
| ARC-05 | Config/env не используется вне config модуля |
| ARC-06 | Внешние зависимости инжектируются (DI) |

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

**Нарушения слоёв:**
- Бизнес-логика в контроллерах/роутерах
- Прямые обращения к БД из presentation layer
- Импорт infrastructure-зависимостей из domain/core
- Круговые зависимости между модулями
- God-классы/файлы (>500 строк с несвязанной логикой)

**Структура папок:**
- Модули без чёткой границы ответственности
- Нарушение co-location (связанный код разнесён далеко)
- Неправильное место для shared/common кода
- Feature folders vs Layer folders — непоследовательное применение

**Dependency direction:**
- Зависимость от конкретных реализаций вместо интерфейсов
- Отсутствие dependency injection для тестируемости
- Прямые импорты между feature-модулями (должны идти через shared)

**Coupling & Cohesion:**
- Высокое coupling (изменение одного модуля ломает другие)
- Низкая cohesion (модуль делает несвязанные вещи)
- Leaky abstractions (внутренние детали протекают наружу)

## Формат вывода

| Check ID | Проверка | Статус | Доказательство | Решение |
|----------|----------|--------|----------------|---------|
| ARC-01 | Бизнес-логика в service слое, не в route handlers | ✅ PASS | — | — |
| ARC-03 | Нет circular dependencies между модулями | ❌ FAIL 🟠 | `modules/order/index.ts:3` | **1. Вынести общий код в shared модуль** \\ 2. Применить dependency inversion \\ 3. Разбить модуль на независимые части |
| ARC-06 | Внешние зависимости инжектируются (DI) | ⏸ ACCEPTED | `services/payment.ts:1` | В baseline: refactor запланирован в Q3 |

Статусы: `✅ PASS` / `❌ FAIL 🔴` / `❌ FAIL 🟠` / `❌ FAIL 🟡` / `❌ FAIL 🟢` / `⏸ ACCEPTED`

Для `❌ FAIL`: ровно 3 варианта решения, разделитель `\\`, вариант 1 жирным.

Если все PASS — выведи: `✅ Архитектурные принципы соблюдены.`

## Сохранение результатов

1. Найди папку сессии:
   ```bash
   ls -dt ./docs/audits/[0-9]*/ 2>/dev/null | head -1 | sed 's|/$||'
   ```
   Если пусто — создай: `mkdir -p ./docs/audits/$(date +"%Y-%m-%d_%H-%M")`
2. Сохрани через Write: `<AUDIT_DIR>/audit-architecture.md`

```
# Audit Report: Architecture & File Structure — <YYYY-MM-DD HH:MM>
<таблица>
```
