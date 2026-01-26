# cleaner Specification

## Purpose
TBD - created by archiving change add-comprehensive-comment-support. Update Purpose after archive.
## Requirements
### Requirement: Recognition of all comment forms
The `clean` command SHALL detect and remove any form of comments in files (`//`, `/* ... */`, HTML `<!-- ... -->` etc.), except for constructs that fall under `keepJSDoc` or user/system whitelist. HTML `<!-- ... -->` should only be removed in `.html`/`.htm` files, so that strings with similar regular expressions or literals in TypeScript/JavaScript remain untouched. Removal SHALL occur without distorting surrounding code and be reflected in `--dry-run` previews.

#### Scenario: Removal of multiline blocks `/* ... */`
- **WHEN** `kodu clean` is run on a TS/JS file containing a `/* ... */` block that is not marked as JSDoc or is not in the whitelist
- **THEN** the block is removed, surrounding code remains syntactically correct, and `--dry-run` shows this range without applying changes

#### Scenario: Removal of HTML comments inside templates
- **WHEN** `kodu clean` is run on a template file (`.html` or `.htm`) with comments of the form `<!-- ... -->`
- **THEN** HTML comments are removed, template structure is preserved, and output/generation does not contain extra blank lines

#### Scenario: HTML comment in TypeScript literal remains untouched
- **WHEN** `kodu clean` is run on a TS/JS file where a template like `` `<!--[\s\S]*?-->` `` is encountered inside a string or regular expression
- **THEN** the text remains unchanged because HTML parsing is limited to `.html`/`.htm` files only

### Requirement: Cleaning only changed files
The `clean` command SHALL accept the `--changed` flag (short form `-c`), which switches the operation to process only those files in which Git has recorded changes. Extension rules continue to apply and output is lost if no files fall under the filter. Switching to `--dry-run` remains possible, it only changes the message without actual writing.

#### Scenario: Cleaning changes in `--changed` mode
- **WHEN** `kodu clean --changed` is run in a repository with changed TypeScript/HTML files
- **THEN** the command removes comments only from changed files, others are skipped, and the count of affected files/comments reflects only this selection

#### Scenario: `--changed` mode without actual files
- **WHEN** `kodu clean --changed` is run but Git does not produce a single changed file with a suitable extension
- **THEN** the command changes nothing and outputs the warning `No changed files to clean.`, and the spinner stops without executing `cleaner.cleanFiles`

