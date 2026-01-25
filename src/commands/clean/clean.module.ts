import { Module } from '@nestjs/common';
import { ConfigModule } from '../../core/config/config.module';
import { FsModule } from '../../core/file-system/fs.module';
import { UiModule } from '../../core/ui/ui.module';
import { CleanerService } from '../../shared/cleaner/cleaner.service';
import { GitModule } from '../../shared/git/git.module';
import { CleanCommand } from './clean.command';

@Module({
  imports: [FsModule, UiModule, ConfigModule, GitModule],
  providers: [CleanCommand, CleanerService],
})
export class CleanModule {}
