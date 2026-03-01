import { Module } from '@nestjs/common';
import { CleanModule } from './commands/clean/clean.module';
import { CommitModule } from './commands/commit/commit.module';
import { InitModule } from './commands/init/init.module';
import { OpsModule } from './commands/ops/ops.module';
import { PackModule } from './commands/pack/pack.module';
import { ReviewModule } from './commands/review/review.module';
import { ConfigModule } from './core/config/config.module';
import { FsModule } from './core/file-system/fs.module';
import { UiModule } from './core/ui/ui.module';
import { AiModule } from './shared/ai/ai.module';
import { GitModule } from './shared/git/git.module';
import { TokenizerModule } from './shared/tokenizer/tokenizer.module';

@Module({
  imports: [
    ConfigModule,
    UiModule,
    FsModule,
    GitModule,
    AiModule,
    TokenizerModule,
    InitModule,
    PackModule,
    CleanModule,
    ReviewModule,
    CommitModule,
    OpsModule,
  ],
})
export class AppModule {}
