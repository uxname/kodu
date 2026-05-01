<div align="center">

# Kodu

**Bundle your codebase for LLMs. Strip noise. Ship faster.**

[![npm version](https://img.shields.io/npm/v/kodu?style=flat-square&color=black)](https://www.npmjs.com/package/kodu)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square&color=black)](LICENSE)

</div>

---

## What it does

| Problem | Kodu |
| :--- | :--- |
| Copy-pasting files one by one into ChatGPT | `kodu pack` bundles your entire project in one command |
| Hitting token limits with comments and docs | `kodu clean` strips comments deterministically via AST |

---

## Install

```bash
npm install -g kodu
```

Or run without installing:

```bash
npx kodu pack --copy
```

---

## kodu init

Add `.kodu/context.txt` to `.gitignore` so generated context files are never committed:

```bash
kodu init
```

Run once after cloning or setting up the project.

---

## kodu pack

Bundle project files into a single context file optimized for LLMs.

```bash
# Pack everything and copy to clipboard
kodu pack --copy

# Pack only specific directories
kodu pack --path src --path tests --copy

# Just see what files would be included
kodu pack --list

# Exclude extra patterns on the fly
kodu pack --exclude "**/*.test.ts" --exclude "docs/" --copy

# Save to a custom path
kodu pack --out /tmp/context.txt

# Use plain text format instead of XML
kodu pack --format text --copy
```

### Output format

By default, kodu wraps each file in XML tags — the format that LLMs parse most reliably:

```xml
<files>
<file path="src/index.ts">
// your code here
</file>

<file path="src/utils.ts">
// more code
</file>
</files>
```

Use `--format text` for legacy `// file: path` style headers.

### Options

| Flag | Description |
|------|-------------|
| `-c, --copy` | Copy result to clipboard |
| `-o, --out <path>` | Output file path (default: `.kodu/context.txt`) |
| `-p, --path <path>` | Include only this directory/glob (repeatable) |
| `-e, --exclude <pattern>` | Additional exclude pattern (repeatable) |
| `-l, --list` | Print file list only, no content |
| `-f, --format <xml\|text>` | Output format (default: `xml`) |
| `--clean` | Strip comments in-memory before packing (files not modified) |
| `-t, --template <name>` | Wrap output in a prompt template from `.kodu/prompts/` |

---

## kodu clean

Remove comments from source files using AST-based parsing. No AI, fully deterministic.

```bash
# Preview what would be removed (with byte/token savings)
kodu clean --dry-run

# Show every removed comment, not just first 3
kodu clean --dry-run --verbose

# Clean only git-staged files
kodu clean --staged

# Clean only git-changed files (staged + unstaged + untracked)
kodu clean --changed

# Target specific files or directories
kodu clean src/utils.ts src/helpers/

# Remove JSDoc too (overrides config)
kodu clean --no-jsdoc

# Backup originals before modifying
kodu clean --backup

# Read from stdin, write to stdout (great for scripting)
cat src/foo.ts | kodu clean --stdin

# Clean all project files
kodu clean
```

Supports `.ts`, `.tsx`, `.js`, `.jsx`, `.mjs`, `.cjs`, `.html`. Respects `cleaner.whitelist` in `kodu.json` (e.g. `//!` to preserve important comments).

### Options

| Flag | Description |
|------|-------------|
| `-d, --dry-run` | Show what will be removed without modifying files |
| `-v, --verbose` | Show all removed comments in dry-run (not just first 3) |
| `-c, --changed` | Clean only git-changed files (staged + unstaged + untracked) |
| `-s, --staged` | Clean only git-staged files |
| `-n, --no-jsdoc` | Remove JSDoc comments (overrides `keepJSDoc` in config) |
| `-b, --backup` | Save originals to `.kodu/backup/` before modifying |
| `--stdin` | Read from stdin, write cleaned result to stdout |

---

## Configuration

Create `kodu.json` in your project root:

```json
{
  "cleaner": {
    "whitelist": ["//!"],
    "keepJSDoc": true,
    "useGitignore": true
  },
  "packer": {
    "ignore": ["package-lock.json", "dist", "coverage"],
    "useGitignore": true
  }
}
```

Both commands work without a config file using sensible defaults.

### Custom pack template

Point `prompts.pack` at a markdown file to wrap packed context in a prompt:

```json
{
  "prompts": {
    "pack": ".kodu/prompts/pack.md"
  }
}
```

Available template variables: `{{context}}`, `{{fileList}}`, `{{tokenCount}}`, `{{usdEstimate}}`.

---

## Privacy

- All processing runs locally
- No data sent anywhere
- API keys are never stored — only read from env vars

---

<div align="center">
  <sub>Built for productive developers.</sub>
  <br>
  <a href="CONTRIBUTING.md">Contributing</a> • <a href="LICENSE">License</a>
</div>
