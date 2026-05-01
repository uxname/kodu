import { z } from 'zod';

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

const promptsSchema = z
  .object({
    pack: z.string().optional(),
  })
  .optional();

export const configSchema = z.object({
  $schema: z.string().optional(),
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
