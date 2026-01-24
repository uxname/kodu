import { Module } from '@nestjs/common';
import { InitModule } from './commands/init/init.module';
import { PackModule } from './commands/pack/pack.module';
import { ConfigModule } from './core/config/config.module';
import { FsModule } from './core/file-system/fs.module';
import { UiModule } from './core/ui/ui.module';
import { TokenizerModule } from './shared/tokenizer/tokenizer.module';

@Module({
  imports: [
    ConfigModule,
    UiModule,
    FsModule,
    TokenizerModule,
    InitModule,
    PackModule,
  ],
})
export class AppModule {}
