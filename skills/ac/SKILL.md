---
name: ac
description: "Updates status.md, records decisions, and makes a clean git commit."
compatibility: opencode
metadata:
  level: multi
  output: .agent-log/ + git commit
---

# What this skill does

Before committing, it does three things:
1. **Checks** — that there are no secrets in the code
2. **Updates** — `status.md` (marks what was done) and writes a decision if needed
3. **Commits** — following Conventional Commits, with the WHY in the body

Runs **automatically** on phrases like "let's commit / done / push / finished / save".
The user can invoke it manually: `/ac`.

---

# Steps

## Step 1 — Find out what changed

```bash
git diff --cached --name-only
```

If the list is empty:
```bash
git diff HEAD --name-only
```

If that's empty too → print "Nothing to commit" and **stop**.

---

## Step 2 — Check for secrets

```bash
git diff --cached
```

Scan the entire diff against these patterns:

- AWS access key
- GitHub PAT
- OpenAI / Anthropic key / ...
- Private key
- Password in code (not a placeholder)
- Secret in code (not a placeholder)

**Found one → STOP.** Show the offending line. Do not commit.

---

## Step 3 — Record an architectural decision (if there was one)

**First answer this:** was there an architectural decision in this session? (options were discussed, an approach was chosen)

**No → skip this step.**

**Yes:**

1. Find the next ID: `glob: .agent-log/decisions/**/*.md` → max number `DEC-NNNN` → +1. If there are no files → `DEC-0001`.

2. Slug from the title:
   ```bash
   python3 -c "import unicodedata,re,sys; s=unicodedata.normalize('NFKD',sys.argv[1]).encode('ascii','ignore').decode(); print(re.sub(r'[^a-z0-9]+','-',s.lower()).strip('-'))" "Decision title"
   ```

3. Create `.agent-log/decisions/YYYY/MM/DD/DEC-NNNN-slug.md`:

   ```
   ---
   id: DEC-NNNN
   type: decision
   title: Decision title
   summary: One line — the essence of the choice
   status: active
   created_at: YYYY-MM-DD
   updated_at: YYYY-MM-DD
   tags: [tag1, tag2]
   confidence: medium
   schema_version: 1
   related_files: [source code only, no .agent-log/]
   origin: generated
   verification_state: unverified
   ---

   ## Context

   Why this question came up. What the problem is. Clear to a new developer.

   ## Decision

   What was chosen and why exactly this.

   ## Alternatives

   What was considered and why it was not chosen. Even the bad options — so they don't get suggested again.
   ```

   Remember the ID (`DEC-NNNN`) — it's needed in step 6.

---

## Step 4 — Update status.md

Read `.agent-log/status.md`.

Determine: which task from "Current focus" did this commit resolve?

- `- [ ] task` → `- [x] task` for the completed one
- If "Upcoming backlog" has tasks — move the first one into "Current focus"
  (remove it from the backlog, add it as `- [ ]` in the focus)
- Update `updated_at` in the frontmatter: current date and time (`YYYY-MM-DD HH:MM`)

If `status.md` does not exist → create an empty template:

```
---
type: status
updated_at: YYYY-MM-DD HH:MM
current_goal: ""
---

# Current focus
- [ ] Task not defined

# Upcoming backlog (Next Steps)

# Known issues / Technical debt
```

---

## Step 5 — Stage the files

```bash
git add .agent-log/status.md
```

If a decision was created in step 3:
```bash
git add .agent-log/decisions/
```

---

## Step 6 — Build the commit message

### Format (Conventional Commits)

```
type(scope): short description

Why this change was made.
What problem was being solved and why this way.

Related: DEC-NNNN
```

### How to choose the type

| Type | When |
|---|---|
| `feat` | New functionality |
| `fix` | Bug fix |
| `refactor` | Rework without changing behavior |
| `docs` | Documentation only |
| `chore` | Configs, dependencies |
| `test` | Tests |
| `perf` | Performance |
| `style` | Formatting |
| `build` | Build, CI/CD |

### Rules for the first line

- Imperative: `add`, `fix`, `remove` (not `added`, not `adding`)
- Lowercase after the colon, no period, ≤72 characters
- **Always in English**

### Body — required

WHY, not WHAT. Why this way exactly. Separate with a blank line, wrap at 72 characters.

### Footer

- `Related: DEC-NNNN` — if a decision was created in step 3
- `BREAKING CHANGE: description` — if it's a breaking change

### Example

```
feat(auth): add session invalidation via Redis

JWT tokens could not be revoked server-side without maintaining
a blacklist. Redis-based sessions solve this with O(1) lookup
and TTL-based cleanup.

Related: DEC-0001
```

---

## Step 7 — Make the commit

```bash
git commit -m "type(scope): description

WHY in the body.

Related: DEC-NNNN"
```

---

## Step 8 — Report to the user

```
✅ Changes committed
Commit: <sha7> — <description>
status.md updated.
Decisions recorded: [N / none]
```

---

# Rules

- **Secrets** → STOP, show the line, do not commit
- **Decision** — only for a real architectural choice, not every commit
- **Commit message** — English, Conventional Commits
- **WHY is unclear** — infer it from the session context, don't ask again
- **status.md** — always update, at minimum `updated_at`
- **Language in markdown files** — simple and clear, short sentences, no unexplained jargon
- **Order**: check → write files → git add → git commit
