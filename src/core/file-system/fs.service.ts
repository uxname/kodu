import { promises as fs } from 'node:fs';
import path from 'node:path';
import { Injectable } from '@nestjs/common';
import { glob } from 'tinyglobby';
import { ConfigService } from '../config/config.service';

@Injectable()
export class FsService {
  constructor(private readonly configService: ConfigService) {}

  async findProjectFiles(): Promise<string[]> {
    const { packer } = this.configService.getConfig();
    const gitignore = await this.readGitignorePatterns();
    const entries = await glob(['**/*'], {
      onlyFiles: true,
      absolute: true,
      ignore: [...packer.ignore, ...gitignore],
    });

    return entries
      .map((entry) => path.relative(process.cwd(), entry))
      .filter((relative) => relative.length > 0)
      .sort((a, b) => a.localeCompare(b));
  }

  async readFileRelative(relativePath: string): Promise<string> {
    const absolute = path.resolve(process.cwd(), relativePath);
    return fs.readFile(absolute, 'utf8');
  }

  private async readGitignorePatterns(): Promise<string[]> {
    const gitignorePath = path.join(process.cwd(), '.gitignore');

    try {
      const content = await fs.readFile(gitignorePath, 'utf8');
      return content
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter((line) => line.length > 0 && !line.startsWith('#'));
    } catch {
      return [];
    }
  }
}
