---
name: skills-best-practices
description: Design reusable AI agent skills in SKILL.md format — activation rules, constraints, workflow, output requirements, failure handling, and YAML frontmatter. Use when the user wants to create, improve, or standardize another skill for an AI agent.
---

## PURPOSE

Мета-скилл для проектирования других skills.
Часть 1 — принципы качества. Часть 2 — пошаговый процесс создания и улучшения.

---

# Часть 1. Принципы

## Skill vs промпт

Skill — это SOP, а не инструкция чат-боту. Хороший skill описывает:

1. Когда использовать / когда НЕ использовать.
2. Как принимать решения.
3. Какие инструменты применять.
4. Что делать при неопределённости.
5. Как выглядит хороший / плохой результат.

## 6 ключевых паттернов

| # | Паттерн | Плохо | Хорошо |
|---|---------|-------|--------|
| 1 | Роль + ограничения раздельно | «Ты эксперт по K8s» | ROLE + CONSTRAINTS |
| 2 | Ветка «не знаю» | «Ответь на вопрос» | Если данных нет — остановись, перечисли |
| 3 | Позитивные инструкции | «Не будь многословным» | «Используй 3-7 предложений» |
| 4 | Few-shot примеры | Только инструкции | INPUT / GOOD OUTPUT / BAD OUTPUT |
| 5 | Явная иерархия | 20 правил | PRIORITY 1 > 2 > 3 |
| 6 | Инструменты отдельно | «Используй поиск если нужно» | SEARCH TOOL: когда, когда нет, что после |

## Антипаттерны

- простыня без структуры;
- десятки запретов подряд;
- расплывчатые слова («будь умным», «по возможности»);
- отсутствие примеров, критериев остановки, ветки неопределённости;
- бизнес-логика и форматирование в одном разделе.

## What not to optimize

- **бесконечные примеры** — 1-2 пар GOOD/BAD достаточно, остальное в appendix;
- **перегруженный YAML** — только name + description, никакой логики;
- **skill-учебник** — не объясняй основы языка или технологии, это SOP, не туториал;
- **идеальную изоляцию** — повторяющийся паттерн лучше вынести в reference, чем дублировать;
- **waterfall-структуру** — если секция не нужна, оставь «(not applicable)», а не удаляй.

---

# Часть 2. Процесс

Сначала определи режим: **CREATE** (новый skill) или **IMPROVE** (существующий).

---

## Pre-flight gate

**Не начинай генерацию финального SKILL.md, пока не определены все критичные поля.**

Критичные (без них — STOP, запроси у пользователя):
- [ ] цель навыка
- [ ] платформа (OpenCode / Claude Code / универсальный)
- [ ] аудитория (разработчик / тимлид / QA / AI-агент)

Второстепенные (если нет — генерируй с явными допущениями):
- [ ] тип навыка (аналитический / генеративный / ревью / диагностика / утилита)
- [ ] нужны ли инструменты
- [ ] ожидаемый формат результата
- [ ] нужен ли контекст существующего проекта

Если хотя бы одно критичное поле отсутствует — **не пиши финальный файл, задай вопросы**.

---

## Skill Type Inference

Если тип навыка не указан пользователем — определи его до начала работы над brief.

| Тип | Характерный паттерн | На что влияет |
|-----|---------------------|---------------|
| Аналитический | Проверка, поиск проблем, сравнение | PROCESS: крайние случаи, ветвление по данным |
| Генеративный | Создание текста, кода, документации | OUTPUT REQUIREMENTS: точная структура результата |
| Диагностический | Поиск причины, расследование сбоев | FAILURE HANDLING: 3+ сценария отказа |
| Ревью | Оценка, классификация, вердикт | DECISION RULES: чёткая шкала severity |
| Утилита | Преобразование, запуск, миграция | TOOL USAGE: точные команды, флаги |

Если тип всё ещё неясен — выбери наиболее вероятный, запиши в brief и отметь как inferred.

---

## CREATE: новый skill

### Step 1. Skill Brief (единственный источник правды)

Заполни brief. Поля имеют жёсткий формат, а не произвольный текст.

```yaml
brief:
  purpose: '<глагол + что делает. Одна фраза.>'
  activation: ['<триггер 1>', '<триггер 2>']        # 2-4 элемента
  do_not_use: ['<условие 1>', '<условие 2>']          # 2+ элемента
  required_inputs: ['<поле 1>', '<поле 2>']           # обязательные + опциональные
  tools: ['<инструмент>: <когда>']                      # или пустой []
  output_format: '<структура результата. Один абзац.>'
  failure_handling: ['<сценарий 1>: <действие>', '<сценарий 2>: <действие>']  # 2+
  examples:                                            # 1-2 элемента
    - input: '<вход>'
      good: '<правильный ответ>'
      bad: '<неправильный ответ>'
```

Пока brief не заполнен целиком — не переходи к рендерингу.

### Step 2. YAML frontmatter

```yaml
---
name: <kebab-case>
description: <глагол + объект>. Use when <триггер>.
---
```

Формула description:
> **глагол в активном залоге + что делает + «. Use when / Use for» + триггер**

Правила:
- глагол + объект: «Analyze Node.js dependencies», «Review code changes»
- 1-2 триггера: «Use when reviewing package.json»
- **запрещены** `help with`, `assist`, `support` без конкретного действия
- без платформенных деталей (кроме strict single-platform)
- без логики, примеров, инструкций
- один абзац, макс 3 предложения

Тест на проверяемость: **прочитай только description. Сразу ясно, когда skill нужен, а когда нет?** Если нет — перепиши.

Запрещено:
```yaml
description: Helps with code stuff. Use when needed.
```
Правильно:
```yaml
description: Review code changes for bugs, security issues, and convention violations. Use when reviewing a PR or diff before merge.
```

### Step 3. Render SKILL.md

Каждое поле brief маппится на секцию. Порядок фиксирован. Это единственный шаблон.

| Brief field | → | Section | Формат содержимого |
|-------------|---|---------|-------------------|
| `purpose` | → | `## PURPOSE` | Один абзац |
| `activation` | → | `## ACTIVATION` | Маркированный список триггеров |
| `do_not_use` | → | `## DO NOT USE WHEN` | Маркированный список условий |
| `required_inputs` | → | `## INPUTS` | Маркированный список полей |
| _(выводится)_ | → | `## PROCESS` | Нумерованный список конкретных действий (не «проанализируй», а «собери данные» → «проверь ограничения» → «выбери вариант»), минимум 3 шага |
| _(выводится)_ | → | `## DECISION RULES` | Маркированный список приоритетов. Минимум одно правило приоритизации (PRIORITY 1 / 2 / 3) |
| `tools` | → | `## TOOL USAGE` | Описание или «(not applicable)» |
| `output_format` | → | `## OUTPUT REQUIREMENTS` | Один абзац или список |
| `failure_handling` | → | `## FAILURE HANDLING` | Маркированный список сценариев |
| `examples` | → | `## EXAMPLES` | INPUT / GOOD OUTPUT / BAD OUTPUT |

Обязательность секций:

| Секция | Всегда содержательна | Может быть «(not applicable)» |
|--------|---------------------|------------------------------|
| PURPOSE | да | нет |
| ACTIVATION | да | нет |
| DO NOT USE WHEN | да | нет |
| INPUTS | да | нет |
| PROCESS | да (>=3 шага) | нет |
| DECISION RULES | да | нет |
| TOOL USAGE | нет | да |
| OUTPUT REQUIREMENTS | да | нет |
| FAILURE HANDLING | да (>=2 сценария) | нет |
| EXAMPLES | нет (но >=1 если есть) | да |

### Step 4. Conflict check

1. **Межсекционные** — ACTIVATION не противоречит DO NOT USE; PROCESS не дублирует DECISION RULES.
2. **С запросом пользователя** — приоритет у запроса, но skill не нарушает принципы качества.
3. **Рантайм-конфликт** — если две инструкции могут противоречить, расставь PRIORITY явно.
4. **Двусмысленность** — нет фраз «по возможности», «если уместно», «будь внимателен».
5. **Длина** — skill должен быть настолько коротким, насколько возможно, но настолько длинным, насколько необходимо. Если документ содержит большие примеры, справочники или команды — рассмотри вынос в reference-файлы.

### Step 5. Modularity check

- [ ] Паттерн повторяется >2 раз? → вынеси в reference-файл
- [ ] Пример >30 строк? → сократи или вынеси в appendix
- [ ] Инструменты с длинными командами? → вынеси в TOOL USAGE reference
- [ ] Skill ссылается на внешние документы? → сделай автономным или пропиши зависимость явно
- [ ] Расползается? → сокращай в порядке: examples → process → пояснения

### Step 6. Quality check

Общие минимумы:
- PURPOSE — одно полное предложение с глаголом
- ACTIVATION — минимум 2 триггера
- DO NOT USE — минимум 2 условия отказа
- PROCESS — минимум 3 нумерованных шага с конкретными действиями, а не с оценками
- DECISION RULES — минимум одно правило приоритизации (PRIORITY 1 / 2 / 3)
- FAILURE HANDLING — минимум 2 сценария с конкретными действиями
- EXAMPLES — минимум 1 пара GOOD / BAD
- OUTPUT REQUIREMENTS — формат вывода описан так, что результат проверяем

Дополнительно по типу навыка:

| Тип | Приоритет | Особое внимание |
|-----|-----------|-----------------|
| Аналитический | Conflict detection | PROCESS покрывает крайние случаи (нет данных, битые данные) |
| Генеративный | Output format | OUTPUT REQUIREMENTS содержит точную структуру результата |
| Диагностический | Failure handling | FAILURE HANDLING покрывает 3+ сценария |
| Ревью | Severity classification | DECISION RULES чётко делит findings по severity |
| Утилита | Tool usage | TOOL USAGE содержит точные команды, флаги, примеры |

### Step 7. Three test scenarios

Если хоть один не проходит — вернись к Step 3.

| Кейс | Проверка |
|------|----------|
| Простой | Идеальные данные → корректный результат без лишних действий |
| Неполные | Данных не хватает → FAILURE HANDLING: запрос конкретных данных, без гадания |
| Вне зоны | Запрос вне DO NOT USE → отказ или делегирование, без выполнения |

### Step 8. Activation Coverage Test

Проверь, что skill срабатывает тогда, когда нужно, и молчит тогда, когда не нужно.

Сгенерируй:

- **3 should-trigger запроса** — пользователь явно находится в зоне skill
- **3 should-not-trigger запроса** — пользователь вне зоны skill или рядом, но не в ней

Проверь:

- все should-trigger запросы активируют skill (по ACTIVATION и description)
- все should-not-trigger запросы **не** активируют skill (попадают в DO NOT USE или не совпадают с триггерами)

Если тест не проходит — перепиши description и ACTIVATION триггеры.

### Step 9. Meta decision rules

- **Brief не закрыт** → не переходи к рендерингу.
- **Конфликт краткости и полноты** → в core секциях (PROCESS, FAILURE HANDLING) предпочитай полноту; в примерах — краткость.
- **Шаблон vs запрос** → приоритет у запроса. Если запрос нарушает принципы качества — объясни why и предложи компромисс.
- **Не знаешь, нужна ли секция** → добавь с пометкой «(not applicable)», а не удаляй.

---

## IMPROVE: улучшить существующий skill

### Step 0. Найди повторы и конфликты ДО переписывания

Прежде чем что-то менять, зафиксируй проблемные места. Не начинай править, пока не составлен полный список дефектов.

### Step 1. Полный аудит

- **Покрытие секций**: какие есть, каких нет.
- **Содержательность**: секция полезна или заглушка («...», «TBD»).
- **Конфликты**: инструкции противоречат друг другу.
- **Неопределённость**: есть ли фразы «по возможности», «если нужно».
- **Длина**: не превышает разумных пределов. Если есть большие примеры или справочники — их место в reference.
- **YAML**: name в kebab-case, description по формуле.
- **Примеры**: отражают реальные use-case или абстрактны.
- **FAILURE HANDLING**: покрывает типичные отказы или пустой.
- **Output contract**: описан ли формат результата.
- **Платформа**: соответствует ли целевой.

### Step 2. Классифицируй дефекты

| Категория | Дефекты | Действие |
|-----------|---------|----------|
| Missing | Нет секции, примеров, YAML | Добавить |
| Shallow | Секция-заглушка, абстрактные примеры | Заполнить по контексту |
| Conflict | PROCESS дублирует DECISION RULES | Переписать, расставить PRIORITY |
| Vague | «по возможности», «будь внимателен» | Заменить на конкретные условия |
| Bloated | Много повторов, большие примеры, справочники внутри | Сократить: examples → process → пояснения; вынести в reference
| Wrong platform | YAML/формат не под платформу | Переписать под целевую платформу |

### Step 3. Примени исправления

По каждому дефекту — одно действие из таблицы. Не меняй то, что работает.

### Step 4. Re-validate

Прогони через шаги 4-9 режима CREATE.

---

## Output contract (обязательный)

При любом ответе выдавай строго этот набор блоков:

```
## Assumptions
<допущения, сделанные в процессе>

## Brief / Analysis
<для CREATE: brief из Step 1>
<для IMPROVE: аудит из Step 1>

## Final SKILL.md
<готовый файл целиком, включая YAML>

## Validation
<результаты проверки: конфликты, модульность, качество (Step 6), 3 теста (Step 7) — пройдены/не пройдены>

## Open Questions
<вопросы, если остались>
```

Если пользователь просит только фрагмент — выдай его с пометкой «это неполный output».

---

# Platform compatibility

Этот мета-скилл проектирует skills. По умолчанию тело skill считается **платформо-независимым**.

Универсальные правила (всегда):
- YAML frontmatter с name + description
- Все секции PURPOSE–EXAMPLES присутствуют
- PROCESS нумерованный, FAILURE HANDLING конкретный
- OUTPUT REQUIREMENTS содержит проверяемый формат

Платформа → формат:

| Платформа | Формат файла | Особенности |
|-----------|-------------|-------------|
| **universal** (default) | `SKILL.md` | name + description, минимум полей |
| **opencode-only** | `SKILL.md` | description максимально лаконичный |
| **claude-code-only** | `CLAUDE.md` / `AGENTS.md` | Возможны доп. поля платформы |

Правила:
- Если платформа не указана → universal
- Если платформа требует специальных секций, ресурсов или структуры — адаптируй skill под неё, но в базовом виде тело считается платформо-независимым
- Если skill строго single-platform, отметь в description: «OpenCode-only: ...» или «Claude Code: ...»

---

# Canonical example

```markdown
---
name: code-reviewer
description: Review code changes for bugs, security issues, and convention violations. Use when reviewing a PR, diff, or changed files before merge.
---

## PURPOSE
Catch regressions, security holes, and style violations before code reaches production.

## ACTIVATION
- User asks to review a PR, diff, or changed files
- User says «review this code» / «check my changes»

## DO NOT USE WHEN
- Code is a generated artifact (protobuf, GraphQL schema)
- User explicitly asks for formatting only (use linter)
- File is lockfile, binary, or vendored dependency

## INPUTS
- git diff or list of changed files (or inline code snippet)
- (optional) severity threshold — skip INFO if asked

## PROCESS
1. Scan diff for security red flags: eval, exec, SQL injection, hardcoded secrets.
2. Check for logic bugs: off-by-one, null dereference, missing await.
3. Check code quality: dead code, excessive complexity, copy-paste.
4. Verify consistency with surrounding code (naming, patterns, imports).
5. Report findings grouped by severity.
6. For each finding: file:line, problem, suggested fix.

## DECISION RULES
- BLOCKER = security or data loss → always report
- CRITICAL = logic bug → breaks in production
- WARNING = maintainability concern
- INFO = style nit — skip if user said «quick review»

## TOOL USAGE
(not applicable — pure analysis)

## OUTPUT REQUIREMENTS
- Grouped: BLOCKER > CRITICAL > WARNING > INFO
- Each finding: location, problem, fix
- Summary line: X BLOCKER + Y CRITICAL + Z WARNING
- If zero findings: «No issues found — lgtm»

## FAILURE HANDLING
- No diff provided → ask for it, don't proceed
- File binary or >500 lines → warn, ask to narrow scope
- Language not recognised → generic review (logic + security only)

## EXAMPLES
INPUT: diff with `eval(userInput)` in express handler
GOOD OUTPUT: [BLOCKER] src/handler.ts:42 — eval() on user input allows arbitrary code execution. Fix: use safe JSON.parse or whitelist-based approach.
BAD OUTPUT: «This is potentially unsafe» — no explanation, no fix.
```

---

## Pareto Optimization (финальный шаг)

Перед выдачей результата:

1. Найди 20% инструкций, которые дают 80% качества (обычно это PROCESS, FAILURE HANDLING, DECISION RULES).
2. Проверь: можно ли удалить остальные 80% инструкций без существенной потери качества?
3. Если можно — сократи skill.

Большинство плохих skills становятся лучше после удаления 20-40% текста.
Не добавляй то, без чего skill будет работать так же хорошо.
```
