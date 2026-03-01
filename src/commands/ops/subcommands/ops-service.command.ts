import path from 'node:path';
import { CommandRunner, Option, SubCommand } from 'nest-commander';
import { ConfigService } from '../../../core/config/config.service';
import { SshService } from '../../../shared/ssh/ssh.service';
import { OpsCliError, type ResolvedServerConfig } from '../ops.types';
import {
  ensureAction,
  ensureRequired,
  printCliError,
  printJson,
  printSshError,
  resolveAppsPath,
  resolveServerOrThrow,
  shellQuote,
} from '../ops.utils';

type OpsServiceAction = 'clone' | 'pull' | 'up' | 'down' | 'logs' | 'status';

type OpsServiceOptions = {
  server?: string;
  action?: string;
  project?: string;
  repo?: string;
};

@SubCommand({
  name: 'service',
  description:
    'Manage project lifecycle using Docker Compose.\nExamples:\n  kodu ops service --server dev --action clone --project temp --repo https://github.com/org/repo.git\n  kodu ops service --server dev --action up --project temp\n  kodu ops service --server dev --action status --project temp\n  kodu ops service --server dev --action logs --project temp\n  kodu ops service --server dev --action pull --project temp\n  kodu ops service --server dev --action down --project temp',
})
export class OpsServiceCommand extends CommandRunner {
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
    description: 'Action to perform: clone | pull | up | down | logs | status',
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

  @Option({ flags: '--repo <url>', description: 'Repository URL for clone' })
  parseRepo(value: string): string {
    return value;
  }

  async run(
    passedParams: string[],
    options: OpsServiceOptions = {},
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

      const action = ensureAction<OpsServiceAction>(
        rawAction,
        ['clone', 'pull', 'up', 'down', 'logs', 'status'],
        'service action',
      );
      const server = await resolveServerOrThrow(
        this.configService.getConfig(),
        serverAlias,
      );
      const projectPath = path.posix.join(resolveAppsPath(server), project);

      if (action === 'status') {
        await this.runStatus(server, projectPath);
        return;
      }

      const command = this.buildActionCommand(action, projectPath, options);
      const result = await this.sshService.execute(server, command);
      if (!result.success) {
        printSshError(result, command);
        return;
      }

      printJson({
        status: 'ok',
        data: {
          action,
          project,
          stdout: result.stdout,
        },
      });
    } catch (error) {
      printCliError(error);
      process.exitCode = 1;
    }
  }

  private buildActionCommand(
    action: Exclude<OpsServiceAction, 'status'>,
    projectPath: string,
    options: OpsServiceOptions,
  ): string {
    const quotedProjectPath = shellQuote(projectPath);

    if (action === 'clone') {
      const repo = ensureRequired(options.repo, 'repo');
      const quotedRepo = shellQuote(repo);
      return [
        `if [ -d ${quotedProjectPath} ]; then echo 'Project already exists' >&2; exit 3; fi`,
        `git clone ${quotedRepo} ${quotedProjectPath}`,
      ].join(' && ');
    }

    if (action === 'pull') {
      return `cd ${quotedProjectPath} && git pull`;
    }

    if (action === 'up') {
      return `cd ${quotedProjectPath} && docker compose up -d`;
    }

    if (action === 'down') {
      return `cd ${quotedProjectPath} && docker compose down`;
    }

    return `cd ${quotedProjectPath} && docker compose logs --no-color --tail=200`;
  }

  private async runStatus(
    server: ResolvedServerConfig,
    projectPath: string,
  ): Promise<void> {
    const quotedProjectPath = shellQuote(projectPath);
    const jsonCommand = `cd ${quotedProjectPath} && docker compose ps --format json`;
    const jsonResult = await this.sshService.execute(server, jsonCommand);

    if (jsonResult.success) {
      printJson({
        status: 'ok',
        data: {
          action: 'status',
          stdout: jsonResult.stdout,
        },
      });
      return;
    }

    const fallbackCommand = `cd ${quotedProjectPath} && docker compose ps`;
    const fallbackResult = await this.sshService.execute(
      server,
      fallbackCommand,
    );

    if (!fallbackResult.success) {
      printSshError(fallbackResult, fallbackCommand);
      return;
    }

    printJson({
      status: 'ok',
      data: {
        action: 'status',
        raw: fallbackResult.stdout,
      },
    });
  }
}
