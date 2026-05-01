import { Module } from '@nestjs/common';
import { ConfigModule } from '../../core/config/config.module';
import { FsModule } from '../../core/file-system/fs.module';
import { UiModule } from '../../core/ui/ui.module';
import { CleanerService } from '../../shared/cleaner/cleaner.service';
import { DepsModule } from '../../shared/deps/deps.module';
import { TokenizerModule } from '../../shared/tokenizer/tokenizer.module';
import { PackCommand } from './pack.command';

@Module({
  imports: [ConfigModule, UiModule, FsModule, TokenizerModule, DepsModule],
  providers: [PackCommand, CleanerService],
})
export class PackModule {}
