#!/usr/bin/env node

import { homedir } from "os";
import { spawn } from "child_process";
import prompts from "prompts";
import {
  mkdirSync,
  readFileSync,
  readdirSync,
  lstatSync,
  openSync,
  readSync,
  closeSync,
  existsSync,
  writeFileSync,
} from "fs";
import { join, relative, basename, extname, dirname, resolve, sep } from "path";
import { program } from "commander";
import ora from "ora";
import ignore, { type Ignore } from "ignore";

const CONFIG = {
  MAX_FILE_SIZE_BYTES: 1 * 1024 * 1024,
  EXCLUDED_DIRS: new Set([
    ".git",
    ".github",
    ".idea",
    ".vscode",
    "node_modules",
    "dist",
    "build",
    "coverage",
    "public",
    "vendor",
    ".cache",
    ".next",
    ".nuxt",
    ".svelte-kit",
    "__pycache__",
    ".venv",
    "env",
    "target",
    "out",
    "bin",
    "obj",
    "log",
    "logs",
  ]),
  EXCLUDED_FILES: new Set([
    "package-lock.json",
    "yarn.lock",
    "pnpm-lock.yaml",
    "bun.lockb",
    ".env",
    ".DS_Store",
  ]),
  EXCLUDED_PATTERNS: new Set([
    "*.svg",
    "*.png",
    "*.jpg",
    "*.jpeg",
    "*.gif",
    "*.webp",
    "*.ico",
    "*.mp4",
    "*.mov",
    "*.avi",
    "*.webm",
    "*.pdf",
    "*.zip",
    "*.rar",
    "*.gz",
    "*.lockb",
  ]),
  INCLUDED_EXTENSIONS_AND_FILENAMES: new Set([
    "package.json",
    "pom.xml",
    "pyproject.toml",
    "requirements.txt",
    "go.mod",
    "Cargo.toml",
    "composer.json",
    "Dockerfile",
    "docker-compose.yml",
    "Makefile",
    "README.md",
    ".js",
    ".mjs",
    ".cjs",
    ".ts",
    ".mts",
    ".cts",
    ".jsx",
    ".tsx",
    ".json",
    ".html",
    ".htm",
    ".css",
    ".scss",
    ".sass",
    ".less",
    ".styl",
    ".vue",
    ".svelte",
    ".py",
    ".go",
    ".rs",
    ".java",
    ".kt",
    ".scala",
    ".c",
    ".cpp",
    ".h",
    ".hpp",
    ".cs",
    ".fs",
    ".rb",
    ".php",
    ".swift",
    ".dart",
    ".sh",
    ".ps1",
    ".sql",
    ".md",
    ".toml",
    ".yaml",
    ".yml",
    ".xml",
    ".ini",
    ".properties",
    ".env.example",
    ".conf",
    ".cfg",
  ]),
};

const BINARY_CHECK_BUFFER_SIZE = 512;
const SYSTEM_ARGS_COUNT = 2;
const MAX_HISTORY_ENTRIES = 20;
const CONFIG_DIR = join(homedir(), ".config", "ctx");
const HISTORY_FILE = join(CONFIG_DIR, "history.json");

interface PriorityRule {
  score: number;
  test: (relativePath: string, fileBasename: string, fileDirname: string) => boolean;
}

interface FileWithScore {
  path: string;
  score: number;
}

interface OutputOptions {
  stdout?: boolean;
  output?: string;
}

interface CommandOptions extends OutputOptions {
  gitignore: boolean;
  exclude: string[];
}

const PRIORITY_RULES: PriorityRule[] = [
  {
    score: 1,
    test: (_: string, fileBasename: string) =>
      [
        "package.json",
        "pom.xml",
        "pyproject.toml",
        "requirements.txt",
        "go.mod",
        "cargo.toml",
        "composer.json",
      ].includes(fileBasename),
  },
  {
    score: 2,
    test: (_: string, fileBasename: string) =>
      fileBasename.startsWith("readme") ||
      fileBasename.startsWith("dockerfile") ||
      fileBasename === "docker-compose.yml" ||
      fileBasename === "makefile",
  },
  {
    score: 3,
    test: (_: string, fileBasename: string) =>
      [
        "main.js",
        "index.js",
        "main.ts",
        "index.ts",
        "app.py",
        "main.go",
        "main.rs",
        "main.java",
      ].includes(fileBasename),
  },
  {
    score: 4,
    test: (_: string, fileBasename: string, fileDirname: string) =>
      fileBasename.includes("config") && !fileDirname.includes("test"),
  },
  {
    score: 5,
    test: (relativePath: string) =>
      ["src", "app", "lib", "core", "cmd"].some((dir) =>
        relativePath.startsWith(dir + sep)
      ),
  },
  {
    score: 10,
    test: (relativePath: string, fileBasename: string, fileDirname: string) =>
      fileBasename.includes(".test.") ||
      fileBasename.includes(".spec.") ||
      fileBasename.endsWith("_test.go") ||
      fileDirname.includes("test"),
  },
];

function createFilter(patterns: Iterable<string>): Ignore {
  return ignore().add(Array.from(patterns));
}

function isBinaryFile(filePath: string): boolean {
  const buffer = Buffer.alloc(BINARY_CHECK_BUFFER_SIZE);
  try {
    const fileDescriptor = openSync(filePath, "r");
    const bytesRead = readSync(fileDescriptor, buffer, 0, BINARY_CHECK_BUFFER_SIZE, 0);
    closeSync(fileDescriptor);
    return buffer.slice(0, bytesRead).includes(0);
  } catch {
    return true;
  }
}

function getPriorityScore(relativePath: string): number {
  const fileBasename = basename(relativePath).toLowerCase();
  const fileDirname = dirname(relativePath);

  for (const rule of PRIORITY_RULES) {
    if (rule.test(relativePath, fileBasename, fileDirname)) {
      return rule.score;
    }
  }

  return 8;
}

function shouldSkipEntry(
  entry: string,
  relativePath: string,
  gitignoreFilter: Ignore,
  fullPath: string
): boolean {
  if (gitignoreFilter.ignores(relativePath)) {
    return true;
  }

  const stat = lstatSync(fullPath);

  if (stat.isSymbolicLink()) {
    return true;
  }

  if (stat.isDirectory() && CONFIG.EXCLUDED_DIRS.has(entry)) {
    return true;
  }

  return false;
}

function isFileIncluded(
  fullPath: string,
  entryName: string,
  exclusionFilter: Ignore,
  stat: { size: number }
): boolean {
  if (stat.size > CONFIG.MAX_FILE_SIZE_BYTES) {
    return false;
  }

  if (exclusionFilter.ignores(entryName) || CONFIG.EXCLUDED_FILES.has(entryName)) {
    return false;
  }

  const extension = extname(entryName);
  const hasIncludedExtensionOrName =
    CONFIG.INCLUDED_EXTENSIONS_AND_FILENAMES.has(extension) ||
    CONFIG.INCLUDED_EXTENSIONS_AND_FILENAMES.has(entryName);

  return hasIncludedExtensionOrName && !isBinaryFile(fullPath);
}

function collectFilesRecursively(
  dirPath: string,
  rootPath: string,
  gitignoreFilter: Ignore,
  exclusionFilter: Ignore
): string[] {
  let collectedFiles: string[] = [];

  try {
    for (const entry of readdirSync(dirPath)) {
      const fullPath = join(dirPath, entry);
      const relativePath = relative(rootPath, fullPath).replace(/\\/g, "/");

      if (shouldSkipEntry(entry, relativePath, gitignoreFilter, fullPath)) {
        continue;
      }

      const stat = lstatSync(fullPath);

      if (stat.isDirectory()) {
        collectedFiles.push(
          ...collectFilesRecursively(
            fullPath,
            rootPath,
            gitignoreFilter,
            exclusionFilter
          )
        );
      } else if (stat.isFile()) {
        const entryName = basename(fullPath);

        if (isFileIncluded(fullPath, entryName, exclusionFilter, stat)) {
          collectedFiles.push(fullPath);
        }
      }
    }
  } catch (error: unknown) {
    const errorCode = error instanceof Error && 'code' in error
      ? (error as NodeJS.ErrnoException).code
      : 'UNKNOWN';
    console.error(
      `\n[Warning] Cannot process path, skipping: ${dirPath} (${errorCode})`
    );
  }

  return collectedFiles;
}

function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) {
    return "0 Bytes";
  }

  const kilobyte = 1024;
  const decimalPlaces = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
  const unitIndex = Math.floor(Math.log(bytes) / Math.log(kilobyte));

  return `${parseFloat((bytes / Math.pow(kilobyte, unitIndex)).toFixed(decimalPlaces))} ${sizes[unitIndex]}`;
}

function generateOutputContent(
  files: FileWithScore[],
  rootDir: string
): string {
  const outputParts: string[] = [];

  for (const file of files) {
    const relativePath = relative(rootDir, file.path).replace(/\\/g, "/");

    try {
      const content = readFileSync(file.path, "utf8");
      outputParts.push(`// File: ${relativePath}\n${content}\n---\n`);
    } catch (error: unknown) {
      const errorCode = error instanceof Error && 'code' in error
        ? (error as NodeJS.ErrnoException).code
        : 'UNKNOWN';
      console.error(
        `[Warning] Cannot read file, skipping: ${relativePath} (${errorCode})`
      );
    }
  }

  return outputParts.join("\n");
}

function writeOutput(
  content: string,
  options: OutputOptions,
  rootDir: string
): void {
  const outputSize = formatBytes(Buffer.from(content).length);

  if (options.stdout) {
    console.error(
      `\n✨ Generated ${outputSize} of content. Piping to stdout...`
    );
    console.log(content);
    return;
  }

  const outputPath = options.output ?? join(process.cwd(), `${basename(rootDir)}.txt`);
  const absoluteOutputPath = resolve(outputPath);

  try {
    writeFileSync(absoluteOutputPath, content);
    console.log(`\n✅ Success!`);
    console.log(`   Scanned Directory: ${rootDir}`);
    console.log(`   Output File:       ${absoluteOutputPath}`);
    console.log(`   Total Size:        ${outputSize}`);
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    console.error(
      `\nError: Failed to write to file ${absoluteOutputPath}: ${errorMessage}`
    );
    process.exit(1);
  }
}

class HistoryManager {
  private history: string[][] = [];

  constructor() {
    this.load();
  }

  private load(): void {
    if (!existsSync(HISTORY_FILE)) {
      return;
    }

    try {
      const data = readFileSync(HISTORY_FILE, "utf-8");
      this.history = JSON.parse(data);
    } catch {
      console.error(
        "[Warning] Could not load history file. It might be corrupted."
      );
      this.history = [];
    }
  }

  private save(): void {
    try {
      if (!existsSync(CONFIG_DIR)) {
        mkdirSync(CONFIG_DIR, { recursive: true });
      }
      writeFileSync(HISTORY_FILE, JSON.stringify(this.history, null, 2));
    } catch {
      console.error("[Warning] Could not save history file.");
    }
  }

  add(args: string[]): void {
    this.history = this.history.filter(
      (historyEntry) => !this.areArgsEqual(historyEntry, args)
    );

    this.history.unshift(args);

    if (this.history.length > MAX_HISTORY_ENTRIES) {
      this.history.pop();
    }

    this.save();
  }

  get(): string[][] {
    return this.history;
  }

  private areArgsEqual(a: string[], b: string[]): boolean {
    return a.length === b.length && a.every((val, index) => val === b[index]);
  }
}

const historyManager = new HistoryManager();

async function showHistoryMenu(): Promise<void> {
  const history = historyManager.get();

  if (history.length === 0) {
    console.log(
      "No history found. Run a command first to create a history entry."
    );
    return;
  }

  const choices = history.map((args) => ({
    title: args.join(" "),
    value: args,
  }));

  const response = await prompts({
    type: "select",
    name: "command",
    message: "Select a recent command to run again",
    choices: choices,
  });

  if (response.command) {
    spawn(process.argv[0]!, [process.argv[1], ...response.command], {
      stdio: "inherit",
    });
  }
}

function isValidDirectory(path: string): boolean {
  return existsSync(path) && lstatSync(path).isDirectory();
}

async function run(dir: string, options: CommandOptions): Promise<void> {
  const absoluteRootDir = resolve(dir);

  if (!isValidDirectory(absoluteRootDir)) {
    console.error(
      `Error: Directory not found or is not a directory: ${absoluteRootDir}`
    );
    process.exit(1);
  }

  const spinner = ora("Scanning files...").start();

  const gitignorePath = join(absoluteRootDir, ".gitignore");
  const gitignorePatterns =
    options.gitignore && existsSync(gitignorePath)
      ? readFileSync(gitignorePath, "utf8")
      : "";

  const gitignoreFilter = createFilter(gitignorePatterns.split("\n"));
  const exclusionPatterns = new Set([
    ...CONFIG.EXCLUDED_PATTERNS,
    ...options.exclude,
  ]);
  const exclusionFilter = createFilter(exclusionPatterns);

  const allFilePaths = collectFilesRecursively(
    absoluteRootDir,
    absoluteRootDir,
    gitignoreFilter,
    exclusionFilter
  );

  const prioritizedFiles = allFilePaths
    .map((filePath) => ({
      path: filePath,
      score: getPriorityScore(relative(absoluteRootDir, filePath)),
    }))
    .sort((a, b) => a.score - b.score);

  spinner.succeed(`Found ${prioritizedFiles.length} files to process.`);

  const finalOutput = generateOutputContent(prioritizedFiles, absoluteRootDir);
  writeOutput(finalOutput, options, absoluteRootDir);

  historyManager.add(process.argv.slice(SYSTEM_ARGS_COUNT));
}

async function main(): Promise<void> {
  if (process.argv.length <= SYSTEM_ARGS_COUNT) {
    await showHistoryMenu();
    return;
  }

  program
    .version("1.4.0")
    .description(
      "Scans a directory and concatenates file contents into a single output. Run without arguments to see history."
    )
    .argument("[directory]", "The directory to scan", ".")
    .option("-o, --output <file>", "Specify output file path")
    .option("-s, --stdout", "Print to stdout instead of a file")
    .option("--no-gitignore", "Disable .gitignore file parsing")
    .option(
      "--exclude <pattern>",
      "Exclude files by pattern (can be used multiple times)",
      (pattern: string, accumulated: string[]) => accumulated.concat([pattern]),
      []
    )
    .action(run);

  await program.parseAsync(process.argv);
}

main();
