## MODIFIED Requirements
### Requirement: Comment Recognition
The `clean` command SHALL detect and remove any form of comments in files (`//`, `/* ... */`, HTML `<!-- ... -->`, etc.), except for constructs covered by `keepJSDoc` or custom/system whitelist. HTML `<!-- ... -->` should only be removed in `.html`/`.htm` files, so that lines with similar regular expressions or literals in TypeScript/JavaScript remain untouched. Removal SHALL occur without distorting surrounding code and be reflected in `--dry-run` previews.

#### Scenario: Remove multi-line `/* ... */` blocks
- **WHEN** `kodu clean` runs on a TS/JS file containing a `/* ... */` block that is not marked as JSDoc or not in the whitelist
- **THEN** the block is removed, surrounding code remains syntactically correct, and `--dry-run` shows this range without applying changes

#### Scenario: Remove HTML comments inside templates
- **WHEN** `kodu clean` runs on a template file (`.html` or `.htm`) with comments of the form `<!-- ... -->`
- **THEN** HTML comments are removed, template structure is preserved, and output/generation does not contain extra whitespace lines

#### Scenario: HTML comment in TypeScript literal remains untouched
- **WHEN** `kodu clean` runs on a TS/JS file where a pattern `` `<!--[\s\S]*?-->` `` appears inside a string or regular expression
- **THEN** the text remains unchanged because HTML parsing is limited to `.html`/`.htm` files only

### Requirement: Clean Only Changed Files
The `clean` command SHALL accept a `--changed` flag (short form `-c`) that switches the operation to process only those files where Git has detected changes. Extension rules continue to apply, and output is lost if no files match the filter. Switching to `--dry-run` remains possible, it only changes the message without actual writing.

#### Scenario: Clean changes in `--changed` mode
- **WHEN** `kodu clean --changed` runs in a repository with changed TypeScript/HTML files
- **THEN** the command removes comments only from changed files, others are skipped, and the count of affected files/comments reflects only this selection

#### Scenario: `--changed` mode with no current files
- **WHEN** `kodu clean --changed` runs but Git produces no changed file with suitable extension
- **THEN** the command makes no changes and outputs warning `No changed files to clean.`, and the spinner stops without executing `cleaner.cleanFiles`
