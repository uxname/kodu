import { promises as fs, type Stats } from 'node:fs';
import path from 'node:path';
import { Injectable } from '@nestjs/common';
import ignore from 'ignore';
import { glob } from 'tinyglobby';
import {
  BINARY_EXTENSIONS,
  KNOWN_TEXT_EXTENSIONS,
  MAX_FILE_SIZE_BYTES,
} from '../../shared/constants';
import { ConfigService } from '../config/config.service';
import { UiService } from '../ui/ui.service';

const BINARY_PROBE_SIZE = 8192;
const GLOB_IGNORE = ['.git/**'];

type FindProjectFilesOptions = {
  ignore?: string[];
  useGitignore?: boolean;
  excludeBinary?: boolean;
  contentBasedBinaryDetection?: boolean;
  maxFileSizeBytes?: number;
  rootPaths?: string[];
};

@Injectable()
export class FsService {
  constructor(
    private readonly configService: ConfigService,
    private readonly ui: UiService,
  ) {}

  async findProjectFiles(
    options: FindProjectFilesOptions = {},
  ): Promise<string[]> {
    const { packer } = this.configService.getConfig();
    const shouldUseGitignore = options.useGitignore ?? packer.useGitignore;
    const gitignorePatterns = shouldUseGitignore
      ? await this.readIgnoreFile('.gitignore')
      : [];
    const koduignorePatterns = await this.readIgnoreFile('.koduignore');

    const baseIgnore = options.ignore ?? packer.ignore ?? [];
    const normalizedBase = this.normalizeIgnorePatterns(baseIgnore);
    const combinedIgnore = [
      ...normalizedBase,
      ...gitignorePatterns,
      ...koduignorePatterns,
    ].map((pattern) => pattern.replace(/\\/g, '/'));

    const ig = ignore();
    if (combinedIgnore.length > 0) {
      ig.add(combinedIgnore);
    }

    const globIgnore = this.buildGlobIgnorePatterns(combinedIgnore);

    const patterns = options.rootPaths?.length
      ? options.rootPaths.map((p) => `${p}/**`)
      : ['**/*'];

    const entries = await glob(patterns, {
      onlyFiles: true,
      absolute: true,
      dot: true,
      ignore: [...GLOB_IGNORE, ...globIgnore],
    });

    const relativePaths = entries
      .map((entry) => path.relative(process.cwd(), entry))
      .map((relative) => this.toPosixPath(relative))
      .filter((relative) => relative.length > 0);

    const filtered = ig
      .filter(relativePaths)
      .sort((a, b) => a.localeCompare(b));

    // By default exclude binary files when collecting project files (so pack will skip them).
    // Consumers can override with options.excludeBinary = false.
    const excludeBinary = options.excludeBinary ?? true;
    const useContentDetection =
      options.contentBasedBinaryDetection ??
      packer.contentBasedBinaryDetection ??
      false;
    const maxFileSize = options.maxFileSizeBytes ?? MAX_FILE_SIZE_BYTES;

    const textFiles: string[] = [];

    for (const rel of filtered) {
      const abs = path.resolve(process.cwd(), rel);
      let stats: Stats;

      try {
        stats = await fs.stat(abs);
      } catch {
        continue;
      }

      if (stats.size > maxFileSize) {
        this.ui.log.warn(
          `Skipping large file: ${rel} (>${(maxFileSize / (1024 * 1024)).toFixed(0)}MB)`,
        );
        continue;
      }

      if (
        excludeBinary &&
        (await this.shouldExcludeBinary(rel, abs, useContentDetection))
      ) {
        continue;
      }

      textFiles.push(rel);
    }

    return textFiles;
  }

  async readFileRelative(relativePath: string): Promise<string> {
    const absolute = path.resolve(process.cwd(), relativePath);
    return fs.readFile(absolute, 'utf8');
  }

  private toPosixPath(relativePath: string): string {
    return relativePath.split(path.sep).join(path.posix.sep);
  }

  private normalizeIgnorePatterns(patterns: string[]): string[] {
    return patterns
      .map((pattern) => pattern.trim())
      .filter((pattern) => pattern.length > 0 && !pattern.startsWith('#'));
  }

  private buildGlobIgnorePatterns(patterns: string[]): string[] {
    const normalized = patterns
      .map((pattern) => pattern.trim())
      .filter(
        (pattern) =>
          pattern.length > 0 &&
          !pattern.startsWith('#') &&
          !pattern.startsWith('!'),
      )
      .map((pattern) => pattern.replace(/\\/g, '/'));

    const result = new Set<string>();

    for (const pattern of normalized) {
      const trimmed = pattern.replace(/\/+$/, '');
      result.add(pattern);

      if (trimmed.length === 0) {
        continue;
      }

      if (!pattern.includes('*')) {
        result.add(`${trimmed}/**`);
        result.add(`**/${trimmed}/**`);
      }

      if (!pattern.startsWith('**/')) {
        result.add(`**/${trimmed}`);
      }

      if (pattern.endsWith('/')) {
        result.add(`${trimmed}/**`);
      }
    }

    return [...result];
  }

  private async readIgnoreFile(fileName: string): Promise<string[]> {
    const target = path.join(process.cwd(), fileName);

    try {
      const content = await fs.readFile(target, 'utf8');
      return this.parseIgnoreContent(content);
    } catch {
      return [];
    }
  }

  private parseIgnoreContent(content: string): string[] {
    return content
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter((line) => line.length > 0 && !line.startsWith('#'));
  }

  private isBinaryExtension(relativePath: string): boolean {
    const ext = path.extname(relativePath).toLowerCase();
    return ext.length > 0 && BINARY_EXTENSIONS.has(ext);
  }

  private isKnownTextFile(relativePath: string): boolean {
    const ext = path.extname(relativePath).toLowerCase();
    if (ext && KNOWN_TEXT_EXTENSIONS.has(ext)) {
      return true;
    }

    const baseName = path.basename(relativePath).toLowerCase();
    return KNOWN_TEXT_EXTENSIONS.has(baseName);
  }

  private async shouldExcludeBinary(
    relativePath: string,
    absolutePath: string,
    detectByContent: boolean,
  ): Promise<boolean> {
    if (this.isKnownTextFile(relativePath)) {
      return false;
    }

    if (this.isBinaryExtension(relativePath)) {
      return true;
    }

    if (!detectByContent) {
      return false;
    }

    return this.hasNullByte(absolutePath);
  }

  private async hasNullByte(absolutePath: string): Promise<boolean> {
    let handle: fs.FileHandle | undefined;

    try {
      handle = await fs.open(absolutePath, 'r');
      const buffer = Buffer.alloc(BINARY_PROBE_SIZE);
      const { bytesRead } = await handle.read(buffer, 0, buffer.length, 0);

      for (let i = 0; i < bytesRead; i += 1) {
        if (buffer[i] === 0) {
          return true;
        }
      }

      return false;
    } catch {
      return true;
    } finally {
      if (handle) {
        await handle.close();
      }
    }
  }
}
