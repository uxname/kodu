import { promises as fs } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { Injectable } from '@nestjs/common';
import {
  type ProjectEntry,
  type Registry,
  registrySchema,
} from './registry.schema';

/**
 * Читает и пишет глобальный реестр проектов `~/.config/kodu/registry.json`
 * (учитывает `$XDG_CONFIG_HOME`). Файл создаётся при первой записи — пока
 * проектов нет, ничего в системе не создаётся.
 */
@Injectable()
export class RegistryService {
  private readonly dir = path.join(
    process.env.XDG_CONFIG_HOME?.trim() || path.join(os.homedir(), '.config'),
    'kodu',
  );
  private readonly file = path.join(this.dir, 'registry.json');

  /** Путь к файлу реестра (для подсказок пользователю). */
  getFilePath(): string {
    return this.file;
  }

  /** Загрузить реестр. Если файла ещё нет — вернуть пустой реестр. */
  async load(): Promise<Registry> {
    let raw: unknown;

    try {
      const content = await fs.readFile(this.file, 'utf8');
      raw = JSON.parse(content);
    } catch (error) {
      if ((error as NodeJS.ErrnoException).code === 'ENOENT') {
        return registrySchema.parse({});
      }
      throw new Error(
        `Не удалось прочитать реестр ${this.file}: ${(error as Error).message}`,
      );
    }

    const parsed = registrySchema.safeParse(raw);

    if (!parsed.success) {
      const issues = parsed.error.issues
        .map(
          (issue) => `- ${issue.path.join('.') || '(root)'}: ${issue.message}`,
        )
        .join('\n');
      throw new Error(`Реестр ${this.file} невалиден:\n${issues}`);
    }

    return parsed.data;
  }

  /** Сохранить реестр на диск (создаёт директорию при необходимости). */
  async save(registry: Registry): Promise<void> {
    const validated = registrySchema.parse(registry);
    await fs.mkdir(this.dir, { recursive: true });
    await fs.writeFile(
      this.file,
      `${JSON.stringify(validated, null, 2)}\n`,
      'utf8',
    );
  }

  async list(): Promise<Registry['projects']> {
    return (await this.load()).projects;
  }

  async get(name: string): Promise<ProjectEntry | undefined> {
    return (await this.load()).projects[name];
  }

  async has(name: string): Promise<boolean> {
    return Boolean(await this.get(name));
  }

  /** Добавить проект. По умолчанию запрещает перезапись существующего имени. */
  async add(
    name: string,
    entry: ProjectEntry,
    options: { overwrite?: boolean } = {},
  ): Promise<void> {
    const registry = await this.load();

    if (registry.projects[name] && !options.overwrite) {
      throw new Error(
        `Проект с именем "${name}" уже есть в реестре. Выбери другое имя или обнови существующий проект.`,
      );
    }

    registry.projects[name] = entry;
    await this.save(registry);
  }

  /** Обновить поля существующего проекта. */
  async update(
    name: string,
    patch: Partial<ProjectEntry>,
  ): Promise<ProjectEntry> {
    const registry = await this.load();
    const existing = registry.projects[name];

    if (!existing) {
      throw new Error(`Проект "${name}" не найден в реестре.`);
    }

    const next: ProjectEntry = { ...existing, ...patch };
    registry.projects[name] = next;
    await this.save(registry);
    return next;
  }

  async remove(name: string): Promise<void> {
    const registry = await this.load();

    if (!registry.projects[name]) {
      throw new Error(`Проект "${name}" не найден в реестре.`);
    }

    delete registry.projects[name];
    await this.save(registry);
  }
}
