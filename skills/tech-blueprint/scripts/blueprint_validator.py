#!/usr/bin/env python3
"""
blueprint_validator.py — валидатор технических контрактов проекта.

Использование:
  python3 blueprint_validator.py validate "ИмяПроекта" [--output PATH] [--update-mode]
"""

import argparse
import re
import sys
from pathlib import Path


# ─── Цвета и вывод ────────────────────────────────────────────────────────────

class C:
    RED    = "\033[0;31m"
    GREEN  = "\033[0;32m"
    YELLOW = "\033[1;33m"
    RESET  = "\033[0m"


def ok(msg: str)   -> None: print(f"{C.GREEN}✓{C.RESET} {msg}")
def err(msg: str)  -> None: print(f"{C.RED}✗{C.RESET} {msg}")
def warn(msg: str) -> None: print(f"{C.YELLOW}⟳{C.RESET} {msg}")


# ─── Константы ────────────────────────────────────────────────────────────────

REQUIRED_FILES = [
    "DATABASE_MODEL.md",
    "API_CONTRACTS.md",
    "ARCHITECTURE.md",
    "TESTING_PLAN.md",
]

_FSD_PATH_RE = re.compile(
    r"src/(entities|features|widgets|pages|shared|app)/|"
    r"app/(entities|features|widgets|pages|shared)/"
)

_LIST_TYPE_RE      = re.compile(r":\s*\[\w")
_PAGINATION_KW_RE  = re.compile(r"\b(first|after|limit|offset|page|cursor)\b", re.IGNORECASE)
_SPEC_REF_RE       = re.compile(r"(SPEC\.md|VISION\.md)")
_PRISMA_MODEL_RE   = re.compile(r"model\s+(\w+)\s*\{([\s\S]*?)\n\}", re.MULTILINE)
_MIGRATION_SEC_RE  = re.compile(
    r"^#{1,3}\s*(план миграции|изменения бд|migration plan|db changes|database changes)",
    re.MULTILINE | re.IGNORECASE,
)


# ─── Вспомогательные функции ──────────────────────────────────────────────────

def _read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def _extract_code_block(content: str, lang: str) -> str:
    """Возвращает содержимое первого блока ```lang ... ```."""
    m = re.search(rf"```{re.escape(lang)}\n([\s\S]*?)```", content)
    return m.group(1) if m else ""


def _extract_prisma_models(prisma_content: str) -> dict[str, str]:
    """Возвращает {ИмяМодели: тело_модели} из Prisma-схемы."""
    return {m.group(1): m.group(2) for m in _PRISMA_MODEL_RE.finditer(prisma_content)}


# ─── Проверки ─────────────────────────────────────────────────────────────────

def check_files(blueprint_dir: Path) -> list[str]:
    """Проверка 1: все обязательные файлы существуют."""
    return [
        f"Файл отсутствует: {f}"
        for f in REQUIRED_FILES
        if not (blueprint_dir / f).exists()
    ]


def check_code_blocks(blueprint_dir: Path) -> list[str]:
    """Проверка 2: DATABASE_MODEL.md содержит ```prisma, API_CONTRACTS.md — ```graphql."""
    errors: list[str] = []
    db = _read(blueprint_dir / "DATABASE_MODEL.md")
    if not _extract_code_block(db, "prisma"):
        errors.append("DATABASE_MODEL.md: отсутствует блок ```prisma")
    api = _read(blueprint_dir / "API_CONTRACTS.md")
    if not _extract_code_block(api, "graphql"):
        errors.append("API_CONTRACTS.md: отсутствует блок ```graphql")
    return errors


def check_fsd_paths(blueprint_dir: Path) -> list[str]:
    """Проверка 3: ARCHITECTURE.md не содержит файловых путей FSD."""
    arch = _read(blueprint_dir / "ARCHITECTURE.md")
    matches = _FSD_PATH_RE.findall(arch)
    if matches:
        return [
            "ARCHITECTURE.md: обнаружены файловые пути FSD. "
            "В архитектурном документе должны быть только логические названия "
            "сущностей и компонентов (не файловые пути)."
        ]
    return []


def check_model_coverage(blueprint_dir: Path) -> list[str]:
    """Проверка 4: большинство Prisma-моделей упомянуты в API_CONTRACTS.md."""
    db  = _read(blueprint_dir / "DATABASE_MODEL.md")
    api = _read(blueprint_dir / "API_CONTRACTS.md")

    prisma_block = _extract_code_block(db, "prisma")
    if not prisma_block:
        return []

    model_names = list(_extract_prisma_models(prisma_block).keys())
    if not model_names:
        return []

    api_lower = api.lower()
    covered = [m for m in model_names if m.lower() in api_lower]
    ratio = len(covered) / len(model_names)

    if ratio < 0.5:
        missing = [m for m in model_names if m.lower() not in api_lower]
        return [
            f"Кросс-чек БД/API: {len(covered)}/{len(model_names)} Prisma-моделей "
            f"найдены в API_CONTRACTS.md. Отсутствуют: {', '.join(missing)}"
        ]
    return []


def check_traceability(blueprint_dir: Path) -> list[str]:
    """Проверка 5: DATABASE_MODEL.md и API_CONTRACTS.md содержат ссылки на SPEC.md/VISION.md."""
    errors: list[str] = []
    for filename in ("DATABASE_MODEL.md", "API_CONTRACTS.md"):
        content = _read(blueprint_dir / filename)
        if not _SPEC_REF_RE.search(content):
            errors.append(
                f"{filename}: отсутствует трассируемость. "
                "Добавьте комментарии со ссылками на бизнес-требования (SPEC.md или VISION.md)"
            )
    return errors


def check_pagination(blueprint_dir: Path) -> list[str]:
    """
    Проверка 6: поля Query/Mutation, возвращающие списки, имеют аргументы пагинации.
    Ищет только внутри type Query / type Mutation / type Subscription.
    Предупреждения, не ошибки (pagination может быть в обёртке-типе).
    """
    issues: list[str] = []
    api = _read(blueprint_dir / "API_CONTRACTS.md")
    graphql_block = _extract_code_block(api, "graphql")
    if not graphql_block:
        return issues

    lines = graphql_block.splitlines()
    in_root_op   = False
    current_type: str | None = None
    depth        = 0

    for i, line in enumerate(lines):
        stripped = line.strip()

        # Определить вход в тип Query/Mutation/Subscription
        m = re.match(r"^type\s+(Query|Mutation|Subscription)\b", stripped)
        if m:
            in_root_op   = True
            current_type = m.group(1)
            depth        = 0

        # Обновить глубину вложенности
        depth += stripped.count("{") - stripped.count("}")
        if depth <= 0 and in_root_op:
            in_root_op   = False
            current_type = None
            continue

        if not in_root_op:
            continue

        # Проверить, возвращает ли строка список
        if not _LIST_TYPE_RE.search(stripped):
            continue

        # Собрать контекст: до 8 предыдущих строк (для многострочных аргументов)
        ctx = "\n".join(lines[max(0, i - 8) : i + 1])
        if not _PAGINATION_KW_RE.search(ctx):
            name_m     = re.search(r"(\w+)\s*[\(:]", stripped)
            field_name = name_m.group(1) if name_m else stripped[:50]
            issues.append(
                f"API_CONTRACTS.md [{current_type}]: поле «{field_name}» возвращает список "
                "без аргументов пагинации — добавьте first/after или limit/offset"
            )

    return issues


def check_prisma_timestamps(blueprint_dir: Path) -> list[str]:
    """
    Проверка 7: каждая Prisma-модель (кроме join-таблиц с '_' в названии)
    содержит createdAt или updatedAt.
    """
    errors: list[str] = []
    db = _read(blueprint_dir / "DATABASE_MODEL.md")
    prisma_block = _extract_code_block(db, "prisma")
    if not prisma_block:
        return errors

    for model_name, body in _extract_prisma_models(prisma_block).items():
        # Join-таблицы (содержат '_' в имени) — пропустить
        if "_" in model_name:
            continue
        if "createdAt" not in body and "updatedAt" not in body:
            errors.append(
                f"DATABASE_MODEL.md: модель «{model_name}» не содержит "
                "обязательных полей createdAt/updatedAt"
            )

    return errors


def check_migration_section(blueprint_dir: Path) -> list[str]:
    """Проверка 8 (--update-mode): DATABASE_MODEL.md содержит раздел «План миграции»."""
    db = _read(blueprint_dir / "DATABASE_MODEL.md")
    if not _MIGRATION_SEC_RE.search(db):
        return [
            "DATABASE_MODEL.md: в режиме обновления требуется описать план миграции. "
            "Добавьте раздел «## План миграции» или «## Изменения БД»"
        ]
    return []


# ─── Команда validate ──────────────────────────────────────────────────────────

def cmd_validate(args: argparse.Namespace) -> int:
    blueprint_dir = Path(args.output) / args.name / "3_TECH_BLUEPRINT"

    if not blueprint_dir.exists():
        err(f"Папка не найдена: {blueprint_dir}")
        print(f"  Ожидается: docs/{args.name}/3_TECH_BLUEPRINT/")
        return 1

    print(f"\nВалидация технического блюпринта: {args.name}\n")

    all_errors:   list[str] = []
    all_warnings: list[str] = []

    # ── 1. Наличие файлов ─────────────────────────────────────────────────────
    print(f"{C.YELLOW}── Наличие файлов ──────────────────────────{C.RESET}")
    file_errors = check_files(blueprint_dir)
    for e in file_errors:
        err(e)
    if file_errors:
        print(
            f"\n{C.RED}Критические ошибки: создайте все обязательные файлы "
            f"прежде чем продолжить.{C.RESET}\n"
        )
        return 1
    ok("Все обязательные файлы присутствуют")

    # ── 2. Блоки кода ─────────────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── Структура контента ──────────────────────{C.RESET}")
    block_errors = check_code_blocks(blueprint_dir)
    for e in block_errors:
        err(e)
    all_errors.extend(block_errors)
    if not block_errors:
        ok("Блоки кода (```prisma, ```graphql) найдены")

    # ── 3. FSD-пути ───────────────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── FSD Paths Check ─────────────────────────{C.RESET}")
    fsd_errors = check_fsd_paths(blueprint_dir)
    for e in fsd_errors:
        err(e)
    all_errors.extend(fsd_errors)
    if not fsd_errors:
        ok("ARCHITECTURE.md не содержит файловых путей FSD")

    # ── 4. Кросс-чек моделей ──────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── Кросс-чек БД / GraphQL ──────────────────{C.RESET}")
    if not block_errors:
        coverage_errors = check_model_coverage(blueprint_dir)
        for e in coverage_errors:
            err(e)
        all_errors.extend(coverage_errors)
        if not coverage_errors:
            ok("Большинство Prisma-моделей упомянуты в API_CONTRACTS.md")

    # ── 5. Трассируемость ─────────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── Трассируемость ──────────────────────────{C.RESET}")
    trace_errors = check_traceability(blueprint_dir)
    for e in trace_errors:
        err(e)
    all_errors.extend(trace_errors)
    if not trace_errors:
        ok("Ссылки на бизнес-требования найдены в DB-модели и API-контрактах")

    # ── 6. Пагинация GraphQL ──────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── Пагинация GraphQL ───────────────────────{C.RESET}")
    if not block_errors:
        pag_issues = check_pagination(blueprint_dir)
        for issue in pag_issues:
            warn(issue)
        all_warnings.extend(pag_issues)
        if not pag_issues:
            ok("Все списочные поля Query/Mutation имеют аргументы пагинации")

    # ── 7. Технические поля Prisma ────────────────────────────────────────────
    print(f"\n{C.YELLOW}── Технические поля Prisma ─────────────────{C.RESET}")
    if not block_errors:
        ts_errors = check_prisma_timestamps(blueprint_dir)
        for e in ts_errors:
            err(e)
        all_errors.extend(ts_errors)
        if not ts_errors:
            ok("Все модели (кроме join-таблиц) содержат createdAt/updatedAt")

    # ── 8. Режим обновления ───────────────────────────────────────────────────
    if args.update_mode:
        print(f"\n{C.YELLOW}── Режим обновления ────────────────────────{C.RESET}")
        mig_errors = check_migration_section(blueprint_dir)
        for e in mig_errors:
            err(e)
        all_errors.extend(mig_errors)
        if not mig_errors:
            ok("Раздел с планом миграции найден")

    # ── Итог ──────────────────────────────────────────────────────────────────
    print()
    if all_errors:
        warn_suffix = f", предупреждений: {len(all_warnings)}" if all_warnings else ""
        print(
            f"{C.RED}Итог: {len(all_errors)} ошибок{warn_suffix}. "
            f"Устраните ошибки перед продолжением.{C.RESET}\n"
        )
        return 1

    if all_warnings:
        print(
            f"{C.YELLOW}Итог: ошибок нет, предупреждений: {len(all_warnings)}. "
            f"Рекомендуется исправить.{C.RESET}\n"
        )
    else:
        print(
            f"{C.GREEN}✅ Блюпринт «{args.name}» прошёл полную проверку.{C.RESET}\n"
        )

    return 0


# ─── CLI ──────────────────────────────────────────────────────────────────────

def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="blueprint_validator.py",
        description="Валидатор технических контрактов (tech-blueprint)",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    val = sub.add_parser("validate", help="Проверить папку 3_TECH_BLUEPRINT/")
    val.add_argument("name", metavar="ИмяПроекта")
    val.add_argument(
        "--output", default="./blueprint", metavar="PATH",
        help="Корневая папка проектов (по умолчанию: ./blueprint)",
    )
    val.add_argument(
        "--update-mode", action="store_true",
        help="Режим обновления: требует раздел «План миграции» в DATABASE_MODEL.md",
    )

    return parser


def main() -> None:
    parser = build_parser()
    args   = parser.parse_args()
    sys.exit(cmd_validate(args))


if __name__ == "__main__":
    main()
