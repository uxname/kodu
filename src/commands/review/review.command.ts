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
  description: 'AI review for staged changes',
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
    description: 'Review mode: bug | style | security | <custom-mode>',
  })
  parseMode(value: string): ReviewMode {
    const availableModes = this.ai.getAvailableReviewModes();

    if (availableModes.includes(value)) {
      return value;
    }

    this.ui.log.warn(
      `Mode "${value}" not found. Available modes: ${availableModes.join(', ')}. Using default mode: ${DEFAULT_MODE}`,
    );
    return DEFAULT_MODE;
  }

  @Option({ flags: '-c, --copy', description: 'Copy result to clipboard' })
  parseCopy(): boolean {
    return true;
  }

  @Option({
    flags: '--ci',
    description: 'CI mode: no spinner and no buffering',
  })
  parseCi(): boolean {
    return true;
  }

  @Option({
    flags: '-o, --output <path>',
    description: 'Save final review to file',
  })
  parseOutput(value: string): string {
    return value;
  }

  async run(_inputs: string[], options: ReviewOptions = {}): Promise<void> {
    const ciMode = Boolean(options.ci);
    const spinner = ciMode
      ? undefined
      : this.ui.createSpinner({ text: 'Collecting diff from git...' }).start();

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
          spinner.stop('AI key not found');
        } else {
          this.ui.log.error('AI key not found');
        }
        this.ui.log.warn(`'review' command requires AI key to work.`);
        this.ui.log.info(`Set key: export ${envName}=<your_key>`);
        this.ui.log.info(
          `Environment variable name is configured via llm.apiKeyEnv in kodu.json`,
        );
        process.exitCode = 1;
        return;
      }

      await this.git.ensureRepo();

      const hasStaged = await this.git.hasStagedChanges();
      if (!hasStaged) {
        if (spinner) {
          spinner.stop('No staged changes');
        } else {
          this.ui.log.info('No staged changes');
        }
        this.ui.log.warn('First run git add for the required files.');
        return;
      }

      const diff = await this.git.getStagedDiff();
      if (!diff.trim()) {
        if (spinner) {
          spinner.stop(
            'Diff is empty - possibly everything excluded by packer.ignore',
          );
        } else {
          this.ui.log.info(
            'Diff is empty - possibly everything excluded by packer.ignore',
          );
        }
        this.ui.log.warn(
          'Diff is empty: all changes fell into packer.ignore exclusions.',
        );
        return;
      }

      const tokens = this.tokenizer.count(diff);
      const warningBudget = 12000;
      if (tokens.tokens > warningBudget) {
        this.ui.log.warn(
          `Large context (${tokens.tokens} tokens, ~$${tokens.usdEstimate.toFixed(2)}). Review may cost more.`,
        );
      }

      logProgress('Requesting AI...');
      const mode = options.mode ?? DEFAULT_MODE;
      const result = await this.ai.reviewDiff(diff, mode);

      finishProgress('Review ready');

      console.log(result.text);
      await this.writeOutput(options.output, result.text, ciMode);

      if (options.copy) {
        await this.copyText(result.text, ciMode);
      }
    } catch (error) {
      if (spinner) {
        spinner.error('Review error');
      } else {
        this.ui.log.error('Review error');
      }
      const message = error instanceof Error ? error.message : 'Unknown error';
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
      this.ui.log.success(`Result saved to ${target}`);
    }
  }

  private async copyText(text: string, ciMode: boolean): Promise<void> {
    if (ciMode) {
      this.ui.log.warn('--copy ignored in CI mode');
      return;
    }
    await clipboard.write(text);
    this.ui.log.success('Result copied to clipboard');
  }
}
