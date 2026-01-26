<div align="center">

# Kodu 🚀

**The AI-First CLI for Modern Developers**

Generate contexts, clean code, review PRs, and draft commits—instantly.

[![npm version](https://img.shields.io/npm/v/kodu?style=flat-square)](https://www.npmjs.com/package/kodu)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![Node.js](https://img.shields.io/badge/node-%3E%3D20-green.svg?style=flat-square)](https://nodejs.org/)
[![Built with NestJS](https://img.shields.io/badge/built%20with-NestJS-E0234E.svg?style=flat-square)](https://nestjs.com/)

[Get Started](#quick-start) • [Documentation](#usage) • [Configuration](#configuration) • [Contributing](#contributing)

</div>

---

## 💡 Why Kodu?

Bridging the gap between your local codebase and LLMs is tedious. Manual copy-pasting, token limit struggles, and formatting issues slow you down.

**Kodu** automates the "last mile" of AI-assisted development:
*   **📦 Context Packing**: Turn your file tree into a single, token-optimized prompt.
*   **🧹 Smart Cleaning**: Strip comments/noise deterministically to save tokens and money.
*   **🤖 AI Agents**: Instant code reviews and commit message generation using your API key.
*   **⚡ Performance**: Built on a modern stack (`tinyglobby`, `oxc`, `tiktoken`) for <0.5s startup.

---

## ⚡ Quick Start

### 1. Install

```bash
npm install -g kodu
```

### 2. Initialize

Run the wizard to generate `kodu.json` and prompt templates:

```bash
kodu init
```

### 3. Connect AI (Optional)

To use `review` and `commit` commands, export your API key:

```bash
# Supports OpenAI, Anthropic, Google, and 70+ others via Mastra
export OPENAI_API_KEY=sk-...
```

---

## 🛠 Usage

### 📦 Context Packing
Bundle your project files into a format ready for ChatGPT/Claude. Respects `.gitignore` automatically.

```bash
# Copy entire project context to clipboard
kodu pack --copy

# Apply a specific prompt template (e.g., for refactoring)
kodu pack --copy --template refactor

# Save to a file instead of clipboard
kodu pack --out context.txt
```

### 🧹 Code Cleaning
Save tokens by stripping comments (`//`, `/** */`, `<!-- -->`) before sending code to AI.
*Note: Keeps `@ts-ignore`, `TODO`, and `biome-ignore` by default.*

```bash
# Preview what would be removed
kodu clean --dry-run

# Clean only files changed in git (perfect for PRs)
kodu clean --changed

# Clean everything
kodu clean
```

### 🔍 AI Code Review
Get an instant second opinion on your **staged** changes.

```bash
# Default review (Bugs & Logic)
kodu review

# Focus on specific aspects
kodu review --mode security
kodu review --mode style

# CI/CD Mode (no spinners, raw output)
kodu review --mode bug --ci --output report.md
```

### 📝 Smart Commit Messages
Generate Conventional Commits based on the actual diff.

```bash
# Print suggested message
kodu commit

# Use directly with git
git commit -m "$(kodu commit --ci)"
```

---

## ⚙️ Configuration

Kodu is fully customizable via `kodu.json`.

```json
{
  "$schema": "https://raw.githubusercontent.com/uxname/kodu/master/kodu.schema.json",
  "llm": {
    "model": "openai/gpt-4o",
    "apiKeyEnv": "OPENAI_API_KEY",
    "commands": {
      "review": {
        "modelSettings": { "maxOutputTokens": 5000 }
      }
    }
  },
  "packer": {
    "ignore": ["package-lock.json", "dist", "coverage"],
    "useGitignore": true
  },
  "cleaner": {
    "whitelist": ["//!"] // Comments starting with //! will be preserved
  },
  "prompts": {
    "review": {
      "bug": ".kodu/prompts/review-bug.md",
      "custom": ".kodu/prompts/my-custom-prompt.md"
    }
  }
}
```

### Prompt Templates
Kodu uses Markdown files for prompts. You can use variables like `{diff}`, `{context}`, or `{fileList}` inside them.
Templates are stored in `.kodu/prompts/` by default.

---

## 🤖 CI/CD Integration

Kodu is designed to run in pipelines (GitHub Actions, GitLab CI).

**Example: Automated PR Review**
```yaml
- name: AI Code Review
  run: |
    npm install -g kodu
    kodu review --mode bug --ci --output pr-review.md
  env:
    OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

---

## 🏗 Architecture & Stack

We believe in a "Fresh & Modern" stack strategy. No legacy dependencies.

| Component | Technology | Why? |
| :--- | :--- | :--- |
| **Framework** | NestJS + Commander | Modular architecture & DI |
| **AI Engine** | Mastra | Model routing & agent orchestration |
| **Parsing** | ts-morph | AST-based safety (no Regex hacking) |
| **Filesystem** | tinyglobby | Fastest glob matching available |
| **UI** | @inquirer + yocto-spinner | Modern, interactive CLI UX |

---

## 🤝 Contributing

Contributions are welcome!
1. Fork the repo
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

<div align="center">
  <sub>Built with ❤️ by Developers for Developers</sub>
</div>
