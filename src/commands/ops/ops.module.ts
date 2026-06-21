import { Module } from '@nestjs/common';
import { OpsCommand } from './ops.command';
import { OpsAddCommand } from './ops-add.command';
import { OpsInitCommand } from './ops-init.command';
import { OpsListCommand } from './ops-list.command';
import { OpsPathCommand } from './ops-path.command';
import { OpsRunbookCommand } from './ops-runbook.command';
import { OpsStatusCommand } from './ops-status.command';
import { OpsUseCommand } from './ops-use.command';

@Module({
  providers: [
    OpsCommand,
    OpsInitCommand,
    OpsListCommand,
    OpsAddCommand,
    OpsStatusCommand,
    OpsUseCommand,
    OpsPathCommand,
    OpsRunbookCommand,
  ],
})
export class OpsModule {}
