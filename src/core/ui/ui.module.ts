import { Global, Module } from '@nestjs/common';
import { UiService } from './ui.service';

@Global()
@Module({
  providers: [UiService],
  exports: [UiService],
})
export class UiModule {}
