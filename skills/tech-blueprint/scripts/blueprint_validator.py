#!/usr/bin/env python3
"""
blueprint_validator.py — validator for a project's technical contracts.

Usage:
  python3 blueprint_validator.py validate "ProjectName" [--output PATH] [--update-mode]
"""

import argparse
import re
import sys
from pathlib import Path


# ─── Colors and output ────────────────────────────────────────────────────────

class C:
    RED    = "\033[0;31m"
    GREEN  = "\033[0;32m"
    YELLOW = "\033[1;33m"
    RESET  = "\033[0m"


def ok(msg: str)   -> None: print(f"{C.GREEN}✓{C.RESET} {msg}")
def err(msg: str)  -> None: print(f"{C.RED}✗{C.RESET} {msg}")
def warn(msg: str) -> None: print(f"{C.YELLOW}⟳{C.RESET} {msg}")


# ─── Constants ──────────────────────────────────────────────────────────────────

REQUIRED_FILES = [
    "IMPLEMENTATION_GUIDE.md",
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
    r"^#{1,3}\s*(migration plan|db changes|database changes)",
    re.MULTILINE | re.IGNORECASE,
)
_IMPL_STACK_RE     = re.compile(r"^#{1,3}\s*(stack)\b", re.MULTILINE | re.IGNORECASE)
_IMPL_DONE_RE      = re.compile(
    r"^#{1,3}\s*(already implemented|what.s already)",
    re.MULTILINE | re.IGNORECASE,
)
_IMPL_LAUNCH_RE    = re.compile(
    r"^#{1,3}\s*(local (setup|run|start)|getting started|quick start)",
    re.MULTILINE | re.IGNORECASE,
)


# ─── Helper functions ─────────────────────────────────────────────────────────

def _read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def _extract_code_block(content: str, lang: str) -> str:
    """Returns the contents of the first ```lang ... ``` block."""
    m = re.search(rf"```{re.escape(lang)}\n([\s\S]*?)```", content)
    return m.group(1) if m else ""


def _extract_prisma_models(prisma_content: str) -> dict[str, str]:
    """Returns {ModelName: model_body} from a Prisma schema."""
    return {m.group(1): m.group(2) for m in _PRISMA_MODEL_RE.finditer(prisma_content)}


# ─── Checks ───────────────────────────────────────────────────────────────────

def check_files(blueprint_dir: Path) -> list[str]:
    """Check 1: all required files exist."""
    return [
        f"Missing file: {f}"
        for f in REQUIRED_FILES
        if not (blueprint_dir / f).exists()
    ]


def check_code_blocks(blueprint_dir: Path) -> list[str]:
    """Check 2: DATABASE_MODEL.md contains ```prisma, API_CONTRACTS.md contains ```graphql."""
    errors: list[str] = []
    db = _read(blueprint_dir / "DATABASE_MODEL.md")
    if not _extract_code_block(db, "prisma"):
        errors.append("DATABASE_MODEL.md: missing ```prisma block")
    api = _read(blueprint_dir / "API_CONTRACTS.md")
    if not _extract_code_block(api, "graphql"):
        errors.append("API_CONTRACTS.md: missing ```graphql block")
    return errors


def check_fsd_paths(blueprint_dir: Path) -> list[str]:
    """Check 3: ARCHITECTURE.md contains no FSD file paths."""
    arch = _read(blueprint_dir / "ARCHITECTURE.md")
    matches = _FSD_PATH_RE.findall(arch)
    if matches:
        return [
            "ARCHITECTURE.md: FSD file paths detected. "
            "The architecture document should contain only the logical names "
            "of entities and components (not file paths)."
        ]
    return []


def check_model_coverage(blueprint_dir: Path) -> list[str]:
    """Check 4: most Prisma models are mentioned in API_CONTRACTS.md."""
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
            f"DB/API cross-check: {len(covered)}/{len(model_names)} Prisma models "
            f"found in API_CONTRACTS.md. Missing: {', '.join(missing)}"
        ]
    return []


def check_traceability(blueprint_dir: Path) -> list[str]:
    """Check 5: DATABASE_MODEL.md and API_CONTRACTS.md contain references to SPEC.md/VISION.md."""
    errors: list[str] = []
    for filename in ("DATABASE_MODEL.md", "API_CONTRACTS.md"):
        content = _read(blueprint_dir / filename)
        if not _SPEC_REF_RE.search(content):
            errors.append(
                f"{filename}: traceability missing. "
                "Add comments referencing the business requirements (SPEC.md or VISION.md)"
            )
    return errors


def check_pagination(blueprint_dir: Path) -> list[str]:
    """
    Check 6: Query/Mutation fields that return lists have pagination arguments.
    Only searches inside type Query / type Mutation / type Subscription.
    Warnings, not errors (pagination may live in a wrapper type).
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

        # Detect entry into a Query/Mutation/Subscription type
        m = re.match(r"^type\s+(Query|Mutation|Subscription)\b", stripped)
        if m:
            in_root_op   = True
            current_type = m.group(1)
            depth        = 0

        # Update nesting depth
        depth += stripped.count("{") - stripped.count("}")
        if depth <= 0 and in_root_op:
            in_root_op   = False
            current_type = None
            continue

        if not in_root_op:
            continue

        # Check whether the line returns a list
        if not _LIST_TYPE_RE.search(stripped):
            continue

        # Gather context: up to 8 preceding lines (for multi-line arguments)
        ctx = "\n".join(lines[max(0, i - 8) : i + 1])
        if not _PAGINATION_KW_RE.search(ctx):
            name_m     = re.search(r"(\w+)\s*[\(:]", stripped)
            field_name = name_m.group(1) if name_m else stripped[:50]
            issues.append(
                f"API_CONTRACTS.md [{current_type}]: field «{field_name}» returns a list "
                "without pagination arguments — add first/after or limit/offset"
            )

    return issues


def check_prisma_timestamps(blueprint_dir: Path) -> list[str]:
    """
    Check 7: every Prisma model (except join tables with '_' in the name)
    contains createdAt or updatedAt.
    """
    errors: list[str] = []
    db = _read(blueprint_dir / "DATABASE_MODEL.md")
    prisma_block = _extract_code_block(db, "prisma")
    if not prisma_block:
        return errors

    for model_name, body in _extract_prisma_models(prisma_block).items():
        # Join tables (contain '_' in the name) — skip
        if "_" in model_name:
            continue
        if "createdAt" not in body and "updatedAt" not in body:
            errors.append(
                f"DATABASE_MODEL.md: model «{model_name}» is missing "
                "the required createdAt/updatedAt fields"
            )

    return errors


def check_implementation_guide(blueprint_dir: Path) -> list[str]:
    """Check 8: IMPLEMENTATION_GUIDE.md contains the required sections."""
    guide_path = blueprint_dir / "IMPLEMENTATION_GUIDE.md"
    if not guide_path.exists():
        return []  # already reported in check_files
    content = _read(guide_path)
    errors: list[str] = []
    if not _IMPL_STACK_RE.search(content):
        errors.append("IMPLEMENTATION_GUIDE.md: missing «## Stack» section")
    if not _IMPL_DONE_RE.search(content):
        errors.append("IMPLEMENTATION_GUIDE.md: missing «## Already Implemented» section")
    if not _IMPL_LAUNCH_RE.search(content):
        errors.append("IMPLEMENTATION_GUIDE.md: missing «## Local Setup» section")
    return errors


def check_migration_section(blueprint_dir: Path) -> list[str]:
    """Check 8 (--update-mode): DATABASE_MODEL.md contains a «Migration Plan» section."""
    db = _read(blueprint_dir / "DATABASE_MODEL.md")
    if not _MIGRATION_SEC_RE.search(db):
        return [
            "DATABASE_MODEL.md: update mode requires a migration plan. "
            "Add a «## Migration Plan» or «## DB Changes» section"
        ]
    return []


# ─── validate command ────────────────────────────────────────────────────────────

def cmd_validate(args: argparse.Namespace) -> int:
    blueprint_dir = Path(args.output) / args.name / "3_TECH_BLUEPRINT"

    if not blueprint_dir.exists():
        err(f"Folder not found: {blueprint_dir}")
        print(f"  Expected: docs/{args.name}/3_TECH_BLUEPRINT/")
        return 1

    print(f"\nValidating technical blueprint: {args.name}\n")

    all_errors:   list[str] = []
    all_warnings: list[str] = []

    # ── 1. File presence ──────────────────────────────────────────────────────
    print(f"{C.YELLOW}── File presence ───────────────────────────{C.RESET}")
    file_errors = check_files(blueprint_dir)
    for e in file_errors:
        err(e)
    if file_errors:
        print(
            f"\n{C.RED}Critical errors: create all required files "
            f"before continuing.{C.RESET}\n"
        )
        return 1
    ok("All required files are present")

    # ── 2. Code blocks ────────────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── Content structure ───────────────────────{C.RESET}")
    block_errors = check_code_blocks(blueprint_dir)
    for e in block_errors:
        err(e)
    all_errors.extend(block_errors)
    if not block_errors:
        ok("Code blocks (```prisma, ```graphql) found")

    # ── 3. FSD paths ──────────────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── FSD Paths Check ─────────────────────────{C.RESET}")
    fsd_errors = check_fsd_paths(blueprint_dir)
    for e in fsd_errors:
        err(e)
    all_errors.extend(fsd_errors)
    if not fsd_errors:
        ok("ARCHITECTURE.md contains no FSD file paths")

    # ── 4. Model cross-check ──────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── DB / GraphQL cross-check ────────────────{C.RESET}")
    if not block_errors:
        coverage_errors = check_model_coverage(blueprint_dir)
        for e in coverage_errors:
            err(e)
        all_errors.extend(coverage_errors)
        if not coverage_errors:
            ok("Most Prisma models are mentioned in API_CONTRACTS.md")

    # ── 5. Traceability ───────────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── Traceability ────────────────────────────{C.RESET}")
    trace_errors = check_traceability(blueprint_dir)
    for e in trace_errors:
        err(e)
    all_errors.extend(trace_errors)
    if not trace_errors:
        ok("References to business requirements found in DB model and API contracts")

    # ── 6. GraphQL pagination ─────────────────────────────────────────────────
    print(f"\n{C.YELLOW}── GraphQL pagination ──────────────────────{C.RESET}")
    if not block_errors:
        pag_issues = check_pagination(blueprint_dir)
        for issue in pag_issues:
            warn(issue)
        all_warnings.extend(pag_issues)
        if not pag_issues:
            ok("All list-returning Query/Mutation fields have pagination arguments")

    # ── 7. Prisma technical fields ────────────────────────────────────────────
    print(f"\n{C.YELLOW}── Prisma technical fields ─────────────────{C.RESET}")
    if not block_errors:
        ts_errors = check_prisma_timestamps(blueprint_dir)
        for e in ts_errors:
            err(e)
        all_errors.extend(ts_errors)
        if not ts_errors:
            ok("All models (except join tables) contain createdAt/updatedAt")

    # ── 8. IMPLEMENTATION_GUIDE.md ────────────────────────────────────────────
    print(f"\n{C.YELLOW}── IMPLEMENTATION_GUIDE.md ─────────────────{C.RESET}")
    guide_errors = check_implementation_guide(blueprint_dir)
    for e in guide_errors:
        err(e)
    all_errors.extend(guide_errors)
    if not guide_errors:
        ok("IMPLEMENTATION_GUIDE.md contains the required sections")

    # ── 9. Update mode ────────────────────────────────────────────────────────
    if args.update_mode:
        print(f"\n{C.YELLOW}── Update mode (migration) ─────────────────{C.RESET}")
        mig_errors = check_migration_section(blueprint_dir)
        for e in mig_errors:
            err(e)
        all_errors.extend(mig_errors)
        if not mig_errors:
            ok("Migration plan section found")

    # ── Summary ─────────────────────────────────────────────────────────────────
    print()
    if all_errors:
        warn_suffix = f", warnings: {len(all_warnings)}" if all_warnings else ""
        print(
            f"{C.RED}Summary: {len(all_errors)} errors{warn_suffix}. "
            f"Fix the errors before continuing.{C.RESET}\n"
        )
        return 1

    if all_warnings:
        print(
            f"{C.YELLOW}Summary: no errors, warnings: {len(all_warnings)}. "
            f"Fixing them is recommended.{C.RESET}\n"
        )
    else:
        print(
            f"{C.GREEN}✅ Blueprint «{args.name}» passed full validation.{C.RESET}\n"
        )

    return 0


# ─── CLI ──────────────────────────────────────────────────────────────────────

def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="blueprint_validator.py",
        description="Validator for technical contracts (tech-blueprint)",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    val = sub.add_parser("validate", help="Validate the 3_TECH_BLUEPRINT/ folder")
    val.add_argument("name", metavar="ProjectName")
    val.add_argument(
        "--output", default="./blueprint", metavar="PATH",
        help="Root folder for projects (default: ./blueprint)",
    )
    val.add_argument(
        "--update-mode", action="store_true",
        help="Update mode: requires a «Migration Plan» section in DATABASE_MODEL.md",
    )

    return parser


def main() -> None:
    parser = build_parser()
    args   = parser.parse_args()
    sys.exit(cmd_validate(args))


if __name__ == "__main__":
    main()
