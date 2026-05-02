import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ConfigService } from '../../../src/core/config/config.service';
import { FsService } from '../../../src/core/file-system/fs.service';
import { UiService } from '../../../src/core/ui/ui.service';

vi.mock('../../../src/core/config/config.service', () => ({
  ConfigService: vi.fn().mockImplementation(() => ({
    getConfig: () => ({
      cleaner: { whitelist: [], keepJSDoc: false },
      packer: { ignore: ['node_modules', 'dist'], useGitignore: false },
    }),
  })),
}));

vi.mock('../../../src/core/ui/ui.service', () => ({
  UiService: vi.fn().mockImplementation(() => ({
    log: { warn: vi.fn() },
  })),
}));

describe('FsService', () => {
  let fsService: FsService;

  beforeEach(() => {
    vi.clearAllMocks();
    const configService = new ConfigService() as never;
    const uiService = new UiService() as never;
    fsService = new FsService(configService, uiService);
  });

  describe('findProjectFiles', () => {
    it('should find ts files in current directory', async () => {
      const files = await fsService.findProjectFiles({
        ignore: ['node_modules', 'dist'],
        useGitignore: false,
      });

      const tsFiles = files.filter((f) => f.endsWith('.ts'));
      expect(tsFiles.length).toBeGreaterThan(0);
    });

    it('should exclude node_modules by default', async () => {
      const files = await fsService.findProjectFiles({
        ignore: ['node_modules', 'dist'],
        useGitignore: false,
      });

      const hasNodeModules = files.some((f) => f.includes('node_modules'));
      expect(hasNodeModules).toBe(false);
    });

    it('should return relative paths', async () => {
      const files = await fsService.findProjectFiles({
        ignore: [],
        useGitignore: false,
      });

      const hasAbsolute = files.some((f) => f.startsWith('/'));
      expect(hasAbsolute).toBe(false);
    });

    it('should return sorted paths', async () => {
      const files = await fsService.findProjectFiles({
        ignore: [],
        useGitignore: false,
      });

      const sorted = [...files].sort((a, b) => a.localeCompare(b));
      expect(files).toEqual(sorted);
    });
  });
});
