#!/usr/bin/env python3
"""
doc_gen.py — product documentation generator and validator.

Usage:
  python3 doc_gen.py generate      "ProjectName" [--only L1|L2] [--update] [--output PATH]
  python3 doc_gen.py validate      "ProjectName" [--output PATH]
  python3 doc_gen.py consistency   "ProjectName" [--output PATH]
  python3 doc_gen.py status       ["ProjectName"] [--output PATH]
  python3 doc_gen.py update-status "ProjectName" "status" [--output PATH]
"""

import argparse
import re
import sys
from datetime import date
from pathlib import Path


# ─── Document structure ───────────────────────────────────────────────────────

STRUCTURE = {
    "INDEX.md": {
        "description": "Table of contents",
        "headings": ["## Navigation", "## Quick links"],
    },
    "1_PRODUCT_VISION/VISION.md": {
        "description": "Product vision",
        "headings": [
            "## Problem", "## Target audience", "## Goal",
            "## Key capabilities", "## Success metrics",
            "## In scope", "## Out of scope",
        ],
        "min_section_chars": 60,
        "check_metrics": True,
    },
    "2_PRODUCT_SPEC/SPEC.md": {
        "description": "Product specification",
        "headings": [
            "## References", "## How the system works", "## Glossary",
            "## Entities", "## Pages and screens", "## Key operations",
            "## Integrations", "## Testing", "## Artifacts",
        ],
        "min_section_chars": 60,
    },
}

# Sections exempt from the minimum-length check
_SECTION_LEN_EXEMPT = {
    "## References", "## Navigation", "## Quick links", "## Integrations",
    "## Artifacts",
}
_MIN_SECTION_CHARS = 60

STATUS_PATTERN      = re.compile(r'\*\*Status:\*\*\s*(draft|in review|approved)')
DATE_PATTERN        = re.compile(r'\*\*Date:\*\*\s*\d{4}-\d{2}-\d{2}')
PLACEHOLDER_PATTERN = re.compile(r'\[[^\]]+\](?!\()')
_UNSAFE_CHARS       = re.compile(r'[/\\:*?"<>|]')

VALID_STATUSES = {"draft", "in review", "approved"}

# Words/phrases banned in document content (vague, promotional, optional).
# Normalized: ё→е (the check runs through _norm()).
_FORBIDDEN: list[str] = [
    # Optionality
    "if needed", "if desired", "optionally", "maybe",
    "perhaps", "possibly", "probably", "likely",
    "при необходимости", "по желанию", "при желании", "может быть",
    "возможно", "опционально", "наверное", "вероятно",
    # Vague qualities
    "fast", "quick", "speedy", "rapid",
    "convenient", "handy", "user-friendly",
    "easy", "simple", "effortless", "straightforward",
    "intuitive", "seamless", "flexible",
    "modern", "innovative", "cutting-edge",
    "best", "optimal", "efficient", "high-quality",
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

# Stopwords for keyword analysis (consistency)
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
    "page", "pages", "entity", "entities", "should", "feature", "features",
    "include", "includes", "contains", "operation", "operations", "screen",
}


# ─── Colors and output ────────────────────────────────────────────────────────

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
        issues.append("Project name cannot be empty.")
    if _UNSAFE_CHARS.search(name):
        issues.append('Project name contains invalid characters: / \\ : * ? " < > |')
    if " " in name:
        issues.append(
            f"Project name contains spaces. Use CamelCase or a hyphen: "
            f"\"{name.replace(' ', '')}\" or \"{name.replace(' ', '-')}\""
        )
    return issues


# ─── Helper functions ─────────────────────────────────────────────────────────

def _norm(text: str) -> str:
    """Normalization: ё→е, lowercase."""
    return text.replace("ё", "е").replace("Ё", "Е").lower()


def _section(content: str, heading: str) -> str:
    """Markdown section text from heading up to the next heading of the same or higher level."""
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
    """Significant words: at least 5 characters, not in the stopword list."""
    words = re.findall(r"\b[а-яёА-ЯЁa-zA-Z]{5,}\b", text)
    return [_norm(w) for w in words if _norm(w) not in _STOPWORDS]


def _list_items(text: str) -> list[str]:
    """Text of bulleted/numbered list items."""
    items = []
    for line in text.splitlines():
        m = re.match(r"^\s*(?:\d+\.|[-*•])\s+(.+)", line)
        if m:
            item = re.sub(r"\*+([^*]+)\*+", r"\1", m.group(1)).strip()
            items.append(item)
    return items


def _table_first_col(section_text: str, skip_headers: set[str] | None = None) -> list[str]:
    """Values of the first table column (skips the header and separator rows)."""
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
    """Checks that metric target values contain numbers."""
    errors = []
    metrics_sec = _section(content, "## Success metrics")
    if not metrics_sec:
        return errors
    for line in metrics_sec.splitlines():
        if "|" not in line:
            continue
        if re.match(r"^\s*\|[-: |]+\|\s*$", line):
            continue
        cols = [c.strip() for c in line.strip("|").split("|")]
        if not cols or cols[0] in ("Metric", "Метрика", ""):
            continue
        val_col = cols[1].strip() if len(cols) > 1 else ""
        if not re.search(r"\d", val_col):
            errors.append(
                f"{rel_path}: metric \"{cols[0]}\" — "
                f"target value \"{val_col}\" contains no number"
            )
    return errors


# ─── File templates ───────────────────────────────────────────────────────────

def _today() -> str:
    return date.today().isoformat()


def _index_template(name: str) -> str:
    return f"""\
# Product documentation: {name}

**Status:** draft | **Date:** {_today()}

## Navigation

| Document | Description |
|----------|----------|
| [Vision](./1_PRODUCT_VISION/VISION.md) | Problem, audience, goal, project boundaries |
| [Specification](./2_PRODUCT_SPEC/SPEC.md) | Entities, pages, operations, testing |

## Quick links

- [Key capabilities](./1_PRODUCT_VISION/VISION.md#key-capabilities)
- [Pages and screens](./2_PRODUCT_SPEC/SPEC.md#pages-and-screens)
- [Testing](./2_PRODUCT_SPEC/SPEC.md#testing)
- [Artifacts](./2_PRODUCT_SPEC/SPEC.md#artifacts)
"""


def _vision_template(name: str) -> str:
    return f"""\
# {name}

**Status:** draft | **Date:** {_today()}

## Problem
[What specifically is inconvenient or broken. Context. 2–4 sentences.]

## Target audience
[Who the user is: their role and work context. Not "all users", but a specific type. 2–4 sentences.]

## Goal
[What exactly we are building and for whom. 2–4 sentences.]

## Key capabilities
1. **[Name]**: the user performs [X] → gets [Y]
2. **[Name]**: the user performs [X] → gets [Y]

## Success metrics
| Metric | Target value |
|---------|------------------|
| [Metric] | [specific number] |

## In scope (project boundaries)
Functionality that will be implemented:
1. [Functionality 1]

## Out of scope
Explicit exclusions to prevent scope creep:
- Does not include [X]
- Does not include [Y]
"""


def _spec_template(name: str) -> str:
    return f"""\
# Product specification: {name}

**Status:** draft | **Date:** {_today()}

## References
- Vision: [VISION.md](../1_PRODUCT_VISION/VISION.md)

## How the system works
[A brief description of the parts the product is made of and how they interact.
Example: "A web application with a personal account area and an admin panel. Data is stored
centrally, accessed through a browser without installing any apps."]

## Glossary
Key product concepts. Each term has a single precise definition with no synonyms.

| Term | Definition |
|--------|-------------|
| [Term] | [Precise definition] |

## Entities
The main objects the system works with. In business terms, not database terms.

| Entity | Description | Key attributes |
|----------|----------|-------------------|
| User | [Description] | [Attribute 1, Attribute 2] |

### Entity lifecycle
For each key entity — the allowed statuses and the transitions between them.

| Entity | Statuses | Transitions |
|----------|---------|----------|
| [Entity] | [status1 → status2 → status3] | [status1→status2: transition condition] |

## Pages and screens
An exhaustive list of the pages and screens that need to be created.

| Page | Purpose | Key elements |
|----------|------------|-------------------|
| Home | [Purpose] | [Element 1, Element 2] |
| Sign-up | [Purpose] | [Element 1, Element 2] |

## Key operations
What users can do in the system.

**[Role or "All users"]:**
- [Operation]: [brief description of the result]

## Integrations
External services the product cannot work without.
If there are no integrations — delete the table and write: "No integrations."

| Service | Purpose |
|--------|------------|
| [Name] | [Why it is needed] |

## Testing
Functionality covered by tests.

**Critical scenarios** (must work without errors):
- [The user performs X → the system returns Y]

**Business rules** (correctness of calculations and constraints):
- [Rule or calculation]

**Negative scenarios** (system behavior on errors and cancellations):
- [The user performs X with invalid data → the system returns message Z, the action is not performed]

## Artifacts
Supporting materials for developing and launching the product.
If there are no artifacts — write: "No artifacts."

No artifacts.
"""


# ─── Generation ─────────────────────────────────────────────────────────────────

def _should_write(path: Path, update_mode: bool) -> bool:
    if update_mode and path.exists():
        warn(f"Skipped (already exists): {path.name}")
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
        print(f"{C.RED}Error:{C.RESET} folder {target_dir} already exists.")
        print("Use --update to add files without overwriting.")
        return 1

    target_dir.mkdir(parents=True, exist_ok=True)
    print(f"\n{C.GREEN}📁{C.RESET} {args.name} → {target_dir}\n")

    # INDEX.md
    index_path = target_dir / "INDEX.md"
    if _should_write(index_path, args.update):
        index_path.write_text(_index_template(args.name), encoding="utf-8")
        ok("INDEX.md")

    # L1 — Vision
    if not only or only == "L1":
        l1_dir = target_dir / "1_PRODUCT_VISION"
        l1_dir.mkdir(exist_ok=True)
        vision_path = l1_dir / "VISION.md"
        if _should_write(vision_path, args.update):
            vision_path.write_text(_vision_template(args.name), encoding="utf-8")
            ok("1_PRODUCT_VISION/VISION.md")

    # L2 — Specification
    if not only or only == "L2":
        l2_dir = target_dir / "2_PRODUCT_SPEC"
        l2_dir.mkdir(exist_ok=True)
        spec_path = l2_dir / "SPEC.md"
        if _should_write(spec_path, args.update):
            spec_path.write_text(_spec_template(args.name), encoding="utf-8")
            ok("2_PRODUCT_SPEC/SPEC.md")

    # 3_ARTIFACTS/ — NOT created automatically.
    # Create the folder and subfolders only when placing a real artifact file.

    print(f"\n{C.GREEN}✅ Done:{C.RESET} {target_dir}")
    print(f"{C.YELLOW}Next step:{C.RESET} fill in the documents, then run validation:")
    print(f"  python3 <path-to-skill>/scripts/doc_gen.py validate {args.name!r} --output {args.output!r}\n")
    return 0


# ─── Structure validation ───────────────────────────────────────────────────────

def _check_file(rel_path: str, file_path: Path, spec: dict) -> list[str]:
    errors: list[str] = []

    if not file_path.exists():
        return [f"File missing: {rel_path}"]

    content = file_path.read_text(encoding="utf-8")
    lines   = content.splitlines()

    # Status and date
    if not STATUS_PATTERN.search(content):
        errors.append(f"{rel_path}: line \"**Status:** draft|in review|approved\" is missing")
    if not DATE_PATTERN.search(content):
        errors.append(f"{rel_path}: line \"**Date:** YYYY-MM-DD\" is missing")

    # Required sections
    for heading in spec["headings"]:
        if not any(line.strip().startswith(heading) for line in lines):
            errors.append(f"{rel_path}: required section \"{heading}\" is missing")

    # Placeholders + banned words (line by line, outside code blocks)
    in_code_block = False
    for lineno, line in enumerate(lines, 1):
        stripped = line.strip()
        if stripped.startswith("```"):
            in_code_block = not in_code_block
            continue
        if in_code_block:
            continue

        if PLACEHOLDER_PATTERN.search(stripped):
            errors.append(f"{rel_path}:{lineno}: unfilled placeholder → {stripped[:80]}")

        line_norm = _norm(stripped)
        for phrase in _FORBIDDEN:
            phrase_norm = _norm(phrase)
            if " " in phrase_norm:
                if phrase_norm in line_norm:
                    errors.append(
                        f"{rel_path}:{lineno}: banned phrase \"{phrase}\" → {stripped[:80]}"
                    )
                    break
            else:
                if re.search(rf"\b{re.escape(phrase_norm)}\b", line_norm):
                    errors.append(
                        f"{rel_path}:{lineno}: banned word \"{phrase}\" → {stripped[:80]}"
                    )
                    break

    # Minimum section length (only files with min_section_chars)
    min_chars = spec.get("min_section_chars")
    if min_chars:
        for heading in spec["headings"]:
            if heading in _SECTION_LEN_EXEMPT:
                continue
            sec = _section(content, heading)
            if sec and len(sec.strip()) < min_chars:
                errors.append(
                    f"{rel_path}: section \"{heading}\" is too short "
                    f"({len(sec.strip())} chars, minimum {min_chars})"
                )

    # Numbers in metrics (only for files with check_metrics)
    if spec.get("check_metrics"):
        errors.extend(_check_metrics(content, rel_path))

    return errors


# ─── Consistency and contradiction analysis ───────────────────────────────────

def _check_consistency(vision_path: Path, spec_path: Path) -> list[str]:
    """
    Pairwise and cross-document consistency analysis.
    Returns a list of the problems found.
    """
    issues: list[str] = []

    if not vision_path.exists() or not spec_path.exists():
        issues.append("Cannot check consistency: one or both files are missing.")
        return issues

    vision = vision_path.read_text(encoding="utf-8")
    spec   = spec_path.read_text(encoding="utf-8")

    includes_text = _section(vision, "## In scope")
    excludes_text = _section(vision, "## Out of scope")
    inc_items     = _list_items(includes_text)
    exc_items     = _list_items(excludes_text)

    # ── 1. Pairwise: "In scope" vs "Out of scope" ───────────────────────────
    # Flag only if ≥2 keywords match AND they cover ≥50% of the exc item.
    # This rules out false positives on different facets of the same topic.
    for exc_item in exc_items:
        exc_kw = set(_keywords(exc_item))
        if len(exc_kw) < 2:
            continue
        for inc_item in inc_items:
            inc_kw = set(_keywords(inc_item))
            overlap = exc_kw & inc_kw
            if len(overlap) >= 2 and len(overlap) / len(exc_kw) >= 0.5:
                issues.append(
                    f"[VISION] Contradiction between \"In scope\" and \"Out of scope\":\n"
                    f"  In scope:     \"{inc_item[:70]}\"\n"
                    f"  Out of scope: \"{exc_item[:70]}\"\n"
                    f"  Shared words: {', '.join(sorted(overlap))}"
                )

    # ── 2. "Out of scope" vs SPEC operations/pages ──────────────────────────
    ops_text   = _section(spec, "## Key operations")
    pages_text = _section(spec, "## Pages and screens")
    spec_func  = _norm(ops_text + " " + pages_text)

    for exc_item in exc_items:
        exc_kw = _keywords(exc_item)
        found  = [kw for kw in exc_kw if kw in spec_func]
        if len(found) >= 2:
            issues.append(
                f"[VISION→SPEC] Contradiction: excluded item \"{exc_item[:70]}\" "
                f"found in SPEC (words: {', '.join(found[:5])})"
            )

    # ── 3. VISION "Key capabilities" → coverage in SPEC ─────────────────────
    capabilities_text = _section(vision, "## Key capabilities")
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
                f"[VISION→SPEC] Inconsistency: capability \"{short}\" "
                f"is weakly reflected in SPEC ({found_n}/{len(cap_kw)} keywords)"
            )

    # ── 4. "In scope" items → coverage in SPEC ──────────────────────────────
    for inc_item in inc_items:
        inc_kw = _keywords(inc_item)
        if not inc_kw:
            continue
        found_n  = sum(1 for kw in inc_kw if kw in spec_norm)
        coverage = found_n / len(inc_kw)
        if coverage < 0.35:
            issues.append(
                f"[VISION→SPEC] Inconsistency: \"In scope\" item \"{inc_item[:60]}\" "
                f"is not reflected in SPEC ({found_n}/{len(inc_kw)} keywords)"
            )

    # ── 5. SPEC entities → mentioned in "Key operations" ────────────────────
    entities_text = _section(spec, "## Entities")
    entity_names  = _table_first_col(entities_text, skip_headers={"Entity", "Сущность"})
    ops_norm      = _norm(ops_text)

    for entity in entity_names:
        if _norm(entity) not in ops_norm:
            issues.append(
                f"[SPEC] Inconsistency: entity \"{entity}\" "
                f"is not mentioned in \"Key operations\""
            )

    # ── 6. "Testing" — three required subsections ───────────────────────────
    testing_text = _section(spec, "## Testing")
    for sub in ("**Critical scenarios**", "**Business rules**", "**Negative scenarios**"):
        if sub not in testing_text:
            issues.append(
                f"[SPEC] \"Testing\" does not contain the required subsection {sub}"
            )

    # ── 7. Glossary → terms are used somewhere in the documentation ─────────
    glossary_text  = _section(spec, "## Glossary")
    glossary_terms = _table_first_col(glossary_text, skip_headers={"Term", "Термин"})
    spec_no_gloss  = _norm(spec.replace(glossary_text, ""))
    vision_norm    = _norm(vision)

    for term in glossary_terms:
        term_norm = _norm(term)
        if term_norm not in spec_no_gloss and term_norm not in vision_norm:
            issues.append(
                f"[SPEC] Glossary: term \"{term}\" is defined "
                f"but used nowhere in the documentation"
            )

    # ── 8. Pairwise: "Goal" vs "Out of scope" ───────────────────────────────
    goal_text  = _section(vision, "## Goal")
    goal_norm  = _norm(goal_text)

    for exc_item in exc_items:
        exc_kw = set(_keywords(exc_item))
        if len(exc_kw) < 2:
            continue
        overlap = {kw for kw in exc_kw if kw in goal_norm}
        if len(overlap) >= 2 and len(overlap) / len(exc_kw) >= 0.5:
            issues.append(
                f"[VISION] Possible contradiction: the \"Goal\" section conflicts with \"Out of scope\":\n"
                f"  Out of scope: \"{exc_item[:70]}\"\n"
                f"  Matching words in \"Goal\": {', '.join(sorted(overlap))}"
            )

    return issues


# ─── Artifact check ───────────────────────────────────────────────────────────

def _check_artifacts(target_dir: Path) -> list[str]:
    """Every file in 3_ARTIFACTS/ must be mentioned in at least one document."""
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
                f"[ARTIFACTS] Artifact \"{rel_str}\" is not mentioned in any document"
            )
    return issues


# ─── Commands ─────────────────────────────────────────────────────────────────

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

    print(f"\nConsistency analysis: {args.name}\n")

    issues = _check_consistency(vision_path, spec_path)
    if not issues:
        ok("No contradictions or inconsistencies found.")
        print(f"\n{C.GREEN}✅ Documentation \"{args.name}\" is consistent.{C.RESET}\n")
        return 0

    for issue in issues:
        err(issue)
    print(
        f"\n{C.RED}Summary: {len(issues)} consistency problems. "
        f"Resolve them before finalizing.{C.RESET}\n"
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
            err(f"Folder not found: {output_dir}")
            return 1
        projects = sorted(d.name for d in output_dir.iterdir() if d.is_dir())
        if not projects:
            warn("No projects found.")
            return 0

    doc_files = ["INDEX.md", "1_PRODUCT_VISION/VISION.md", "2_PRODUCT_SPEC/SPEC.md"]
    col1, col2, col3 = 22, 32, 14

    print()
    hdr = f"{'Project':<{col1}} {'File':<{col2}} {'Status':<{col3}} Date"
    print(hdr)
    print("─" * (col1 + col2 + col3 + 14))

    for proj_name in projects:
        target_dir = output_dir / proj_name
        for rel_path in doc_files:
            fp = target_dir / rel_path
            if not fp.exists():
                print(f"{proj_name:<{col1}} {rel_path:<{col2}} {'MISSING':<{col3}}")
                continue
            content  = fp.read_text(encoding="utf-8")
            status_m = STATUS_PATTERN.search(content)
            date_m   = DATE_PATTERN.search(content)
            status   = status_m.group(1) if status_m else "?"
            date_val = date_m.group(0).replace("**Date:**", "").strip() if date_m else "?"
            print(f"{proj_name:<{col1}} {rel_path:<{col2}} {status:<{col3}} {date_val}")

    print()
    return 0


def cmd_update_status(args: argparse.Namespace) -> int:
    if args.status not in VALID_STATUSES:
        err(
            f"Invalid status: \"{args.status}\". "
            f"Allowed: {', '.join(sorted(VALID_STATUSES))}"
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
        err(f"Project folder not found: {target_dir}")
        return 1

    today   = _today()
    updated = 0

    print(f"\nUpdating status \"{args.name}\" → {args.status}\n")

    for rel_path in ("INDEX.md", "1_PRODUCT_VISION/VISION.md", "2_PRODUCT_SPEC/SPEC.md"):
        fp = target_dir / rel_path
        if not fp.exists():
            warn(f"Skipped (does not exist): {rel_path}")
            continue
        content     = fp.read_text(encoding="utf-8")
        new_content = STATUS_PATTERN.sub(f"**Status:** {args.status}", content)
        new_content = DATE_PATTERN.sub(f"**Date:** {today}", new_content)
        if new_content != content:
            fp.write_text(new_content, encoding="utf-8")
            ok(f"{rel_path} → {args.status} | {today}")
            updated += 1
        else:
            warn(f"{rel_path}: status lines not found, skipped")

    if updated:
        print(f"\n{C.GREEN}✅ Files updated: {updated}.{C.RESET}\n")
        print("Don't forget to create a git commit:")
        print(f"  git add {target_dir}/")
        print(f'  git commit -m "docs: status {args.name} → {args.status}"\n')
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
        err(f"Project folder not found: {target_dir}")
        return 1

    print(f"\nValidating documentation: {args.name}\n")

    # ── Block 1: structure, placeholders, banned words, metrics ──────────────
    print(f"{C.YELLOW}── Structure and content ───────────────────{C.RESET}")
    all_errors: list[str] = []
    for rel_path, spec in STRUCTURE.items():
        file_path   = target_dir / rel_path
        file_errors = _check_file(rel_path, file_path, spec)
        if file_errors:
            all_errors.extend(file_errors)
        else:
            ok(f"{rel_path} — {spec['description']}")

    if all_errors:
        print(f"\n{C.RED}Errors:{C.RESET}")
        for e in all_errors:
            err(e)
        print(
            f"\n{C.RED}Summary: {len(all_errors)} errors. "
            f"Fix them before the consistency check.{C.RESET}\n"
        )
        return 1

    # ── Block 2: consistency and contradictions ───────────────────────────────
    print(f"\n{C.YELLOW}── Consistency and contradictions ──────────{C.RESET}")
    vision_path = target_dir / "1_PRODUCT_VISION" / "VISION.md"
    spec_path   = target_dir / "2_PRODUCT_SPEC" / "SPEC.md"
    c_issues    = _check_consistency(vision_path, spec_path)

    if c_issues:
        for issue in c_issues:
            err(issue)
        print(
            f"\n{C.RED}Summary: {len(c_issues)} consistency problems. "
            f"Documentation is not ready for finalization.{C.RESET}\n"
        )
        return 1

    ok("No contradictions or inconsistencies found.")

    # ── Block 3: artifacts not mentioned in the documentation ──────────────────
    artifacts_dir = target_dir / "3_ARTIFACTS"
    has_artifacts = artifacts_dir.exists() and any(
        f for f in artifacts_dir.rglob("*") if f.is_file()
    )
    if has_artifacts:
        print(f"\n{C.YELLOW}── Artifacts ────────────────────────────────{C.RESET}")
        a_issues = _check_artifacts(target_dir)
        if a_issues:
            for issue in a_issues:
                err(issue)
            print(
                f"\n{C.RED}Summary: {len(a_issues)} artifacts not mentioned in the documentation. "
                f"Add references in SPEC.md → ## Artifacts.{C.RESET}\n"
            )
            return 1
        ok("All artifacts are documented.")

    print(
        f"\n{C.GREEN}✅ Documentation \"{args.name}\" passed the full check "
        f"(structure + content + consistency + artifacts).{C.RESET}\n"
    )
    return 0


# ─── CLI ──────────────────────────────────────────────────────────────────────

def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="doc_gen.py",
        description="Product documentation generator and validator",
    )
    sub = parser.add_subparsers(dest="command", required=True)

    # generate
    gen = sub.add_parser("generate", help="Create the document structure")
    gen.add_argument("name", metavar="ProjectName")
    gen.add_argument("--only", choices=["L1", "L2"], metavar="L1|L2",
                     help="Create only one level")
    gen.add_argument("--update", action="store_true",
                     help="Do not overwrite existing files")
    gen.add_argument("--output", default="./docs", metavar="PATH",
                     help="Output folder (default: ./docs)")

    # validate
    val = sub.add_parser("validate", help="Full check: structure + content + consistency")
    val.add_argument("name", metavar="ProjectName")
    val.add_argument("--output", default="./docs", metavar="PATH",
                     help="Documentation folder (default: ./docs)")

    # consistency
    con = sub.add_parser("consistency", help="Consistency and contradiction analysis only")
    con.add_argument("name", metavar="ProjectName")
    con.add_argument("--output", default="./docs", metavar="PATH",
                     help="Documentation folder (default: ./docs)")

    # status
    sta = sub.add_parser("status", help="Document status (all projects or one)")
    sta.add_argument("name", metavar="ProjectName", nargs="?", default=None)
    sta.add_argument("--output", default="./docs", metavar="PATH",
                     help="Documentation folder (default: ./docs)")

    # update-status
    upd = sub.add_parser("update-status", help="Atomically update the status across all project files")
    upd.add_argument("name",   metavar="ProjectName")
    upd.add_argument("status", metavar="status",
                     choices=sorted(VALID_STATUSES),
                     help="draft | in review | approved")
    upd.add_argument("--output", default="./docs", metavar="PATH",
                     help="Documentation folder (default: ./docs)")

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
