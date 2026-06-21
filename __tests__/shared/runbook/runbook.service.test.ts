import { promises as fs } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { RunbookService } from '../../../src/shared/runbook/runbook.service';

describe('RunbookService', () => {
  let root: string;
  const service = new RunbookService();

  beforeEach(async () => {
    root = await fs.mkdtemp(path.join(os.tmpdir(), 'kodu-runbook-'));
  });

  afterEach(async () => {
    await fs.rm(root, { recursive: true, force: true });
  });

  async function readGitignore(): Promise<string> {
    return fs.readFile(path.join(root, '.gitignore'), 'utf8');
  }

  describe('scaffold', () => {
    it('creates config.json and runbook.md', async () => {
      await service.scaffold(
        { project: 'demo', stands: ['local', 'dev'], activeStand: 'local' },
        root,
      );

      const config = await service.readConfig(root);
      expect(config.project).toBe('demo');
      expect(config.activeStand).toBe('local');

      const runbook = await service.readRunbook(root);
      expect(runbook).toContain('## Стенд: dev');
    });

    it('does not overwrite an existing runbook.md', async () => {
      await service.scaffold(
        { project: 'demo', stands: ['local'], activeStand: 'local' },
        root,
      );
      await fs.writeFile(service.runbookPath(root), 'ручные правки', 'utf8');

      await service.scaffold(
        { project: 'demo', stands: ['local'], activeStand: 'dev' },
        root,
      );

      expect(await service.readRunbook(root)).toBe('ручные правки');
    });
  });

  describe('ensureGitignore', () => {
    it('returns no-git outside a git repository', async () => {
      expect(await service.ensureGitignore(root)).toBe('no-git');
      await expect(
        fs.access(path.join(root, '.gitignore')),
      ).rejects.toBeTruthy();
    });

    it('creates .gitignore when missing in a git repo', async () => {
      await fs.mkdir(path.join(root, '.git'));

      expect(await service.ensureGitignore(root)).toBe('created');
      expect(await readGitignore()).toContain('/.runbook/');
    });

    it('appends to an existing .gitignore', async () => {
      await fs.mkdir(path.join(root, '.git'));
      await fs.writeFile(
        path.join(root, '.gitignore'),
        'node_modules\n',
        'utf8',
      );

      expect(await service.ensureGitignore(root)).toBe('added');
      const content = await readGitignore();
      expect(content).toContain('node_modules');
      expect(content).toContain('/.runbook/');
    });

    it('is idempotent — does not duplicate the entry', async () => {
      await fs.mkdir(path.join(root, '.git'));
      await fs.writeFile(path.join(root, '.gitignore'), '.runbook/\n', 'utf8');

      expect(await service.ensureGitignore(root)).toBe('present');
      const occurrences = (await readGitignore()).split('runbook').length - 1;
      expect(occurrences).toBe(1);
    });
  });

  describe('detectStack', () => {
    it('detects docker-compose and .env.example', async () => {
      await fs.writeFile(path.join(root, 'docker-compose.yml'), '', 'utf8');
      await fs.writeFile(path.join(root, '.env.example'), '', 'utf8');

      const detected = await service.detectStack(root);
      expect(detected.compose).toBe(true);
      expect(detected.envExample).toBe(true);
      expect(detected.dockerfile).toBe(false);
    });
  });
});
