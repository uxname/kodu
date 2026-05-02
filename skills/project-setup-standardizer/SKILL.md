---
name: project-setup-standardizer
description: Enforces a unified, production-ready project setup (scripts, linting, testing, biome, lefthook) across any JS/TS project (backend, frontend, CLI, etc.). This skill MUST be invoked when initializing a new project.
---

**This skill is MANDATORY and must be invoked when creating a new project.**

You are responsible for transforming an existing JavaScript/TypeScript project into a **production-ready, standardized
structure**.

Your goal is NOT to suggest — you MUST **apply and enforce** the conventions below.

---

## 1. PROJECT TYPE DETECTION

Before making any changes, detect the project type using these signals:

| Signal                                                           | Type     |
|------------------------------------------------------------------|----------|
| `dependencies` contains `react` or `vue`                         | frontend |
| `devDependencies` contains `vite` (without react/vue)            | frontend |
| `dependencies` contains `@nestjs/core` or `express` or `fastify` | backend  |
| `bin` field exists in package.json                               | CLI      |
| None of the above                                                | library  |

Frontend projects get additional tools (stylelint, steiger). All other types share the base setup.

---

## 2. MIGRATION: REMOVE CONFLICTING TOOLS FIRST

Before installing Biome, you MUST remove ESLint and Prettier if present:

1. Delete config files: `.eslintrc*`, `.eslintignore`, `.prettierrc*`, `.prettierignore`
2. Remove from `package.json` dependencies/devDependencies: `eslint`, `prettier`, and all `eslint-*` / `prettier-*`
   packages
3. Only after removal, proceed with Biome setup

---

## 3. PACKAGE.JSON SCRIPTS (MANDATORY STRUCTURE)

You MUST normalize `package.json/scripts` into clearly separated sections using visual separators.

Use this exact pattern:

```json
"scripts": {
"________________ BUILD AND RUN ________________": "",
"build": "...",
"start:dev": "...",
"start:prod": "...",
"________________ FORMAT AND LINT ________________": "",
"lint": "biome check",
"lint:fix": "biome check --write",
"lint:fix:unsafe": "biome check --write --unsafe",
"ts:check": "tsc --noEmit",
"knip": "knip --production",
"check": "run-p ts:check lint:fix knip",
"________________ TEST ________________": "",
"test": "vitest run",
"test:watch": "vitest",
"test:cov": "vitest run --coverage",

"________________ OTHER ________________": "",
"prepare": "is-ci || lefthook install",
"update": "npx npm-check-updates -u && rimraf node_modules package-lock.json && npm i",
"postupdate": "npm run lint:fix && npm run check"
}
```

If the project has distinct unit and e2e test layers, add:

```json
"test:unit": "...",
"test:e2e": "...",
"test:all": "run-s test:unit test:e2e"
```

Otherwise omit `test:all` — do not duplicate `test` under a different name.

### Rules:

- ALWAYS include `check` script → central quality gate
- ALWAYS include `ts:check`; if no `tsconfig.json` exists, create a minimal one (see section 7)
- ALWAYS include `knip` for unused exports/deps detection
- ALWAYS include `lint:fix` in `check`
- Use `run-p` (from `npm-run-all`) for parallel execution — never `npx run-p`
- `prepare` MUST use `is-ci || lefthook install` to avoid CI failures when lefthook is not yet installed

---

## 4. FRONTEND-SPECIFIC REQUIREMENTS

If project type is **frontend**, additionally include:

```json
"lint:fsd": "steiger ./src",
"lint:style": "stylelint '**/*.{css,scss}'",
"lint:style:fix": "npm run lint:style -- --fix",
"check": "run-p lint:style ts:check lint:fix knip lint:fsd"
```

### Mandatory:

- FSD validation via `steiger`
- stylelint integration
- include ALL checks in `check`

---

## 5. TESTING STANDARD (MANDATORY)

Use **vitest only**.

Rules:

- NO jest
- NO mixed frameworks
- coverage MUST be supported via `test:cov`

---

## 6. BIOME CONFIG (MANDATORY)

You MUST ensure `biome.json` exists with this content:

```json
{
  "$schema": "./node_modules/@biomejs/biome/configuration_schema.json",
  "formatter": {
    "enabled": true,
    "indentStyle": "space",
    "lineEnding": "lf"
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "single"
    }
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "correctness": {
        "noUnusedImports": "error",
        "noUnusedVariables": "error",
        "noUnusedFunctionParameters": "error"
      }
    }
  },
  "files": {
    "includes": [
      "**",
      "!node_modules",
      "!dist"
    ]
  },
  "vcs": {
    "enabled": true,
    "clientKind": "git",
    "useIgnoreFile": true
  }
}
```

---

## 7. LEFTHOOK (MANDATORY)

You MUST write `lefthook.yml` with this exact content:

```yaml
pre-commit:
  parallel: false
  commands:
    check:
      run: npm run check

pre-push:
  parallel: false
  commands:
    test-all:
      run: npm run test:all
```

Rules:

- `check` (lint + types + knip) gates every commit
- tests gate every push — `pre-push` does NOT re-run `check` (already gated at commit)
- no bypassing allowed

---

## 8. MINIMAL TSCONFIG (if missing)

If `tsconfig.json` does not exist, create it:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "noEmit": true,
    "skipLibCheck": true
  },
  "include": [
    "src",
    "*.ts"
  ]
}
```

---

## 9. REQUIRED DEV DEPENDENCIES

Ensure these are installed (add any that are missing):

Core (all project types):

- `@biomejs/biome`
- `typescript`
- `vitest`
- `lefthook`
- `is-ci`
- `npm-run-all`
- `knip`
- `rimraf`
- `npm-check-updates`

Frontend only:

- `stylelint`
- `steiger`

---

## 10. ENFORCEMENT LOGIC

You MUST follow this order:

1. Detect project type (section 1)
2. Remove conflicting tools (section 2)
3. Rewrite `package.json` scripts (section 3, or 4 for frontend)
4. Write/update `biome.json` (section 6)
5. Write/update `lefthook.yml` (section 7)
6. Create `tsconfig.json` if missing (section 8)
7. Install missing dev dependencies (section 9)

NEVER:

- keep inconsistent scripts
- mix tools (eslint + biome)
- leave partial setup

---

## 11. OUTPUT FORMAT

You must output:

1. Updated `package.json` (only the `scripts` and `devDependencies` sections)
2. `biome.json`
3. `lefthook.yml`
4. `tsconfig.json` (if created)
5. List of removed files and packages

All configs must be **complete and copy-paste ready**.

---

## 12. PRINCIPLE

Your job is to enforce:

- determinism
- reproducibility
- zero-config onboarding
- strict quality gates

If something is ambiguous — choose the stricter option.
