import { Module } from '@nestjs/common';
import { ConfigModule } from '../../core/config/config.module';
import { SshModule } from '../../shared/ssh/ssh.module';
import { OpsCommand } from './ops.command';
import { OpsEnvCommand } from './subcommands/ops-env.command';
import { OpsRoutesCommand } from './subcommands/ops-routes.command';
import { OpsServiceCommand } from './subcommands/ops-service.command';
import { OpsSysinfoCommand } from './subcommands/ops-sysinfo.command';

@Module({
  imports: [ConfigModule, SshModule],
  providers: [
    OpsCommand,
    OpsSysinfoCommand,
    OpsEnvCommand,
    OpsRoutesCommand,
    OpsServiceCommand,
  ],
})
export class OpsModule {}
