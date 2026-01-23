**koda** is a blazing fast CLI tool designed to prepare your codebase for Large Language Models (LLMs).
# koda

**koda** is a blazing fast CLI tool designed to prepare your codebase for Large Language Models (LLMs).

It scans your directory, ignores the junk (binaries, locks, `node_modules`), intelligently prioritizes critical files (
like `package.json` or `README`), and concatenates everything into a single text output.

**New:** Now includes a **Code Cleanup** tool to strip comments from your project!

![License](https://img.shields.io/npm/l/koda)
![Version](https://img.shields.io/npm/v/koda)

## ✨ Features

- **🚀 Fast & Lightweight:** Built with Bun, runs anywhere Node.js runs.
- **🧠 Context-Aware:** Automatically puts `package.json`, docs, and entry points at the top of the file.
- **🛡️ Smart Filtering:**
    - Respects `.gitignore`.
    - Automatically excludes binary files, lockfiles, and huge directories (`node_modules`, `.git`, `dist`).
    - Skips files larger than 1MB by default.
- **🧹 Code Cleanup:** Safely remove all comments from JS/TS files (`.ts`, `.js`, `.tsx`, `.jsx`) to reduce token usage
  or minify code.
- **📜 Interactive History:** Run `koda` without arguments to select from your recent commands.
- **📋 clipboard-ready:** Output to a file or pipe directly to stdout.

## 📦 Installation

Install globally via npm:

```bash
npm install -g koda
```

## 🚀 Usage

### 1. Bundle Context (Default)

Scan the current directory and generate a text file for LLMs.

```bash
# Basic usage
koda

# Specify output file
koda ./backend -o context.txt

# Pipe to clipboard (macOS)
koda . --stdout | pbcopy
```

### 2. Strip Comments (Cleanup)

Remove all comments (`// ...`, `/* ... */`) from JavaScript and TypeScript files in a directory.

> **⚠️ Warning:** This command **modifies files in place**. Always commit your changes before running!

**Dry Run (Check what will happen):**
See which files will be modified without actually touching them.

```bash
koda strip --dry-run
```

**Execute Cleanup:**
This will ask for confirmation before proceeding.

```bash
koda strip
# or
koda strip ./src
```

**Force Execute (No Prompt):**
Useful for scripts or CI/CD.

```bash
koda strip -y
```

## ⚙️ Options & Flags

### Bundle Command (`koda [path]`)

| Flag                  | Description                                                             |
|-----------------------|-------------------------------------------------------------------------|
| `-o, --output <file>` | Specify the output file path.                                           |
| `-s, --stdout`        | Print to stdout instead of creating a file.                             |
| `--exclude <pattern>` | Add custom glob patterns to exclude (e.g. `--exclude "*.css"`).         |
| `--no-gitignore`      | Disable `.gitignore` parsing (scan everything except default excludes). |

### Strip Command (`koda strip [path]`)

| Flag             | Description                                                 |
|------------------|-------------------------------------------------------------|
| `-d, --dry-run`  | Show which files would be processed without modifying them. |
| `-y, --yes`      | Skip confirmation prompt (Danger!).                         |
| `-e, --exclude`  | Exclude patterns from stripping.                            |
| `--no-gitignore` | Disable `.gitignore` parsing.                               |

## 🧠 How it Sorts (Priority Rules)

`koda` sorts files to maximize LLM understanding:

1. **Manifests:** `package.json`, `Cargo.toml`, etc.
2. **Documentation:** `README.md`, `Dockerfile`.
3. **Entry Points:** `index.ts`, `main.go`.
4. **Config:** Configuration files.
5. **Source Code:** Files in `src/`, `lib/`.
6. **Tests:** Test files are placed last.

## 🚫 Default Exclusions

`koda` automatically ignores:

- **Directories:** `.git`, `node_modules`, `dist`, `build`, `coverage`, `.vscode`, `__pycache__`, etc.
- **Files:** Lockfiles (`package-lock.json`, `yarn.lock`), `.DS_Store`, `.env`.
- **Extensions:** Images, videos, archives, binaries.

## 🛠 Development

If you want to contribute or build locally:

```bash
# Clone repository
git clone https://github.com/uxname/pp.git
cd pp

# Install dependencies
bun install

# Run in dev mode
bun run dev

# Build for production
bun run build
```

## License

MIT
