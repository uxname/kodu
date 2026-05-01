import { Module } from '@nestjs/common';
import { DepsService } from './deps.service';

@Module({
  providers: [DepsService],
  exports: [DepsService],
})
export class DepsModule {}
