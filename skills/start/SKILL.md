---
name: start
description: The single entry point for developing a product from scratch. Determines the project's current state, explains what's happening at each step, and invokes the right skills on the user's behalf. Run this first — this skill drives everything. The other skills (doc-gen, tech-blueprint, implement-project, etc.) don't need to be invoked directly.
license: MIT
compatibility: opencode
metadata:
  level: multi
  output: docs/ + prototype/ + blueprint/ + projects/
---

## Purpose

This skill is **the only one** you need to interact with. It guides you through the entire development pipeline, invoking the other skills at the right moment:

```
[1] Documentation  →  doc-gen              → docs/<name>/
[2] Prototype      →  litefront-prototype  → prototype/         (recommended, not required)
[3] Spec           →  tech-blueprint       → blueprint/<name>/3_TECH_BLUEPRINT/
[4] Implementation →  implement-project    → projects/<name>/
```

Each stage produces artifacts for the next. You don't need to know how the other skills work.

---

## On first run

1. Determine the project state (see "Determining the state")
2. If there's no project at all → explain the pipeline, ask for a name and a description of the idea
3. If a project is found → show the status and offer to continue

**Explanation for a new user (say once at the start):**

> Product development goes through 4 stages:
>
> **1. Documentation** — we'll describe the product: why, for whom, what it consists of.
> Result: VISION.md (goals) + SPEC.md (requirements and scenarios).
>
> **2. Prototype** — we'll build a clickable UI mockup. All data is simulated,
> no backend needed. Helps validate the UX before writing code.
>
> **3. Spec** — we'll design the database, the GraphQL API,
> and the frontend and backend architecture. 5 documents.
>
> **4. Implementation** — we'll write the full working project with tests.
>
> Before moving to each next stage, I'll wait for your "continue".

---

## Workspace structure

```
docs/<ProjectName>/                    ← [Stage 1] product documentation
├── INDEX.md
├── 1_PRODUCT_VISION/VISION.md
└── 2_PRODUCT_SPEC/SPEC.md

prototype/                             ← [Stage 2] UI prototype (shared across all projects)
└── src/...

blueprint/<ProjectName>/               ← [Stage 3] the spec
└── 3_TECH_BLUEPRINT/
    ├── IMPLEMENTATION_GUIDE.md
    ├── DATABASE_MODEL.md
    ├── API_CONTRACTS.md
    ├── ARCHITECTURE.md
    └── TESTING_PLAN.md

projects/<ProjectName>/                ← [Stage 4] implementation
├── backend/
└── frontend/
```

---

## Determining the project state

Check for the presence of files (bash):

```bash
# Stage 1 complete:
test -f docs/<name>/2_PRODUCT_SPEC/SPEC.md

# Stage 2 complete:
test -d prototype/src

# Stage 3 created (files exist):
test -f blueprint/<name>/3_TECH_BLUEPRINT/IMPLEMENTATION_GUIDE.md

# Stage 3 valid (run it and look at the exit code):
python3 ~/.config/opencode/skills/tech-blueprint/scripts/blueprint_validator.py \
  validate "<name>" --output ./blueprint

# Stage 4 complete:
test -d projects/<name>/backend/src && test -d projects/<name>/frontend/src
```

**Multiple projects:** if more than one folder is found in `docs/` — show the list and ask which one we're working on.

**Consistency:** if `docs/<name>/SPEC.md` is newer than `blueprint/<name>/3_TECH_BLUEPRINT/DATABASE_MODEL.md` — warn:
> ⚠️ SPEC.md was changed after the spec was created. Recommend regenerating the spec with the "regenerate spec" command.

---

## Status output format

On a request for "status" / "where are we" / "progress":

```
## Project: <ProjectName>

| Stage | Status | Details |
|------|--------|--------|
| 1. Documentation   | ✅ Done       | docs/<name>/SPEC.md |
| 2. Prototype       | ✅ Done       | prototype/src/ |
| 3. Spec            | ⚠️ Needs fixes | blueprint_validator: 2 errors |
| 4. Implementation  | ⏳ Not started | — |

➡️ Next step: fix the spec errors → say "validate" → then "continue".
```

Statuses: `✅ Done` / `⚠️ Needs fixes` / `🔄 In progress` / `⏳ Not started`

---

## Stage 1: Documentation (doc-gen)

**Say before running:**
> We'll start by describing the product. I'll ask questions — the more precisely you describe
> the idea, the more precise the document will be. Usually takes 5–10 minutes of dialogue.

**Action:** run the `doc-gen` skill
(`~/.config/opencode/skills/doc-gen/SKILL.md`)

The output documents must live in `docs/<ProjectName>/`.

**After completion — show a brief summary:**
```
✅ Documentation created: docs/<name>/

Key entities (from SPEC.md §Entities): <list>
Key operations: <3-5 operations>
Pages: <list>
User roles: <list>
```

**Review gate — wait for confirmation:**
> Read docs/<name>/2_PRODUCT_SPEC/SPEC.md.
> Are all the entities and operations described correctly?
> - Need fixes → describe what to change, I'll update it
> - All correct → say **"continue"**

---

## Stage 2: Prototype (litefront-prototype) — recommended

**Say before running:**
> We'll build a clickable UI prototype — all the pages will work,
> but the data is simulated (no real backend).
> It will help you visually validate SPEC.md and refine the UX before writing code.
>
> This stage can be skipped — say **"skip"**, and the spec
> will be created from the SPEC.md text alone.

**Action:** run the `litefront-prototype` skill
(`~/.config/opencode/skills/litefront-prototype/SKILL.md`)

**After completion:**
```
✅ Prototype created: prototype/

Run: cd prototype && npm run start:dev  →  http://localhost:3000

Implemented pages: <list from the prototype>
```

**Review gate — wait for confirmation:**
> Run the prototype and check all the pages and transitions.
> - Need UI changes → describe them, I'll update the prototype
> - Need to refine SPEC.md → say **"back to documentation"**
> - All correct → say **"continue"**

---

## Stage 3: Spec (tech-blueprint)

**Say before running:**
> We'll create 5 technical documents — a precise description of how the project will be built:
> - **DATABASE_MODEL.md** — DB schema (Prisma ORM, PostgreSQL)
> - **API_CONTRACTS.md** — GraphQL API: operations, types, access rights
> - **ARCHITECTURE.md** — NestJS modules and the frontend's FSD slices
> - **TESTING_PLAN.md** — concrete test scenarios
> - **IMPLEMENTATION_GUIDE.md** — the developer's guide: stack, what's already done, commands
>
> After generation, the documents are automatically checked by the validator against 9 criteria.

**Action:** run the `tech-blueprint` skill
(`~/.config/opencode/skills/tech-blueprint/SKILL.md`)

**After completion — run the validator:**
```bash
python3 ~/.config/opencode/skills/tech-blueprint/scripts/blueprint_validator.py \
  validate "<name>" --output ./blueprint
```

If there are errors → **fix them before showing the user** (don't move to the next stage with errors).

**After successful validation — show a summary:**
```
✅ Spec created and valid: blueprint/<name>/3_TECH_BLUEPRINT/

Prisma models: <list>
GraphQL operations by domain:
  <Domain 1>: query1, mutation1
  <Domain 2>: query2, mutation2
NestJS modules to implement: <list>
New FSD slices: <list (without boilerplate slices)>
```

**Review gate — wait for confirmation:**
> Review the spec, especially DATABASE_MODEL.md and API_CONTRACTS.md.
> - Want to change something → edit the files in blueprint/<name>/3_TECH_BLUEPRINT/ and say **"validate"**
> - Want to fully regenerate → say **"regenerate spec"**
> - All correct → say **"continue"**

---

## Stage 4: Implementation (implement-project)

**Say before running:**
> The final and longest stage — full implementation per the approved spec.
> On completion:
> - projects/<name>/backend/ — a working NestJS API
> - projects/<name>/frontend/ — a React SPA
> - all tests from TESTING_PLAN.md implemented
> - npm run check && npm run test:all green in both projects

**Final check before running:**
- [ ] SPEC.md approved by the user
- [ ] Prototype reviewed (or explicitly skipped)
- [ ] Spec passed validation without errors (`blueprint_validator.py` → exit 0)

If even one item isn't satisfied → warn and ask for confirmation:
> ⚠️ I recommend resolving [the issue] first. Continue anyway? ("yes" / "no")

**Action:** run the `implement-project` skill
(`~/.config/opencode/skills/implement-project/SKILL.md`)

**After completion:**
```
🎉 Project implemented!

Run the backend:
  cd projects/<name>/backend && npm run start:dev

Run the frontend:
  cd projects/<name>/frontend && npm run start:dev

All checks: npm run check && npm run test:all → ✅

What's next:
  1. Configure the OIDC provider (Logto/Keycloak) in the .env of both projects
  2. Bring up PostgreSQL + Redis (docker compose up -d)
  3. npm run db:migrations:apply
```

---

## All recognized commands

| Command | Action |
|---------|---------|
| "start", "begin", "new project" | Determine the state → start or continue |
| "status", "where are we", "progress" | A state table for all 4 stages |
| "continue", "go", "next step" | Run the next unfinished stage |
| "skip" | Skip the prototype (stage 2 only) |
| "back to documentation" | Re-run doc-gen |
| "regenerate prototype" | Re-run litefront-prototype |
| "regenerate spec" | Re-run tech-blueprint |
| "validate", "check the spec" | Run blueprint_validator → show the result |
| "show the documents" | List all created files with their paths |
| "how to run" | Run commands for the project's current state |
| "what is [skill]" | Explain the purpose of a specific skill |
| "help" | Show this table |

---

## Error handling

**Documentation doesn't reflect the idea:**
→ User describes the fixes → update SPEC.md → show what changed

**Prototype won't run:**
→ `cd prototype && npm install && npm run start:dev` → show the full error text

**blueprint_validator errors:**
→ Show each error + an explanation → fix automatically where possible → re-run the validation

**SPEC.md updated after the spec was created:**
→ Warn that it's stale → suggest "regenerate spec"

**Tests don't pass:**
→ Show the failing tests → suggest fixing them → re-run `npm run test:all`

**Unfamiliar user question:**
→ Answer in the context of the current stage + hint at the next step

---

## Key behavior principles

1. **Never move to the next stage without an explicit "continue"** — always wait
2. **Explain WHAT and WHY** before each action — one paragraph, no extra detail
3. **Show progress**: "Stage 2 of 4 — Prototype" at the start of each block
4. **On an error — don't continue**: fix → verify → only then move forward
5. **Brief summaries**: 5–10 bullet points of key entities/operations — don't retell the whole document
6. **One question at a time** — don't overload the user with a questionnaire
7. **Remember the boilerplate constraints** — don't offer to implement what's already done (`Profile`, `features/auth`, `pages/404`)
