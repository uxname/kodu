## 1. Translation Preparation
- [x] 1.1 Inventory all Russian text in the codebase
- [x] 1.2 Create translation mapping for key terms
- [x] 1.3 Identify user-facing vs internal strings

## 2. Core Configuration Translation
- [x] 2.1 Translate `DEFAULT_REVIEW_PROMPTS` in `src/core/config/default-prompts.ts`
- [x] 2.2 Translate prompt templates in `.kodu/prompts/` directory
- [x] 2.3 Update configuration schema if needed

## 3. Command Files Translation
- [x] 3.1 Translate Russian strings in `src/commands/review/review.command.ts`
- [x] 3.2 Translate Russian strings in `src/commands/commit/commit.command.ts`
- [x] 3.3 Translate Russian strings in `src/commands/init/init.command.ts`
- [x] 3.4 Translate Russian strings in `src/commands/pack/pack.command.ts`
- [x] 3.5 Translate Russian strings in `src/commands/clean/clean.command.ts`

## 4. Service Files Translation
- [x] 4.1 Translate Russian strings in `src/shared/ai/ai.service.ts`
- [x] 4.2 Translate Russian strings in `src/core/config/config.service.ts`
- [x] 4.3 Translate Russian strings in `src/core/config/prompt.service.ts`
- [x] 4.4 Translate Russian strings in `src/shared/git/git.service.ts`

## 5. Documentation Translation
- [x] 5.1 Translate `openspec/specs/cleaner/spec.md`
- [x] 5.2 Update archived changes documentation
- [x] 5.3 Translate `docs/todo.md`
- [x] 5.4 Update Russian examples in `AGENTS.md`

## 6. Validation
- [x] 6.1 Test all commands with English output
- [x] 6.2 Verify no Russian text remains in user-facing strings
- [x] 6.3 Run `npm run check` to ensure no breaking changes
