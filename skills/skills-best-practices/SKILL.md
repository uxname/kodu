---
name: skills-best-practices
description: Design reusable AI agent skills in SKILL.md format — activation rules, constraints, workflow, output requirements, failure handling, and YAML frontmatter. Use when the user wants to create, improve, or standardize another skill for an AI agent.
---

## PURPOSE

A meta-skill for designing other skills.
Part 1 — quality principles. Part 2 — a step-by-step process for creating and improving.

---

# Part 1. Principles

## Skill vs prompt

A skill is an SOP, not an instruction to a chatbot. A good skill describes:

1. When to use it / when NOT to use it.
2. How to make decisions.
3. Which tools to apply.
4. What to do under uncertainty.
5. What a good / bad result looks like.

## 6 key patterns

| # | Pattern | Bad | Good |
|---|---------|-------|--------|
| 1 | Role + constraints separated | "You're a K8s expert" | ROLE + CONSTRAINTS |
| 2 | A "don't know" branch | "Answer the question" | If there's no data — stop, list what's missing |
| 3 | Positive instructions | "Don't be verbose" | "Use 3–7 sentences" |
| 4 | Few-shot examples | Instructions only | INPUT / GOOD OUTPUT / BAD OUTPUT |
| 5 | Explicit hierarchy | 20 rules | PRIORITY 1 > 2 > 3 |
| 6 | Tools separated | "Use search if needed" | SEARCH TOOL: when, when not, what after |

## Anti-patterns

- a wall of text with no structure;
- dozens of prohibitions in a row;
- vague words ("be smart", "where possible");
- no examples, stop criteria, or uncertainty branch;
- business logic and formatting in one section.

## What not to optimize

- **endless examples** — 1–2 GOOD/BAD pairs are enough, the rest goes in an appendix;
- **bloated YAML** — only name + description, no logic;
- **a tutorial-skill** — don't explain the basics of the language or technology, this is an SOP, not a tutorial;
- **perfect isolation** — a recurring pattern is better extracted into a reference than duplicated;
- **a waterfall structure** — if a section isn't needed, leave "(not applicable)" rather than deleting it.

---

# Part 2. Process

First, determine the mode: **CREATE** (a new skill) or **IMPROVE** (an existing one).

---

## Pre-flight gate

**Don't start generating the final SKILL.md until all the critical fields are determined.**

Critical (without them — STOP, ask the user):
- [ ] the skill's purpose
- [ ] the platform (OpenCode / Claude Code / universal)
- [ ] the audience (developer / team lead / QA / AI agent)

Secondary (if missing — generate with explicit assumptions):
- [ ] the skill type (analytical / generative / review / diagnostic / utility)
- [ ] whether tools are needed
- [ ] the expected result format
- [ ] whether the context of an existing project is needed

If even one critical field is missing — **don't write the final file, ask questions**.

---

## Skill Type Inference

If the user didn't specify the skill type — determine it before starting work on the brief.

| Type | Characteristic pattern | What it affects |
|-----|---------------------|---------------|
| Analytical | Checking, finding problems, comparing | PROCESS: edge cases, branching on the data |
| Generative | Creating text, code, documentation | OUTPUT REQUIREMENTS: the exact structure of the result |
| Diagnostic | Finding the cause, investigating failures | FAILURE HANDLING: 3+ failure scenarios |
| Review | Evaluation, classification, verdict | DECISION RULES: a clear severity scale |
| Utility | Transformation, running, migration | TOOL USAGE: exact commands, flags |

If the type is still unclear — pick the most likely one, record it in the brief, and mark it as inferred.

---

## CREATE: a new skill

### Step 1. Skill Brief (the single source of truth)

Fill in the brief. The fields have a strict format, not free-form text.

```yaml
brief:
  purpose: '<verb + what it does. One phrase.>'
  activation: ['<trigger 1>', '<trigger 2>']        # 2-4 items
  do_not_use: ['<condition 1>', '<condition 2>']          # 2+ items
  required_inputs: ['<field 1>', '<field 2>']           # required + optional
  tools: ['<tool>: <when>']                      # or empty []
  output_format: '<the result structure. One paragraph.>'
  failure_handling: ['<scenario 1>: <action>', '<scenario 2>: <action>']  # 2+
  examples:                                            # 1-2 items
    - input: '<input>'
      good: '<correct answer>'
      bad: '<incorrect answer>'
```

Until the brief is filled in completely — don't move on to rendering.

### Step 2. YAML frontmatter

```yaml
---
name: <kebab-case>
description: <verb + object>. Use when <trigger>.
---
```

Description formula:
> **active-voice verb + what it does + ". Use when / Use for" + trigger**

Rules:
- verb + object: "Analyze Node.js dependencies", "Review code changes"
- 1–2 triggers: "Use when reviewing package.json"
- `help with`, `assist`, `support` without a concrete action are **forbidden**
- no platform details (except strict single-platform)
- no logic, examples, or instructions
- one paragraph, max 3 sentences

Testability test: **read only the description. Is it immediately clear when the skill is needed and when not?** If not — rewrite it.

Forbidden:
```yaml
description: Helps with code stuff. Use when needed.
```
Correct:
```yaml
description: Review code changes for bugs, security issues, and convention violations. Use when reviewing a PR or diff before merge.
```

### Step 3. Render SKILL.md

Each brief field maps to a section. The order is fixed. This is the only template.

| Brief field | → | Section | Content format |
|-------------|---|---------|-------------------|
| `purpose` | → | `## PURPOSE` | One paragraph |
| `activation` | → | `## ACTIVATION` | A bulleted list of triggers |
| `do_not_use` | → | `## DO NOT USE WHEN` | A bulleted list of conditions |
| `required_inputs` | → | `## INPUTS` | A bulleted list of fields |
| _(derived)_ | → | `## PROCESS` | A numbered list of concrete actions (not "analyze", but "gather data" → "check constraints" → "choose an option"), at least 3 steps |
| _(derived)_ | → | `## DECISION RULES` | A bulleted list of priorities. At least one prioritization rule (PRIORITY 1 / 2 / 3) |
| `tools` | → | `## TOOL USAGE` | A description or "(not applicable)" |
| `output_format` | → | `## OUTPUT REQUIREMENTS` | One paragraph or a list |
| `failure_handling` | → | `## FAILURE HANDLING` | A bulleted list of scenarios |
| `examples` | → | `## EXAMPLES` | INPUT / GOOD OUTPUT / BAD OUTPUT |

Section requirements:

| Section | Always substantive | May be "(not applicable)" |
|--------|---------------------|------------------------------|
| PURPOSE | yes | no |
| ACTIVATION | yes | no |
| DO NOT USE WHEN | yes | no |
| INPUTS | yes | no |
| PROCESS | yes (>=3 steps) | no |
| DECISION RULES | yes | no |
| TOOL USAGE | no | yes |
| OUTPUT REQUIREMENTS | yes | no |
| FAILURE HANDLING | yes (>=2 scenarios) | no |
| EXAMPLES | no (but >=1 if present) | yes |

### Step 4. Conflict check

1. **Cross-section** — ACTIVATION doesn't contradict DO NOT USE; PROCESS doesn't duplicate DECISION RULES.
2. **With the user's request** — the request takes priority, but the skill doesn't violate the quality principles.
3. **Runtime conflict** — if two instructions might contradict, set PRIORITY explicitly.
4. **Ambiguity** — no phrases like "where possible", "if appropriate", "be careful".
5. **Length** — the skill should be as short as possible but as long as necessary. If the document contains large examples, references, or commands — consider extracting them into reference files.

### Step 5. Modularity check

- [ ] A pattern repeats >2 times? → extract it into a reference file
- [ ] An example >30 lines? → shorten it or move it to an appendix
- [ ] Tools with long commands? → move them into a TOOL USAGE reference
- [ ] Does the skill reference external documents? → make it self-contained or declare the dependency explicitly
- [ ] Sprawling? → trim in this order: examples → process → explanations

### Step 6. Quality check

General minimums:
- PURPOSE — one complete sentence with a verb
- ACTIVATION — at least 2 triggers
- DO NOT USE — at least 2 refusal conditions
- PROCESS — at least 3 numbered steps with concrete actions, not assessments
- DECISION RULES — at least one prioritization rule (PRIORITY 1 / 2 / 3)
- FAILURE HANDLING — at least 2 scenarios with concrete actions
- EXAMPLES — at least 1 GOOD / BAD pair
- OUTPUT REQUIREMENTS — the output format is described so the result is verifiable

Additionally, by skill type:

| Type | Priority | Special attention |
|-----|-----------|-----------------|
| Analytical | Conflict detection | PROCESS covers edge cases (no data, broken data) |
| Generative | Output format | OUTPUT REQUIREMENTS contains the exact result structure |
| Diagnostic | Failure handling | FAILURE HANDLING covers 3+ scenarios |
| Review | Severity classification | DECISION RULES clearly divides findings by severity |
| Utility | Tool usage | TOOL USAGE contains exact commands, flags, examples |

### Step 7. Three test scenarios

If even one fails — go back to Step 3.

| Case | Check |
|------|----------|
| Simple | Ideal data → correct result with no extra actions |
| Incomplete | Not enough data → FAILURE HANDLING: request specific data, no guessing |
| Out of scope | Request outside DO NOT USE → refusal or delegation, no execution |

### Step 8. Activation Coverage Test

Check that the skill fires when it should and stays silent when it shouldn't.

Generate:

- **3 should-trigger requests** — the user is clearly within the skill's scope
- **3 should-not-trigger requests** — the user is outside the skill's scope, or near it but not in it

Check:

- all should-trigger requests activate the skill (per ACTIVATION and description)
- all should-not-trigger requests do **not** activate the skill (they fall under DO NOT USE or don't match the triggers)

If the test fails — rewrite the description and the ACTIVATION triggers.

### Step 9. Meta decision rules

- **Brief not closed** → don't move on to rendering.
- **Conflict between brevity and completeness** → in the core sections (PROCESS, FAILURE HANDLING) prefer completeness; in examples — brevity.
- **Template vs request** → the request takes priority. If the request violates the quality principles — explain why and propose a compromise.
- **Unsure whether a section is needed** → add it with a "(not applicable)" note rather than deleting it.

---

## IMPROVE: improve an existing skill

### Step 0. Find repetitions and conflicts BEFORE rewriting

Before changing anything, record the problem spots. Don't start fixing until a full defect list is compiled.

### Step 1. Full audit

- **Section coverage**: which exist, which are missing.
- **Substance**: is the section useful or a stub ("...", "TBD").
- **Conflicts**: instructions contradict each other.
- **Uncertainty**: are there phrases like "where possible", "if needed".
- **Length**: doesn't exceed reasonable limits. If there are large examples or references — their place is in a reference file.
- **YAML**: name in kebab-case, description per the formula.
- **Examples**: reflect real use cases or are abstract.
- **FAILURE HANDLING**: covers typical failures or is empty.
- **Output contract**: is the result format described.
- **Platform**: does it match the target.

### Step 2. Classify the defects

| Category | Defects | Action |
|-----------|---------|----------|
| Missing | No section, examples, YAML | Add |
| Shallow | A stub section, abstract examples | Fill in from context |
| Conflict | PROCESS duplicates DECISION RULES | Rewrite, set PRIORITY |
| Vague | "where possible", "be careful" | Replace with concrete conditions |
| Bloated | Lots of repetition, large examples, references inside | Trim: examples → process → explanations; extract into a reference
| Wrong platform | YAML/format not for the platform | Rewrite for the target platform |

### Step 3. Apply the fixes

For each defect — one action from the table. Don't change what works.

### Step 4. Re-validate

Run it through steps 4–9 of the CREATE mode.

---

## Output contract (mandatory)

For any response, output exactly this set of blocks:

```
## Assumptions
<assumptions made in the process>

## Brief / Analysis
<for CREATE: the brief from Step 1>
<for IMPROVE: the audit from Step 1>

## Final SKILL.md
<the finished file in full, including the YAML>

## Validation
<check results: conflicts, modularity, quality (Step 6), the 3 tests (Step 7) — passed/failed>

## Open Questions
<questions, if any remain>
```

If the user asks for only a fragment — provide it with a note "this is an incomplete output".

---

# Platform compatibility

This meta-skill designs skills. By default, the skill body is considered **platform-independent**.

Universal rules (always):
- YAML frontmatter with name + description
- All sections PURPOSE–EXAMPLES are present
- PROCESS is numbered, FAILURE HANDLING is concrete
- OUTPUT REQUIREMENTS contains a verifiable format

Platform → format:

| Platform | File format | Specifics |
|-----------|-------------|-------------|
| **universal** (default) | `SKILL.md` | name + description, minimal fields |
| **opencode-only** | `SKILL.md` | description as concise as possible |
| **claude-code-only** | `CLAUDE.md` / `AGENTS.md` | Additional platform fields possible |

Rules:
- If the platform isn't specified → universal
- If the platform requires special sections, resources, or structure — adapt the skill to it, but in its base form the body is considered platform-independent
- If the skill is strictly single-platform, note it in the description: "OpenCode-only: ..." or "Claude Code: ..."

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

## Pareto Optimization (final step)

Before delivering the result:

1. Find the 20% of instructions that produce 80% of the quality (usually this is PROCESS, FAILURE HANDLING, DECISION RULES).
2. Check: can the other 80% of instructions be removed without a significant loss of quality?
3. If so — trim the skill.

Most bad skills get better after removing 20–40% of the text.
Don't add anything the skill would work just as well without.
```
