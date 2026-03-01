import { CommandRunner, SubCommand } from 'nest-commander';
import { ConfigService } from '../../../core/config/config.service';
import { SshService } from '../../../shared/ssh/ssh.service';
import {
  printCliError,
  printJson,
  printSshError,
  resolveServerOrThrow,
} from '../ops.utils';

@SubCommand({
  name: 'sysinfo',
  description: 'Collect remote server diagnostics',
  arguments: '<alias>',
})
export class OpsSysinfoCommand extends CommandRunner {
  constructor(
    private readonly configService: ConfigService,
    private readonly sshService: SshService,
  ) {
    super();
  }

  async run(passedParams: string[]): Promise<void> {
    const [alias] = passedParams;

    try {
      const server = await resolveServerOrThrow(
        this.configService.getConfig(),
        alias,
      );
      const payload = `echo "{"uptime": "$(uptime -p)", "disk_usage": "$(df -h / | tail -1 | awk '{print $5}')", "mem_free": "$(free -m | grep Mem | awk '{print $4}')MB"}"`;

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
