import { writeFile } from 'node:fs/promises';
import clipboard from 'clipboardy';
import { Command, CommandRunner, Option } from 'nest-commander';
import { UiService } from '../../core/ui/ui.service';
import { AiService, type ReviewMode } from '../../shared/ai/ai.service';
import { GitService } from '../../shared/git/git.service';
import { TokenizerService } from '../../shared/tokenizer/tokenizer.service';

type ReviewOptions = {
  mode?: ReviewMode;
  copy?: boolean;
  ci?: boolean;
  output?: string;
};

const DEFAULT_MODE: ReviewMode = 'bug';

@Command({
  name: 'review',
  description: 'AI ревью для застейдженных изменений',
})
export class ReviewCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly git: GitService,
    private readonly tokenizer: TokenizerService,
    private readonly ai: AiService,
  ) {
    super();
  }

  @Option({
    flags: '-m, --mode <mode>',
    description: 'Режим проверки: bug | style | security',
  })
  parseMode(value: string): ReviewMode {
    if (value === 'bug' || value === 'style' || value === 'security') {
      return value;
    }
    return DEFAULT_MODE;
  }

  @Option({ flags: '-c, --copy', description: 'Скопировать результат в буфер' })
  parseCopy(): boolean {
    return true;
  }

  @Option({ flags: '--ci', description: 'CI-режим: без спиннера и без буфера' })
  parseCi(): boolean {
    return true;
  }

  @Option({
    flags: '-o, --output <path>',
    description: 'Сохранить итоговый ревью в файл',
  })
  parseOutput(value: string): string {
    return value;
  }

  async run(_inputs: string[], options: ReviewOptions = {}): Promise<void> {
    const ciMode = Boolean(options.ci);
    const spinner = ciMode
      ? undefined
      : this.ui.createSpinner({ text: 'Собираю diff из git...' }).start();

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
      if (!this.ai.hasApiKey()) {
        const envName = this.ai.getApiKeyEnvName();
        if (spinner) {
          spinner.stop('AI ключ не найден');
        } else {
          this.ui.log.error('AI ключ не найден');
        }
        this.ui.log.warn(`Команда 'review' требует AI ключ для работы.`);
        this.ui.log.info(`Установите ключ: export ${envName}=<ваш_ключ>`);
        this.ui.log.info(
          `Имя переменной окружения настраивается через llm.apiKeyEnv в kodu.json`,
        );
        process.exitCode = 1;
        return;
      }

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

      const tokens = this.tokenizer.count(diff);
      const warningBudget = 12000;
      if (tokens.tokens > warningBudget) {
        this.ui.log.warn(
          `Большой контекст (${tokens.tokens} токенов, ~$${tokens.usdEstimate.toFixed(2)}). Ревью может стоить дороже.`,
        );
      }

      logProgress('Запрос к AI...');
      const mode = options.mode ?? DEFAULT_MODE;
      const result = await this.ai.reviewDiff(diff, mode, false);

      finishProgress('Ревью готово');

      console.log(result.text);
      await this.writeOutput(options.output, result.text, ciMode);

      if (options.copy) {
        await this.copyText(result.text, ciMode);
      }
    } catch (error) {
      if (spinner) {
        spinner.error('Ошибка ревью');
      } else {
        this.ui.log.error('Ошибка ревью');
      }
      const message =
        error instanceof Error ? error.message : 'Неизвестная ошибка';
      this.ui.log.error(message);
      process.exitCode = 1;
    }
  }

  private async writeOutput(
    target: string | undefined,
    payload: string,
    ciMode?: boolean,
  ): Promise<void> {
    if (!target) {
      return;
    }
    await writeFile(target, payload, { encoding: 'utf8' });
    if (!ciMode) {
      this.ui.log.success(`Результат сохранён в ${target}`);
    }
  }

  private async copyText(text: string, ciMode: boolean): Promise<void> {
    if (ciMode) {
      this.ui.log.warn('--copy игнорируется в CI режиме');
      return;
    }
    await clipboard.write(text);
    this.ui.log.success('Результат скопирован в буфер обмена');
  }
}
