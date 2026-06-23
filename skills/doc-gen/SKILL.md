---
name: doc-gen
description: Generates MVP+ level product documentation. Run when the user is explicitly planning a new product/service/application or asks for help with project structure. Do NOT run for questions about existing code, one-off tasks ("write a script", "fix a bug"), or learning exercises.
license: MIT
compatibility: opencode
metadata:
  level: multi
  output: folder with markdown files
---

## Purpose

The skill creates two documents that fully describe a product for development:
- **Vision** (VISION.md) — why the product is being built and for whom
- **Specification** (SPEC.md) — what the product consists of, enough to begin development

**When to run:**
- The user describes a **new product** or **service** and plans to build it
- The user asks to **structure a project** or describe its architecture
- The user says "I want to build X" in the context of a product or service

**When NOT to run:**
- A question about the code of an existing project
- A one-off task ("write a script", "add a function")
- A learning or test project with no intention to develop it further

---

## Inputs

The user provides a **minimal description**:
1. What problem the product solves
2. Who will use it
3. What needs to work first

If there isn't enough information — ask **no more than 2 clarifying questions**.

---

## Output structure

```
<project name>/
├── INDEX.md                             — table of contents
├── 1_PRODUCT_VISION/
│   └── VISION.md                        — product vision
├── 2_PRODUCT_SPEC/
│   └── SPEC.md                          — product specification
└── 3_ARTIFACTS/                         — supporting materials
    ├── legal/                           — legal documents (contracts, terms of service, privacy policies)
    ├── media/
    │   ├── images/                      — images (logos, screenshots, diagrams, mockups)
    │   └── video/                       — video materials (demos, tutorials)
    ├── content/                         — page copy and marketing materials
    └── examples/                        — content samples, templates, test data
```

---

## Artifacts (3_ARTIFACTS/)

Artifacts are supporting materials that the documents reference. They are not source code and not additional documentation — they are concrete files needed to develop or launch the product.

### Folder organization

| Folder | What to store |
|-------|-------------|
| `legal/` | Contracts, terms of service, privacy policies, terms of use, license agreements |
| `media/images/` | Logos, branding, interface screenshots, architecture diagrams, page mockups |
| `media/video/` | Product demos, tutorial videos, presentations |
| `content/` | Site page copy, marketing copy, email templates, FAQs |
| `examples/` | Sample form submissions, content templates, test data sets |

### Rules for working with artifacts

- **Create the `3_ARTIFACTS/` folder and its subfolders only when there is an actual file to place there.** Empty folders are not allowed.
- **Every artifact** must be explicitly mentioned in at least one document (VISION.md, SPEC.md, or INDEX.md)
- **SPEC.md** must contain an `## Artifacts` section with links to all files in `3_ARTIFACTS/`
- **INDEX.md** must contain quick links to the key artifacts
- Links are **relative**, for example `../3_ARTIFACTS/legal/privacy-policy.md`
- File names use Latin letters and hyphens, no spaces: `privacy-policy.pdf`, `logo-main.svg`
- **Do not add placeholder links** in SPEC.md to files that don't exist yet

---

## Workflow

> **Script:** `~/.config/opencode/skills/doc-gen/scripts/doc_gen.py`
> The script is NOT copied into the project — it is used directly from the skill folder.
> In the examples below: `{DOC_GEN}` = the full path above.

### Step 1. Gather information
Ask no more than **2 clarifying questions**. If the requested functionality is clearly excessive for the first version — point this out before starting generation.

### Step 2. Create the file structure
```bash
python3 {DOC_GEN} generate "ProjectName"

# Vision only (VISION.md):
python3 {DOC_GEN} generate "ProjectName" --only L1

# Specification only (SPEC.md):
python3 {DOC_GEN} generate "ProjectName" --only L2

# Add to existing files without overwriting them:
python3 {DOC_GEN} generate "ProjectName" --update
```

The `generate` command does **not** create the `3_ARTIFACTS/` folder automatically. The folder and its subfolders are created **only when an actual artifact file is placed there**. If there are no artifacts — the `## Artifacts` section in SPEC.md **must still be present** with an explicit note: "No artifacts."

### Step 3. Fill in the documents
Fill in **strictly in this order**: VISION.md → SPEC.md. All links are relative.

**Mandatory filling rules:**
- There must be no unfilled placeholders (text in `[square brackets]`)
- Each section is either **filled with concrete content** or **explicitly contains a negation** (for example: "No integrations.")
- It is **forbidden** to write "if needed", "optionally", "possibly" — only concrete statements
- No optional decisions: either something **is** in the product, or it's **explicitly stated** that it isn't

### Step 4. Full check: structure + consistency **(mandatory)**
After **any** change to the documentation, run the full validation:
```bash
python3 {DOC_GEN} validate "ProjectName"
```
**The documentation is not considered ready until validation passes without errors.**

The script runs **two blocks of checks in sequence**:

**Block 1 — Structure and content:**
- All mandatory files exist
- All mandatory sections are present in each file
- The **Status** and **Date** lines are present
- There are no unfilled placeholders `[...]`
- Forbidden words/phrases are absent ("possibly", "convenient", "fast", "optionally", etc.)
- VISION/SPEC sections are not too short (minimum 60 characters)
- Target values in "Success metrics" contain numbers
- The `## Artifacts` section is present in SPEC.md (either with links or with "No artifacts.")
- Every file in `3_ARTIFACTS/` is mentioned in at least one document (no "dead" artifacts)

**Block 2 — Consistency and contradictions (automatic):**
- `[VISION]` Pairwise check: "What's included" ↔ "What's NOT included" (≥2 shared words in a pair → flag)
- `[VISION]` Pairwise: "Goal" conflicts with "What's NOT included"
- `[VISION→SPEC]` Excluded VISION items found in SPEC operations/pages
- `[VISION→SPEC]` Key VISION capabilities not reflected in SPEC (<40% of keywords)
- `[VISION→SPEC]` "What's included" items not covered in SPEC (<35% of keywords)
- `[SPEC]` Entities not mentioned in "Key operations"
- `[SPEC]` Mandatory "Testing" subsections missing
- `[SPEC]` Glossary terms used nowhere

To run only the consistency analysis in isolation:
```bash
python3 {DOC_GEN} consistency "ProjectName"
```

### Step 5. AI quality check **(mandatory)**
After a successful `validate` — perform this manually. The script does not check these things.

**Readability and formatting:**
1. Each capability in "Key capabilities" has a **bold name** (`**Name**:`) and an arrow `→` — the format "user does X → gets Y"
2. The "Problem", "Target audience", "Goal", "How the system works" sections contain at least one **bold** key statement for quick scanning
3. No monolithic paragraphs — sentences are broken up; no sentence is longer than 2–3 lines
4. The "Key capabilities", "What's included", "What's NOT included", "Key operations" sections are formatted as lists, not solid text
5. Bold is not applied to everything — only to key words/actions (no more than 30% of lines in a section)

**Logic and meaning:**
6. The "Goal" concretely solves the "Problem" — no gap between the pain described and the product described
7. Each success metric measures the stated "Goal", not general business growth
8. "What's NOT included" does not contradict the "Key capabilities" in meaning (not just by keywords — check the actual meaning)
9. The test scenarios cover the real risks of this domain, not just boilerplate cases

**Language:**
10. "Key operations" describe the **result for the user**, not a technical process ("the user sees a list of orders", not "the system returns an array of objects")
11. The terms in the "Glossary" genuinely need explanation — don't add obvious words like "user" or "system"
12. There is no technical jargon anywhere (endpoint, CRUD, ORM, REST, JSON, etc.) — exception: an API product, where this is acceptable in the "Pages and screens" section

### Step 6. Git commit **(mandatory)**
After every change to the documentation, create a **commit**:
```bash
git add docs/
git commit -m "docs: update documentation <ProjectName>"
```
If there is **no git repository** — create one before the first commit:
```bash
git init
git add .
git commit -m "docs: initial documentation <ProjectName>"
```
**A git repository is required.** Without it, the change history is unavailable.

---

## Management commands

### Documentation status
```bash
# All projects in the docs/ folder:
python3 {DOC_GEN} status

# A specific project:
python3 {DOC_GEN} status "ProjectName"
```
Prints a table: project / file / status / date.

### Status update
```bash
python3 {DOC_GEN} update-status "ProjectName" "in review"
python3 {DOC_GEN} update-status "ProjectName" "approved"
python3 {DOC_GEN} update-status "ProjectName" "draft"
```
Atomically updates **Status** and **Date** in all three files (INDEX.md, VISION.md, SPEC.md). Afterwards — create a git commit.

### Consistency analysis only
```bash
python3 {DOC_GEN} consistency "ProjectName"
```
Runs only Block 2 without the structural check. Useful during iterative content edits.

---

## Managing document status

Each document contains a status line at the top:
```
**Status:** draft | **Date:** YYYY-MM-DD
```

| Status | Transition condition |
|--------|-----------------|
| `draft` | Document created, not reviewed by the team |
| `in review` | Document submitted for review |
| `approved` | Document agreed upon and ready for development |

Update the **Date** on every significant change. Change history — `git log -- <file>`.

---

## Rules for writing the vision (VISION.md)

**Bad example:**
- "Build authentication"
- "An app for notes"

**Good example:**
- "Sign-in via external services (Google, GitHub) without storing passwords"
- "A mobile notes app with offline support for users with unstable connectivity"

### Mandatory structure of VISION.md

```markdown
# <Product name>

**Status:** draft | **Date:** YYYY-MM-DD

## Problem
What exactly is inconvenient or doesn't work. The context. 2–4 sentences.

## Target audience
The user's role and work context. Not "all users", but a specific type. 2–4 sentences.

## Goal
What exactly we're building and for whom. 2–4 sentences.

## Key capabilities
1. **<Name>**: the user does X → gets Y
2. ...

## Success metrics
| Metric | Target value |
|---------|------------------|
| ... | a specific number |

## What's included (project scope)
Functionality that will be implemented:
1. ...

## What's NOT included
Explicit exclusions to prevent scope creep:
- Does not include X
- Does not include Y
```

### Section rules

- No mention of **development technologies or stack**
- **Concrete metrics** with numbers (not "faster" — but "under 2 seconds")
- Each capability explains the **value to the user**
- Both project-scope sections are filled in

---

## Rules for writing the specification (SPEC.md)

The document describes **what the product consists of** in business terms. No technical implementation details — but **concrete enough** that a developer can start without additional questions.

**Mandatory sections:** Links, How the system works, Glossary, Entities, Pages and screens, Key operations, Integrations, Testing.

### Mandatory structure of SPEC.md

```markdown
# Product specification: <Product name>

**Status:** draft | **Date:** YYYY-MM-DD

## Links
- Vision: [VISION.md](../1_PRODUCT_VISION/VISION.md)

## How the system works
A brief description of what parts the product consists of and how they interact.
Example: "A web application with a personal dashboard and an admin panel. Data is stored
centrally, access is through a browser with no app installation required."

## Glossary
The product's key concepts. Each term has one precise definition with no synonyms.

| Term | Definition |
|--------|-------------|
| ... | ... |

## Entities
The main objects the system works with. In business terms, not database terms.

| Entity | Description | Key properties |
|----------|----------|-------------------|
| User | A registered participant in the system | Name, email, role, registration date |
| ... | ... | ... |

### Entity lifecycle
For each key entity — the allowed statuses and the transitions between them.

| Entity | Statuses | Transitions |
|----------|---------|----------|
| Order | draft → confirmed → completed → cancelled | draft→confirmed: the user clicks "Place order" |

## Pages and screens
An exhaustive list of the pages and screens that must be built.

| Page | Purpose | Key elements |
|----------|------------|-------------------|
| Home | Entry point, overview of capabilities | Navigation, benefits block, call-to-action button |
| Sign-up | Account creation | Form, validation, email confirmation |
| ... | ... | ... |

## Key operations
What users can do in the system. Group by role when there are several types.

**<Role or "All users">:**
- Operation 1: a brief description of the result
- Operation 2: a brief description of the result

## Integrations
External services without which the product doesn't work.
If there are no integrations — write: "No integrations."

| Service | Purpose |
|--------|------------|
| ... | ... |

## Testing
Functionality covered by tests.

**Critical scenarios** (must work without errors):
- The user signs up and logs into the system
- ...

**Business rules** (correctness of calculations and constraints):
- Calculation of X under condition Y
- Validation of Z
- ...

**Negative scenarios** (system behavior on errors and cancellations):
- The user enters a wrong password → the system shows "Invalid login or password", access is denied
- The user cancels an order → the status changes to "cancelled", the money is refunded within 3 business days
- ...

## Artifacts
Supporting materials for developing and launching the product.
If there are no artifacts — write: "No artifacts."

| File | Type | Purpose |
|------|-----|-----------|
| [privacy-policy.md](../3_ARTIFACTS/legal/privacy-policy.md) | Legal | Privacy policy |
| [logo-main.svg](../3_ARTIFACTS/media/images/logo-main.svg) | Image | The product's main logo |
| [homepage-copy.md](../3_ARTIFACTS/content/homepage-copy.md) | Content | Copy for the home page |
| ... | ... | ... |
```

### Section rules

- Business language, not developer language: "list of products", not "array of objects"
- Pages — an **exhaustive** list, none omitted
- Entities — only those the product **actually needs**
- Operations describe the **result**, not the technical implementation
- Testing describes **what to verify**, not how to implement it
- No stack, frameworks, or architectural patterns
- **Negative scenarios are mandatory**: what happens on errors, cancellations, conflicts

---

## INDEX.md — Table of contents

```markdown
# Product documentation: <Name>

**Status:** draft | **Date:** YYYY-MM-DD

## Navigation

| Document | Description |
|----------|----------|
| [Vision](./1_PRODUCT_VISION/VISION.md) | Problem, audience, goal, project scope |
| [Specification](./2_PRODUCT_SPEC/SPEC.md) | Entities, pages, operations, testing |

## Quick links

- [Key capabilities](./1_PRODUCT_VISION/VISION.md#key-capabilities)
- [Pages and screens](./2_PRODUCT_SPEC/SPEC.md#pages-and-screens)
- [Testing](./2_PRODUCT_SPEC/SPEC.md#testing)
- [Artifacts](./2_PRODUCT_SPEC/SPEC.md#artifacts)

## Artifacts

| File | Description |
|------|----------|
| [Privacy policy](./3_ARTIFACTS/legal/privacy-policy.md) | Terms of personal data processing |
| ... | ... |
```

*If there are no artifacts, delete the "Artifacts" section and its table. Also delete the "Artifacts" line from "Quick links".*

---

## Edge Cases

### Product without a UI (API, CLI, library)
Replace the "Pages and screens" section with "Endpoints and commands" using a similar table structure. At the top of SPEC.md, state explicitly: "The product has no user interface." The validator accepts the renamed section — the mandatory `## Pages and screens` heading must still be present, with the content replaced.

### Multiple user types (roles)
In "Key operations" — a subsection for each role. In the Glossary — a definition of each role. The "User" entity reflects the role split (a "role" property). Success metrics — separately for each role.

### Contradiction between VISION.md and SPEC.md
The script (`validate` / `consistency`) detects contradictions automatically. Resolution priority — VISION.md. Clarify with the user before finalizing SPEC.md. The documentation cannot be considered ready if `validate` finishes with errors in the "Consistency" block.

### The user changes requirements after generation
1. Update VISION.md.
2. Check consistency with SPEC.md (the checklist from Step 5).
3. Run the validation.
4. Create a commit describing what changed and why.
Do not regenerate the files — use `--update` or edit them directly.

### Project name with spaces or special characters
Use CamelCase or a hyphen: `MyProject`, `my-project`. Spaces and the characters `/ \ : * ? " < > |` are not allowed — doc_gen.py will reject them with an error.

### Product without integrations
Delete the table. State explicitly: "No integrations." This is the only section where an explicit negation is acceptable instead of a table.

### Product with multiple sub-products (monorepo)
Run `doc_gen.py generate` separately for each sub-product with different names. The shared top-level INDEX.md is created manually with links to each folder.

### Concept change (pivot)
Create a git branch. Rewrite VISION.md from scratch. Revise SPEC.md completely. Don't mix the old and new concepts in one document — the old version is kept in the git history.

### Artifacts not yet created
Until a file physically exists in `3_ARTIFACTS/` — **do not mention it in SPEC.md and do not create a subfolder for it**. The `## Artifacts` section contains only existing links or "No artifacts." The rule: first there's a real file — then a folder is created and a link is added. Never the other way around.

### Too many artifacts of one type
When there are many files of one type — create nested subfolders inside the standard ones:
`media/images/screenshots/`, `media/images/mockups/`, `legal/agreements/`, `content/emails/`.
The structure should be self-evident without explanation — don't create a subfolder for a single file.

### Artifacts without documentation
An artifact cannot exist without being mentioned in a document. If a file in `3_ARTIFACTS/` is referenced nowhere — either delete it or add a link in SPEC.md. "Dead" artifacts are not allowed.

### Validator false positives on placeholders
A `[text]` pattern not followed by `(` is treated as a placeholder. Exceptions: links `[text](url)` — not a placeholder. If the content genuinely needs square brackets not as a placeholder (for example, `[RFC 7231]`), use HTML escapes: `&#91;RFC 7231&#93;`.

---

## Key constraints

- No technologies or stack — the document is implementable in **any** language and platform
- The language is understandable to a business owner **without a technical background**
- Each page in SPEC.md is **linked** to a capability from VISION.md
- There must be no unfilled placeholders
- **No optional decisions**: either it's explicitly described, or it's explicitly stated as "no"
- The documentation language is **Russian**
- All links are **relative**
- After every change — **validation** (`{DOC_GEN} validate`) and a **git commit**
- **`3_ARTIFACTS/` is not created automatically** — only when an actual file is placed there
- **Every artifact** in `3_ARTIFACTS/` is mentioned in at least one document
- **SPEC.md must** contain an `## Artifacts` section (either with a table of links or with "No artifacts.")
