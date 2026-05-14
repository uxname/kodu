# AGENTS.md

This file provides guidelines and instructions for AI assistants working on the Kodu project.

## 1. Project Overview

**Kodu** is a high-performance CLI utility that bridges local development environments with LLMs. It automates context preparation and code cleaning.

- **Key Goals:** Speed (<0.5s startup), Determinism (no AI for critical file ops), DX (Developer Experience)
- **Available Commands:** `pack`, `clean`

## 2. Technology Stack (Enforced)

| Category | USE | DO NOT USE |
|----------|-----|------------|
| Framework | NestJS + nest-commander | Pure Node.js, Oclif |
| File System | node:fs/promises + tinyglobby | fs-extra, glob, rimraf |
| Config | lilconfig | cosmiconfig, rc |
| Validation | zod | class-validator, joi |
| CLI UI | @inquirer/prompts + picocolors | inquirer (legacy), chalk |
| Spinners | yocto-spinner | ora, cli-spinners |
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
│   └── cleaner/            # CleanerService (AST)
└── commands/               # Feature Commands
    ├── pack/               # kodu pack
    └── clean/              # kodu clean
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

- Config validated via Zod on startup
- `kodu.json` must be in current working directory

## 7. Commands Reference

### `kodu init`

Add `.kodu/context.txt` to `.gitignore` (if `.gitignore` exists). Run once after cloning or setting up a project.

### `kodu pack`

Bundle project files into a single context file for LLMs.

| Option | Description |
|--------|-------------|
| `-c, --copy` | Copy result to clipboard |
| `-o, --out <path>` | Path to save result (default: `.kodu/context.txt`) |
| `-p, --path <path>` | Directory or glob to include (repeatable) |
| `-e, --exclude <pattern>` | Additional exclude pattern (repeatable) |
| `-l, --list` | Print file list only, without content |
| `-f, --format <format>` | Output format: `xml` (default) or `text` |
| `-t, --template <name>` | Template name from `.kodu/prompts` |

Output format `xml` wraps each file in `<file path="...">` tags — recommended for LLM consumption. Format `text` uses `// file: ...` header comments.

### `kodu clean`

Remove comments from source files using AST-based parsing (deterministic, no AI).

| Option | Description |
|--------|-------------|
| `-d, --dry-run` | Show what will be removed without modifying files |
| `-c, --changed` | Clean only git-changed files |

## 8. Critical Constraints

1. **No AI:** Both commands are deterministic — no AI integration
2. **Validation First:** Invalid `kodu.json` causes graceful crash with Zod error
3. **Performance:** Mindful of import costs. Use lightweight libraries
4. **Config Location:** `kodu.json` must be in current working directory

## 9. Development Workflow

### Adding a New Command
1. Create `src/commands/<name>/`
2. Create `<name>.command.ts` and `<name>.module.ts`
3. Implement `run()` extending `CommandRunner`
4. Decorate with `@Command()` from `nest-commander`
5. Register module in `app.module.ts`
6. Test: `npm run build && node dist/src/main.js <name>`

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

## 11. Release Process

### Prerequisites
- Working directory must be clean (`git status` — no dirty files)
- All changes committed and pushed
- You have npm publish access

### Steps
```bash
# 1. Ensure clean working directory
git status

# 2. Bump version, build, publish
npm version patch && npm run build && npm publish --access public

# 3. Push the version bump commit and tag
git push && git push --tags
```

- Use `npm version minor` for new features, `npm version major` for breaking changes
- The `npm version` command creates a git tag automatically (e.g. `v2.1.3`)

## 12. Handling Uncertainties

- Unclear requirements? Ask the user first
- Library not in Tech Stack section? Prefer native Node.js APIs
- New dependency? Ensure it follows "Fresh & Modern" strategy
- Breaking changes? Create an OpenSpec proposal
