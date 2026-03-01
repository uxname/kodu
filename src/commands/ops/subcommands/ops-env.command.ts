import path from 'node:path';
import { CommandRunner, Option, SubCommand } from 'nest-commander';
import { ConfigService } from '../../../core/config/config.service';
import { SshService } from '../../../shared/ssh/ssh.service';
import { OpsCliError } from '../ops.types';
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
  server?: string;
  action?: string;
  project?: string;
  key?: string;
  val?: string;
};

@SubCommand({
  name: 'env',
  description:
    'Manage remote .env files for a specific project.\nExamples:\n  kodu ops env --server dev --action get --project my-app\n  kodu ops env --server dev --action set --project my-app --key PORT --val 3000\n  kodu ops env --server dev --action unset --project my-app --key PORT',
})
export class OpsEnvCommand extends CommandRunner {
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

  @Option({
    flags: '-a, --action <type>',
    description: 'Action to perform: get | set | unset',
  })
  parseAction(value: string): string {
    return value;
  }

  @Option({
    flags: '-p, --project <name>',
    description: 'Target project directory name',
  })
  parseProject(value: string): string {
    return value;
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
    try {
      if (passedParams.length > 0) {
        throw new OpsCliError(
          'VALIDATION_ERROR',
          'Positional arguments are not supported. Use named flags (e.g., --server, --action). Run with --help for examples.',
        );
      }

      const serverAlias = ensureRequired(options.server, 'server');
      const rawAction = ensureRequired(options.action, 'action');
      const project = ensureRequired(options.project, 'project');

      const action = ensureAction<OpsEnvAction>(
        rawAction,
        ['get', 'set', 'unset'],
        'env action',
      );
      const server = await resolveServerOrThrow(
        this.configService.getConfig(),
        serverAlias,
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
      const scriptLines = [
        `ENV_FILE=${quotedPath}`,
        `KEY=${quotedKey}`,
        `VAL=${quotedVal}`,
        'mkdir -p "$(dirname "$ENV_FILE")"',
        'touch "$ENV_FILE"',
        'awk -v k="$KEY" -v v="$VAL" \'BEGIN{found=0} $0 ~ "^" k "=" { print k "=" v; found=1; next } { print } END { if (!found) print k "=" v }\' "$ENV_FILE" > "$ENV_FILE.tmp"',
        'mv "$ENV_FILE.tmp" "$ENV_FILE"',
      ];
      return this.buildBashScript(scriptLines);
    }

    const scriptLines = [
      `ENV_FILE=${quotedPath}`,
      `KEY=${quotedKey}`,
      'if [ ! -f "$ENV_FILE" ]; then exit 0; fi',
      'grep -v "^$KEY=" "$ENV_FILE" > "$ENV_FILE.tmp"',
      'mv "$ENV_FILE.tmp" "$ENV_FILE"',
    ];
    return this.buildBashScript(scriptLines);
  }

  private buildBashScript(scriptLines: string[]): string {
    const script = scriptLines.join(' && ');
    return `bash -lc ${shellQuote(script)}`;
  }
}
