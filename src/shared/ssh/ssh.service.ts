import { Injectable } from '@nestjs/common';
import { execa } from 'execa';
import type { ServerConfig } from '../../core/config/config.schema';

export type SshResult = {
  success: boolean;
  stdout: string;
  stderr: string;
  exitCode: number;
  error?: string;
};

@Injectable()
export class SshService {
  async execute(
    serverConfig: ServerConfig,
    command: string,
  ): Promise<SshResult> {
    const args = [
      '-i',
      serverConfig.sshKeyPath,
      '-p',
      String(serverConfig.port ?? 22),
      '-o',
      'StrictHostKeyChecking=no',
      '-o',
      'ConnectTimeout=10',
      `${serverConfig.user}@${serverConfig.host}`,
      command,
    ];

    try {
      const { stdout, stderr, exitCode } = await execa('ssh', args, {
        env: serverConfig.env,
      });

      return {
        success: (exitCode ?? 0) === 0,
        stdout,
        stderr,
        exitCode: exitCode ?? 0,
      };
    } catch (error) {
      const failure = error as {
        stdout?: string;
        stderr?: string;
        exitCode?: number;
        shortMessage?: string;
        message?: string;
      };

      return {
        success: false,
        stdout: failure.stdout ?? '',
        stderr: failure.stderr ?? '',
        exitCode: failure.exitCode ?? -1,
        error: failure.shortMessage ?? failure.message ?? 'Unknown SSH error',
      };
    }
  }
}
