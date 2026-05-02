import { execa } from 'execa';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { GitService } from '../../../src/shared/git/git.service';

vi.mock('execa');

const mockExeca = vi.mocked(execa);

describe('GitService', () => {
  let gitService: GitService;

  beforeEach(() => {
    vi.clearAllMocks();
    gitService = new GitService();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe('ensureRepo', () => {
    it('should resolve when inside git repo', async () => {
      mockExeca.mockResolvedValue({ stdout: '' } as never);

      await expect(gitService.ensureRepo()).resolves.not.toThrow();
    });

    it('should reject when not in git repo', async () => {
      mockExeca.mockRejectedValue(new Error('fatal: not a git repository'));

      await expect(gitService.ensureRepo()).rejects.toThrow();
    });
  });

  describe('getChangedFiles', () => {
    it('should return empty array when no changes', async () => {
      mockExeca.mockResolvedValue({ stdout: '' } as never);

      const files = await gitService.getChangedFiles();

      expect(files).toEqual([]);
    });

    it('should return changed files from all sources', async () => {
      const mockCalls = [
        { stdout: '' }, // ensureRepo
        { stdout: 'src/a.ts\nsrc/b.ts' }, // diff
        { stdout: '' }, // diff --staged
        { stdout: '' }, // ls-files
      ];
      mockExeca
        .mockResolvedValueOnce(mockCalls[0] as never)
        .mockResolvedValueOnce(mockCalls[1] as never)
        .mockResolvedValueOnce(mockCalls[2] as never)
        .mockResolvedValueOnce(mockCalls[3] as never);

      const files = await gitService.getChangedFiles();

      expect(files).toEqual(['src/a.ts', 'src/b.ts']);
    });
  });

  describe('getStagedFiles', () => {
    it('should return empty array when no staged files', async () => {
      mockExeca
        .mockResolvedValueOnce({ stdout: '' } as never)
        .mockResolvedValueOnce({ stdout: '' } as never);

      const files = await gitService.getStagedFiles();

      expect(files).toEqual([]);
    });

    it('should return staged files', async () => {
      mockExeca
        .mockResolvedValueOnce({ stdout: '' } as never)
        .mockResolvedValueOnce({ stdout: 'src/new.ts' } as never);

      const files = await gitService.getStagedFiles();

      expect(files).toEqual(['src/new.ts']);
    });
  });
});
