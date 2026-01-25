import { Command, CommandRunner, Option } from 'nest-commander';
import { ConfigService } from '../../core/config/config.service';
import { FsService } from '../../core/file-system/fs.service';
import { UiService } from '../../core/ui/ui.service';
import { CleanerService } from '../../shared/cleaner/cleaner.service';
import { GitService } from '../../shared/git/git.service';

type CleanOptions = {
  dryRun?: boolean;
  changed?: boolean;
};

@Command({ name: 'clean', description: 'Удалить комментарии из кода' })
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

  @Option({
    flags: '-d, --dry-run',
    description: 'Показать, что будет удалено',
  })
  parseDryRun(): boolean {
    return true;
  }

  @Option({
    flags: '-c, --changed',
    description: 'Очистить только изменённые файлы',
  })
  parseChanged(): boolean {
    return true;
  }

  async run(_inputs: string[], options: CleanOptions = {}): Promise<void> {
    const spinner = this.ui
      .createSpinner({ text: this.buildSpinnerText(options) })
      .start();

    try {
      const { cleaner: cleanerConfig } = this.config.getConfig();
      const allFiles = await this.fsService.findProjectFiles({
        useGitignore: cleanerConfig.useGitignore,
      });
      const targets = await this.collectTargets(allFiles, options);

      if (targets.length === 0) {
        const noFilesMessage = options.changed
          ? 'Нет изменённых файлов для очистки.'
          : 'Нет файлов для очистки.';
        spinner.stop(noFilesMessage);
        this.ui.log.warn(noFilesMessage);
        return;
      }

      const summary = await this.cleaner.cleanFiles(targets, {
        dryRun: options.dryRun,
      });

      spinner.success(options.dryRun ? 'Анализ завершен' : 'Очистка завершена');

      if (options.dryRun) {
        this.ui.log.info(
          `Будет затронуто файлов: ${summary.filesChanged}, комментариев: ${summary.commentsRemoved}`,
        );
        summary.reports
          .filter((report) => report.removed > 0)
          .forEach((report) => {
            const previews = report.previews
              .map((item) => `"${item}"`)
              .join(', ');
            this.ui.log.info(
              `- ${report.file} (${report.removed}): ${previews}`,
            );
          });
        return;
      }

      this.ui.log.success(
        `Очищено файлов: ${summary.filesChanged}, удалено комментариев: ${summary.commentsRemoved}`,
      );
    } catch (error) {
      spinner.error('Ошибка при очистке');
      const message =
        error instanceof Error ? error.message : 'Неизвестная ошибка';
      this.ui.log.error(message);
      process.exitCode = 1;
    }
  }

  private buildSpinnerText(options: CleanOptions): string {
    const action = options.dryRun ? 'Анализ' : 'Очистка';
    const target = options.changed ? ' изменённых файлов' : ' комментариев';
    return `${action}${target}...`;
  }

  private async collectTargets(
    allFiles: string[],
    options: CleanOptions,
  ): Promise<string[]> {
    const matcher = /\.(ts|tsx|js|jsx|html)$/i;
    const filtered = allFiles.filter((file) => matcher.test(file));

    if (!options.changed) {
      return filtered;
    }

    const changedFiles = await this.git.getChangedFiles();
    const changedSet = new Set(changedFiles);

    return filtered.filter((file) => changedSet.has(file));
  }
}
