## ADDED Requirements
### Requirement: English Review Prompts
The AI service SHALL provide English-language review prompts for bug, style, and security analysis modes.

#### Scenario: Bug review in English
- **WHEN** `kodu review --mode bug` is executed
- **THEN** the AI receives an English prompt asking it to act as a strict code reviewer and find potential bugs, logical errors, and regressions

#### Scenario: Style review in English
- **WHEN** `kodu review --mode style` is executed
- **THEN** the AI receives an English prompt asking it to check readability, consistency, formatting, and naming

#### Scenario: Security review in English
- **WHEN** `kodu review --mode security` is executed
- **THEN** the AI receives an English prompt asking it to find vulnerabilities, secret leaks, and improper permission checks

### Requirement: English Error Messages
The AI service SHALL provide English error messages for API key validation and model configuration.

#### Scenario: Missing API key in English
- **WHEN** AI command is executed without API key
- **THEN** the system displays English error message about missing API key and setup instructions

#### Scenario: Invalid model format in English
- **WHEN** model configuration is invalid
- **THEN** the system displays English error message about expected model format
