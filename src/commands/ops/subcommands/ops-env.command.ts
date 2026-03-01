import path from 'node:path';
import { CommandRunner, Option, SubCommand } from 'nest-commander';
import { ConfigService } from '../../../core/config/config.service';
import { SshService } from '../../../shared/ssh/ssh.service';
import {
  ensureAction,
  ensureEnvKey,
  ensureRequired,
  printCliError,
  printJson,
  printSshError,
  resolveAppsPath,
  resolveServerOrThrow,
  shellQuote,
} from '../ops.utils';

type OpsEnvAction = 'get' | 'set' | 'unset';

type OpsEnvOptions = {
  key?: string;
  val?: string;
};

@SubCommand({
  name: 'env',
  description: 'Manage remote .env files',
  arguments: '<alias> <action> <project>',
})
export class OpsEnvCommand extends CommandRunner {
  constructor(
    private readonly configService: ConfigService,
    private readonly sshService: SshService,
  ) {
    super();
  }

  @Option({ flags: '--key <key>', description: 'Environment key' })
  parseKey(value: string): string {
    return value;
  }

  @Option({ flags: '--val <value>', description: 'Environment value' })
  parseVal(value: string): string {
    return value;
  }

  async run(
    passedParams: string[],
    options: OpsEnvOptions = {},
  ): Promise<void> {
    const [alias, rawAction, project] = passedParams;

    try {
      const action = ensureAction<OpsEnvAction>(
        rawAction,
        ['get', 'set', 'unset'],
        'env action',
      );
      const server = await resolveServerOrThrow(
        this.configService.getConfig(),
        alias,
      );
      const envPath = path.posix.join(resolveAppsPath(server), project, '.env');

      const command = this.buildCommand(action, envPath, options);
      const result = await this.sshService.execute(server, command);

      if (!result.success) {
        printSshError(result, command);
        return;
      }

      if (action === 'get') {
        printJson({ status: 'ok', data: { content: result.stdout } });
        return;
      }

      printJson({ status: 'ok', message: 'Env updated' });
    } catch (error) {
      printCliError(error);
      process.exitCode = 1;
    }
  }

  private buildCommand(
    action: OpsEnvAction,
    envPath: string,
    options: OpsEnvOptions,
  ): string {
    const quotedPath = shellQuote(envPath);

    if (action === 'get') {
      return `cat ${quotedPath}`;
    }

    const key = ensureEnvKey(options.key);
    const quotedKey = shellQuote(key);

    if (action === 'set') {
      const val = ensureRequired(options.val, 'val');
      const quotedVal = shellQuote(val);
      return [
        `ENV_FILE=${quotedPath}`,
        `KEY=${quotedKey}`,
        `VAL=${quotedVal}`,
        'mkdir -p "$(dirname "$ENV_FILE")"',
        'touch "$ENV_FILE"',
        'awk -v k="$KEY" -v v="$VAL" \'BEGIN{found=0} $0 ~ "^" k "=" { print k "=" v; found=1; next } { print } END { if (!found) print k "=" v }\' "$ENV_FILE" > "$ENV_FILE.tmp"',
        'mv "$ENV_FILE.tmp" "$ENV_FILE"',
      ].join(' && ');
    }

    return [
      `ENV_FILE=${quotedPath}`,
      `KEY=${quotedKey}`,
      'if [ ! -f "$ENV_FILE" ]; then exit 0; fi',
      'grep -v "^$KEY=" "$ENV_FILE" > "$ENV_FILE.tmp"',
      'mv "$ENV_FILE.tmp" "$ENV_FILE"',
    ].join(' && ');
  }
}
