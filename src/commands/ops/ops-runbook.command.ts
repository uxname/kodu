import { CommandRunner, SubCommand } from 'nest-commander';
import { RegistryService } from '../../core/registry/registry.service';
import { UiService } from '../../core/ui/ui.service';
import { RunbookService } from '../../shared/runbook/runbook.service';
import { resolveProjectRoot } from './ops.helpers';

@SubCommand({
  name: 'runbook',
  description: 'Напечатать runbook проекта (или секцию конкретного стенда)',
  arguments: '<name> [stand]',
})
export class OpsRunbookCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly registry: RegistryService,
    private readonly runbook: RunbookService,
  ) {
    super();
  }

  async run(inputs: string[]): Promise<void> {
    const name = inputs[0];
    const stand = inputs[1];

    if (!name) {
      this.ui.log.error('Укажи имя проекта: kodu ops runbook <name> [stand]');
      process.exitCode = 1;
      return;
    }

    try {
      const root = await resolveProjectRoot(this.registry, name);

      if (!(await this.runbook.exists(root))) {
        this.ui.log.warn(
          `В проекте "${name}" нет .runbook/. Запусти: kodu ops init`,
        );
        process.exitCode = 1;
        return;
      }

      const markdown = await this.runbook.readRunbook(root);
      const output = stand ? this.extractStand(markdown, stand) : markdown;

      if (!output) {
        this.ui.log.warn(`Секция для стенда "${stand}" не найдена в runbook.`);
        process.exitCode = 1;
        return;
      }

      process.stdout.write(`${output.trimEnd()}\n`);
    } catch (error) {
      this.ui.log.error((error as Error).message);
      process.exitCode = 1;
    }
  }

  /** Возвращает блок `## Стенд: <stand> ...` до следующего заголовка `## `. */
  private extractStand(markdown: string, stand: string): string | undefined {
    const lines = markdown.split(/\r?\n/);
    const startPrefix = `## Стенд: ${stand}`;
    const start = lines.findIndex((line) => line.startsWith(startPrefix));

    if (start === -1) {
      return undefined;
    }

    const rest = lines.slice(start + 1);
    const nextHeading = rest.findIndex((line) => line.startsWith('## '));
    const end = nextHeading === -1 ? lines.length : start + 1 + nextHeading;

    return lines.slice(start, end).join('\n');
  }
}
