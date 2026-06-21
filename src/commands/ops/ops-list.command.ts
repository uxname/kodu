import { CommandRunner, SubCommand } from 'nest-commander';
import { RegistryService } from '../../core/registry/registry.service';
import { UiService } from '../../core/ui/ui.service';
import { RunbookService } from '../../shared/runbook/runbook.service';

@SubCommand({
  name: 'list',
  aliases: ['ls'],
  description: 'Показать все проекты из реестра и их активные стенды',
})
export class OpsListCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly registry: RegistryService,
    private readonly runbook: RunbookService,
  ) {
    super();
  }

  async run(): Promise<void> {
    try {
      const projects = await this.registry.list();
      const names = Object.keys(projects).sort((a, b) => a.localeCompare(b));

      if (names.length === 0) {
        this.ui.log.info(
          'Реестр пуст. Добавь проект: kodu ops add <name> --path <dir>',
        );
        this.ui.log.info(`Файл реестра: ${this.registry.getFilePath()}`);
        return;
      }

      for (const name of names) {
        const entry = projects[name];
        const active = await this.readActiveStand(entry.path);
        const activeLabel = active ? ` [активный: ${active}]` : '';
        this.ui.log.info(`${name}${activeLabel}`);
        this.ui.log.info(`  путь:   ${entry.path}`);
        this.ui.log.info(`  стенды: ${entry.stands.join(', ')}`);
      }
    } catch (error) {
      this.ui.log.error((error as Error).message);
      process.exitCode = 1;
    }
  }

  private async readActiveStand(root: string): Promise<string | undefined> {
    try {
      if (!(await this.runbook.exists(root))) {
        return undefined;
      }
      return (await this.runbook.readConfig(root)).activeStand;
    } catch {
      return undefined;
    }
  }
}
