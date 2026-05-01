import { Module } from '@nestjs/common';
import { UiModule } from '../../core/ui/ui.module';
import { InitCommand } from './init.command';

@Module({
  imports: [UiModule],
  providers: [InitCommand],
})
export class InitModule {}
