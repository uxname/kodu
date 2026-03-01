import { CommandRunner, Option, SubCommand } from 'nest-commander';
import { ConfigService } from '../../../core/config/config.service';
import { SshService } from '../../../shared/ssh/ssh.service';
import { OpsCliError } from '../ops.types';
import {
  ensureRequired,
  printCliError,
  printJson,
  printSshError,
  resolveServerOrThrow,
} from '../ops.utils';

type OpsSysinfoOptions = {
  server?: string;
};

@SubCommand({
  name: 'sysinfo',
  description:
    'Collect remote server diagnostics.\nExample: kodu ops sysinfo --server dev',
})
export class OpsSysinfoCommand extends CommandRunner {
  constructor(
    private readonly configService: ConfigService,
    private readonly sshService: SshService,
  ) {
    super();
  }

  @Option({
    flags: '-s, --server <name>',
    description: 'Server alias defined in kodu.json (e.g., dev)',
  })
  parseServer(value: string): string {
    return value;
  }

  async run(
    passedParams: string[],
    options: OpsSysinfoOptions = {},
  ): Promise<void> {
    try {
      if (passedParams.length > 0) {
        throw new OpsCliError(
          'VALIDATION_ERROR',
          'Positional arguments are not supported. Use named flags (e.g., --server, --action). Run with --help for examples.',
        );
      }

      const serverAlias = ensureRequired(options.server, 'server');

      const server = await resolveServerOrThrow(
        this.configService.getConfig(),
        serverAlias,
      );
      const payload = `echo "{"uptime": "$(uptime -p)", "disk_usage": "$(df -h / | tail -1 | awk '{print $5}')", "mem_free": "$(free -m | grep Mem | awk '{print $4}')MB"}`;

      const result = await this.sshService.execute(server, payload);
      if (!result.success) {
        printSshError(result, payload);
        return;
      }

      let data: Record<string, string> = {};
      try {
        data = JSON.parse(result.stdout) as Record<string, string>;
      } catch {
        data = { raw: result.stdout.trim() };
      }

      printJson({ status: 'ok', data });
    } catch (error) {
      printCliError(error);
      process.exitCode = 1;
    }
  }
}
