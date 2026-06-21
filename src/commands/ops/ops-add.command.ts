import path from 'node:path';
import { CommandRunner, Option, SubCommand } from 'nest-commander';
import { DEFAULT_STANDS } from '../../core/registry/registry.schema';
import { RegistryService } from '../../core/registry/registry.service';
import { UiService } from '../../core/ui/ui.service';

type AddOptions = {
  path?: string;
  repo?: string;
  // Свойство называется по длинному флагу `--stand` (см. nest-commander/commander).
  stand?: string[];
};

@SubCommand({
  name: 'add',
  description:
    'Зарегистрировать проект в глобальном реестре по уникальному имени',
  arguments: '<name>',
})
export class OpsAddCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly registry: RegistryService,
  ) {
    super();
  }

  @Option({
    flags: '-p, --path <dir>',
    description: 'Путь к репозиторию (по умолчанию текущая директория)',
  })
  parsePath(value: string): string {
    return value;
  }

  @Option({
    flags: '-r, --repo <url>',
    description: 'URL репозитория (git clone)',
  })
  parseRepo(value: string): string {
    return value;
  }

  @Option({
    flags: '-s, --stand <stand>',
    description: 'Стенд проекта (можно повторять)',
  })
  parseStand(value: string, previous: string[] = []): string[] {
    return [...previous, value];
  }

  async run(inputs: string[], options: AddOptions): Promise<void> {
    const name = inputs[0];

    if (!name) {
      this.ui.log.error(
        'Укажи имя проекта: kodu ops add <name> [--path <dir>]',
      );
      process.exitCode = 1;
      return;
    }

    const projectPath = path.resolve(options.path ?? process.cwd());
    const stands =
      options.stand && options.stand.length > 0
        ? options.stand
        : [...DEFAULT_STANDS];

    try {
      await this.registry.add(name, {
        path: projectPath,
        repo: options.repo,
        stands,
      });
      this.ui.log.success(`Проект "${name}" добавлен в реестр.`);
      this.ui.log.info(`Путь: ${projectPath}`);
      this.ui.log.info(`Стенды: ${stands.join(', ')}`);
    } catch (error) {
      this.ui.log.error((error as Error).message);
      process.exitCode = 1;
    }
  }
}
