import type { ServerConfig } from '../../core/config/config.schema';

export type OpsErrorCode =
  | 'CLI_ERROR'
  | 'CONFIG_ERROR'
  | 'NOT_FOUND'
  | 'VALIDATION_ERROR';

export type ResolvedServerConfig = ServerConfig & {
  sshKeyPath: string;
  paths?: {
    apps: string;
    caddy?: string;
  };
};

export class OpsCliError extends Error {
  constructor(
    public readonly code: OpsErrorCode,
    message: string,
  ) {
    super(message);
  }
}
