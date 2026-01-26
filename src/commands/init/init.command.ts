import { promises as fs } from 'node:fs';
import path from 'node:path';
import { Command, CommandRunner } from 'nest-commander';
import { type KoduConfig } from '../../core/config/config.schema';
import {
  DEFAULT_COMMIT_PROMPT,
  DEFAULT_PACK_PROMPT,
  DEFAULT_REVIEW_PROMPTS,
} from '../../core/config/default-prompts';
import { UiService } from '../../core/ui/ui.service';

const buildDefaultCommandSettings = () => ({
  commit: { modelSettings: { maxOutputTokens: 150 } },
  review: { modelSettings: { maxOutputTokens: 5000 } },
});

@Command({ name: 'init', description: 'Initialize Kodu configuration' })
export class InitCommand extends CommandRunner {
  constructor(private readonly ui: UiService) {
    super();
  }

  async run(): Promise<void> {
    const configPath = path.join(process.cwd(), 'kodu.json');

    const defaultLlmConfig = {
      model: 'openai/gpt-5-mini',
      apiKeyEnv: 'OPENAI_API_KEY',
    };

    const defaultConfig: KoduConfig = {
      $schema:
        'https://raw.githubusercontent.com/uxname/kodu/refs/heads/master/kodu.schema.json',
      llm: defaultLlmConfig,
      cleaner: { whitelist: ['//!'], keepJSDoc: true, useGitignore: true },
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
        useGitignore: true,
      },
    };

    const useAi = await this.ui.promptConfirm({
      message: 'Will you use AI functions?',
      default: true,
    });

    let llmConfig: KoduConfig['llm'] | undefined;
    if (useAi) {
      const useCustomModel = await this.ui.promptConfirm({
        message: 'Use your own model?',
        default: false,
      });

      let model: string;
      if (useCustomModel) {
        model = await this.ui.promptInput({
          message:
            'Enter model in format provider/model-name (e.g., openai/gpt-5-mini):',
          default: defaultLlmConfig.model,
          validate: (input) => {
            if (!input.includes('/')) {
              return 'Model must be in format provider/model-name';
            }
            return true;
          },
        });
      } else {
        model = await this.ui.promptSelect<string>(
          this.buildModelQuestion(defaultLlmConfig.model),
        );
      }

      llmConfig = {
        model,
        apiKeyEnv: defaultLlmConfig.apiKeyEnv,
        commands: buildDefaultCommandSettings(),
      };
    }

    const extendIgnore = await this.ui.promptConfirm({
      message: 'Modify standard ignore list?',
      default: false,
    });

    const ignoreList = extendIgnore
      ? await this.askIgnoreList(defaultConfig.packer.ignore)
      : defaultConfig.packer.ignore;

    const additionalWhitelist = await this.ui.promptInput({
      message:
        'Additional whitelist prefixes (comma-separated, empty - keep default):',
      default: '',
    });

    const whitelist = this.mergeWhitelist(
      defaultConfig.cleaner.whitelist,
      additionalWhitelist,
    );

    const promptPaths = this.buildPromptPaths();

    const configToSave: KoduConfig = {
      $schema: defaultConfig.$schema,
      ...(llmConfig && { llm: llmConfig }),
      cleaner: {
        whitelist,
        keepJSDoc: defaultConfig.cleaner.keepJSDoc,
        useGitignore: defaultConfig.cleaner.useGitignore,
      },
      packer: {
        ignore: ignoreList,
        useGitignore: defaultConfig.packer.useGitignore,
      },
      prompts: {
        review: {
          bug: promptPaths.review.bug,
          style: promptPaths.review.style,
          security: promptPaths.review.security,
        },
        commit: promptPaths.commit,
        pack: promptPaths.pack,
      },
    };

    await this.writeConfig(configPath, configToSave);
    await this.ensurePromptFiles(promptPaths);
    await this.ensureGitignore();

    this.ui.log.success('Kodu configuration created.');
    if (useAi) {
      this.ui.log.info('🎉 Kodu initialized! Run `kodu pack` to continue.');
    } else {
      this.ui.log.info('🎉 Kodu initialized! Available commands: pack, clean.');
      this.ui.log.info(
        'To use AI functions (review, commit) add llm section to kodu.json.',
      );
    }
  }

  private buildModelQuestion(defaultModel: string) {
    return {
      message: 'Select AI model',
      choices: [
        {
          name: 'OpenAI GPT-5 Mini (recommended)',
          value: 'openai/gpt-5-mini',
        },
        { name: 'OpenAI GPT-4o Mini', value: 'openai/gpt-4o-mini' },
        { name: 'OpenAI GPT-4o', value: 'openai/gpt-4o' },
        {
          name: 'Anthropic Claude 3.5 Sonnet',
          value: 'anthropic/claude-3-5-sonnet-20241022',
        },
        { name: 'Google Gemini 2.5 Flash', value: 'google/gemini-2.5-flash' },
      ],
      default: defaultModel,
    };
  }

  private async askIgnoreList(defaultIgnore: string[]): Promise<string[]> {
    const answer = await this.ui.promptInput({
      message: 'Specify ignore patterns (comma-separated)',
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
        message: 'kodu.json already exists. Overwrite?',
        default: false,
      });

      if (!overwrite) {
        this.ui.log.warn(
          'Initialization cancelled: kodu.json file already exists.',
        );
        return;
      }
    }

    await fs.writeFile(
      configPath,
      `${JSON.stringify(config, null, 2)}\n`,
      'utf8',
    );
    this.ui.log.success(`Saved ${configPath}`);
  }

  private async ensurePromptFiles(
    paths: ReturnType<InitCommand['buildPromptPaths']>,
  ): Promise<void> {
    const promptDir = path.join(process.cwd(), '.kodu', 'prompts');
    await fs.mkdir(promptDir, { recursive: true });

    const keepFile = path.join(promptDir, '.keep');
    if (!(await this.fileExists(keepFile))) {
      await fs.writeFile(keepFile, '');
    }

    await Promise.all([
      this.writePromptIfMissing(paths.review.bug, DEFAULT_REVIEW_PROMPTS.bug),
      this.writePromptIfMissing(
        paths.review.style,
        DEFAULT_REVIEW_PROMPTS.style,
      ),
      this.writePromptIfMissing(
        paths.review.security,
        DEFAULT_REVIEW_PROMPTS.security,
      ),
      this.writePromptIfMissing(paths.commit, DEFAULT_COMMIT_PROMPT),
      this.writePromptIfMissing(paths.pack, DEFAULT_PACK_PROMPT),
    ]);
  }

  private buildPromptPaths() {
    return {
      review: {
        bug: path.posix.join('.kodu', 'prompts', 'review-bug.md'),
        style: path.posix.join('.kodu', 'prompts', 'review-style.md'),
        security: path.posix.join('.kodu', 'prompts', 'review-security.md'),
      },
      commit: path.posix.join('.kodu', 'prompts', 'commit.md'),
      pack: path.posix.join('.kodu', 'prompts', 'pack.md'),
    } as const;
  }

  private async writePromptIfMissing(
    target: string,
    content: string,
  ): Promise<void> {
    const absolute = path.isAbsolute(target)
      ? target
      : path.join(process.cwd(), target);

    if (await this.fileExists(absolute)) {
      return;
    }

    await fs.mkdir(path.dirname(absolute), { recursive: true });
    await fs.writeFile(absolute, `${content}\n`, 'utf8');
  }

  private async ensureGitignore(): Promise<void> {
    const gitignorePath = path.join(process.cwd(), '.gitignore');
    const content = (await this.fileExists(gitignorePath))
      ? await fs.readFile(gitignorePath, 'utf8')
      : '';

    const lines = content.split(/\r?\n/);
    const additions: string[] = [];

    if (!lines.some((line) => line.trim() === '.env')) {
      const addEnv = await this.ui.promptConfirm({
        message: '.env not in .gitignore. Add it?',
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
    this.ui.log.success('Updated .gitignore');
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
