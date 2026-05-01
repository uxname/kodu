import { Module } from '@nestjs/common';
import { CleanModule } from './commands/clean/clean.module';
import { InitModule } from './commands/init/init.module';
import { PackModule } from './commands/pack/pack.module';
import { ConfigModule } from './core/config/config.module';
import { FsModule } from './core/file-system/fs.module';
import { UiModule } from './core/ui/ui.module';
import { GitModule } from './shared/git/git.module';
import { TokenizerModule } from './shared/tokenizer/tokenizer.module';

@Module({
  imports: [
    ConfigModule,
    UiModule,
    FsModule,
    GitModule,
    TokenizerModule,
    InitModule,
    PackModule,
    CleanModule,
  ],
})
export class AppModule {}
