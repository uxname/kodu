# Kodu 🚀

**Kodu** is a high-performance CLI tool designed to bridge the gap between your local codebase and Large Language Models (LLMs). It automates the tedious parts of AI-assisted development: preparing context, stripping "noise" from code, performing instant reviews, and drafting commit messages.

Built for speed and developer experience, Kodu helps you get the best out of AI without the manual "copy-paste" overhead.

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
npm install -g kodu-cli # (Or your package name)

# Or run via npx
npx kodu-cli init
```

### 2. Initialization

Set up Kodu in your project:

```bash
kodu init
```
This creates a `kodu.json` configuration file and a `.kodu/` folder for your custom prompt templates.

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
```

### Clean Code
Remove comments from your JS/TS files to reduce token usage. It uses safe parsing to ensure `@ts-ignore`, `TODO`, and `biome-ignore` comments are preserved.

```bash
# See what will be removed without changing files
kodu clean --dry-run

# Clean the code
kodu clean
```

### AI Code Review
Analyze your **staged** changes before committing.

```bash
# Check for bugs (default)
kodu review

# Check for security vulnerabilities or style issues
kodu review --mode security
kodu review --mode style
```

### AI Commit Messages
Generate a concise Conventional Commit message based on your staged diff.

```bash
kodu commit
```

---

## Configuration

Kodu is controlled by `kodu.json`. You can customize:
- **LLM Settings**: Choose your model (e.g., `gpt-4o`) and provider.
- **Ignored Patterns**: Files that should never be sent to the AI (e.g., lockfiles, binaries).
- **Cleaner Whitelist**: Specific comment prefixes you want to keep.

Example `kodu.json`:
```json
{
  "llm": {
    "model": "gpt-4o"
  },
  "packer": {
    "ignore": ["*.log", "dist/**"]
  },
  "cleaner": {
    "whitelist": ["//!"]
  }
}
```

## Why Kodu?

1. **Speed**: Optimized for near-instant startup.
2. **Privacy & Control**: You decide exactly what code leaves your machine.
3. **Deterministic**: Code cleaning is performed via logic, not AI, ensuring your actual code logic is never accidentally altered.
4. **CI/CD Ready**: Use `--ci` and `--json` flags to integrate Kodu reviews into your automation pipelines.

---

**Happy Coding!** 🦄