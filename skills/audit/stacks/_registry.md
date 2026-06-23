# Stack Profiles — registry

A mapping table of "marker file → profile id". Used by the
Runtime Detection block (see `../runtime-detect.md`). Checked top to bottom,
selecting the first matching runtime (one run = one runtime).

| Priority | Marker file(s) | runtime id | Profile | Tier |
|-----------|----------------|------------|---------|------|
| 1 | `package.json` | `node` | `node.md` | first-class |
| 2 | `go.mod` | `go` | `go.md` | first-class |
| 3 | `pyproject.toml` / `requirements.txt` / `setup.py` | `python` | `python.md` | general |
| 4 | `Cargo.toml` | `rust` | `rust.md` | general |
| 5 | `pom.xml` / `build.gradle*` / `settings.gradle*` | `java` | `java.md` | general |
| — | none of the above | `generic` | `_generic.md` | fallback |

## Tier — what it means

- **first-class** — the profile contains concrete tools and idioms; findings
  may be `❌ FAIL` with precise evidence.
- **general** — the profile provides neutral phrasings and general idioms, but without
  guaranteed tools; mark stack-specific findings without unambiguous
  evidence as `🔍 UNVERIFIED`.
- **fallback** (`generic`) — no tools; only stack-neutral
  checks work (doc links, secrets, naming readability), everything else → `🔍 UNVERIFIED`.

## Tool categories (common to all profiles)

The "Tooling by category" section in each profile uses the same set
of keys, so the skills reference a category rather than a command:

`unused-code`, `clone-detection`, `dep-audit`, `env-extraction`, `arch-lint`,
`lint/format`, `type-check`, `test-run`, `secret-scan`.

## Profile structure

Each `stacks/<id>.md` has the same sections:
1. **Detection signals** — marker files.
2. **Tooling by category** — a table by the categories above.
3. **Idioms** — what "correct" looks like (PASS expectations).
4. **Anti-patterns** — what FAIL looks like (1 line of code per item).
5. **Check-ID hints** — targeted hints by Check ID prefix.

## Adding a new stack

1. Create `stacks/<id>.md` following the structure above.
2. Add a row to the registry table.
3. Add a branch to the bash detection in `../runtime-detect.md` and in the skills' inline blocks.
