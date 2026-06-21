---
name: ops
description: >
  Work with project stands (local/dev/stage/prod) from anywhere: a global project
  registry, the active stand, and deploys driven by a runbook. Trigger on /ops or
  requests like "deploy X to dev", "switch the stand", "how do I deploy this project",
  "add a project", "where do I deploy", "show my projects".
metadata:
  level: multi
  output: ~/.config/kodu/registry.json + <project>/.runbook/
---

# ops skill — stands and deploys

This skill helps you work with project **stands**: local, dev, stage, prod.
It is designed for juniors: **you never need to edit any JSON file by hand** — just
tell the agent what you want, and the agent runs the right `kodu ops` command or edits
the file for you.

## Core concepts (in plain words)

- **Stand** — an environment you run things into: `local` (your machine), `dev` (test
  server), `stage` (pre-prod), `prod` (production, be careful!).
- **Registry** — the shared list of all projects on your machine. File:
  `~/.config/kodu/registry.json`. It is created automatically. Thanks to it `kodu`
  can find a project by name **from any directory**.
- **Runbook** — the file `.runbook/runbook.md` inside a project, describing step by
  step how to work with each stand (where to go, which `docker compose` to run, how to
  clone the repo). It lives in `.gitignore` and is **never committed**.
- **Active stand** — the stand you are currently working with. Stored in
  `.runbook/config.json`.

## `kodu` commands this skill relies on

| Command | What it does |
|---|---|
| `kodu ops init` | Set up stands in the current project (creates `.runbook/`, fixes `.gitignore`, adds the project to the registry) |
| `kodu ops list` | Show all projects and their active stands |
| `kodu ops add <name> --path <dir>` | Register a project under a unique name |
| `kodu ops status [name]` | Show the project's active stand |
| `kodu ops use <stand>` | Switch the active stand (in the current project) |
| `kodu ops use <name> <stand>` | Switch the active stand of a named project (from anywhere) |
| `kodu ops path <name>` | Print the repository path (handy: `cd $(kodu ops path my-api)`) |
| `kodu ops runbook <name> [stand]` | Show the instructions for a stand |

## Creating and changing config by request (key section)

**Never make the user edit JSON by hand.** When they ask to change something — run
the command or edit the file yourself, then report the result in plain language.

| What the user says | What the agent does |
|---|---|
| "add project my-api, it's in ~/work/my-api" | `kodu ops add my-api --path ~/work/my-api` |
| "set up stands here" / "init the project" | `kodu ops init` (from the project folder), help fill in the runbook |
| "switch me to dev" | `kodu ops use dev` |
| "switch billing to prod" | `kodu ops use billing prod` (ask for confirmation — it's prod) |
| "add a stage stand to this project" | `kodu ops use stage` (the stand is added automatically), or edit `.runbook/config.json` |
| "change the path of project billing to …" | edit the entry in `~/.config/kodu/registry.json` for the user |
| "what projects / stands do I have?" | `kodu ops list` or `kodu ops status` |
| "where is project X?" | `kodu ops path X` |

After any change — briefly say **what you changed and why**, and show `kodu ops status`
when useful.

## How to deploy a project to a stand

1. Find the project: `kodu ops path <name>` (gives you the path) or work in its folder.
2. Read the stand instructions: `kodu ops runbook <name> <stand>` (or open
   `.runbook/runbook.md`).
3. Run the runbook steps in order (git, `docker compose`, ssh, etc.).
   **Show every dangerous command to the user before running it.**
4. If the runbook still has `<...>` placeholders — ask the user for the details and
   help fill in the file.

## Bootstrapping a new project

If the project does not have `.runbook/` yet:
1. From the project folder run `kodu ops init` (use `--name <name>` if the folder name
   doesn't fit).
2. In plain words explain to the junior what appeared: a `.runbook/` folder with
   `config.json` (active stand) and `runbook.md` (instructions), and that `.runbook/`
   is already added to `.gitignore`.
3. Help fill in `runbook.md` for the real infrastructure.

## Guard-rails (mandatory)

- **prod — only with confirmation.** Before any action on the `prod` stand, explicitly
  ask: "Are we sure we're deploying to prod?".
- **`.gitignore` — automatically.** The `.runbook/` folder must always be in
  `.gitignore`; `kodu ops init` does this itself. Never commit `.runbook/`.
- **Show commands.** Before `docker compose down`, deletions, migrations, and other
  irreversible steps — show the command to the user.
- **Explain changes.** For every config change — one line of "what and why".
