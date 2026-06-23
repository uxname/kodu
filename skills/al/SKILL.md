---
name: al
description: "Loads working context from .agent-log/ and the git log at the start of a session."
compatibility: opencode
metadata:
  level: single
  output: context summary
---

# What this skill does

Gives the agent up-to-date working context before starting work:
- what we're doing right now (from `status.md`)
- what has already been done (from `git log`)
- which architectural decisions have been made (from `decisions/`)

Runs **automatically** at the start of every session. The user can invoke it manually: `/al`.

---

# Steps

**Step 1 — Check the knowledge base**

```bash
ls .agent-log/
```

> ⚠️ Never use the `read` tool on a directory — it will return "File not found". Always go through `bash`.

- Directory exists → go to step 2
- Directory missing → report: _"AKMS is not initialized. Run `akms init`?"_ — stop

---

**Step 2 — Read the current focus**

Read `.agent-log/status.md` in full — the file is always short.

Remember:
- `current_goal` from the frontmatter — the project's overall goal
- the first open `- [ ]` item from the "Current focus" section

If `status.md` does not exist → skip, continue without a focus.

---

**Step 3 — Load active decisions**

```bash
glob: .agent-log/decisions/**/*.md
```

For each, read **only the first 15 lines** — the frontmatter is enough.

Keep only: `status: active` or `status: needs_review`. Maximum **7 entries**.

If there are no decisions → skip.

---

**Step 4 — Read history from Git**

```bash
git log -n 5 --pretty=format:"* %h — %s (%ar)"
```

The last 5 commits, one line each. Fast, doesn't touch the filesystem.

If git is not initialized or the repository is empty → skip.

---

**Step 5 — Print the context to the user**

```
📚 AKMS context loaded
Current goal: [current_goal from status.md / "not set"]
Active task: [first - [ ] item / "no open tasks"]
Recent commits:
  * a1b2c3d — feat(auth): add redis support (2 hours ago)
  * ...
Active decisions: [DEC-0001: title, ...] / none

Let's get to work.
```

After printing — go straight to the user's task.

---

# Rules

- **Read status.md in full** — it's deliberately short, this is not wasteful
- **decisions** — only the first 15 lines (frontmatter), the body isn't needed
- **git log instead of commit files** — faster and always current
- **Check the directory via `bash ls`** — not via `read`
- **Don't ask questions** — only read, change nothing
