import { promises as fs } from 'node:fs';
import path from 'node:path';
import { Injectable } from '@nestjs/common';
import { replacePromptVariables } from './default-prompts';

type Variables = Record<string, string | number>;

@Injectable()
export class PromptService {
  private readonly cache = new Map<string, string>();

  async load(source: string, variables?: Variables): Promise<string> {
    const raw = await this.readSource(source);
    return variables
      ? replacePromptVariables(raw, this.normalize(variables))
      : raw;
  }

  async loadFromPromptsDir(name: string): Promise<string> {
    const candidates = this.buildCandidates(name);

    for (const candidate of candidates) {
      if (await this.exists(candidate)) {
        return this.readAndCache(candidate);
      }
    }

    throw new Error(
      `Template ${name} not found. Expected files: ${candidates
        .map((c) => path.relative(process.cwd(), c))
        .join(', ')}`,
    );
  }

  private async readSource(source: string): Promise<string> {
    const resolved = path.isAbsolute(source)
      ? source
      : path.resolve(process.cwd(), source);

    if (await this.exists(resolved)) {
      return this.readAndCache(resolved);
    }

    if (this.looksLikeInline(source)) {
      return source;
    }

    throw new Error(
      `Prompt file not found: ${path.relative(process.cwd(), resolved)}`,
    );
  }

  private async readAndCache(target: string): Promise<string> {
    const cached = this.cache.get(target);
    if (cached) {
      return cached;
    }
    const content = await fs.readFile(target, 'utf8');
    this.cache.set(target, content);
    return content;
  }

  private looksLikeInline(value: string): boolean {
    if (value.includes('\n')) {
      return true;
    }
    const hasPathSegments = value.includes('/') || value.includes('\\');
    const hasExtension = path.extname(value) !== '';

    return value.trim().length > 0 && !hasPathSegments && !hasExtension;
  }

  private buildCandidates(name: string): string[] {
    const names = path.extname(name) ? [name] : [`${name}.md`, `${name}.txt`];

    const roots = [path.join(process.cwd(), '.kodu', 'prompts')];

    const candidates: string[] = [];
    for (const root of roots) {
      for (const variant of names) {
        candidates.push(path.join(root, variant));
      }
    }

    return candidates;
  }

  private normalize(variables: Variables): Record<string, string> {
    return Object.fromEntries(
      Object.entries(variables).map(([key, value]) => [key, value.toString()]),
    );
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
