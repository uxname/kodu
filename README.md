# pp (PrintProject)

**PrintProject** is a blazing fast CLI tool designed to prepare your codebase for Large Language Models (LLMs).

It scans your directory, ignores the junk (binaries, locks, `node_modules`), intelligently prioritizes critical files (like `package.json` or `README`), and concatenates everything into a single text output.

Ideal for pasting context into ChatGPT, Claude, or DeepSeek to get better code assistance.

![License](https://img.shields.io/npm/l/@uxname/pp)
![Version](https://img.shields.io/npm/v/@uxname/pp)

## ✨ Features

- **🚀 Fast & Lightweight:** Built with Bun, runs anywhere Node.js runs.
- **🧠 Context-Aware:** Automatically puts `package.json`, docs, and entry points at the top of the file so the LLM understands the project structure first.
- **🛡️ Smart Filtering:**
  - Respects `.gitignore`.
  - Automatically excludes binary files, lockfiles, and huge directories (`node_modules`, `.git`, `dist`).
  - Skips files larger than 1MB by default.
- **📜 Interactive History:** Run `pp` without arguments to select from your recent commands.
- **📋 clipboard-ready:** Output to a file or pipe directly to stdout.

## 📦 Installation

Install globally via npm:

```bash
npm install -g @uxname/pp
```

The command `pp` will now be available in your terminal.

## 🚀 Usage

### Basic Usage

Run in the current directory. By default, it creates a file named after the directory (e.g., `my-project.txt`).

```bash
pp
```

*If you run it without arguments, it may also show a history menu of previous commands.*

### Specify Directory and Output

Scan a specific folder and save to a specific file:

```bash
pp ./backend -o context.txt
```

### Pipe to Clipboard (macOS/Linux)

Use the stdout flag to pipe content directly to your clipboard:

```bash
# macOS
pp . --stdout | pbcopy

# Linux (xclip)
pp . --stdout | xclip -selection clipboard
```

## ⚙️ Options & Flags

| Flag | Description |
|------|-------------|
| `[directory]` | The directory to scan (default: `.`) |
| `-o, --output <file>` | Specify the output file path. |
| `-s, --stdout` | Print to stdout instead of creating a file. |
| `--exclude <pattern>` | Add custom glob patterns to exclude (e.g. `--exclude "*.css"`). |
| `--no-gitignore` | Disable `.gitignore` parsing (scan everything except default excludes). |

### Examples

**Exclude all CSS and test files:**
```bash
pp . --exclude "*.css" --exclude "*.test.ts"
```

**Ignore gitignore rules (include everything):**
```bash
pp . --no-gitignore
```

## 🧠 How it Sorts (Priority Rules)

`pp` doesn't just dump files alphabetically. It sorts them to maximize LLM understanding:

1.  **Manifests:** `package.json`, `Cargo.toml`, `go.mod`, etc.
2.  **Documentation:** `README.md`, `Dockerfile`.
3.  **Entry Points:** `index.ts`, `main.go`, `app.py`.
4.  **Config:** Configuration files.
5.  **Source Code:** Files in `src/`, `lib/`, etc.
6.  **Tests:** Test files are placed last.

## 🚫 Default Exclusions

`pp` automatically ignores:
- **Directories:** `.git`, `node_modules`, `dist`, `build`, `coverage`, `.vscode`, `__pycache__`, etc.
- **Files:** Lockfiles (`package-lock.json`, `yarn.lock`), `.DS_Store`, `.env`.
- **Extensions:** Images, videos, archives, binaries (`.png`, `.exe`, `.zip`, etc.).

## 🛠 Development

If you want to contribute or build locally:

```bash
# Clone repository
git clone https://github.com/YOUR_USERNAME/pp.git
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
