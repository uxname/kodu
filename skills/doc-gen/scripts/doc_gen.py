#!/usr/bin/env python3
"""
doc_gen.py — генератор и валидатор документации продукта.

Использование:
  python3 doc_gen.py generate      "НазваниеПроекта" [--only L1|L2] [--update] [--output PATH]
  python3 doc_gen.py validate      "НазваниеПроекта" [--output PATH]
  python3 doc_gen.py consistency   "НазваниеПроекта" [--output PATH]
  python3 doc_gen.py status       ["НазваниеПроекта"] [--output PATH]
  python3 doc_gen.py update-status "НазваниеПроекта" "статус" [--output PATH]
"""

import argparse
import re
import sys
from datetime import date
from pathlib import Path


# ─── Структура документов ─────────────────────────────────────────────────────

STRUCTURE = {
    "INDEX.md": {
        "description": "Оглавление",
        "headings": ["## Навигация", "## Быстрые ссылки"],
    },
    "1_PRODUCT_VISION/VISION.md": {
        "description": "Концепция продукта",
        "headings": [
            "## Проблема", "## Целевая аудитория", "## Цель",
            "## Ключевые возможности", "## Метрики успеха",
            "## Что входит", "## Что НЕ входит",
        ],
        "min_section_chars": 60,
        "check_metrics": True,
    },
    "2_PRODUCT_SPEC/SPEC.md": {
        "description": "Спецификация продукта",
        "headings": [
            "## Ссылки", "## Как устроена система", "## Глоссарий",
            "## Сущности", "## Страницы и экраны", "## Ключевые операции",
            "## Интеграции", "## Тестирование", "## Артефакты",
        ],
        "min_section_chars": 60,
    },
}

# Секции, освобождённые от проверки минимальной длины
_SECTION_LEN_EXEMPT = {
    "## Ссылки", "## Навигация", "## Быстрые ссылки", "## Интеграции",
    "## Артефакты",
}
_MIN_SECTION_CHARS = 60

STATUS_PATTERN      = re.compile(r'\*\*Статус:\*\*\s*(черновик|на ревью|утверждён)')
DATE_PATTERN        = re.compile(r'\*\*Дата:\*\*\s*\d{4}-\d{2}-\d{2}')
PLACEHOLDER_PATTERN = re.compile(r'\[[^\]]+\](?!\()')
_UNSAFE_CHARS       = re.compile(r'[/\\:*?"<>|]')

VALID_STATUSES = {"черновик", "на ревью", "утверждён"}

# Слова/фразы, запрещённые в содержимом документа (расплывчатые, рекламные, опциональные).
# Нормализованы: е вместо ё (проверка идёт через _norm()).
_FORBIDDEN: list[str] = [
    # Опциональность
    "при необходимости", "по желанию", "при желании", "может быть",
    "возможно", "опционально", "наверное", "вероятно",
    # Расплывчатые качества
    "быстро", "быстрый", "быстрая", "быстрое",
    "удобно", "удобный", "удобная", "удобное",
    "легко", "легкий", "легкая", "легкое",
    "просто", "простой", "простая", "простое",
    "интуитивно", "интуитивный", "интуитивная",
    "гибко", "гибкий", "гибкая",
    "современный", "современная", "современное",
    "инновационный", "инновационная",
    "лучший", "лучшая", "лучшее",
    "оптимальный", "оптимальная", "оптимальное",
    "эффективно", "эффективный", "эффективная",
    "качественный", "качественная", "качественное",
]

# Стоп-слова для анализа ключевых слов (согласованность)
_STOPWORDS: set[str] = {
    "будет", "этого", "этот", "этой", "этом", "этому", "этими",
    "который", "которой", "которая", "которые", "которых", "которым", "которого",
    "также", "чтобы", "должен", "может", "каждый", "каждая", "каждое",
    "должна", "должно", "через", "после", "перед", "между",
    "пользователь", "пользователя", "пользователей", "пользователи",
    "системы", "система", "продукт", "продукта", "список", "раздел",
    "данные", "данных", "только", "более", "менее", "того",
    "всего", "всех", "всем", "свою", "своих", "своем",
    "такой", "такая", "такое", "такие", "таких", "таким",
    "одного", "одному", "одним", "один", "одна", "одно",
    "иметь", "делать", "делает", "делают", "выполнять", "выполняет",
    "включает", "включают", "содержит", "содержат",
    "возможность", "возможности", "функция", "функции", "функционал",
    "позволяет", "нужно", "нужен", "нужна", "нужны", "можно", "нельзя",
    "является", "являются",
    "that", "this", "with", "from", "have", "will", "been",
    "they", "their", "which", "when", "user", "users", "system",
}


# ─── Цвета и вывод ────────────────────────────────────────────────────────────

class C:
    RED    = "\033[0;31m"
    GREEN  = "\033[0;32m"
    YELLOW = "\033[1;33m"
    RESET  = "\033[0m"


def ok(msg: str)   -> None: print(f"{C.GREEN}✓{C.RESET} {msg}")
def err(msg: str)  -> None: print(f"{C.RED}✗{C.RESET} {msg}")
def warn(msg: str) -> None: print(f"{C.YELLOW}⟳{C.RESET} {msg}")


def _validate_name(name: str) -> list[str]:
    issues = []
    if not name.strip():
        issues.append("Имя проекта не может быть пустым.")
    if _UNSAFE_CHARS.search(name):
        issues.append('Имя проекта содержит недопустимые символы: / \\ : * ? " < > |')
    if " " in name:
        issues.append(
            f"Имя проекта содержит пробелы. Используйте CamelCase или дефис: "
            f"«{name.replace(' ', '')}» или «{name.replace(' ', '-')}»"
        )
    return issues


# ─── Вспомогательные функции ──────────────────────────────────────────────────

def _norm(text: str) -> str:
    """Нормализация: ё→е, нижний регистр."""
    return text.replace("ё", "е").replace("Ё", "Е").lower()


def _section(content: str, heading: str) -> str:
    """Текст секции markdown от heading до следующего заголовка того же или выше уровня."""
    pattern = re.compile(rf"^{re.escape(heading)}\b.*$", re.MULTILINE)
    match = pattern.search(content)
    if not match:
        return ""
    start = match.end()
    level = len(re.match(r"^(#+)", heading).group(1))
    next_h = re.search(r"^#{1," + str(level) + r"} ", content[start:], re.MULTILINE)
    end = start + next_h.start() if next_h else len(content)
    return content[start:end].strip()


def _keywords(text: str) -> list[str]:
    """Значимые слова: минимум 5 символов, не из стоп-списка."""
    words = re.findall(r"\b[а-яёА-ЯЁa-zA-Z]{5,}\b", text)
    return [_norm(w) for w in words if _norm(w) not in _STOPWORDS]


def _list_items(text: str) -> list[str]:
    """Текст элементов маркированного/нумерованного списка."""
    items = []
    for line in text.splitlines():
        m = re.match(r"^\s*(?:\d+\.|[-*•])\s+(.+)", line)
        if m:
            item = re.sub(r"\*+([^*]+)\*+", r"\1", m.group(1)).strip()
            items.append(item)
    return items


def _table_first_col(section_text: str, skip_headers: set[str] | None = None) -> list[str]:
    """Значения первого столбца таблицы (пропускает заголовок и разделитель)."""
    if skip_headers is None:
        skip_headers = set()
    result = []
    for line in section_text.splitlines():
        if "|" not in line:
            continue
        if re.match(r"^\s*\|[-: |]+\|\s*$", line):
            continue
        cols = [c.strip() for c in line.strip("|").split("|")]
        if cols and cols[0] and cols[0] not in skip_headers:
            result.append(cols[0])
    return result


def _check_metrics(content: str, rel_path: str) -> list[str]:
    """Проверяет, что целевые значения метрик содержат числа."""
    errors = []
    metrics_sec = _section(content, "## Метрики успеха")
    if not metrics_sec:
        return errors
    for line in metrics_sec.splitlines():
        if "|" not in line:
            continue
        if re.match(r"^\s*\|[-: |]+\|\s*$", line):
            continue
        cols = [c.strip() for c in line.strip("|").split("|")]
        if not cols or cols[0] in ("Метрика", "Metric", ""):
            continue
        val_col = cols[1].strip() if len(cols) > 1 else ""
        if not re.search(r"\d", val_col):
            errors.append(
                f"{rel_path}: метрика «{cols[0]}» — "
                f"целевое значение «{val_col}» не содержит числа"
            )
    return errors


# ─── Шаблоны файлов ───────────────────────────────────────────────────────────

def _today() -> str:
    return date.today().isoformat()


def _index_template(name: str) -> str:
    return f"""\
# Документация продукта: {name}

**Статус:** черновик | **Дата:** {_today()}

## Навигация

| Документ | Описание |
|----------|----------|
| [Концепция](./1_PRODUCT_VISION/VISION.md) | Проблема, аудитория, цель, границы проекта |
| [Спецификация](./2_PRODUCT_SPEC/SPEC.md) | Сущности, страницы, операции, тестирование |

## Быстрые ссылки

- [Ключевые возможности](./1_PRODUCT_VISION/VISION.md#ключевые-возможности)
- [Страницы и экраны](./2_PRODUCT_SPEC/SPEC.md#страницы-и-экраны)
- [Тестирование](./2_PRODUCT_SPEC/SPEC.md#тестирование)
- [Артефакты](./2_PRODUCT_SPEC/SPEC.md#артефакты)
"""


def _vision_template(name: str) -> str:
    return f"""\
# {name}

**Статус:** черновик | **Дата:** {_today()}

## Проблема
[Что конкретно неудобно или не работает. Контекст. 2–4 предложения.]

## Целевая аудитория
[Кто пользователь: роль и контекст работы. Не «все пользователи», а конкретный тип. 2–4 предложения.]

## Цель
[Что именно создаём и для кого. 2–4 предложения.]

## Ключевые возможности
1. **[Название]**: пользователь выполняет [X] → получает [Y]
2. **[Название]**: пользователь выполняет [X] → получает [Y]

## Метрики успеха
| Метрика | Целевое значение |
|---------|------------------|
| [Метрика] | [конкретное число] |

## Что входит (границы проекта)
Функциональность, которая будет реализована:
1. [Функциональность 1]

## Что НЕ входит
Явные исключения для предотвращения расширения границ проекта:
- Не включает [X]
- Не включает [Y]
"""


def _spec_template(name: str) -> str:
    return f"""\
# Продуктовая спецификация: {name}

**Статус:** черновик | **Дата:** {_today()}

## Ссылки
- Концепция: [VISION.md](../1_PRODUCT_VISION/VISION.md)

## Как устроена система
[Краткое описание из каких частей состоит продукт и как они взаимодействуют.
Пример: «Веб-приложение с личным кабинетом и административной панелью. Данные хранятся
централизованно, доступ — через браузер без установки приложений.»]

## Глоссарий
Ключевые понятия продукта. Каждый термин имеет одно точное определение без синонимов.

| Термин | Определение |
|--------|-------------|
| [Термин] | [Точное определение] |

## Сущности
Основные объекты, с которыми работает система. В терминах бизнеса, не базы данных.

| Сущность | Описание | Ключевые свойства |
|----------|----------|-------------------|
| Пользователь | [Описание] | [Свойство 1, Свойство 2] |

### Жизненный цикл сущностей
Для каждой ключевой сущности — допустимые статусы и переходы между ними.

| Сущность | Статусы | Переходы |
|----------|---------|----------|
| [Сущность] | [статус1 → статус2 → статус3] | [статус1→статус2: условие перехода] |

## Страницы и экраны
Исчерпывающий список страниц и экранов, которые должны быть созданы.

| Страница | Назначение | Ключевые элементы |
|----------|------------|-------------------|
| Главная | [Назначение] | [Элемент 1, Элемент 2] |
| Регистрация | [Назначение] | [Элемент 1, Элемент 2] |

## Ключевые операции
Что пользователи могут делать в системе.

**[Роль или «Все пользователи»]:**
- [Операция]: [краткое описание результата]

## Интеграции
Внешние сервисы, без которых продукт не работает.
Если интеграций нет — удалить таблицу и написать: «Интеграций нет.»

| Сервис | Назначение |
|--------|------------|
| [Название] | [Зачем нужен] |

## Тестирование
Функциональность, покрытая тестами.

**Критические сценарии** (обязаны работать без ошибок):
- [Пользователь выполняет X → система возвращает Y]

**Бизнес-правила** (корректность расчётов и ограничений):
- [Правило или расчёт]

**Негативные сценарии** (поведение системы при ошибках и отменах):
- [Пользователь выполняет X с ошибочными данными → система возвращает сообщение Z, действие не выполняется]

## Артефакты
Вспомогательные материалы для разработки и запуска продукта.
Если артефактов нет — написать: «Артефактов нет.»

Артефактов нет.
"""


# ─── Генерация ────────────────────────────────────────────────────────────────

def _should_write(path: Path, update_mode: bool) -> bool:
    if update_mode and path.exists():
        warn(f"Пропущен (уже существует): {path.name}")
        return False
    return True


def cmd_generate(args: argparse.Namespace) -> int:
    name_errors = _validate_name(args.name)
    if name_errors:
        for e in name_errors:
            err(e)
        return 1

    output_dir = Path(args.output)
    target_dir = output_dir / args.name
    only = (args.only or "").upper()

    if not args.update and target_dir.exists() and not only:
        print(f"{C.RED}Ошибка:{C.RESET} папка {target_dir} уже существует.")
        print("Используйте --update для дополнения без перезаписи.")
        return 1

    target_dir.mkdir(parents=True, exist_ok=True)
    print(f"\n{C.GREEN}📁{C.RESET} {args.name} → {target_dir}\n")

    # INDEX.md
    index_path = target_dir / "INDEX.md"
    if _should_write(index_path, args.update):
        index_path.write_text(_index_template(args.name), encoding="utf-8")
        ok("INDEX.md")

    # L1 — Концепция
    if not only or only == "L1":
        l1_dir = target_dir / "1_PRODUCT_VISION"
        l1_dir.mkdir(exist_ok=True)
        vision_path = l1_dir / "VISION.md"
        if _should_write(vision_path, args.update):
            vision_path.write_text(_vision_template(args.name), encoding="utf-8")
            ok("1_PRODUCT_VISION/VISION.md")

    # L2 — Спецификация
    if not only or only == "L2":
        l2_dir = target_dir / "2_PRODUCT_SPEC"
        l2_dir.mkdir(exist_ok=True)
        spec_path = l2_dir / "SPEC.md"
        if _should_write(spec_path, args.update):
            spec_path.write_text(_spec_template(args.name), encoding="utf-8")
            ok("2_PRODUCT_SPEC/SPEC.md")

    # 3_ARTIFACTS/ — НЕ создаётся автоматически.
    # Папку и подпапки создавать только при размещении реального файла-артефакта.

    print(f"\n{C.GREEN}✅ Готово:{C.RESET} {target_dir}")
    print(f"{C.YELLOW}Следующий шаг:{C.RESET} заполните документы, затем запустите валидацию:")
    print(f"  python3 <путь-к-скиллу>/scripts/doc_gen.py validate {args.name!r} --output {args.output!r}\n")
    return 0


# ─── Валидация структуры ──────────────────────────────────────────────────────

def _check_file(rel_path: str, file_path: Path, spec: dict) -> list[str]:
    errors: list[str] = []

    if not file_path.exists():
        return [f"Файл отсутствует: {rel_path}"]

    content = file_path.read_text(encoding="utf-8")
    lines   = content.splitlines()

    # Статус и дата
    if not STATUS_PATTERN.search(content):
        errors.append(f"{rel_path}: строка «**Статус:** черновик|на ревью|утверждён» отсутствует")
    if not DATE_PATTERN.search(content):
        errors.append(f"{rel_path}: строка «**Дата:** YYYY-MM-DD» отсутствует")

    # Обязательные разделы
    for heading in spec["headings"]:
        if not any(line.strip().startswith(heading) for line in lines):
            errors.append(f"{rel_path}: отсутствует обязательный раздел «{heading}»")

    # Заглушки + запрещённые слова (построчно, вне code-блоков)
    in_code_block = False
    for lineno, line in enumerate(lines, 1):
        stripped = line.strip()
        if stripped.startswith("```"):
            in_code_block = not in_code_block
            continue
        if in_code_block:
            continue

        if PLACEHOLDER_PATTERN.search(stripped):
            errors.append(f"{rel_path}:{lineno}: незаполненная заглушка → {stripped[:80]}")

        line_norm = _norm(stripped)
        for phrase in _FORBIDDEN:
            phrase_norm = _norm(phrase)
            if " " in phrase_norm:
                if phrase_norm in line_norm:
                    errors.append(
                        f"{rel_path}:{lineno}: запрещённая фраза «{phrase}» → {stripped[:80]}"
                    )
                    break
            else:
                if re.search(rf"\b{re.escape(phrase_norm)}\b", line_norm):
                    errors.append(
                        f"{rel_path}:{lineno}: запрещённое слово «{phrase}» → {stripped[:80]}"
                    )
                    break

    # Минимальная длина разделов (только файлы с min_section_chars)
    min_chars = spec.get("min_section_chars")
    if min_chars:
        for heading in spec["headings"]:
            if heading in _SECTION_LEN_EXEMPT:
                continue
            sec = _section(content, heading)
            if sec and len(sec.strip()) < min_chars:
                errors.append(
                    f"{rel_path}: раздел «{heading}» слишком короткий "
                    f"({len(sec.strip())} симв., минимум {min_chars})"
                )

    # Числа в метриках (только для файлов с check_metrics)
    if spec.get("check_metrics"):
        errors.extend(_check_metrics(content, rel_path))

    return errors


# ─── Анализ согласованности и противоречий ────────────────────────────────────

def _check_consistency(vision_path: Path, spec_path: Path) -> list[str]:
    """
    Попарный и кросс-документный анализ согласованности.
    Возвращает список найденных проблем.
    """
    issues: list[str] = []

    if not vision_path.exists() or not spec_path.exists():
        issues.append("Невозможно проверить согласованность: один или оба файла отсутствуют.")
        return issues

    vision = vision_path.read_text(encoding="utf-8")
    spec   = spec_path.read_text(encoding="utf-8")

    includes_text = _section(vision, "## Что входит")
    excludes_text = _section(vision, "## Что НЕ входит")
    inc_items     = _list_items(includes_text)
    exc_items     = _list_items(excludes_text)

    # ── 1. Попарно: «Что входит» vs «Что НЕ входит» ──────────────────────────
    # Флажок только если ≥2 ключевых слов совпадают И они покрывают ≥50% exc-элемента.
    # Это исключает ложные срабатывания на разные грани одной темы.
    for exc_item in exc_items:
        exc_kw = set(_keywords(exc_item))
        if len(exc_kw) < 2:
            continue
        for inc_item in inc_items:
            inc_kw = set(_keywords(inc_item))
            overlap = exc_kw & inc_kw
            if len(overlap) >= 2 and len(overlap) / len(exc_kw) >= 0.5:
                issues.append(
                    f"[VISION] Противоречие между «Что входит» и «Что НЕ входит»:\n"
                    f"  Входит:    «{inc_item[:70]}»\n"
                    f"  НЕ входит: «{exc_item[:70]}»\n"
                    f"  Общие слова: {', '.join(sorted(overlap))}"
                )

    # ── 2. «Что НЕ входит» vs SPEC операции/страницы ─────────────────────────
    ops_text   = _section(spec, "## Ключевые операции")
    pages_text = _section(spec, "## Страницы и экраны")
    spec_func  = _norm(ops_text + " " + pages_text)

    for exc_item in exc_items:
        exc_kw = _keywords(exc_item)
        found  = [kw for kw in exc_kw if kw in spec_func]
        if len(found) >= 2:
            issues.append(
                f"[VISION→SPEC] Противоречие: исключённый элемент «{exc_item[:70]}» "
                f"обнаружен в SPEC (слова: {', '.join(found[:5])})"
            )

    # ── 3. «Ключевые возможности» VISION → покрытие в SPEC ───────────────────
    capabilities_text = _section(vision, "## Ключевые возможности")
    spec_norm = _norm(spec)

    for cap_item in _list_items(capabilities_text):
        cap_kw = _keywords(cap_item)
        if not cap_kw:
            continue
        found_n   = sum(1 for kw in cap_kw if kw in spec_norm)
        coverage  = found_n / len(cap_kw)
        if coverage < 0.4:
            short = re.sub(r"^\d+\.\s*", "", cap_item)[:70]
            issues.append(
                f"[VISION→SPEC] Несогласованность: возможность «{short}» "
                f"слабо отражена в SPEC ({found_n}/{len(cap_kw)} ключевых слов)"
            )

    # ── 4. Пункты «Что входит» → покрытие в SPEC ─────────────────────────────
    for inc_item in inc_items:
        inc_kw = _keywords(inc_item)
        if not inc_kw:
            continue
        found_n  = sum(1 for kw in inc_kw if kw in spec_norm)
        coverage = found_n / len(inc_kw)
        if coverage < 0.35:
            issues.append(
                f"[VISION→SPEC] Несогласованность: «Что входит» «{inc_item[:60]}» "
                f"не отражён в SPEC ({found_n}/{len(inc_kw)} ключевых слов)"
            )

    # ── 5. Сущности SPEC → упоминание в «Ключевые операции» ──────────────────
    entities_text = _section(spec, "## Сущности")
    entity_names  = _table_first_col(entities_text, skip_headers={"Сущность", "Entity"})
    ops_norm      = _norm(ops_text)

    for entity in entity_names:
        if _norm(entity) not in ops_norm:
            issues.append(
                f"[SPEC] Несогласованность: сущность «{entity}» "
                f"не упоминается в «Ключевые операции»"
            )

    # ── 6. «Тестирование» — обязательные три подраздела ──────────────────────
    testing_text = _section(spec, "## Тестирование")
    for sub in ("**Критические сценарии**", "**Бизнес-правила**", "**Негативные сценарии**"):
        if sub not in testing_text:
            issues.append(
                f"[SPEC] «Тестирование» не содержит обязательный подраздел {sub}"
            )

    # ── 7. Глоссарий → термины используются где-то в документации ────────────
    glossary_text  = _section(spec, "## Глоссарий")
    glossary_terms = _table_first_col(glossary_text, skip_headers={"Термин", "Term"})
    spec_no_gloss  = _norm(spec.replace(glossary_text, ""))
    vision_norm    = _norm(vision)

    for term in glossary_terms:
        term_norm = _norm(term)
        if term_norm not in spec_no_gloss and term_norm not in vision_norm:
            issues.append(
                f"[SPEC] Глоссарий: термин «{term}» определён, "
                f"но нигде не используется в документации"
            )

    # ── 8. Попарно: «Цель» vs «Что НЕ входит» ────────────────────────────────
    goal_text  = _section(vision, "## Цель")
    goal_norm  = _norm(goal_text)

    for exc_item in exc_items:
        exc_kw = set(_keywords(exc_item))
        if len(exc_kw) < 2:
            continue
        overlap = {kw for kw in exc_kw if kw in goal_norm}
        if len(overlap) >= 2 and len(overlap) / len(exc_kw) >= 0.5:
            issues.append(
                f"[VISION] Возможное противоречие: раздел «Цель» конфликтует с «Что НЕ входит»:\n"
                f"  НЕ входит: «{exc_item[:70]}»\n"
                f"  Совпавшие слова в «Цели»: {', '.join(sorted(overlap))}"
            )

    return issues


# ─── Проверка артефактов ──────────────────────────────────────────────────────

def _check_artifacts(target_dir: Path) -> list[str]:
    """Каждый файл в 3_ARTIFACTS/ должен быть упомянут хотя бы в одном документе."""
    artifacts_dir = target_dir / "3_ARTIFACTS"
    if not artifacts_dir.exists():
        return []

    doc_texts: list[str] = []
    for rel in ("INDEX.md", "1_PRODUCT_VISION/VISION.md", "2_PRODUCT_SPEC/SPEC.md"):
        fp = target_dir / rel
        if fp.exists():
            doc_texts.append(fp.read_text(encoding="utf-8"))
    combined = "\n".join(doc_texts)

    issues: list[str] = []
    for artifact in sorted(artifacts_dir.rglob("*")):
        if not artifact.is_file():
            continue
        rel_str = str(artifact.relative_to(target_dir)).replace("\\", "/")
        if artifact.name not in combined and rel_str not in combined:
            issues.append(
                f"[ARTIFACTS] Артефакт «{rel_str}» не упомянут ни в одном документе"
            )
    return issues


# ─── Команды ──────────────────────────────────────────────────────────────────

def cmd_consistency(args: argparse.Namespace) -> int:
    name_errors = _validate_name(args.name)
    if name_errors:
        for e in name_errors:
            err(e)
        return 1

    output_dir  = Path(args.output)
    target_dir  = output_dir / args.name
    vision_path = target_dir / "1_PRODUCT_VISION" / "VISION.md"
    spec_path   = target_dir / "2_PRODUCT_SPEC" / "SPEC.md"

    print(f"\nАнализ согласованности: {args.name}\n")

    issues = _check_consistency(vision_path, spec_path)
    if not issues:
        ok("Противоречий и несогласованностей не обнаружено.")
        print(f"\n{C.GREEN}✅ Документация «{args.name}» согласована.{C.RESET}\n")
        return 0

    for issue in issues:
        err(issue)
    print(
        f"\n{C.RED}Итог: {len(issues)} проблем согласованности. "
        f"Устраните их перед финализацией.{C.RESET}\n"
    )
    return 1


def cmd_status(args: argparse.Namespace) -> int:
    output_dir = Path(args.output)

    if args.name:
        name_errors = _validate_name(args.name)
        if name_errors:
            for e in name_errors:
                err(e)
            return 1
        projects = [args.name]
    else:
        if not output_dir.exists():
            err(f"Папка не найдена: {output_dir}")
            return 1
        projects = sorted(d.name for d in output_dir.iterdir() if d.is_dir())
        if not projects:
            warn("Проекты не найдены.")
            return 0

    doc_files = ["INDEX.md", "1_PRODUCT_VISION/VISION.md", "2_PRODUCT_SPEC/SPEC.md"]
    col1, col2, col3 = 22, 32, 14

    print()
    hdr = f"{'Проект':<{col1}} {'Файл':<{col2}} {'Статус':<{col3}} Дата"
    print(hdr)
    print("─" * (col1 + col2 + col3 + 14))

    for proj_name in projects:
        target_dir = output_dir / proj_name
        for rel_path in doc_files:
            fp = target_dir / rel_path
            if not fp.exists():
                print(f"{proj_name:<{col1}} {rel_path:<{col2}} {'ОТСУТСТВУЕТ':<{col3}}")
                continue
            content  = fp.read_text(encoding="utf-8")
            status_m = STATUS_PATTERN.search(content)
            date_m   = DATE_PATTERN.search(content)
            status   = status_m.group(1) if status_m else "?"
            date_val = date_m.group(0).replace("**Дата:**", "").strip() if date_m else "?"
            print(f"{proj_name:<{col1}} {rel_path:<{col2}} {status:<{col3}} {date_val}")

    print()
    return 0


def cmd_update_status(args: argparse.Namespace) -> int:
    if args.status not in VALID_STATUSES:
        err(
            f"Недопустимый статус: «{args.status}». "
            f"Допустимые: {', '.join(sorted(VALID_STATUSES))}"
        )
        return 1

    name_errors = _validate_name(args.name)
    if name_errors:
        for e in name_errors:
            err(e)
        return 1

    output_dir = Path(args.output)
    target_dir = output_dir / args.name

    if not target_dir.exists():
        err(f"Папка проекта не найдена: {target_dir}")
        return 1

    today   = _today()
    updated = 0

    print(f"\nОбновление статуса «{args.name}» → {args.status}\n")

    for rel_path in ("INDEX.md", "1_PRODUCT_VISION/VISION.md", "2_PRODUCT_SPEC/SPEC.md"):
        fp = target_dir / rel_path
        if not fp.exists():
            warn(f"Пропущен (не существует): {rel_path}")
            continue
        content     = fp.read_text(encoding="utf-8")
        new_content = STATUS_PATTERN.sub(f"**Статус:** {args.status}", content)
        new_content = DATE_PATTERN.sub(f"**Дата:** {today}", new_content)
        if new_content != content:
            fp.write_text(new_content, encoding="utf-8")
            ok(f"{rel_path} → {args.status} | {today}")
            updated += 1
        else:
            warn(f"{rel_path}: строки статуса не найдены, пропущен")

    if updated:
        print(f"\n{C.GREEN}✅ Обновлено файлов: {updated}.{C.RESET}\n")
        print("Не забудьте создать git-коммит:")
        print(f"  git add {target_dir}/")
        print(f'  git commit -m "docs: статус {args.name} → {args.status}"\n')
    return 0


def cmd_validate(args: argparse.Namespace) -> int:
    name_errors = _validate_name(args.name)
    if name_errors:
        for e in name_errors:
            err(e)
        return 1

    output_dir = Path(args.output)
    target_dir = output_dir / args.name

    if not target_dir.exists():
        err(f"Папка проекта не найдена: {target_dir}")
        return 1

    print(f"\nВалидация документации: {args.name}\n")

    # ── Блок 1: структура, заглушки, запрещённые слова, метрики ──────────────
    print(f"{C.YELLOW}── Структура и содержимое ──────────────────{C.RESET}")
    all_errors: list[str] = []
    for rel_path, spec in STRUCTURE.items():
        file_path   = target_dir / rel_path
        file_errors = _check_file(rel_path, file_path, spec)
        if file_errors:
            all_errors.extend(file_errors)
        else:
            ok(f"{rel_path} — {spec['description']}")

    if all_errors:
        print(f"\n{C.RED}Ошибки:{C.RESET}")
        for e in all_errors:
            err(e)
        print(
            f"\n{C.RED}Итог: {len(all_errors)} ошибок. "
            f"Исправьте их перед проверкой согласованности.{C.RESET}\n"
        )
        return 1

    # ── Блок 2: согласованность и противоречия ────────────────────────────────
    print(f"\n{C.YELLOW}── Согласованность и противоречия ──────────{C.RESET}")
    vision_path = target_dir / "1_PRODUCT_VISION" / "VISION.md"
    spec_path   = target_dir / "2_PRODUCT_SPEC" / "SPEC.md"
    c_issues    = _check_consistency(vision_path, spec_path)

    if c_issues:
        for issue in c_issues:
            err(issue)
        print(
            f"\n{C.RED}Итог: {len(c_issues)} проблем согласованности. "
            f"Документация не готова к финализации.{C.RESET}\n"
        )
        return 1

    ok("Противоречий и несогласованностей не обнаружено.")

    # ── Блок 3: артефакты без упоминания в документации ───────────────────────
    artifacts_dir = target_dir / "3_ARTIFACTS"
    has_artifacts = artifacts_dir.exists() and any(
        f for f in artifacts_dir.rglob("*") if f.is_file()
    )
    if has_artifacts:
        print(f"\n{C.YELLOW}── Артефакты ────────────────────────────────{C.RESET}")
        a_issues = _check_artifacts(target_dir)
        if a_issues:
            for issue in a_issues:
                err(issue)
            print(
                f"\n{C.RED}Итог: {len(a_issues)} артефактов без упоминания в документации. "
                f"Добавьте ссылки в SPEC.md → ## Артефакты.{C.RESET}\n"
            )
            return 1
        ok("Все артефакты задокументированы.")

    print(
        f"\n{C.GREEN}✅ Документация «{args.name}» прошла полную проверку "
        f"(структура + содержимое + согласованность + артефакты).{C.RESET}\n"
    )
    return 0


# ─── CLI ──────────────────────────────────────────────────────────────────────

def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="doc_gen.py",
        description="Генератор и валидатор документации продукта",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    # generate
    gen = sub.add_parser("generate", help="Создать структуру документов")
    gen.add_argument("name", metavar="НазваниеПроекта")
    gen.add_argument("--only", choices=["L1", "L2"], metavar="L1|L2",
                     help="Создать только один уровень")
    gen.add_argument("--update", action="store_true",
                     help="Не перезаписывать существующие файлы")
    gen.add_argument("--output", default="./docs", metavar="PATH",
                     help="Папка вывода (по умолчанию: ./docs)")

    # validate
    val = sub.add_parser("validate", help="Полная проверка: структура + содержимое + согласованность")
    val.add_argument("name", metavar="НазваниеПроекта")
    val.add_argument("--output", default="./docs", metavar="PATH",
                     help="Папка с документацией (по умолчанию: ./docs)")

    # consistency
    con = sub.add_parser("consistency", help="Только анализ согласованности и противоречий")
    con.add_argument("name", metavar="НазваниеПроекта")
    con.add_argument("--output", default="./docs", metavar="PATH",
                     help="Папка с документацией (по умолчанию: ./docs)")

    # status
    sta = sub.add_parser("status", help="Статус документов (всех или одного проекта)")
    sta.add_argument("name", metavar="НазваниеПроекта", nargs="?", default=None)
    sta.add_argument("--output", default="./docs", metavar="PATH",
                     help="Папка с документацией (по умолчанию: ./docs)")

    # update-status
    upd = sub.add_parser("update-status", help="Атомарно обновить статус во всех файлах проекта")
    upd.add_argument("name",   metavar="НазваниеПроекта")
    upd.add_argument("status", metavar="статус",
                     choices=sorted(VALID_STATUSES),
                     help="черновик | на ревью | утверждён")
    upd.add_argument("--output", default="./docs", metavar="PATH",
                     help="Папка с документацией (по умолчанию: ./docs)")

    return parser


def main() -> None:
    parser  = build_parser()
    args    = parser.parse_args()
    handlers = {
        "generate":      cmd_generate,
        "validate":      cmd_validate,
        "consistency":   cmd_consistency,
        "status":        cmd_status,
        "update-status": cmd_update_status,
    }
    sys.exit(handlers[args.command](args))


if __name__ == "__main__":
    main()
