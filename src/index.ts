#!/usr/bin/env node

import { resolve } from 'node:path';
import { cwd } from 'node:process';
import { program } from 'commander';
import ora from 'ora';
import pkg from '../package.json';
import { addToHistory, showHistoryMenu } from './cli/history';
import { writeOutput } from './cli/output';
import { runStrip } from './commands/strip';
import { DEFAULT_CONFIG } from './config/defaults';
import { PRIORITY_RULES } from './config/priority';
import { generateOutputContent } from './core/processor';
import { collectFilesRecursively } from './core/scanner';

async function runBundle(args: string[] = []): Promise<void> {
  const spinner = ora('Scanning files...').start();
  try {
    const targetPath = args[0] || '.';
    const rootPath = resolve(cwd(), targetPath);

    const files = collectFilesRecursively(
      rootPath,
      DEFAULT_CONFIG,
      PRIORITY_RULES,
      program.opts().gitignore,
      program.opts().exclude,
    );

    spinner.succeed(`Found ${files.length} files`);

    const { content, skippedFiles } = generateOutputContent(
      files,
      rootPath,
      DEFAULT_CONFIG.MAX_FILE_SIZE_BYTES,
    );

    if (skippedFiles.length > 0) {
      console.log(
        `\nSkipped ${skippedFiles.length} files (too large, binary, or unreadable)`,
      );
    }

    writeOutput(content, rootPath, {
      stdout: program.opts().stdout,
      output: program.opts().output,
    });

    if (!program.opts().noHistory) {
      addToHistory(args);
    }
  } catch (error: unknown) {
    const msg = error instanceof Error ? error.message : String(error);
    spinner.fail(`Error: ${msg}`);
    process.exit(1);
  }
}

async function main() {
  const SYSTEM_ARGS_COUNT = 2;

  program.version(pkg.version).description('koda — Codebase bundler & utility');

  program
    .command('strip [path]')
    .description('Remove all comments from JS/TS files')
    .option('--no-gitignore', 'Disable .gitignore parsing')
    .option('-e, --exclude <patterns...>', 'Exclude patterns', [])
    .option('-y, --yes', 'Skip confirmation prompt')
    .option(
      '-a, --all',
      'Process ALL files (default: only git changed/untracked)',
    )
    .option(
      '-d, --dry-run',
      'Show which files would be processed without modifying them',
    )
    .option('--no-history', 'Disable history')
    .action(async (path, options) => {
      await runStrip(path, {
        gitignore: options.gitignore,
        exclude: options.exclude,
        yes: options.yes,
        dryRun: options.dryRun,
        noHistory: options.history === false,
        all: options.all,
      });
    });

  program
    .arguments('[path]')
    .option('-o, --output <file>', 'Output file')
    .option('--stdout', 'Print output to stdout')
    .option('--no-gitignore', 'Disable .gitignore')
    .option('-e, --exclude <patterns...>', 'Exclude patterns', [])
    .option('--no-history', 'Disable history', false)
    .action(async (path, options) => {
      const args = [path || '.'];
      if (options.output) args.push('--output', options.output);
      if (options.stdout) args.push('--stdout');
      if (options.gitignore === false) args.push('--no-gitignore');
      if (options.exclude && options.exclude.length > 0) {
        options.exclude.forEach((pattern: string) => {
          args.push('--exclude', pattern);
        });
      }
      if (options.noHistory) args.push('--no-history');

      await runBundle(args);
    });

  if (process.argv.length <= SYSTEM_ARGS_COUNT) {
    const historyArgs = await showHistoryMenu();
    if (historyArgs) {
      if (historyArgs[0] === 'strip') {
        const execPath = process.argv[0] ?? 'node';
        const scriptPath = process.argv[1] ?? 'koda';
        await program.parseAsync([execPath, scriptPath, ...historyArgs]);
      } else {
        await runBundle([...historyArgs, '--no-history']);
      }
    }
    return;
  }

  program.parse(process.argv);
}

main().catch(console.error);
