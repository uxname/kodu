---
name: skills-best-practices
description: ...
---

## Что отличает хороший Skill от обычного промпта

Большинство людей пишут Skill как инструкцию для чат-бота:

> Ты эксперт по X. Помогай пользователю.

Для агента это плохо.

Хороший Skill должен описывать:

1. Когда его использовать.
2. Когда его НЕ использовать.
3. Как принимать решения.
4. Какие инструменты применять.
5. Как действовать при неопределённости.
6. Как выглядит хороший результат.
7. Как выглядит плохой результат. ([AgentGuide][2])

Фактически Skill больше похож на мини-ТЗ или SOP (Standard Operating Procedure), чем на промпт.

## Структура, которую я считаю оптимальной

```text
# SKILL: название

## PURPOSE
Зачем существует этот навык.

## ACTIVATION
Когда нужно применять.

## DO NOT USE WHEN
Когда нельзя применять.

## INPUTS
Что ожидается на входе.

## PROCESS
Пошаговый алгоритм действий.

## DECISION RULES
Правила принятия решений.

## TOOL USAGE
Как использовать инструменты.

## OUTPUT REQUIREMENTS
Каким должен быть результат.

## FAILURE HANDLING
Что делать при нехватке данных.

## EXAMPLES
Примеры хорошей работы.
```

Такая структура намного стабильнее огромных простыней текста. ([AgentGuide][2])

## Самые важные паттерны

### 1. Разделять роль и ограничения

Плохо:

```text
Ты эксперт по Kubernetes.
```

Лучше:

```text
ROLE:
Эксперт по Kubernetes.

CONSTRAINTS:
Никогда не придумывай параметры команд.
Если информации недостаточно — запроси данные.
```

Этот паттерн постоянно встречается в production-агентах. ([Reddit][3])

### 2. Явно прописывать ветку "не знаю"

Плохо:

```text
Ответь на вопрос.
```

Лучше:

```text
Если данных недостаточно:
- остановись
- перечисли недостающие данные
- не делай предположений
```

Это резко снижает галлюцинации. ([Reddit][3])

### 3. Использовать позитивные инструкции

Плохо:

```text
Не будь многословным.
Не пиши воду.
Не усложняй.
```

Лучше:

```text
Используй 3-7 предложений.
Давай конкретные рекомендации.
```

Модели лучше выполняют позитивные инструкции. ([OpenAI Help Center][1])

### 4. Давать примеры

Один хороший пример часто полезнее 100 строк инструкций.

```text
INPUT:
...

GOOD OUTPUT:
...

BAD OUTPUT:
...
```

Few-shot примеры остаются одной из самых эффективных техник. ([OpenAI Help Center][1])

### 5. Использовать явную иерархию

Плохой Skill:

```text
20 правил вперемешку.
```

Хороший Skill:

```text
PRIORITY 1
Безопасность

PRIORITY 2
Точность

PRIORITY 3
Полезность
```

Это уменьшает конфликты инструкций. ([OpenAI Platform][4])

### 6. Описывать инструменты отдельно

Очень частая ошибка:

```text
Используй поиск если нужно.
```

Правильно:

```text
SEARCH TOOL

Используй когда:
- нужны свежие данные
- нужна проверка фактов

Не используй когда:
- вопрос полностью покрывается контекстом

После поиска:
- процитируй источники
```

Для агентов описание инструмента зачастую важнее описания роли. ([Agent Mag][5])

## Антипаттерны

Почти всегда ухудшают качество:

* длинная простыня без структуры;
* десятки запретов подряд;
* расплывчатые слова вроде "будь умным";
* отсутствие примеров;
* отсутствие критериев остановки;
* отсутствие инструкции для случая неопределённости;
* смешивание бизнес-логики и форматирования в одном разделе. ([AgentGuide][2])

## Что я буду делать при создании Skill для тебя

Я буду генерировать Skills примерно такого уровня:

```text
# SKILL: Kubernetes Troubleshooter

## ACTIVATION
Использовать при вопросах о Kubernetes,
kubectl, Helm, ingress, networking.

## DO NOT USE WHEN
Вопрос не связан с Kubernetes.

## PROCESS

1. Определи компонент.
2. Определи симптом.
3. Собери недостающие данные.
4. Предложи диагностику.
5. Только потом предлагай исправление.

## FAILURE HANDLING

Если логов недостаточно:
запроси конкретные команды.

## OUTPUT FORMAT

Симптом
Причина
Проверка
Исправление
Риски
```

То есть не просто «промпт», а полноценный операционный регламент для агента.

## YAML FRONTMATTER БЛОК

В начале файла `SKILL.md` используй YAML frontmatter. Он задаёт метаданные навыка и помогает агенту быстро понять, что это за skill и когда его применять.

Пример:

```yaml
---
name: post-call-task-builder
description: Analyze meeting transcripts and project codebase, extract decisions and action items, propose Pareto-optimal implementation approaches, ask required clarification questions, then generate detailed technical specifications and development tasks. Use when user provides a meeting transcript, call notes, discussion log, or asks to create tasks/TZ/backlog from a team discussion.
---
```

Правила:

* `name` — короткое имя навыка в `kebab-case`.
* `description` — чётко описывает назначение навыка и триггеры его использования.
* YAML-блок должен стоять в самом начале файла, до основного текста.
* Не добавляй внутрь YAML длинные инструкции, логику или примеры — это должно быть в основном содержимом `SKILL.md`.
* Если инструмент поддерживает дополнительные поля, добавляй их только когда они реально нужны.


[1]: https://help.openai.com/en/articles/6654000-playground-and-prompt-engineering?utm_source=chatgpt.com "Best practices for prompt engineering with the OpenAI API | OpenAI Help Center"
[2]: https://agentguides.dev/prompt-engineering/?utm_source=chatgpt.com "Prompt Engineering for AI Agents: Techniques and Best Practices | AI Agents & Agentic Workflows Guide"
[3]: https://www.reddit.com/r/PromptEngineering/comments/1t63e41/guide_8_prompt_patterns_we_use_in_production_ai/?utm_source=chatgpt.com "[Guide] 8 prompt patterns we use in production AI agents (lessons from shipping 22+ projects in 2025)"
[4]: https://platform.openai.com/docs/guides/prompt-engineering?utm_source=chatgpt.com "Prompt engineering | OpenAI API"
[5]: https://agentmag.dev/resources/complete-prompt-engineering-guide-for-agents?utm_source=chatgpt.com "The Complete Prompt Engineering Guide for Agents — Prompt Guides"
