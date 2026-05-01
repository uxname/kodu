import { promises as fs } from 'node:fs';
import path from 'node:path';
import { Command, CommandRunner } from 'nest-commander';
import { UiService } from '../../core/ui/ui.service';

const GITIGNORE_ENTRY = '.kodu/context.txt';

@Command({ name: 'init', description: 'Add kodu output to .gitignore' })
export class InitCommand extends CommandRunner {
  constructor(private readonly ui: UiService) {
    super();
  }

  async run(): Promise<void> {
    await this.updateGitignore();
    this.ui.log.success('Done.');
  }

  private async updateGitignore(): Promise<void> {
    const gitignorePath = path.join(process.cwd(), '.gitignore');

    if (!(await this.exists(gitignorePath))) {
      this.ui.log.warn('.gitignore not found, skipping.');
      return;
    }

    const content = await fs.readFile(gitignorePath, 'utf8');
    const lines = content.split(/\r?\n/);

    if (lines.some((line) => line.trim() === GITIGNORE_ENTRY)) {
      this.ui.log.info(`${GITIGNORE_ENTRY} already in .gitignore`);
      return;
    }

    const trimmed = content.trimEnd();
    const next =
      trimmed.length > 0 ? `${trimmed}\n${GITIGNORE_ENTRY}` : GITIGNORE_ENTRY;
    await fs.writeFile(gitignorePath, `${next}\n`, 'utf8');
    this.ui.log.success(`Added ${GITIGNORE_ENTRY} to .gitignore`);
  }

  private async exists(target: string): Promise<boolean> {
    try {
      await fs.access(target);
      return true;
    } catch {
      return false;
    }
  }
}
