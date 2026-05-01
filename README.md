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
| `-t, --template <name>` | Wrap output in a prompt template from `.kodu/prompts/` |

---

## kodu clean

Remove comments from source files using AST-based parsing. No AI, fully deterministic.

```bash
# Preview what would be removed
kodu clean --dry-run

# Clean only git-changed files (great before committing)
kodu clean --changed

# Clean all project files
kodu clean
```

Supports `.ts`, `.tsx`, `.js`, `.jsx`, `.html`. Respects `cleaner.whitelist` in `kodu.json` (e.g. `//!` to preserve important comments).

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
