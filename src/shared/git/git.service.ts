import { Injectable } from '@nestjs/common';
import { execa } from 'execa';

@Injectable()
export class GitService {
  async ensureRepo(): Promise<void> {
    try {
      await execa('git', ['rev-parse', '--is-inside-work-tree']);
    } catch (error) {
      const message =
        error instanceof Error && 'stderr' in error
          ? String((error as { stderr?: string }).stderr ?? error.message)
          : 'Git repository not found. Initialize git before running the command.';
      throw new Error(message);
    }
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

  async getStagedFiles(): Promise<string[]> {
    await this.ensureRepo();
    const { stdout } = await execa('git', ['diff', '--name-only', '--staged']);
    return stdout
      .split('\n')
      .map((entry) => entry.trim())
      .filter((entry) => entry.length > 0)
      .sort();
  }
}
