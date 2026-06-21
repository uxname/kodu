import { CommandRunner, SubCommand } from 'nest-commander';
import { RegistryService } from '../../core/registry/registry.service';
import { UiService } from '../../core/ui/ui.service';
import { RunbookService } from '../../shared/runbook/runbook.service';
import { resolveProjectRoot } from './ops.helpers';

@SubCommand({
  name: 'use',
  aliases: ['switch'],
  description:
    'Переключить активный стенд: "kodu ops use <stand>" в текущем проекте или "kodu ops use <name> <stand>"',
  arguments: '<args...>',
})
export class OpsUseCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly registry: RegistryService,
    private readonly runbook: RunbookService,
  ) {
    super();
  }

  async run(inputs: string[]): Promise<void> {
    try {
      const { root, stand } = await this.resolveTarget(inputs);

      if (!stand) {
        this.ui.log.error(
          'Укажи стенд: kodu ops use <stand> или kodu ops use <name> <stand>',
        );
        process.exitCode = 1;
        return;
      }

      if (!(await this.runbook.exists(root))) {
        this.ui.log.warn(
          `В ${root} нет .runbook/. Инициализируй проект: kodu ops init`,
        );
        process.exitCode = 1;
        return;
      }

      const config = await this.runbook.readConfig(root);
      const stands = config.stands.includes(stand)
        ? config.stands
        : [...config.stands, stand];

      if (!config.stands.includes(stand)) {
        this.ui.log.info(`Стенд "${stand}" добавлен в список стендов проекта.`);
      }

      await this.runbook.writeConfig(
        { ...config, activeStand: stand, stands },
        root,
      );
      this.ui.log.success(
        `Активный стенд проекта "${config.project}" → ${stand}`,
      );
    } catch (error) {
      this.ui.log.error((error as Error).message);
      process.exitCode = 1;
    }
  }

  /** 1 аргумент → стенд в текущей папке; 2 аргумента → <name> <stand>. */
  private async resolveTarget(
    inputs: string[],
  ): Promise<{ root: string; stand?: string }> {
    if (inputs.length >= 2) {
      const root = await resolveProjectRoot(this.registry, inputs[0]);
      return { root, stand: inputs[1] };
    }

    return { root: process.cwd(), stand: inputs[0] };
  }
}
