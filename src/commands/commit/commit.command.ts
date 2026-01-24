import { writeFile } from 'node:fs/promises';
import { Command, CommandRunner, Option } from 'nest-commander';
import { UiService } from '../../core/ui/ui.service';
import { AiService } from '../../shared/ai/ai.service';
import { GitService } from '../../shared/git/git.service';

type CommitOptions = {
  ci?: boolean;
  output?: string;
};

@Command({
  name: 'commit',
  description: 'Сгенерировать и применить сообщение коммита',
})
export class CommitCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly git: GitService,
    private readonly ai: AiService,
  ) {
    super();
  }

  @Option({ flags: '--ci', description: 'CI-режим: без спиннеров и диалогов' })
  parseCi(): boolean {
    return true;
  }

  @Option({
    flags: '-o, --output <path>',
    description: 'Сохранить сообщение в файл',
  })
  parseOutput(value: string): string {
    return value;
  }

  async run(_inputs: string[], options: CommitOptions = {}): Promise<void> {
    const ciMode = Boolean(options.ci);
    const spinner = ciMode
      ? undefined
      : this.ui.createSpinner({ text: 'Собираю diff...' }).start();

    const logProgress = (text: string): void => {
      if (ciMode) {
        return;
      }
      if (spinner) {
        spinner.text = text;
        return;
      }
      this.ui.log.info(text);
    };

    const finishProgress = (text: string): void => {
      if (ciMode) {
        return;
      }
      if (spinner) {
        spinner.success(text);
        return;
      }
      this.ui.log.success(text);
    };

    try {
      await this.git.ensureRepo();

      const hasStaged = await this.git.hasStagedChanges();
      if (!hasStaged) {
        if (spinner) {
          spinner.stop('Нет застейдженных изменений');
        } else {
          this.ui.log.info('Нет застейдженных изменений');
        }
        this.ui.log.warn('Сначала выполните git add для нужных файлов.');
        return;
      }

      const diff = await this.git.getStagedDiff();
      if (!diff.trim()) {
        if (spinner) {
          spinner.stop('Diff пуст — возможно, всё исключено packer.ignore');
        } else {
          this.ui.log.info('Diff пуст — возможно, всё исключено packer.ignore');
        }
        this.ui.log.warn(
          'Diff пустой: все изменения попали в исключения packer.ignore.',
        );
        return;
      }

      logProgress('Генерирую сообщение коммита...');
      const commitMessage = await this.ai.generateCommitMessage(diff);

      finishProgress('Сообщение готово');
      if (!ciMode) {
        this.ui.log.info(`Предложение: ${commitMessage}`);
      }

      console.log(commitMessage);
      if (options.output) {
        await writeFile(options.output, commitMessage, { encoding: 'utf8' });
        if (!ciMode) {
          this.ui.log.success(`Сообщение сохранено в ${options.output}`);
        }
      }
    } catch (error) {
      if (spinner) {
        spinner.error('Ошибка при создании коммита');
      } else {
        this.ui.log.error('Ошибка при создании коммита');
      }
      const message =
        error instanceof Error ? error.message : 'Неизвестная ошибка';
      this.ui.log.error(message);
      process.exitCode = 1;
    }
  }
}
