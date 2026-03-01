# AGENTS.md

This file provides guidelines and instructions for AI assistants working on the Kodu project.

## 1. Project Overview

**Kodu** is a high-performance CLI utility that bridges local development environments with LLMs. It automates context preparation, code cleaning, reviews, and commit drafting.

- **Key Goals:** Speed (<0.5s startup), Determinism (no AI for critical file ops), DX (Developer Experience)
- **Current Phase:** Phase 4 - AI Integration with Mastra/Git
- **Available Commands:** `init`, `pack`, `clean`, `review`, `commit`

## 2. Technology Stack (Enforced)

| Category | USE | DO NOT USE |
|----------|-----|------------|
| Framework | NestJS + nest-commander | Pure Node.js, Oclif |
| File System | node:fs/promises + tinyglobby | fs-extra, glob, rimraf |
| Config | lilconfig | cosmiconfig, rc |
| Validation | zod | class-validator, joi |
| Process/Git | execa | child_process, shelljs |
| CLI UI | @inquirer/prompts + picocolors | inquirer (legacy), chalk |
| Spinners | yocto-spinner | ora, cli-spinners |
| AI Agent | mastra | Direct openai SDK |
| AST/Parsing | ts-morph | Regex, babel |
| Tokens | js-tiktoken | gpt-3-encoder |
| Clipboard | clipboardy | Native APIs |

## 3. Architecture

```
src/
├── app.module.ts           # Root Orchestrator
├── main.ts                 # Entry Point
├── core/                   # Global Infrastructure
│   ├── config/             # ConfigModule (Zod + lilconfig)
│   ├── file-system/        # FsModule (tinyglobby)
│   └── ui/                 # UiModule (spinners, loggers)
├── shared/                 # Shared Business Logic
│   ├── tokenizer/          # TokenizerModule
│   ├── git/                # GitModule
│   ├── ai/                 # AiModule (Mastra)
│   └── cleaner/            # CleanerService (AST)
└── commands/               # Feature Commands
    ├── init/               # kodu init
    ├── pack/               # kodu pack
    ├── clean/              # kodu clean
    ├── review/             # kodu review
    └── commit/             # kodu commit
```

## 4. Build, Lint & Test Commands

### Essential Commands
```bash
# Build the project
npm run build              # Full build (Nest build) + make executable

# Run the built artifact
npm run start:prod         # Run from dist/

# Type check
npm run ts:check           # TypeScript compilation check

# Lint and format
npm run lint               # Run Biome linter
npm run lint:fix           # Biome with auto-fix

# Full check (required before commit)
npm run check              # TypeCheck + Biome + Knip

# Run tests (if configured)
npm test                   # Run all tests
npm test -- <pattern>      # Run tests matching pattern
npm test -- --testPathPattern=<pattern>  # Alternative
```

### Running a Single Test
```bash
# By file name
npm test -- filename

# By pattern
npm test -- --testNamePattern="test name"

# Watch mode
npm test -- --watch

# With coverage
npm test -- --coverage
```

## 5. Code Style Guidelines

### General
- **ESM Only:** Use `import` statements (nodenext mode)
- **Strict Mode:** `strictNullChecks` is ON. Avoid `any`; use `unknown` with narrowing
- **Quotes:** Single quotes preferred
- **Indentation:** 2 spaces
- **No Comments:** Unless explicitly requested by user

### Imports
- Use explicit relative imports: `import { Foo } from './foo'`
- Avoid barrel exports (`index.ts`) unless necessary
- No circular dependencies (NestJS module structure enforces this)

### Types
- Prefer explicit types over `any`
- Use `unknown` and narrow with type guards or Zod validation
- Interface over type for object shapes
- Use readonly for immutable data

### Naming Conventions
- **Files:** kebab-case (`my-file.ts`)
- **Classes:** PascalCase (`MyClass`)
- **Functions:** camelCase (`myFunction`)
- **Constants:** UPPER_SNAKE_CASE for compile-time constants
- **Interfaces:** PascalCase, no `I` prefix (`User` not `IUser`)

### Error Handling
- Use custom error classes extending `Error`
- Never swallow errors silently
- Provide meaningful error messages
- Use try/catch with specific error types
- Validate inputs with Zod schemas

### NestJS Specifics
- All commands extend `CommandRunner` from `nest-commander`
- Use Dependency Injection - never import services directly
- Register modules in `app.module.ts`
- Use `@Injectable()` decorator for services

## 6. Configuration (kodu.json)

```json
{
  "llm": {
    "model": "openai/gpt-4o",
    "apiKeyEnv": "OPENAI_API_KEY"
  },
  "cleaner": {
    "whitelist": ["//!"],
    "keepJSDoc": true,
    "useGitignore": true
  },
  "packer": {
    "ignore": ["*.lock", "node_modules", "dist"],
    "useGitignore": true
  }
}
```

- Model format: `provider/model-name` (e.g., `openai/gpt-4o`, `anthropic/claude-4-5-sonnet`)
- API keys from env vars only (never store in config)
- Config validated via Zod on startup

## 7. Commands Reference

| Command | Description | Key Options |
|---------|-------------|-------------|
| `kodu init` | Interactive setup wizard | - |
| `kodu pack` | Bundle files for LLM | `--copy`, `--template`, `--out` |
| `kodu clean` | Remove comments (AST-based) | `--dry-run` |
| `kodu review` | AI code review | `--mode`, `--ci`, `--output` |
| `kodu commit` | Generate commit message | `--ci`, `--output` |

## 8. Critical Constraints

1. **No AI in Cleaner:** `clean` command is deterministic (AST-based), never AI-generated
2. **Validation First:** Invalid `kodu.json` causes graceful crash with Zod error
3. **Secrets:** Never commit secrets. API keys from env vars only
4. **Performance:** Mindful of import costs. Use lightweight libraries
5. **Git Preconditions:** AI commands require git repo with staged changes
6. **Config Location:** `kodu.json` must be in current working directory

## 9. Development Workflow

### Adding a New Command
1. Create `src/commands/<name>/`
2. Create `<name>.command.ts` and `<name>.module.ts`
3. Implement `run()` extending `CommandRunner`
4. Decorate with `@Command()` from `nest-commander`
5. Register module in `app.module.ts`
6. Test: `npm run build && node dist/main.js <name>`

### Before Commit
Always run:
```bash
npm run check
```

This executes: TypeScript check + Biome lint + Knip dead code detection.

## 10. Testing Strategy

- **Primary Gate:** Static analysis (TypeScript + Biome + Knip)
- **No Legacy Tests:** Project relies on strict static typing
- If tests exist: place in `__tests__/` or `*.test.ts` files
- Use Vitest or Jest (check package.json for actual test runner)

## 11. Handling Uncertainties

- Unclear requirements? Ask the user first
- Library not in Tech Stack section? Prefer native Node.js APIs
- New dependency? Ensure it follows "Fresh & Modern" strategy
- Breaking changes? Create an OpenSpec proposal
