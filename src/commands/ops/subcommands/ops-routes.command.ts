import path from 'node:path';
import { CommandRunner, Option, SubCommand } from 'nest-commander';
import { ConfigService } from '../../../core/config/config.service';
import { SshService } from '../../../shared/ssh/ssh.service';
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
  domain?: string;
  upstream?: string;
};

@SubCommand({
  name: 'routes',
  description: 'Manage remote Caddy routes',
  arguments: '<alias> <action>',
})
export class OpsRoutesCommand extends CommandRunner {
  constructor(
    private readonly configService: ConfigService,
    private readonly sshService: SshService,
  ) {
    super();
  }

  @Option({ flags: '--domain <domain>', description: 'Domain name' })
  parseDomain(value: string): string {
    return value;
  }

  @Option({ flags: '--upstream <upstream>', description: 'Upstream host:port' })
  parseUpstream(value: string): string {
    return value;
  }

  async run(
    passedParams: string[],
    options: OpsRoutesOptions = {},
  ): Promise<void> {
    const [alias, rawAction] = passedParams;

    try {
      const action = ensureAction<OpsRoutesAction>(
        rawAction,
        ['list', 'add', 'remove', 'update'],
        'routes action',
      );
      const server = await resolveServerOrThrow(
        this.configService.getConfig(),
        alias,
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
    const script = this.buildNodeScript(params);
    return `node -e ${shellQuote(script)}`;
  }

  private buildNodeScript(params: {
    action: Exclude<OpsRoutesAction, 'list'>;
    caddyfilePath: string;
    domain: string;
    upstream: string;
  }): string {
    return [
      "const fs = require('node:fs');",
      "const p = require('node:path');",
      `const filePath = ${JSON.stringify(params.caddyfilePath)};`,
      `const action = ${JSON.stringify(params.action)};`,
      `const domain = ${JSON.stringify(params.domain)};`,
      `const upstream = ${JSON.stringify(params.upstream)};`,
      "const esc = (s) => s.replace(/[.*+?^\\$\\{\\}()|[\\]\\\\]/g, '\\\\$&');",
      "const read = fs.readFileSync(filePath, 'utf8');",
      "const domainRe = new RegExp('(^|\\n)\\\\s*' + esc(domain) + '\\s*\\\\{[\\\\s\\\\S]*?\\\\n\\\\}', 'm');",
      'let text = read;',
      "if (action === 'add') {",
      '  if (domainRe.test(text)) {',
      "    process.stderr.write('Route already exists for domain');",
      '    process.exit(3);',
      '  }',
      "  const block = '\\n' + domain + ' {\\n  reverse_proxy ' + upstream + '\\n}\\n';",
      '  text = text.replace(/\\s*$/g, "") + block;',
      "} else if (action === 'remove') {",
      '  if (!domainRe.test(text)) {',
      "    process.stderr.write('Route not found');",
      '    process.exit(4);',
      '  }',
      "  text = text.replace(domainRe, '\\n').replace(/\\n{3,}/g, '\\n\\n').replace(/^\\n+/, '');",
      "} else if (action === 'update') {",
      '  const match = text.match(domainRe);',
      '  if (!match) {',
      "    process.stderr.write('Route not found');",
      '    process.exit(4);',
      '  }',
      '  const block = match[0];',
      "  const updated = block.replace(/reverse_proxy\\s+[^\\n]+/, 'reverse_proxy ' + upstream);",
      '  text = text.replace(domainRe, updated);',
      '}',
      "const tmp = p.join(p.dirname(filePath), '.tmp-' + Date.now() + '-' + Math.random().toString(36).slice(2));",
      "fs.writeFileSync(tmp, text, 'utf8');",
      'fs.renameSync(tmp, filePath);',
      "process.stdout.write(JSON.stringify({ status: 'ok' }));",
    ].join('\n');
  }
}
