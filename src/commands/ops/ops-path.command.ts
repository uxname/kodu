import { CommandRunner, SubCommand } from 'nest-commander';
import { RegistryService } from '../../core/registry/registry.service';
import { UiService } from '../../core/ui/ui.service';
import { resolveProjectRoot } from './ops.helpers';

@SubCommand({
  name: 'path',
  description:
    'Напечатать путь к репозиторию проекта (удобно для cd $(kodu ops path <name>))',
  arguments: '<name>',
})
export class OpsPathCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly registry: RegistryService,
  ) {
    super();
  }

  async run(inputs: string[]): Promise<void> {
    const name = inputs[0];

    if (!name) {
      this.ui.log.error('Укажи имя проекта: kodu ops path <name>');
      process.exitCode = 1;
      return;
    }

    try {
      const root = await resolveProjectRoot(this.registry, name);
      // Чистый вывод пути в stdout — чтобы работало в подстановке команды.
      process.stdout.write(`${root}\n`);
    } catch (error) {
      this.ui.log.error((error as Error).message);
      process.exitCode = 1;
    }
  }
}
