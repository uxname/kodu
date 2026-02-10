import { z } from 'zod';
import {
  DEFAULT_COMMIT_TOKENS,
  DEFAULT_LLM_MODEL,
  DEFAULT_REVIEW_TOKENS,
} from '../../shared/constants';

// Model ID format: provider/model-name (e.g., "openai/gpt-4o", "anthropic/claude-4-5-sonnet")
const modelIdSchema = z.string().regex(/^[a-zA-Z0-9-_]+\/[a-zA-Z0-9-_.]+$/, {
  message:
    "Model must be in format 'provider/model-name' (e.g., 'openai/gpt-4o')",
});

const llmCommandSettingsSchema = z
  .object({
    maxOutputTokens: z.number().int().positive().optional(),
  })
  .passthrough();

const llmCommandSchema = z.object({
  modelSettings: llmCommandSettingsSchema.optional(),
});

const createDefaultCommandSettings = () => ({
  commit: { modelSettings: { maxOutputTokens: DEFAULT_COMMIT_TOKENS } },
  review: { modelSettings: { maxOutputTokens: DEFAULT_REVIEW_TOKENS } },
});

const llmCommandsSchema = z
  .object({
    commit: llmCommandSchema.optional(),
    review: llmCommandSchema.optional(),
  })
  .default(() => createDefaultCommandSettings());

const llmSchema = z.object({
  model: modelIdSchema.default(`openai/${DEFAULT_LLM_MODEL}`),
  apiKeyEnv: z.string().default('OPENAI_API_KEY'),
  commands: llmCommandsSchema.optional(),
});

const cleanerSchema = z.object({
  whitelist: z.array(z.string()).default(['//!']),
  keepJSDoc: z.boolean().default(true),
  useGitignore: z.boolean().default(true),
  ignore: z.array(z.string()).default([]),
});

const packerSchema = z.object({
  ignore: z
    .array(z.string())
    .default([
      'package-lock.json',
      'yarn.lock',
      'pnpm-lock.yaml',
      '.git',
      '.kodu',
      'node_modules',
      'dist',
      'coverage',
    ]),
  useGitignore: z.boolean().default(true),
  contentBasedBinaryDetection: z.boolean().default(false),
});

const promptSourceSchema = z.string();

const promptsSchema = z
  .object({
    review: z.record(z.string(), promptSourceSchema).optional(),
    commit: promptSourceSchema.optional(),
    pack: promptSourceSchema.optional(),
  })
  .optional();

export const configSchema = z.object({
  $schema: z.string().optional(),
  llm: llmSchema.optional(),
  cleaner: cleanerSchema.default({
    whitelist: ['//!'],
    keepJSDoc: true,
    useGitignore: true,
    ignore: [],
  }),
  packer: packerSchema.default({
    ignore: [
      'package-lock.json',
      'yarn.lock',
      'pnpm-lock.yaml',
      '.git',
      '.kodu',
      'node_modules',
      'dist',
      'coverage',
    ],
    useGitignore: true,
    contentBasedBinaryDetection: false,
  }),
  prompts: promptsSchema,
});

export type KoduConfig = z.infer<typeof configSchema>;
