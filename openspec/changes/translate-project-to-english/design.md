## Context
The Kodu CLI project currently uses mixed languages with Russian prompts and error messages. This creates inconsistency and limits international adoption. The project needs comprehensive translation to English while maintaining functionality.

## Goals / Non-Goals
- Goals: Complete translation of all user-facing text to English
- Goals: Maintain all existing functionality
- Goals: Ensure consistency across all prompts and messages
- Non-Goals: Translate code comments (keep technical comments)
- Non-Goals: Change project structure or architecture

## Decisions
- Decision: Translate all user-facing strings to English
- Decision: Keep technical code comments in original form for developer context
- Decision: Maintain existing prompt structure and variable substitution
- Alternatives considered: Keep bilingual support (rejected for complexity)

## Risks / Trade-offs
- Breaking change for existing Russian-speaking users
- Risk of translation errors affecting prompt quality
- Trade-off: Better international adoption vs current user base disruption

## Migration Plan
- Translate all files systematically
- Test each command individually
- Update documentation
- Release with migration guide

## Open Questions
- How to handle existing Russian user configurations?
- Should we provide fallback support for Russian prompts?
