# Kodu 🚀

**Kodu** is a high-performance CLI tool designed to bridge the gap between your local codebase and Large Language Models (LLMs). It automates the tedious parts of AI-assisted development: preparing context, stripping "noise" from code, performing instant reviews, and drafting commit messages.

Built for speed and developer experience, Kodu helps you get the best out of AI without the manual "copy-paste" overhead.

---

## TLDR

```bash
# Install
npm install -g kodu

# Setup
kodu init
export OPENAI_API_KEY=your_key_here

# Quick commands
kodu pack --copy              # Copy project context to clipboard
kodu clean                    # Remove comments from code
kodu review                   # Review staged changes
kodu commit                   # Generate commit message
```

---

## Key Features

- **📦 Smart Context Packing**: Bundle your entire project (or specific parts) into a single, LLM-friendly format.
- **🧹 Deterministic Code Cleaning**: Strip comments and unnecessary metadata to save tokens while keeping critical logic and system instructions.
- **🔍 Instant AI Review**: Get immediate feedback on your staged changes (Bugs, Style, or Security).
- **📝 Automated Commits**: Generate meaningful, Conventional Commit messages based on your actual code changes.
- **💰 Token Budgeting**: Always know how many tokens you are sending and get an estimated cost before hitting the API.

---

## Quick Start

### 1. Installation

```bash
# Install globally
npm install -g kodu

# Or run via npx
npx kodu init
```

### 2. Initialization

Set up Kodu in your project:

```bash
kodu init
```

This creates a `kodu.json` configuration file and a `.kodu/prompts/` folder for your custom prompt templates.

### 3. Configure AI (Optional)

For AI-powered features (`review`, `commit`), set your API key in your environment:

```bash
export OPENAI_API_KEY=your_key_here
```

---

## Usage

### Pack Context
Collect your project files into one file or directly to your clipboard. Kodu respects your `.gitignore` automatically.

```bash
# Copy context to clipboard with a specific prompt template
kodu pack --copy --template refactor

# Save context to a specific file
kodu pack --out context.txt

# Output to stdout
kodu pack
```

### Clean Code
Remove comments from your JS/TS files and `.html`/`.htm` templates (including `//`, `/* ... */`, and `<!-- ... -->`) to reduce token usage. HTML/HTM files gain `<!-- ... -->` stripping, while other file types stay untouched to avoid removing regex literals or strings; `@ts-ignore`, `TODO`, and `biome-ignore` comments are preserved and `--dry-run` previews the impact.

```bash
# See what will be removed without changing files
kodu clean --dry-run

# Clean the code globally
kodu clean

# Clean only files with Git changes
kodu clean --changed
```

### AI Code Review
Analyze your **staged** changes before committing.

```bash
# Check for bugs (default)
kodu review

# Check for security vulnerabilities or style issues
kodu review --mode security
kodu review --mode style

# Custom review modes (if added to kodu.json)
kodu review --mode performance
kodu review --mode accessibility

# CI/CD mode with JSON output
kodu review --mode bug --ci --output review.txt
```

Standard modes (`bug`, `style`, `security`) are always available. Custom modes require configuration in `prompts.review` section of `kodu.json`.

### AI Commit Messages
Generate a concise Conventional Commit message based on your staged diff.

```bash
# Generate commit message
kodu commit

# Save to file
kodu commit --output commit-message.txt

# CI/CD mode
kodu commit --ci
```

**Note:** `kodu commit` only generates the message. You still need to run `git commit -m "$(kodu commit)"` to commit.

---

## Configuration

Kodu is controlled by `kodu.json`. You can customize:
- **LLM Settings**: Choose your model (e.g., `gpt-4o`) and provider.
- **Ignored Patterns**: Files that should never be sent to the AI (e.g., lockfiles, binaries).
- **Cleaner Whitelist**: Specific comment prefixes you want to keep.
- **AI Prompts**: Customize prompts for review and commit commands, including custom review modes.

Example `kodu.json`:
```json
{
  "llm": {
    "provider": "openai",
    "model": "gpt-4o",
    "apiKeyEnv": "OPENAI_API_KEY"
  },
  "packer": {
    "ignore": ["*.log", "dist/**"]
  },
  "cleaner": {
    "whitelist": ["//!"],
    "keepJSDoc": true
  },
  "prompts": {
    "review": {
      "bug": ".kodu/prompts/review-bug.md",
      "style": ".kodu/prompts/review-style.md",
      "security": ".kodu/prompts/review-security.md"
    },
    "commit": ".kodu/prompts/commit.md",
    "pack": ".kodu/prompts/pack.md"
  }
}
```

**Prompts:**
- `prompts.review`: Paths to prompt files for review modes. Supports standard modes (`bug`, `style`, `security`) and custom modes. Use `{diff}` and `{mode}` variables inside files.
- `prompts.commit`: Path to prompt file for commit message generation. Use `{diff}` variable.
- `prompts.pack`: Path to prompt file applied by default in `kodu pack`. Supports `{{context}}`, `{{fileList}}`, `{{tokenCount}}`, `{{usdEstimate}}` placeholders.

---

## Why Kodu?

1. **Speed**: Optimized for near-instant startup (< 0.5s).
2. **Privacy & Control**: You decide exactly what code leaves your machine.
3. **Deterministic**: Code cleaning is performed via logic, not AI, ensuring your actual code logic is never accidentally altered.
4. **CI/CD Ready**: Use `--ci` flag to integrate Kodu reviews into your automation pipelines.

---

**Happy Coding!** 🦄
