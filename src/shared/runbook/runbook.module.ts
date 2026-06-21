import { Global, Module } from '@nestjs/common';
import { RunbookService } from './runbook.service';

@Global()
@Module({
  providers: [RunbookService],
  exports: [RunbookService],
})
export class RunbookModule {}
