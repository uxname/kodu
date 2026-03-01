import { Command, CommandRunner } from 'nest-commander';
import { printJson } from './ops.utils';
import { OpsEnvCommand } from './subcommands/ops-env.command';
import { OpsRoutesCommand } from './subcommands/ops-routes.command';
import { OpsServiceCommand } from './subcommands/ops-service.command';
import { OpsSysinfoCommand } from './subcommands/ops-sysinfo.command';

@Command({
  name: 'ops',
  description: 'Remote server operations for agents',
  subCommands: [
    OpsSysinfoCommand,
    OpsEnvCommand,
    OpsRoutesCommand,
    OpsServiceCommand,
  ],
})
export class OpsCommand extends CommandRunner {
  async run(): Promise<void> {
    printJson(
      {
        status: 'error',
        code: 'VALIDATION_ERROR',
        error: 'Missing ops subcommand',
      },
      true,
    );
    process.exitCode = 1;
  }
}
