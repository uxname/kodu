import { promises as fs } from 'node:fs';
import path from 'node:path';
import { Command, CommandRunner } from 'nest-commander';
import { type KoduConfig } from '../../core/config/config.schema';
import { UiService } from '../../core/ui/ui.service';

@Command({ name: 'init', description: 'Инициализация конфигурации Kodu' })
export class InitCommand extends CommandRunner {
  constructor(private readonly ui: UiService) {
    super();
  }

  async run(): Promise<void> {
    const configPath = path.join(process.cwd(), 'kodu.json');

    const defaultLlmConfig = {
      provider: 'openai' as const,
      model: 'gpt-4o',
      apiKeyEnv: 'OPENAI_API_KEY',
    };

    const defaultConfig: KoduConfig = {
      $schema: 'https://uxna.me/kodu/schema.json',
      llm: defaultLlmConfig,
      cleaner: { whitelist: ['//!'], keepJSDoc: true },
      packer: {
        ignore: [
          'package-lock.json',
          'yarn.lock',
          'pnpm-lock.yaml',
          '.git',
          '.kodu',
          'node_modules',
          'dist',
          'coverage',
        ],
      },
    };

    const useAi = await this.ui.promptConfirm({
      message: 'Будете использовать AI функции?',
      default: true,
    });

    let llmConfig: KoduConfig['llm'] | undefined;
    if (useAi) {
      const provider = await this.ui.promptSelect<'openai'>(
        this.buildProviderQuestion(defaultLlmConfig.provider),
      );
      llmConfig = {
        provider,
        model: defaultLlmConfig.model,
        apiKeyEnv: defaultLlmConfig.apiKeyEnv,
      };
    }

    const extendIgnore = await this.ui.promptConfirm({
      message: 'Изменить стандартный ignore-список?',
      default: false,
    });

    const ignoreList = extendIgnore
      ? await this.askIgnoreList(defaultConfig.packer.ignore)
      : defaultConfig.packer.ignore;

    const additionalWhitelist = await this.ui.promptInput({
      message:
        'Дополнительные префиксы для whitelist (через запятую, пусто — оставить дефолт):',
      default: '',
    });

    const whitelist = this.mergeWhitelist(
      defaultConfig.cleaner.whitelist,
      additionalWhitelist,
    );

    const configToSave: KoduConfig = {
      $schema: defaultConfig.$schema,
      ...(llmConfig && { llm: llmConfig }),
      cleaner: { whitelist, keepJSDoc: defaultConfig.cleaner.keepJSDoc },
      packer: { ignore: ignoreList },
    };

    await this.writeConfig(configPath, configToSave);
    await this.ensureKoduFolders();
    await this.ensureGitignore();

    this.ui.log.success('Конфигурация Kodu создана.');
    if (useAi) {
      this.ui.log.info(
        '🎉 Kodu initialized! Запустите `kodu pack`, чтобы продолжить.',
      );
    } else {
      this.ui.log.info('🎉 Kodu initialized! Доступны команды: pack, clean.');
      this.ui.log.info(
        'Для использования AI функций (review, commit) добавьте секцию llm в kodu.json.',
      );
    }
  }

  private buildProviderQuestion(defaultProvider: 'openai') {
    return {
      message: 'Выберите AI-провайдера',
      choices: [{ name: 'OpenAI', value: 'openai' as const }],
      default: defaultProvider,
    };
  }

  private async askIgnoreList(defaultIgnore: string[]): Promise<string[]> {
    const answer = await this.ui.promptInput({
      message: 'Укажите ignore-паттерны через запятую',
      default: defaultIgnore.join(', '),
    });

    return answer
      .split(',')
      .map((item) => item.trim())
      .filter((item) => item.length > 0);
  }

  private mergeWhitelist(defaultWhitelist: string[], extra: string): string[] {
    if (!extra.trim()) {
      return defaultWhitelist;
    }

    const additions = extra
      .split(',')
      .map((item) => item.trim())
      .filter((item) => item.length > 0);

    return Array.from(new Set([...defaultWhitelist, ...additions]));
  }

  private async writeConfig(
    configPath: string,
    config: KoduConfig,
  ): Promise<void> {
    if (await this.fileExists(configPath)) {
      const overwrite = await this.ui.promptConfirm({
        message: 'kodu.json уже существует. Перезаписать?',
        default: false,
      });

      if (!overwrite) {
        this.ui.log.warn('Инициализация отменена: файл kodu.json уже есть.');
        return;
      }
    }

    await fs.writeFile(
      configPath,
      `${JSON.stringify(config, null, 2)}\n`,
      'utf8',
    );
    this.ui.log.success(`Сохранен ${configPath}`);
  }

  private async ensureKoduFolders(): Promise<void> {
    const koduDir = path.join(process.cwd(), '.kodu');
    const promptsDir = path.join(koduDir, 'prompts');

    await fs.mkdir(promptsDir, { recursive: true });

    const keepFile = path.join(promptsDir, '.keep');
    if (!(await this.fileExists(keepFile))) {
      await fs.writeFile(keepFile, '');
    }
  }

  private async ensureGitignore(): Promise<void> {
    const gitignorePath = path.join(process.cwd(), '.gitignore');
    const content = (await this.fileExists(gitignorePath))
      ? await fs.readFile(gitignorePath, 'utf8')
      : '';

    const lines = content.split(/\r?\n/);
    const additions: string[] = [];

    if (
      !lines.some((line) => line.trim() === '.kodu' || line.trim() === '.kodu/')
    ) {
      additions.push('.kodu/');
    }

    if (!lines.some((line) => line.trim() === '.env')) {
      const addEnv = await this.ui.promptConfirm({
        message: 'В .gitignore нет .env. Добавить?',
        default: true,
      });

      if (addEnv) {
        additions.push('.env');
      }
    }

    if (additions.length === 0) {
      return;
    }

    const trimmed = content.trimEnd();
    const next =
      trimmed.length > 0
        ? `${trimmed}\n${additions.join('\n')}`
        : additions.join('\n');
    await fs.writeFile(gitignorePath, `${next}\n`, 'utf8');
    this.ui.log.success('Обновлен .gitignore');
  }

  private async fileExists(targetPath: string): Promise<boolean> {
    try {
      await fs.access(targetPath);
      return true;
    } catch {
      return false;
    }
  }
}
