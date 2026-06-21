import { Global, Module } from '@nestjs/common';
import { RegistryService } from './registry.service';

@Global()
@Module({
  providers: [RegistryService],
  exports: [RegistryService],
})
export class RegistryModule {}
