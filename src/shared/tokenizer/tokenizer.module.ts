import { Module } from '@nestjs/common';
import { ConfigModule } from '../../core/config/config.module';
import { TokenizerService } from './tokenizer.service';

@Module({
  imports: [ConfigModule],
  providers: [TokenizerService],
  exports: [TokenizerService],
})
export class TokenizerModule {}
