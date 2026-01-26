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
  description: 'Generate and apply commit message',
})
export class CommitCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly git: GitService,
    private readonly ai: AiService,
  ) {
    super();
  }

  @Option({ flags: '--ci', description: 'CI mode: no spinners and dialogs' })
  parseCi(): boolean {
    return true;
  }

  @Option({
    flags: '-o, --output <path>',
    description: 'Save message to file',
  })
  parseOutput(value: string): string {
    return value;
  }

  async run(_inputs: string[], options: CommitOptions = {}): Promise<void> {
    const ciMode = Boolean(options.ci);
    const spinner = ciMode
      ? undefined
      : this.ui.createSpinner({ text: 'Collecting diff...' }).start();

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
        this.ui.log.warn(`'commit' command requires AI key to work.`);
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

      logProgress('Generating commit message...');
      const commitMessage = await this.ai.generateCommitMessage(diff);

      finishProgress('Message ready');
      if (!ciMode) {
        this.ui.log.info(`Suggestion: ${commitMessage}`);
      }

      console.log(commitMessage);
      if (options.output) {
        await writeFile(options.output, commitMessage, { encoding: 'utf8' });
        if (!ciMode) {
          this.ui.log.success(`Message saved to ${options.output}`);
        }
      }
    } catch (error) {
      if (spinner) {
        spinner.error('Error creating commit');
      } else {
        this.ui.log.error('Error creating commit');
      }
      const message = error instanceof Error ? error.message : 'Unknown error';
      this.ui.log.error(message);
      process.exitCode = 1;
    }
  }
}
