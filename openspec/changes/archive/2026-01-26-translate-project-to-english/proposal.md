# Change: Translate all project content from Russian to English

## Why
The project currently contains mixed language content with Russian text in prompts, error messages, specifications, and user-facing strings. This creates inconsistency and limits the project's accessibility to international contributors and users.

## What Changes
- Translate all Russian prompts in `src/core/config/default-prompts.ts`
- Translate all Russian prompt templates in `.kodu/prompts/` directory
- Translate Russian specifications in `openspec/specs/cleaner/spec.md`
- Translate Russian error messages and user-facing strings in command files
- Update all Russian comments and documentation to English
- **BREAKING**: Changes all user-facing text from Russian to English

## Impact
- Affected specs: cleaner, ai, config, ui
- Affected code: All command files, configuration services, prompt templates
- User experience: All CLI output and prompts will be in English
