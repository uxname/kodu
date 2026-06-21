import { CommandRunner, SubCommand } from 'nest-commander';
import { RegistryService } from '../../core/registry/registry.service';
import { UiService } from '../../core/ui/ui.service';
import { RunbookService } from '../../shared/runbook/runbook.service';
import { resolveProjectRoot } from './ops.helpers';

@SubCommand({
  name: 'status',
  description:
    'Показать активный стенд и стенды проекта (по имени или в текущей папке)',
  arguments: '[name]',
})
export class OpsStatusCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly registry: RegistryService,
    private readonly runbook: RunbookService,
  ) {
    super();
  }

  async run(inputs: string[]): Promise<void> {
    try {
      const name = inputs[0];
      const root = name
        ? await resolveProjectRoot(this.registry, name)
        : process.cwd();

      if (!(await this.runbook.exists(root))) {
        this.ui.log.warn(
          `В ${root} нет .runbook/. Инициализируй проект: kodu ops init`,
        );
        process.exitCode = 1;
        return;
      }

      const config = await this.runbook.readConfig(root);
      this.ui.log.info(`Проект:        ${config.project}`);
      this.ui.log.info(`Активный стенд: ${config.activeStand}`);
      this.ui.log.info(`Стенды:        ${config.stands.join(', ')}`);
      this.ui.log.info(`Путь:          ${root}`);
    } catch (error) {
      this.ui.log.error((error as Error).message);
      process.exitCode = 1;
    }
  }
}
