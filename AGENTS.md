<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# AGENTS

## 1. Project Overview & Philosophy
- **Kodu** is a CLI assistant for developers (JS/TS focus) to streamline interactions with LLMs.
- **Key Goals:** Speed (0.5s startup), Determinism (no AI for critical file ops), and DX (Developer Experience).
- **Source of Truth:** 
  - Functional scope: `docs/project_charter.md`.
  - Roadmap & Tech Stack: `docs/plan.md`.
- **Current Phase:** **Phase 4: AI Integration** — Mastra/Git integration, `review` and `commit` commands, diff filtering via `packer.ignore`.
- **Available Commands:** `init`, `pack`, `clean`, `review`, `commit` (all registered in `app.module.ts`).

## 2. Technology Stack (Enforced)
We strictly follow the "Fresh & Modern" stack strategy. Do not install legacy libraries.

| Category | **USE THIS** ✅ | **DO NOT USE** ❌ |
| :--- | :--- | :--- |
| **Framework** | `NestJS` + `nest-commander` | Pure Node.js scripts, Oclif |
| **File System** | `node:fs/promises` + `tinyglobby` | `fs-extra`, `glob`, `fast-glob`, `rimraf` |
| **Config Loading** | `lilconfig` | `cosmiconfig`, `rc` |
| **Validation** | `zod` | `class-validator`, `joi` |
| **Process/Git** | `execa` | `child_process`, `shelljs` |
| **CLI UI** | `@inquirer/prompts` + `picocolors` | `inquirer` (legacy), `chalk`, `colors` |
| **Spinners** | `yocto-spinner` | `ora`, `cli-spinners` |
| **AI Agent** | `mastra` | Direct `openai` SDK calls (unless inside Mastra) |
| **AST/Parsing** | `ts-morph` | Regex for code parsing, `babel` |
| **Tokens** | `js-tiktoken` | `gpt-3-encoder` |
| **Git** | `git` via `execa` | `child_process`, `shelljs` |
| **Clipboard** | `clipboardy` | Native clipboard APIs |

## 3. Architecture & Module Map
The project is NOT a flat structure. Use the following Module Map as a guide for where to place files.

```text
src/
├── app.module.ts            # Root Orchestrator
├── main.ts                  # Entry Point
│
├── core/                    # GLOBAL Infrastructure (Global Modules)
│   ├── config/              # ConfigModule (Zod schemas, lilconfig for kodu.json)
│   ├── file-system/         # FsModule (tinyglobby wrappers)
│   └── ui/                  # UiModule (Spinners, colored loggers)
│
├── shared/                  # Shared Business Logic
│   ├── tokenizer/           # TokenizerModule (js-tiktoken)
│   ├── git/                 # GitModule (execa git helpers, diff filters)
│   ├── ai/                  # AiModule (Mastra agents, API key from llm.apiKeyEnv)
│   └── cleaner/             # CleanerService (AST-based comment removal, used by CleanModule)
│
└── commands/                # Feature Modules (The actual commands)
    ├── init/                # InitModule (kodu init)
    ├── pack/                # PackModule (kodu pack)
    ├── clean/               # CleanModule (kodu clean)
    ├── review/              # ReviewModule (kodu review)
    └── commit/              # CommitModule (kodu commit)
```

**Note:** `CleanerService` is a shared service (not a module) used by `CleanModule`. It's located in `shared/cleaner/` for code reuse.

## 4. Coding Standards & Conventions

### 4.1. General
- **ESM Only:** The project runs in `nodenext` mode. Use `import` statements.
- **Strictness:** `strictNullChecks` is ON. No `any` allowed unless absolutely necessary (use `unknown` and refine).
- **Async/Await:** Prefer `node:fs/promises` over sync methods where possible, but for CLI startup (Config loading), sync operations are acceptable if they improve perceived performance.

### 4.2. NestJS Specifics
- **Dependency Injection:** Always use DI. Do not import services directly into other services without providing them in the Module.
- **CommandRunner:** All commands extend `CommandRunner` from `nest-commander`.
- **Zod Config:** Configuration is loaded ONCE in `ConfigModule` and validated. Other modules inject `ConfigService` to access typed settings.
- **Module Registration:** All command modules must be imported in `src/app.module.ts`.

### 4.3. Code Style (Biome)
- We use **Biome** for linting and formatting.
- **Quotes:** Single quotes.
- **Indent:** 2 spaces.
- **Run Check:** Always run `npm run lint` before finishing a task.

## 5. Development Workflow

### 5.1. Adding a New Command
1. Create a folder in `src/commands/<name>`.
2. Create `<name>.command.ts` and `<name>.module.ts`.
3. Implement `run()` method extending `CommandRunner`.
4. Decorate the class with `@Command()` from `nest-commander`.
5. Register the module in `src/app.module.ts`.
6. **Test:** Run `npm run build && node dist/main.js <name>` to verify.

### 5.2. Scripts
- `npm run build`: Full build (Nest build) + make executable.
- `npm run start:prod`: Run the built artifact.
- `npm run check`: Run TypeCheck + Biome + Knip (Dead code detection).
- `npm run lint`: Run Biome linter.
- `npm run lint:fix`: Run Biome with auto-fix.

## 6. Configuration (`kodu.json`)

Configuration structure:
```json
{
  "llm": {
    "model": "openai/gpt-5-mini",
    "apiKeyEnv": "OPENAI_API_KEY"
  },
  "cleaner": {
    "whitelist": ["//!"],
    "keepJSDoc": true
  },
  "packer": {
    "ignore": [
      "package-lock.json",
      "yarn.lock",
      "pnpm-lock.yaml",
      ".git",
      ".kodu",
      "node_modules",
      "dist",
      "coverage"
    ]
  },
  "prompts": {
    "review": {
      "bug": "You are a strict code reviewer...\n\nDiff:\n{diff}",
      "style": "Check readability...\n\nDiff:\n{diff}",
      "security": "Find vulnerabilities...\n\nDiff:\n{diff}",
      "performance": "Check performance...\n\nDiff:\n{diff}"
    },
    "commit": "You generate Conventional Commit messages...\n\nDiff:\n{diff}"
  }
}
```

**Options:**
- `llm.model`: Model identifier in format `"provider/model-name"` (e.g., `"openai/gpt-4o"`, `"anthropic/claude-4-5-sonnet"`, `"google/gemini-2.5-flash"`).
- `llm.apiKeyEnv`: Environment variable name for API key (default: `"OPENAI_API_KEY"`).

**Supported Providers:**
Kodu supports 72+ providers through Mastra's model router:

- **OpenAI**: `openai/gpt-4o`, `openai/gpt-4o-mini`, `openai/gpt-5`, etc.
- **Anthropic**: `anthropic/claude-4-5-sonnet`, `anthropic/claude-opus-4-1`, etc.
- **Google Gemini**: `google/gemini-2.5-flash`, `google/gemini-2.5-pro`, etc.
- **xAI**: `xai/grok-4`, etc.
- **OpenRouter**: `openrouter/anthropic/claude-haiku-4-5`, etc.
- **And 65+ more providers**

**API Key Environment Variables:**
- OpenAI: `OPENAI_API_KEY`
- Anthropic: `ANTHROPIC_API_KEY`
- Google: `GOOGLE_API_KEY`
- xAI: `XAI_API_KEY`
- OpenRouter: `OPENROUTER_API_KEY`
- Others: See provider documentation

**Cleaner Whitelist Behavior:**
- *System:* Automatically preserved: `@ts-ignore`, `@ts-expect-error`, `eslint-disable`, `prettier-ignore`, `biome-ignore`, `TODO`, `FIXME`.
- *Biome:* Special support for `// biome-ignore ...` comments.
- *User:* Custom list in `cleaner.whitelist` (e.g., `//!`).

**Config Loading:**
- Uses `lilconfig` to search for `kodu.json` in the current working directory.
- Validated via Zod schema on every app startup.
- App exits with error message if config is missing or invalid.

## 7. Commands Reference

### 7.1. `kodu init`
- Interactive setup wizard for creating `kodu.json`.
- Creates `.kodu/` directory for prompt templates.

### 7.2. `kodu pack [options]`
- Bundles project files into a single text file or clipboard.
- Options:
  - `--copy`: Copy to clipboard instead of stdout.
  - `--template <name>`: Use prompt template from `.kodu/prompts/<name>.txt`.
  - `--out <file>`: Write to file instead of stdout.
- Respects `.gitignore` and `packer.ignore` patterns.
- Always displays token count and estimated cost.

### 7.3. `kodu clean [options]`
- Removes comments from JS/TS files using AST parsing.
- Options:
  - `--dry-run`: Show what would be removed without modifying files.
- Supports: `.ts`, `.js`, `.tsx`, `.jsx`.
- Deterministic (no AI) - uses `ts-morph` for safe parsing.
- Preserves system comments and user whitelist.

### 7.4. `kodu review [options]`
- Analyzes staged git diff via AI.
- Options:
  - `--mode <mode>`: Review mode - `bug` (default), `style`, `security`, or custom mode from `prompts.review`. Standard modes (`bug`, `style`, `security`) are always available. Custom modes require configuration in `prompts.review` section of `kodu.json`.
  - `--ci`: CI/CD mode (no spinners, no buffering).
  - `--output <file>`: Save output to file.
- Requires: Git repository, staged changes, API key.

### 7.5. `kodu commit [options]`
- Generates Conventional Commit message from staged diff.
- Options:
  - `--ci`: CI/CD mode (no spinners).
  - `--output <file>`: Save message to file.
- Outputs to stdout (does NOT execute `git commit`).
- Requires: Git repository, staged changes, API key.

## 8. Critical Constraints
1. **No AI in Cleaner:** The `clean` command is purely deterministic (AST-based). Never use AI for code modification.
2. **Validation First:** The app must crash gracefully with a helpful message if `kodu.json` is invalid (handled by Zod).
3. **Secrets:** Never commit secrets. Assume `.env` usage for API keys. API keys are read from environment variables, not config files.
4. **Performance:** Be mindful of import costs. We use `tinyglobby` and `picocolors` to keep startup time low (< 0.5s target).
5. **AI Keys:** `llm.apiKeyEnv` (default `OPENAI_API_KEY`) must be set before running AI commands (`review`, `commit`).
6. **Diff Filtering:** `GitService` excludes paths from `packer.ignore` when building staged diff for AI to filter out lock/dist noise.
7. **Git Preconditions:** AI commands require a git repository and staged changes (otherwise exit with helpful message).
8. **Config Location:** `kodu.json` must be in the current working directory (searched via `lilconfig`).

## 9. Handling Uncertainties
- If a task involves logic not defined in `docs/project_charter.md`, ask the user.
- If unsure about a library, check Section 2 of this file. If not listed, prefer native Node.js APIs or check `docs/plan.md`.
- When adding new dependencies, verify they align with the "Fresh & Modern" stack strategy and are not legacy libraries.
