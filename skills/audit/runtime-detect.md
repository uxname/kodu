# Runtime Detection & Stack Profile (shared canon)

This file is the single source of truth for how the audit skill detects a project's
runtime and loads the corresponding **stack profile** from `./skills/audit/stacks/`.

Every code-audit skill carries an inline copy of the block below (the
"Runtime Detection & Stack Profile" section). The inline copy is required — it guarantees
the skill works standalone (`/audit-<name>`), even when the `/audit` orchestrator
was not run. This file is the reference and the place to edit the canon in sync.

## Principle

One audit run = **exactly one runtime**. The skill:
1. accepts the runtime injected by the orchestrator IF it was passed; otherwise
2. detects exactly one runtime in the current directory; then
3. reads the profile and uses its tools/idioms/anti-patterns.

Polyglot repositories (e.g. a Go backend + a Node frontend in one tree)
are audited per subproject: run the audit separately inside each subdirectory.

## Canonical block (embedded into every code skill)

```
## Runtime Detection & Stack Profile

This audit is stack-agnostic: the checks are phrased neutrally, while the specifics
(tools, idioms, anti-patterns, examples) come from the stack profile.

1. **Profile passed as context?** If the `/audit` orchestrator passed
   `runtime=<id>` and/or the profile contents — use it, skip steps 2–3.

2. **Otherwise detect EXACTLY ONE runtime** for this directory:
   ```bash
   if   [ -f package.json ]; then echo "runtime=node"
   elif [ -f go.mod ]; then echo "runtime=go"
   elif [ -f pyproject.toml ] || [ -f requirements.txt ] || [ -f setup.py ]; then echo "runtime=python"
   elif [ -f Cargo.toml ]; then echo "runtime=rust"
   elif [ -f pom.xml ] || ls build.gradle* settings.gradle* >/dev/null 2>&1; then echo "runtime=java"
   else echo "runtime=generic"; fi
   ```
   One run = one runtime; do not mix backend and frontend. If several
   markers are found (monorepo) — pick the one matching the current scope/analyzed
   files and record the choice in the Audit Coverage section.

3. **Load the profile** via Read: `./skills/audit/stacks/<runtime>.md`
   (fallback `./skills/audit/stacks/_generic.md` if the file is not found).

Then use the profile:
- **Tools** — from the profile's "Tooling by category" section (the
  "Tooling support" section below refers to categories, not commands).
- **PASS expectations** — from "Idioms"; **FAIL phrasings** — from "Anti-patterns".
- **Targeted hints** — from "Check-ID hints" by this audit's prefix.
- If the profile is `tier: general` or `runtime=generic` → mark stack-specific findings
  without unambiguous evidence as `🔍 UNVERIFIED`, not `❌ FAIL`. Checks
  whose mechanism does not exist in the runtime should be marked `N/A`.
```

## Injection from the orchestrator

The master skill `audit/SKILL.md` detects the runtime once (Step 0), reads the profile
once, and passes `runtime=<id>` + the profile contents to every `Skill()` through the same
channel as the baseline. The sub-skills see this in step 1 and skip their own
detection. Standalone operation is not lost: the inline block runs on its own if
there is no injection.
