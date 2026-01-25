import { Injectable } from '@nestjs/common';
import { execa } from 'execa';
import { ConfigService } from '../../core/config/config.service';

const EXCLUDE_PREFIX = ':(exclude)';

@Injectable()
export class GitService {
  constructor(private readonly configService: ConfigService) {}

  async ensureRepo(): Promise<void> {
    try {
      await execa('git', ['rev-parse', '--is-inside-work-tree']);
    } catch (error) {
      const message =
        error instanceof Error && 'stdout' in error
          ? String((error as { stderr?: string }).stderr ?? error.message)
          : 'Git репозиторий не найден. Инициализируйте git перед выполнением команды.';
      throw new Error(message);
    }
  }

  async hasStagedChanges(): Promise<boolean> {
    const { stdout } = await execa('git', ['diff', '--staged', '--name-only']);
    return stdout.trim().length > 0;
  }

  async getStagedDiff(): Promise<string> {
    await this.ensureRepo();
    const excludeArgs = this.buildExcludeArgs();
    const args = ['diff', '--staged', '--unified=3', '--', '.', ...excludeArgs];
    const { stdout } = await execa('git', args);
    return stdout;
  }

  async getChangedFiles(): Promise<string[]> {
    await this.ensureRepo();
    const changed = new Set<string>();
    const load = async (args: string[]) => {
      const { stdout } = await execa('git', args);
      stdout
        .split('\n')
        .map((entry) => entry.trim())
        .filter((entry) => entry.length > 0)
        .forEach((entry) => {
          changed.add(entry);
        });
    };

    await load(['diff', '--name-only']);
    await load(['diff', '--name-only', '--staged']);
    await load(['ls-files', '--others', '--exclude-standard']);
    return [...changed].sort();
  }

  async getStatusShort(): Promise<string> {
    const { stdout } = await execa('git', ['status', '--short']);
    return stdout.trim();
  }

  async commit(message: string): Promise<void> {
    await execa('git', ['commit', '-m', message]);
  }

  private buildExcludeArgs(): string[] {
    const ignore = this.configService.getConfig().packer.ignore ?? [];
    return ignore.map((pattern) => `${EXCLUDE_PREFIX}${pattern}`);
  }
}
