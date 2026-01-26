<div align="center">

# Kodu 🦄

**The AI-First CLI for Modern Developers**

Generate contexts, clean code, review PRs, and draft commits—instantly.

[![npm version](https://img.shields.io/npm/v/kodu?style=flat-square&color=black)](https://www.npmjs.com/package/kodu)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square&color=black)](LICENSE)
[![Privacy](https://img.shields.io/badge/Privacy-First-green.svg?style=flat-square)](https://github.com/uxname/kodu)

</div>

---

## ⚡ The Problem vs. Kodu

| You (Manual) ❌ | Kodu (Automated) ✅ |
| :--- | :--- |
| Copy-pasting 10 files one by one | **`kodu pack`** bundles context in 1 click |
| Hitting token limits with comments | **`kodu clean`** strips noise deterministically |
| Context switching for code reviews | **`kodu review`** checks logic inside your terminal |
| Writing boring commit messages | **`kodu commit`** generates semantic git messages |

---

## 🚀 Instant Start

Don't want to install? Run it directly:

```bash
# 1. Initialize config (creates kodu.json)
npx kodu init

# 2. Bundle your project to clipboard
npx kodu pack --copy
```

**Or install globally for speed (<0.5s startup):**

```bash
npm install -g kodu
```

---

## 🛡️ Privacy & Security

We know trust is paramount when dealing with code.

*   **Local Execution:** Code analysis runs locally.
*   **Zero Data Retention:** We don't store your code.
*   **Explicit Control:** `.env`, `node_modules`, and lockfiles are ignored by default.
*   **You Own the Keys:** Your API key (`OPENAI_API_KEY`) goes directly to the provider.

---

## 💡 Common Workflows

### 1. "I need to ask ChatGPT about my project"
Pack your entire source code (minus ignored files) into the clipboard, optimized for tokens.

```bash
kodu pack --copy
# Output: Copied 45 files (12k tokens) to clipboard. Paste into ChatGPT!
```

### 2. "I want to save money on API costs"
Strip comments and docs to reduce token count by ~30% before sending.

```bash
# Preview savings
kodu clean --dry-run

# Clean only what you changed (Great for PRs)
kodu clean --changed
```

### 3. "Check my code before I push"
Get an AI review of your **staged** changes without leaving the terminal.

```bash
# Detect bugs and logical errors
kodu review --mode bug

# Check for security leaks
kodu review --mode security
```

---

## ⚙️ Configuration

Kodu creates a `kodu.json` in your root. It's pre-configured, but fully hackable.

<details>
<summary><b>Click to see example configuration</b></summary>

```json
{
  "llm": {
    "model": "openai/gpt-4o",
    "apiKeyEnv": "OPENAI_API_KEY"
  },
  "cleaner": {
    // Kodu will NEVER remove comments starting with these:
    "whitelist": ["//!"]
  },
  "packer": {
    // Files to strictly ignore
    "ignore": ["package-lock.json", "dist", "coverage"]
  },
  "prompts": {
    // Use your own prompts for reviews
    "review": {
      "bug": ".kodu/prompts/review-bug.md"
    }
  }
}
```
</details>

---

## 🏎️ Tech Stack (Fresh & Modern)

Built for speed and maintainability.

*   **Runtime:** Node.js (ESM)
*   **Framework:** NestJS + Commander
*   **Parsing:** `ts-morph` (AST-based, not Regex)
*   **Globbing:** `tinyglobby`
*   **UI:** `@inquirer` + `yocto-spinner`

---

<div align="center">
  <sub>Built with ❤️ for productive developers.</sub>
  <br>
  <a href="CONTRIBUTING.md">Contributing</a> • <a href="LICENSE">License</a>
</div>
