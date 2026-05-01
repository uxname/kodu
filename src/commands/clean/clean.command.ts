import { createReadStream } from 'node:fs';
import { Command, CommandRunner, Option } from 'nest-commander';
import { ConfigService } from '../../core/config/config.service';
import { FsService } from '../../core/file-system/fs.service';
import { UiService } from '../../core/ui/ui.service';
import { CleanerService } from '../../shared/cleaner/cleaner.service';
import { GitService } from '../../shared/git/git.service';

const SUPPORTED_EXTENSIONS = /\.(ts|tsx|js|jsx|mjs|cjs|html|htm)$/i;

type CleanOptions = {
  dryRun?: boolean;
  changed?: boolean;
  staged?: boolean;
  backup?: boolean;
  noJsdoc?: boolean;
  verbose?: boolean;
  stdin?: boolean;
};

@Command({ name: 'clean', description: 'Remove comments from code' })
export class CleanCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly fsService: FsService,
    private readonly cleaner: CleanerService,
    private readonly config: ConfigService,
    private readonly git: GitService,
  ) {
    super();
  }

  @Option({ flags: '-d, --dry-run', description: 'Show what will be removed' })
  parseDryRun(): boolean {
    return true;
  }

  @Option({
    flags: '-c, --changed',
    description: 'Clean only git-changed files (staged + unstaged + untracked)',
  })
  parseChanged(): boolean {
    return true;
  }

  @Option({
    flags: '-s, --staged',
    description: 'Clean only git-staged files',
  })
  parseStaged(): boolean {
    return true;
  }

  @Option({
    flags: '-b, --backup',
    description: 'Save originals to .kodu/backup/ before modifying',
  })
  parseBackup(): boolean {
    return true;
  }

  @Option({
    flags: '-n, --no-jsdoc',
    description: 'Remove JSDoc comments (overrides config keepJSDoc)',
  })
  parseNoJsdoc(): boolean {
    return true;
  }

  @Option({
    flags: '-v, --verbose',
    description: 'Show all removed comments in dry-run (not just first 3)',
  })
  parseVerbose(): boolean {
    return true;
  }

  @Option({
    flags: '--stdin',
    description: 'Read from stdin, write cleaned result to stdout',
  })
  parseStdin(): boolean {
    return true;
  }

  async run(inputs: string[], options: CleanOptions = {}): Promise<void> {
    if (options.stdin) {
      await this.runStdin(options);
      return;
    }

    const spinner = this.ui
      .createSpinner({ text: this.buildSpinnerText(options) })
      .start();

    try {
      const { cleaner: cleanerConfig, packer } = this.config.getConfig();
      const ignorePatterns = [
        ...(packer.ignore ?? []),
        ...(cleanerConfig.ignore ?? []),
      ];
      const allFiles = await this.fsService.findProjectFiles({
        useGitignore: cleanerConfig.useGitignore,
        ignore: ignorePatterns,
      });

      const targets = await this.collectTargets(allFiles, inputs, options);

      if (targets.length === 0) {
        const msg = this.noFilesMessage(options);
        spinner.stop(msg);
        this.ui.log.warn(msg);
        return;
      }

      const summary = await this.cleaner.cleanFiles(targets, {
        dryRun: options.dryRun,
        backup: options.backup,
        keepJSDoc: options.noJsdoc ? false : undefined,
        onProgress: (current, total) => {
          spinner.text = `${this.buildSpinnerText(options)} (${current}/${total})`;
        },
      });

      spinner.success(
        options.dryRun ? 'Analysis complete' : 'Cleaning complete',
      );

      const bytesSaved = summary.bytesBefore - summary.bytesAfter;
      const tokensSaved = Math.round(bytesSaved / 4);

      if (options.dryRun) {
        this.ui.log.info(
          `Files affected: ${summary.filesChanged}/${summary.filesProcessed}, comments: ${summary.commentsRemoved}`,
        );
        this.ui.log.info(`Bytes saved: ${bytesSaved} (~${tokensSaved} tokens)`);

        const limit = options.verbose ? Number.POSITIVE_INFINITY : 3;
        for (const report of summary.reports.filter((r) => r.removed > 0)) {
          const previews = options.verbose
            ? report.previews
            : report.previews.slice(0, limit);
          const more =
            !options.verbose && report.previews.length > limit
              ? ` +${report.previews.length - limit} more`
              : '';
          this.ui.log.info(
            `  ${report.file} (${report.removed}): ${previews.map((p) => `"${p}"`).join(', ')}${more}`,
          );
        }
        return;
      }

      this.ui.log.success(
        `Files cleaned: ${summary.filesChanged}, comments removed: ${summary.commentsRemoved}`,
      );
      this.ui.log.info(`Bytes saved: ${bytesSaved} (~${tokensSaved} tokens)`);

      if (options.backup && summary.filesChanged > 0) {
        this.ui.log.info('Originals backed up to .kodu/backup/');
      }
    } catch (error) {
      spinner.error('Error during cleaning');
      const message = error instanceof Error ? error.message : 'Unknown error';
      this.ui.log.error(message);
      process.exitCode = 1;
    }
  }

  private async runStdin(options: CleanOptions): Promise<void> {
    try {
      const input = await this.readStdin();
      const cleaned = this.cleaner.cleanContent(
        'stdin.ts',
        input,
        options.noJsdoc ? false : undefined,
      );
      process.stdout.write(cleaned);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      this.ui.log.error(message);
      process.exitCode = 1;
    }
  }

  private readStdin(): Promise<string> {
    return new Promise((resolve, reject) => {
      const chunks: Buffer[] = [];
      const stream = createReadStream('/dev/stdin');
      stream.on('data', (chunk) =>
        chunks.push(typeof chunk === 'string' ? Buffer.from(chunk) : chunk),
      );
      stream.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')));
      stream.on('error', reject);
    });
  }

  private buildSpinnerText(options: CleanOptions): string {
    if (options.staged) return 'Cleaning staged files...';
    if (options.changed) return 'Cleaning changed files...';
    return options.dryRun ? 'Analysing...' : 'Cleaning...';
  }

  private noFilesMessage(options: CleanOptions): string {
    if (options.staged) return 'No staged files to clean.';
    if (options.changed) return 'No changed files to clean.';
    return 'No files to clean.';
  }

  private async collectTargets(
    allFiles: string[],
    inputs: string[],
    options: CleanOptions,
  ): Promise<string[]> {
    const supported = allFiles.filter((f) => SUPPORTED_EXTENSIONS.test(f));

    if (inputs.length > 0) {
      return supported.filter((f) =>
        inputs.some((i) => f === i || f.startsWith(`${i.replace(/\/$/, '')}/`)),
      );
    }

    if (options.staged) {
      const staged = new Set(await this.git.getStagedFiles());
      return supported.filter((f) => staged.has(f));
    }

    if (options.changed) {
      const changed = new Set(await this.git.getChangedFiles());
      return supported.filter((f) => changed.has(f));
    }

    return supported;
  }
}
