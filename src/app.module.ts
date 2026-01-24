import { Module } from '@nestjs/common';
import { InitModule } from './commands/init/init.module';
import { ConfigModule } from './core/config/config.module';
import { UiModule } from './core/ui/ui.module';

@Module({
  imports: [ConfigModule, UiModule, InitModule],
})
export class AppModule {}
