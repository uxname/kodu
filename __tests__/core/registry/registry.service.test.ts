import { promises as fs } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { RegistryService } from '../../../src/core/registry/registry.service';

describe('RegistryService', () => {
  let tmpDir: string;
  let originalXdg: string | undefined;
  let service: RegistryService;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'kodu-registry-'));
    originalXdg = process.env.XDG_CONFIG_HOME;
    process.env.XDG_CONFIG_HOME = tmpDir;
    // Путь реестра читается в конструкторе — создаём после установки env.
    service = new RegistryService();
  });

  afterEach(async () => {
    if (originalXdg === undefined) {
      delete process.env.XDG_CONFIG_HOME;
    } else {
      process.env.XDG_CONFIG_HOME = originalXdg;
    }
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('returns empty registry when file does not exist', async () => {
    const registry = await service.load();
    expect(registry.projects).toEqual({});
  });

  it('does not create the file until something is saved', async () => {
    await service.load();
    await expect(fs.access(service.getFilePath())).rejects.toBeTruthy();
  });

  it('adds a project and persists it to disk', async () => {
    await service.add('my-api', { path: '/work/my-api', stands: ['dev'] });

    const onDisk = JSON.parse(await fs.readFile(service.getFilePath(), 'utf8'));
    expect(onDisk.projects['my-api']).toEqual({
      path: '/work/my-api',
      stands: ['dev'],
    });
  });

  it('rejects a duplicate project name by default', async () => {
    await service.add('my-api', { path: '/work/my-api', stands: ['dev'] });

    await expect(
      service.add('my-api', { path: '/other', stands: ['dev'] }),
    ).rejects.toThrow(/уже есть в реестре/);
  });

  it('overwrites when overwrite option is set', async () => {
    await service.add('my-api', { path: '/work/my-api', stands: ['dev'] });
    await service.add(
      'my-api',
      { path: '/new', stands: ['prod'] },
      { overwrite: true },
    );

    const entry = await service.get('my-api');
    expect(entry?.path).toBe('/new');
  });

  it('updates an existing project', async () => {
    await service.add('my-api', { path: '/work/my-api', stands: ['dev'] });
    await service.update('my-api', { path: '/moved' });

    expect((await service.get('my-api'))?.path).toBe('/moved');
  });

  it('throws when reading an invalid registry file', async () => {
    await fs.mkdir(path.dirname(service.getFilePath()), { recursive: true });
    await fs.writeFile(service.getFilePath(), '{ not json', 'utf8');

    await expect(service.load()).rejects.toThrow();
  });
});
