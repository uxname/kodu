# AGENTS

## 1. Project Overview & Philosophy
- **Kodu** is a CLI assistant for developers (JS/TS focus) to streamline interactions with LLMs.
- **Key Goals:** Speed (0.5s startup), Determinism (no AI for critical file ops), and DX (Developer Experience).
- **Source of Truth:** 
  - Functional scope: `docs/project_charter.md`.
  - Roadmap & Tech Stack: `docs/plan.md`.
- **Current Phase:** **Phase 1: The Foundation** (Core Setup, ConfigModule, Init Command).

## 2. Technology Stack (Enforced)
We strictly follow the "Fresh & Modern" stack strategy. Do not install legacy libraries.

| Category | **USE THIS** ✅ | **DO NOT USE** ❌ |
| :--- | :--- | :--- |
| **Framework** | `NestJS` + `nest-commander` | Pure Node.js scripts, Oclif |
| **File System** | `node:fs/promises` + `tinyglobby` | `fs-extra`, `glob`, `fast-glob`, `rimraf` |
| **Validation** | `zod` | `class-validator`, `joi` |
| **Process/Git** | `execa` | `child_process`, `shelljs` |
| **CLI UI** | `@inquirer/prompts` + `picocolors` | `inquirer` (legacy), `chalk`, `colors` |
| **Spinners** | `yocto-spinner` | `ora`, `cli-spinners` |
| **AI Agent** | `mastra` | Direct `openai` SDK calls (unless inside Mastra) |
| **AST/Parsing** | `ts-morph` | Regex for code parsing, `babel` |
| **Tokens** | `js-tiktoken` | `gpt-3-encoder` |

## 3. Architecture & Module Map
The project is NOT a flat structure. Use the following Module Map as a guide for where to place files.

```text
src/
├── app.module.ts            # Root Orchestrator
├── main.ts                  # Entry Point
│
├── core/                    # GLOBAL Infrastructure (Global Modules)
│   ├── config/              # ConfigModule (Zod schemas, loading kodu.json)
│   ├── file-system/         # FsModule (tinyglobby wrappers)
│   └── ui/                  # UiModule (Spinners, colored loggers)
│
├── shared/                  # Shared Business Logic
│   ├── tokenizer/           # TokenizerModule
│   ├── git/                 # GitModule
│   └── ai/                  # AiModule (Mastra setup)
│
└── commands/                # Feature Modules (The actual commands)
    ├── init/                # InitModule
    ├── pack/                # PackModule
    ├── clean/               # CleanModule
    ├── review/              # ReviewModule
    └── commit/              # CommitModule
```

## 4. Coding Standards & Conventions

### 4.1. General
- **ESM Only:** The project runs in `nodenext` mode. Use `import` statements.
- **Strictness:** `strictNullChecks` is ON. No `any` allowed unless absolutely necessary (use `unknown` and refine).
- **Async/Await:** Prefer `node:fs/promises` over sync methods where possible, but for CLI startup (Config loading), sync operations are acceptable if they improve perceived performance.

### 4.2. NestJS Specifics
- **Dependency Injection:** Always use DI. Do not import services directly into other services without providing them in the Module.
- **CommandRunner:** All commands extend `CommandRunner` from `nest-commander`.
- **Zod Config:** Configuration is loaded ONCE in `ConfigModule` and validated. Other modules inject `ConfigService` to access typed settings.

### 4.3. Code Style (Biome)
- We use **Biome** for linting and formatting.
- **Quotes:** Single quotes.
- **Indent:** 2 spaces.
- **Run Check:** Always run `npm run lint` before finishing a task.

## 5. Development Workflow

### 5.1. Adding a New Command
1.  Create a folder in `src/commands/<name>`.
2.  Create `<name>.command.ts` and `<name>.module.ts`.
3.  Implement `run()` method.
4.  Register the module in `src/app.module.ts`.
5.  **Test:** Run `npm run build && node dist/main.js <name>` to verify.

### 5.2. Scripts
- `npm run build`: Full build (Nest build).
- `npm run start:prod`: Run the built artifact.
- `npm run check`: Run TypeCheck + Biome + Knip (Dead code).

## 6. Critical Constraints
1.  **No AI in Cleaner:** The `clean` command must be purely deterministic (AST-based).
2.  **Validation First:** The app must crash gracefully with a helpful message if `kodu.json` is invalid (handled by Zod).
3.  **Secrets:** Never commit secrets. Assume `.env` usage for API keys.
4.  **Performance:** Be mindful of import costs. We use `tinyglobby` and `picocolors` to keep startup time low.

## 7. Handling Uncertainties
- If a task involves logic not defined in `docs/project_charter.md`, ask the user.
- If unsure about a library, check Section 2 of this file. If not listed, prefer native Node.js APIs or check `docs/plan.md`.
