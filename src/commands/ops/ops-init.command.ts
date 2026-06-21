import path from 'node:path';
import { CommandRunner, Option, SubCommand } from 'nest-commander';
import { DEFAULT_STANDS } from '../../core/registry/registry.schema';
import { RegistryService } from '../../core/registry/registry.service';
import { UiService } from '../../core/ui/ui.service';
import { RunbookService } from '../../shared/runbook/runbook.service';

type InitOptions = {
  name?: string;
  // Свойства называются по длинным флагам `--active` / `--stand`.
  active?: string;
  stand?: string[];
};

@SubCommand({
  name: 'init',
  description:
    'Настроить стенды в текущем проекте: создать .runbook/, .gitignore и зарегистрировать проект',
})
export class OpsInitCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly registry: RegistryService,
    private readonly runbook: RunbookService,
  ) {
    super();
  }

  @Option({
    flags: '-n, --name <name>',
    description: 'Уникальное имя проекта (по умолчанию — имя папки)',
  })
  parseName(value: string): string {
    return value;
  }

  @Option({
    flags: '-a, --active <stand>',
    description: 'Активный стенд по умолчанию (по умолчанию local)',
  })
  parseActive(value: string): string {
    return value;
  }

  @Option({
    flags: '-s, --stand <stand>',
    description: 'Стенд проекта (можно повторять)',
  })
  parseStand(value: string, previous: string[] = []): string[] {
    return [...previous, value];
  }

  async run(_inputs: string[], options: InitOptions): Promise<void> {
    const root = process.cwd();
    const name = options.name ?? path.basename(root);
    const stands =
      options.stand && options.stand.length > 0
        ? options.stand
        : [...DEFAULT_STANDS];
    const activeStand = options.active ?? stands[0] ?? 'local';

    try {
      // 1. Скаффолд .runbook/ (config.json + runbook.md).
      await this.runbook.scaffold({ project: name, stands, activeStand }, root);
      this.ui.log.success(`Создан .runbook/ для проекта "${name}".`);
      this.ui.log.info(`Активный стенд: ${activeStand}`);
      this.ui.log.info(
        `Заполни шаги деплоя в ${this.runbook.runbookPath(root)}`,
      );

      // 2. Гарантируем игнор .runbook/ — автоматически.
      const gitignore = await this.runbook.ensureGitignore(root);
      this.reportGitignore(gitignore);

      // 3. Регистрируем проект в глобальном реестре (имя уникально).
      await this.registerProject(name, root, stands);
    } catch (error) {
      this.ui.log.error((error as Error).message);
      process.exitCode = 1;
    }
  }

  private reportGitignore(result: string): void {
    switch (result) {
      case 'created':
        this.ui.log.success('Создан .gitignore с записью /.runbook/');
        break;
      case 'added':
        this.ui.log.success('Добавил /.runbook/ в .gitignore');
        break;
      case 'present':
        this.ui.log.info('/.runbook/ уже в .gitignore');
        break;
      case 'no-git':
        this.ui.log.warn(
          'Это не git-репозиторий — .gitignore не настроен. Не коммить .runbook/ вручную.',
        );
        break;
    }
  }

  private async registerProject(
    name: string,
    root: string,
    stands: string[],
  ): Promise<void> {
    const existing = await this.registry.get(name);

    if (!existing) {
      await this.registry.add(name, { path: root, stands });
      this.ui.log.success(`Проект "${name}" добавлен в реестр.`);
      return;
    }

    if (existing.path === root) {
      this.ui.log.info(`Проект "${name}" уже в реестре.`);
      return;
    }

    this.ui.log.warn(
      `Имя "${name}" уже занято другим путём (${existing.path}). ` +
        'Запусти заново с другим именем: kodu ops init --name <другое-имя>',
    );
  }
}
