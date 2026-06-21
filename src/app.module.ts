import { Module } from '@nestjs/common';
import { CleanModule } from './commands/clean/clean.module';
import { InitModule } from './commands/init/init.module';
import { OpsModule } from './commands/ops/ops.module';
import { PackModule } from './commands/pack/pack.module';
import { ConfigModule } from './core/config/config.module';
import { FsModule } from './core/file-system/fs.module';
import { RegistryModule } from './core/registry/registry.module';
import { UiModule } from './core/ui/ui.module';
import { GitModule } from './shared/git/git.module';
import { RunbookModule } from './shared/runbook/runbook.module';
import { TokenizerModule } from './shared/tokenizer/tokenizer.module';

@Module({
  imports: [
    ConfigModule,
    UiModule,
    FsModule,
    GitModule,
    TokenizerModule,
    RegistryModule,
    RunbookModule,
    InitModule,
    PackModule,
    CleanModule,
    OpsModule,
  ],
})
export class AppModule {}
