## ADDED Requirements
### Requirement: English CLI Commands
All CLI commands SHALL provide English descriptions, option help text, and user-facing messages.

#### Scenario: English command descriptions
- **WHEN** users run `kodu --help` or command-specific help
- **THEN** all command descriptions and option help text are in English

#### Scenario: English command output
- **WHEN** users execute any kodu command
- **THEN** all progress messages, spinners, and results are in English

### Requirement: English Error Handling
All error messages and warnings SHALL be displayed in English.

#### Scenario: English error messages
- **WHEN** an error occurs during command execution
- **THEN** the error message is in English with clear guidance

#### Scenario: English warning messages
- **WHEN** a warning is displayed (e.g., missing staged changes)
- **THEN** the warning message is in English

### Requirement: English Interactive Prompts
All interactive prompts and user dialogs SHALL be in English.

#### Scenario: English confirmation prompts
- **WHEN** a command requires user confirmation
- **THEN** the prompt text is in English

#### Scenario: English input prompts
- **WHEN** a command requires user input
- **THEN** the prompt text and validation messages are in English
