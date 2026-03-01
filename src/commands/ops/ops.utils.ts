import { access } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import type { KoduConfig, ServerConfig } from '../../core/config/config.schema';
import type { SshResult } from '../../shared/ssh/ssh.service';
import {
  OpsCliError,
  type OpsErrorCode,
  type ResolvedServerConfig,
} from './ops.types';

const DEFAULT_APPS_PATH = '/var/agent-apps';

export function printJson(payload: unknown, isError = false): void {
  const line = JSON.stringify(payload);
  if (isError) {
    console.error(line);
    return;
  }
  console.log(line);
}

export function printCliError(error: unknown): void {
  const cliError = toCliError(error);
  printJson(
    {
      status: 'error',
      code: cliError.code,
      error: cliError.message,
    },
    true,
  );
}

export function printSshError(result: SshResult, command: string): void {
  printJson({
    status: 'error',
    code: result.exitCode,
    stderr: result.stderr || result.error || 'Unknown SSH error',
    command,
  });
}

export async function resolveServerOrThrow(
  config: KoduConfig,
  alias: string,
): Promise<ResolvedServerConfig> {
  const servers = config.ops?.servers;
  if (!servers) {
    throw new OpsCliError(
      'CONFIG_ERROR',
      'ops.servers not configured in kodu.json',
    );
  }

  const server = servers[alias];
  if (!server) {
    throw new OpsCliError(
      'VALIDATION_ERROR',
      `Server alias '${alias}' not found in kodu.json`,
    );
  }

  const resolved = normalizeServer(server);
  await assertSshKeyExists(resolved.sshKeyPath);
  return resolved;
}

export function resolveAppsPath(server: ResolvedServerConfig): string {
  return server.paths?.apps ?? DEFAULT_APPS_PATH;
}

export function resolveCaddyPath(server: ResolvedServerConfig): string {
  return (
    server.paths?.caddy ?? path.posix.join(resolveAppsPath(server), 'caddy')
  );
}

export function shellQuote(value: string): string {
  return `'${value.replace(/'/g, `'"'"'`)}'`;
}

export function ensureEnvKey(key: string | undefined): string {
  if (!key) {
    throw new OpsCliError('VALIDATION_ERROR', 'Option --key is required');
  }

  if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(key)) {
    throw new OpsCliError(
      'VALIDATION_ERROR',
      'Invalid env key format. Allowed: [A-Za-z_][A-Za-z0-9_]*',
    );
  }

  return key;
}

export function ensureRequired(
  value: string | undefined,
  name: string,
): string {
  if (!value) {
    throw new OpsCliError('VALIDATION_ERROR', `Option --${name} is required`);
  }
  return value;
}

export function ensureAction<T extends string>(
  value: string,
  allowed: readonly T[],
  context: string,
): T {
  if (!allowed.includes(value as T)) {
    throw new OpsCliError(
      'VALIDATION_ERROR',
      `Unsupported ${context}: '${value}'. Allowed: ${allowed.join(', ')}`,
    );
  }
  return value as T;
}

function normalizeServer(server: ServerConfig): ResolvedServerConfig {
  const sshKeyPath = server.sshKeyPath.startsWith('~')
    ? path.join(os.homedir(), server.sshKeyPath.slice(1))
    : server.sshKeyPath;
  const apps = server.paths?.apps ?? DEFAULT_APPS_PATH;
  return {
    ...server,
    sshKeyPath: path.isAbsolute(sshKeyPath)
      ? sshKeyPath
      : path.resolve(process.cwd(), sshKeyPath),
    paths: {
      apps,
      caddy: server.paths?.caddy,
    },
  };
}

async function assertSshKeyExists(sshKeyPath: string): Promise<void> {
  try {
    await access(sshKeyPath);
  } catch {
    throw new OpsCliError(
      'VALIDATION_ERROR',
      `SSH key file not found or inaccessible: ${sshKeyPath}`,
    );
  }
}

function toCliError(error: unknown): { code: OpsErrorCode; message: string } {
  if (error instanceof OpsCliError) {
    return { code: error.code, message: error.message };
  }

  if (error instanceof Error) {
    return { code: 'CLI_ERROR', message: error.message };
  }

  return { code: 'CLI_ERROR', message: 'Unknown CLI error' };
}
