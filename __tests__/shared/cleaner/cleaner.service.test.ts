import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ConfigService } from '../../../src/core/config/config.service';

vi.mock('../../../src/core/config/config.service', () => ({
  ConfigService: vi.fn().mockImplementation(() => ({
    getConfig: () => ({
      cleaner: {
        whitelist: [],
        keepJSDoc: false,
        useGitignore: false,
      },
      packer: {
        ignore: [],
        useGitignore: false,
      },
    }),
  })),
}));

vi.mock('../../../src/core/file-system/fs.service', () => ({
  FsService: vi.fn().mockImplementation(() => ({})),
}));

describe('CleanerService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should remove single-line comments', async () => {
    const { CleanerService } = await import(
      '../../../src/shared/cleaner/cleaner.service'
    );
    const configService = new ConfigService() as never;
    const fsService = {} as never;
    const cleaner = new CleanerService(configService, fsService);

    const input = `const x = 1; // comment
const y = 2;`;
    const result = cleaner.cleanContent('test.ts', input);

    expect(result).toBe('const x = 1; \nconst y = 2;');
  });

  it('should remove multi-line comments', async () => {
    const { CleanerService } = await import(
      '../../../src/shared/cleaner/cleaner.service'
    );
    const configService = new ConfigService() as never;
    const fsService = {} as never;
    const cleaner = new CleanerService(configService, fsService);

    const input = `/* comment */
const x = 1;`;
    const result = cleaner.cleanContent('test.ts', input);

    expect(result).toBe('\nconst x = 1;');
  });

  it('should preserve comments in whitelist', async () => {
    const { CleanerService } = await import(
      '../../../src/shared/cleaner/cleaner.service'
    );
    const configService = new ConfigService() as never;
    const fsService = {} as never;
    const cleaner = new CleanerService(configService, fsService);

    const input = `const x = 1; // eslint-disable-line
const y = 2;`;
    const result = cleaner.cleanContent('test.ts', input);

    expect(result).toBe('const x = 1; // eslint-disable-line\nconst y = 2;');
  });

  it('should keep JSDoc when option is set', async () => {
    const { CleanerService } = await import(
      '../../../src/shared/cleaner/cleaner.service'
    );
    const configService = new ConfigService() as never;
    const fsService = {} as never;
    const cleaner = new CleanerService(configService, fsService);

    const input = `/** JSDoc */
const x = 1;`;
    const result = cleaner.cleanContent('test.ts', input, true);

    expect(result).toBe('/** JSDoc */\nconst x = 1;');
  });

  it('should remove JSX expression without content', async () => {
    const { CleanerService } = await import(
      '../../../src/shared/cleaner/cleaner.service'
    );
    const configService = new ConfigService() as never;
    const fsService = {} as never;
    const cleaner = new CleanerService(configService, fsService);

    const input = `const x = <>{/* comment */}</>;`;
    const result = cleaner.cleanContent('test.tsx', input);

    expect(result).toBe('const x = <></>;');
  });
});
