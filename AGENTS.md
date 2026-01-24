# AGENTS

## 1. Repository Context
- The main goal of the project is a CLI utility that facilitates interaction with code and AI: comment cleanup,
  project build into a single `.txt`, generation of helper commands to improve DX.
- Everything we write (code, scripts, docs) must reduce cognitive load and boost DX: clear names, explicit defaults, fewer extra steps.
- Architecture is built on Nest Command (nest-commander) and runs via `CommandFactory`.
- TypeScript targets `ES2023`, modules and plugin resolution work in `nodenext` mode, build goes to `dist/`.
- Command generation scenarios (`new:command`, `new:question`) help extend the CLI parser.
- Automated checks are not yet configured as a separate framework, so temporary verifications are done by running single
  CLI commands after build.

## 2. Build, Run and Check Commands
### 2.1 Building and Running the Production CLI
- `npm run build` — runs `nest build`, then sets the `+x` flag on `dist/main.js`. Always performs a complete TypeScript project rebuild.
- `npm run start:prod` — runs the built CLI (`node dist/main.js`). Used in delivery and smoke validation runs.
- In the project, it's often more convenient to run the CLI directly: `npm run build && node dist/main.js <command>`.

### 2.2 Formatting and Linting
- `npm run lint` — runs `biome check`, includes correctness and style checks according to `biome.json`.
- `npm run lint:fix` — same, but overwrites formatting.
- `npm run lint:fix:unsafe` — uses additional rules that can be applied during major refactors.
- `npm run ts:check` — compilation check `tsc --noEmit`, useful before premium build.
- `npm run knip` — dead code search, especially before releases.
- `npm run check` — combined sequence: `tsc`, `biome`, `knip`.

### 2.3 CLI Artifact Generators
- `npm run new:command` — creates a new command template via schematics for `nest-commander`.
- `npm run new:question` — similar template for question/subcommand entities.

## 3. Code Style and Conventions
### 3.1 Imports and Modules
- We use ES modules (`import ... from ...`), as configured in `tsconfig` (`module: nodenext`).
- Import order: external packages → internal directories (`src/`) → relative modules.
- Use `single` quotes per Biome rules; avoid backticks without templates and double quotes.
- Import `* as something` is not allowed without real necessity; prefer named elements.
- Import organization is proprietary managed by Biome (`organizeImports` enabled in `biome.json`). No manual rearrangements except in special cases.

### 3.2 Formatting
- Indentation: 2 spaces (Biomes in `formatter.indentStyle = space`).
- End of line `LF`. Biome refactors automatically during lint/fixes.
- Don't put extra empty lines between class members, only logical blocks (injections, methods).
- After imports, leave one empty line before class or function declaration.
- `biome check --write` is sufficient, manual `prettier` is not required.

### 3.3 Typing and Strictness
- We use `strictNullChecks`, `esModuleInterop`, `emitDecoratorMetadata`. Types should reflect command semantics.
- Parameter declarations without types are not allowed: each field, option, flag must have exact `string`, `boolean`, `Record<string, unknown>` etc.
- Use `readonly` where the value doesn't change after creation (configurations, `Command` options).
- `any` and `unknown` are allowed only in explicit form: `options: Record<string, unknown>` and only if the type cannot be narrowed.
- Forced assertions (`!`) are prohibited (in `biome` rule `noNonNullAssertion: off`, but we don't use it). Instead, recheck and throw errors.

### 3.4 Naming conventions
- Classes and commands are named `PascalCase` (`DebugCommand`, `AppModule`).
- Methods and variables — `camelCase`, especially `run`, `execute`, `handle`, `options`.
- Constants `const` — `UPPER_SNAKE_CASE` only for flags/constants (move to `constants.ts` module).
- CLI subcommands use `kebab-case` (`debug`, `clean-project`).
- Interfaces are described with `I` only if at API level with external contract; otherwise — `CamelCase` without `I`.

### 3.5 Error handling and logging
- Preference is given to throwing `Error` with a clear message (`throw new Error('...')`).
- In `nest-commander` commands we wrap with `try/catch`, log and exit with non-zero code.
- For user errors we create our own classes (e.g., `class CliProcessingError extends Error`).
- Do not suppress exceptions; if necessary, catch and repackage with additional context.
- `console.log` is allowed only for debug commands (like `DebugCommand`), other scenarios use `this.logger` or `CommandRunner`.

### 3.6 Comments and documentation
- Write comments only for non-trivial business logic.
- Use `/** ... */` for public API, `//` for short explanations.
- Don't leave TODO without a GitHub issue — if we describe `TODO`, always specify `#issue-number`.

## 4. CLI and Architecture Features
- Entry point: `src/main.ts` creates `CommandFactory.run(AppModule)` and waits for all commands to run.
- Each command inherits `CommandRunner` and describes `run(inputs, options)`.
- To add new commands: create a module in `src/*`, export and add to `AppModule`.
- `AppModule` must export an array of commands so Nest can discover them at startup.
- Checks focus on interactive experience: `debug` simply outputs `inputs`, `options`, but useful when developing new commands.
- When extending CLI, ensure support for `--help` (comes from Nest Commander), do not override.

## 5. Quality Maintenance
- `knip --production` searches for unused code — mandatory check before release.
- `biome` manages linting and formatting, so any local IDE settings should not conflict (i.e., disable built-in linter).
- `tsc --noEmit` controls compilation compliance, especially when working with decorators and `experimentalDecorators`.
- `package-lock.json` fixes dependencies for Biome, Nest and `ts-node` — keep lockfile up to date after `npm install`.
- `source-map-support` and `reflect-metadata` are included — import strictly in `main.ts` and entry points.

## 6. Agent Workflow
- Before changing anything always run `npm run lint` + `npm run ts:check` and optionally `npm run knip`.
- For quick debugging run `npm run build && node dist/main.js debug` to inspect the command payload.
- When adding a new command, update `AppModule`, ensure it is exported in `exports`, and register dependencies if needed.
- When introducing new dependencies, clarify whether they are required at runtime (`dependencies`) or only for build tooling (`devDependencies`).
- Add extra scripts to `package.json` with the `npm run` prefix and keep the existing separators (`________________ ...`) consistent.
- Document ongoing plans in `docs/plan.md` (currently focused on the `--version` task) so contributors can quickly learn what to build next.

## 7. Balanced Automation Recommendations
- `biome.json` already allows `unsafeParameterDecorators`, so Nest decorators can rely on parameter metadata without extra annotations.
- Touching `dist/` requires a rebuild; run `rm -rf dist && npm run build`.
- Keep in mind the CLI sometimes runs as the `kodu` binary (`package.json`). During releases `npm pack` should include `dist/main.js`.

## 8. Cursor / Copilot Rules
- The repository does not have a `.cursor/rules` folder or a `.cursorrules` file — no special Cursor restrictions apply.
- There is also no `.github/copilot-instructions.md`, so Copilot does not impose additional constraints.
- If such configuration appears later, add it to this section and keep it up to date.

## 9. Handling Uncertainties
- If it is unclear which command or module to edit, run the CLI in `debug` mode and inspect how the options resolve.
- When experimenting with new flows, add a short helper script and describe it in `AGENTS.md` so future agents understand the expectations.
- Always document new rules directly in `AGENTS.md` so subsequent agents see the updates.
