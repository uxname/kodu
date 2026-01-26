## MODIFIED Requirements
### Requirement: English Configuration Validation
The configuration service SHALL provide English error messages for missing or invalid configuration files.

#### Scenario: Missing configuration file in English
- **WHEN** `kodu.json` is not found
- **THEN** the system displays English error message suggesting to run `kodu init`

#### Scenario: Invalid configuration in English
- **WHEN** `kodu.json` contains validation errors
- **THEN** the system displays English error messages for each validation issue

### Requirement: English Prompt Loading
The prompt service SHALL provide English error messages for missing or invalid prompt templates.

#### Scenario: Missing prompt template in English
- **WHEN** a prompt template file is not found
- **THEN** the system displays English error message with expected file paths

#### Scenario: Invalid prompt file in English
- **WHEN** a prompt file path is invalid
- **THEN** the system displays English error message about file not found
