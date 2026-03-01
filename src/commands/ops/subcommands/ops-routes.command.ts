import path from 'node:path';
import { CommandRunner, Option, SubCommand } from 'nest-commander';
import { ConfigService } from '../../../core/config/config.service';
import { SshService } from '../../../shared/ssh/ssh.service';
import { OpsCliError } from '../ops.types';
import {
  ensureAction,
  ensureRequired,
  printCliError,
  printJson,
  printSshError,
  resolveCaddyPath,
  resolveServerOrThrow,
  shellQuote,
} from '../ops.utils';

type OpsRoutesAction = 'list' | 'add' | 'remove' | 'update';

type OpsRoutesOptions = {
  server?: string;
  action?: string;
  domain?: string;
  upstream?: string;
};

@SubCommand({
  name: 'routes',
  description:
    'Manage remote Caddy reverse proxy routes.\nExamples:\n  kodu ops routes --server dev --action list\n  kodu ops routes --server dev --action add --domain api.example.com --upstream 127.0.0.1:3000\n  kodu ops routes --server dev --action update --domain api.example.com --upstream 127.0.0.1:4000\n  kodu ops routes --server dev --action remove --domain api.example.com',
})
export class OpsRoutesCommand extends CommandRunner {
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
    description: 'Action to perform: list | add | remove | update',
  })
  parseAction(value: string): string {
    return value;
  }

  @Option({
    flags: '--domain <domain>',
    description: 'Domain name (required for add, update, remove)',
  })
  parseDomain(value: string): string {
    return value;
  }

  @Option({
    flags: '--upstream <upstream>',
    description: 'Upstream host:port (required for add, update)',
  })
  parseUpstream(value: string): string {
    return value;
  }

  async run(
    passedParams: string[],
    options: OpsRoutesOptions = {},
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
      const action = ensureAction<OpsRoutesAction>(
        rawAction,
        ['list', 'add', 'remove', 'update'],
        'routes action',
      );
      const server = await resolveServerOrThrow(
        this.configService.getConfig(),
        serverAlias,
      );

      const caddyRoot = resolveCaddyPath(server);
      const caddyfilePath = path.posix.join(caddyRoot, 'data', 'Caddyfile');

      if (action === 'list') {
        const command = `cat ${shellQuote(caddyfilePath)}`;
        const result = await this.sshService.execute(server, command);
        if (!result.success) {
          printSshError(result, command);
          return;
        }

        printJson({ status: 'ok', data: { caddyfile: result.stdout } });
        return;
      }

      const domain = ensureRequired(options.domain, 'domain');
      const upstream =
        action === 'add' || action === 'update'
          ? ensureRequired(options.upstream, 'upstream')
          : '';

      const editCommand = this.buildEditCommand({
        action,
        caddyfilePath,
        domain,
        upstream,
      });
      const editResult = await this.sshService.execute(server, editCommand);

      if (!editResult.success) {
        if (editResult.exitCode === 4) {
          printJson({
            status: 'error',
            code: 'NOT_FOUND',
            stderr: editResult.stderr || editResult.stdout || 'Route not found',
            command: editCommand,
          });
          return;
        }

        printSshError(editResult, editCommand);
        return;
      }

      const applyCommand = `cd ${shellQuote(caddyRoot)} && ./caddy.sh`;
      const applyResult = await this.sshService.execute(server, applyCommand);
      if (!applyResult.success) {
        printSshError(applyResult, applyCommand);
        return;
      }

      printJson({
        status: 'ok',
        message: 'Routes updated',
        data: {
          action,
          domain,
          upstream: upstream || undefined,
          caddyOutput: applyResult.stdout,
        },
      });
    } catch (error) {
      printCliError(error);
      process.exitCode = 1;
    }
  }

  private buildEditCommand(params: {
    action: Exclude<OpsRoutesAction, 'list'>;
    caddyfilePath: string;
    domain: string;
    upstream: string;
  }): string {
    const script = this.buildPythonScript(params);
    return `python3 -c ${shellQuote(script)}`;
  }

  private buildPythonScript(params: {
    action: Exclude<OpsRoutesAction, 'list'>;
    caddyfilePath: string;
    domain: string;
    upstream: string;
  }): string {
    return [
      'import json',
      'import os',
      'import random',
      'import re',
      'import sys',
      'import time',
      'from pathlib import Path',
      `file_path = ${JSON.stringify(params.caddyfilePath)}`,
      `action = ${JSON.stringify(params.action)}`,
      `domain = ${JSON.stringify(params.domain)}`,
      `upstream = ${JSON.stringify(params.upstream)}`,
      'with Path(file_path).open("r", encoding="utf-8") as handle:',
      '  text = handle.read()',
      'domain_re = re.compile(r"(^|\\n)\\s*" + re.escape(domain) + r"\\s*\\{[\\s\\S]*?\\n\\}", re.MULTILINE)',
      'if action == "add":',
      '  if domain_re.search(text):',
      '    sys.stderr.write("Route already exists for domain")',
      '    sys.exit(3)',
      '  block = "\\n" + domain + " {\\n  reverse_proxy " + upstream + "\\n}\\n"',
      '  text = re.sub(r"\\s*$", "", text) + block',
      'elif action == "remove":',
      '  if not domain_re.search(text):',
      '    sys.stderr.write("Route not found")',
      '    sys.exit(4)',
      '  text = domain_re.sub("\n", text, count=1)',
      '  text = re.sub(r"\\n{3,}", "\n\n", text)',
      '  text = re.sub(r"^\\n+", "", text)',
      'elif action == "update":',
      '  match = domain_re.search(text)',
      '  if not match:',
      '    sys.stderr.write("Route not found")',
      '    sys.exit(4)',
      '  block = match.group(0)',
      '  updated = re.sub(r"reverse_proxy\\s+[^\\n]+", "reverse_proxy " + upstream, block, count=1)',
      '  text = domain_re.sub(updated, text, count=1)',
      'tmp = os.path.join(os.path.dirname(file_path), ".tmp-" + str(int(time.time() * 1000)) + "-" + format(random.getrandbits(48), "x"))',
      'with open(tmp, "w", encoding="utf-8") as handle:',
      '  handle.write(text)',
      'os.replace(tmp, file_path)',
      'print(json.dumps({"status": "ok"}))',
    ].join('\n');
  }
}
