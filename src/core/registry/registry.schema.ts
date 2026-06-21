import { z } from 'zod';

/**
 * Стандартный набор стендов. Можно использовать и любые другие имена —
 * стенды хранятся как обычные строки, это лишь значение по умолчанию.
 */
export const DEFAULT_STANDS = ['local', 'dev', 'stage', 'prod'] as const;

/** Один проект в глобальном реестре `~/.config/kodu/registry.json`. */
export const projectEntrySchema = z.object({
  /** Абсолютный путь к репозиторию проекта на этой машине. */
  path: z.string().min(1),
  /** URL репозитория (git clone), необязательно. */
  repo: z.string().optional(),
  /** Доступные стенды проекта. */
  stands: z.array(z.string()).default([...DEFAULT_STANDS]),
});

export type ProjectEntry = z.infer<typeof projectEntrySchema>;

/**
 * Глобальный реестр всех проектов. Ключ объекта `projects` — уникальное имя
 * проекта, по которому агент/CLI понимает, с каким репозиторием работать.
 */
export const registrySchema = z.object({
  $schema: z.string().optional(),
  projects: z.record(z.string(), projectEntrySchema).default({}),
});

export type Registry = z.infer<typeof registrySchema>;

/**
 * Per-project конфиг `.runbook/config.json` (лежит в `.gitignore`).
 * Хранит текущий активный стенд конкретного разработчика.
 */
export const projectConfigSchema = z.object({
  $schema: z.string().optional(),
  /** Имя проекта — должно совпадать с ключом в глобальном реестре. */
  project: z.string().min(1),
  /** Текущий активный стенд. */
  activeStand: z.string().default('local'),
  /** Доступные стенды проекта. */
  stands: z.array(z.string()).default([...DEFAULT_STANDS]),
});

export type ProjectConfig = z.infer<typeof projectConfigSchema>;
