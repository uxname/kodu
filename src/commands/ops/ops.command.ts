import { Command, CommandRunner } from 'nest-commander';
import { UiService } from '../../core/ui/ui.service';
import { OpsAddCommand } from './ops-add.command';
import { OpsInitCommand } from './ops-init.command';
import { OpsListCommand } from './ops-list.command';
import { OpsPathCommand } from './ops-path.command';
import { OpsRunbookCommand } from './ops-runbook.command';
import { OpsStatusCommand } from './ops-status.command';
import { OpsUseCommand } from './ops-use.command';

@Command({
  name: 'ops',
  description:
    'Работа с проектами и стендами (local/dev/stage/prod) из любого места',
  subCommands: [
    OpsInitCommand,
    OpsListCommand,
    OpsAddCommand,
    OpsStatusCommand,
    OpsUseCommand,
    OpsPathCommand,
    OpsRunbookCommand,
  ],
})
export class OpsCommand extends CommandRunner {
  constructor(private readonly ui: UiService) {
    super();
  }

  async run(): Promise<void> {
    this.ui.log.info('Использование: kodu ops <команда>');
    this.ui.log.info(
      '  init                  — настроить стенды в текущем проекте',
    );
    this.ui.log.info('  list                  — список всех проектов');
    this.ui.log.info('  add <name> --path <d> — зарегистрировать проект');
    this.ui.log.info('  status [name]         — активный стенд проекта');
    this.ui.log.info('  use <stand>           — переключить активный стенд');
    this.ui.log.info('  path <name>           — путь к репозиторию проекта');
    this.ui.log.info('  runbook <name> [stand]— показать инструкции по стенду');
  }
}
