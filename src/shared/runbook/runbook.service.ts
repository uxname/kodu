import { promises as fs } from 'node:fs';
import path from 'node:path';
import { Injectable } from '@nestjs/common';
import {
  type ProjectConfig,
  projectConfigSchema,
} from '../../core/registry/registry.schema';
import { type DetectedStack, renderRunbook } from './runbook.templates';

const RUNBOOK_DIR = '.runbook';
const GITIGNORE_ENTRY = '/.runbook/';
/** Эквивалентные записи `.runbook/` в `.gitignore` — считаем дублем. */
const GITIGNORE_EQUIVALENTS = new Set([
  '/.runbook/',
  '/.runbook',
  '.runbook/',
  '.runbook',
]);

/** Результат настройки `.gitignore`. */
export type GitignoreResult = 'created' | 'added' | 'present' | 'no-git';

/**
 * Работа с per-project директорией `.runbook/`: скаффолд, чтение/запись
 * `config.json`, чтение `runbook.md` и идемпотентная настройка `.gitignore`.
 */
@Injectable()
export class RunbookService {
  dirPath(root: string = process.cwd()): string {
    return path.join(root, RUNBOOK_DIR);
  }

  configPath(root: string = process.cwd()): string {
    return path.join(root, RUNBOOK_DIR, 'config.json');
  }

  runbookPath(root: string = process.cwd()): string {
    return path.join(root, RUNBOOK_DIR, 'runbook.md');
  }

  /** Инициализирован ли проект (есть `.runbook/config.json`). */
  async exists(root: string = process.cwd()): Promise<boolean> {
    return this.pathExists(this.configPath(root));
  }

  async readConfig(root: string = process.cwd()): Promise<ProjectConfig> {
    const content = await fs.readFile(this.configPath(root), 'utf8');
    return projectConfigSchema.parse(JSON.parse(content));
  }

  async writeConfig(
    config: ProjectConfig,
    root: string = process.cwd(),
  ): Promise<void> {
    const validated = projectConfigSchema.parse(config);
    await fs.mkdir(this.dirPath(root), { recursive: true });
    await fs.writeFile(
      this.configPath(root),
      `${JSON.stringify(validated, null, 2)}\n`,
      'utf8',
    );
  }

  async readRunbook(root: string = process.cwd()): Promise<string> {
    return fs.readFile(this.runbookPath(root), 'utf8');
  }

  /**
   * Создаёт `.runbook/` с config.json и (если ещё нет) runbook.md.
   * Существующий runbook.md не перезаписывается — ручные правки сохраняются.
   */
  async scaffold(
    options: { project: string; stands: string[]; activeStand: string },
    root: string = process.cwd(),
  ): Promise<void> {
    await fs.mkdir(this.dirPath(root), { recursive: true });

    await this.writeConfig(
      {
        project: options.project,
        activeStand: options.activeStand,
        stands: options.stands,
      },
      root,
    );

    if (!(await this.pathExists(this.runbookPath(root)))) {
      const detected = await this.detectStack(root);
      const markdown = renderRunbook(options.project, options.stands, detected);
      await fs.writeFile(this.runbookPath(root), markdown, 'utf8');
    }
  }

  /** Определяет наличие docker-compose / Dockerfile / .env.example. */
  async detectStack(root: string = process.cwd()): Promise<DetectedStack> {
    const [composeYml, composeYaml, dockerfile, envExample] = await Promise.all(
      [
        this.pathExists(path.join(root, 'docker-compose.yml')),
        this.pathExists(path.join(root, 'docker-compose.yaml')),
        this.pathExists(path.join(root, 'Dockerfile')),
        this.pathExists(path.join(root, '.env.example')),
      ],
    );

    return {
      compose: composeYml || composeYaml,
      dockerfile,
      envExample,
    };
  }

  /**
   * Гарантирует, что `.runbook/` игнорируется git — автоматически и идемпотентно.
   * - не git-репозиторий → ничего не пишем, возвращаем 'no-git';
   * - запись уже есть → 'present';
   * - `.gitignore` отсутствует → создаём с записью → 'created';
   * - иначе дописываем в конец → 'added'.
   */
  async ensureGitignore(
    root: string = process.cwd(),
  ): Promise<GitignoreResult> {
    if (!(await this.pathExists(path.join(root, '.git')))) {
      return 'no-git';
    }

    const gitignorePath = path.join(root, '.gitignore');
    let content = '';
    let fileExists = true;

    try {
      content = await fs.readFile(gitignorePath, 'utf8');
    } catch {
      fileExists = false;
    }

    const alreadyIgnored = content
      .split(/\r?\n/)
      .map((line) => line.trim())
      .some((line) => GITIGNORE_EQUIVALENTS.has(line));

    if (alreadyIgnored) {
      return 'present';
    }

    const trimmed = content.trimEnd();
    const next =
      trimmed.length > 0
        ? `${trimmed}\n${GITIGNORE_ENTRY}\n`
        : `${GITIGNORE_ENTRY}\n`;

    await fs.writeFile(gitignorePath, next, 'utf8');

    return fileExists ? 'added' : 'created';
  }

  private async pathExists(target: string): Promise<boolean> {
    try {
      await fs.access(target);
      return true;
    } catch {
      return false;
    }
  }
}
