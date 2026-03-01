#!/usr/bin/env node

import { CommandFactory } from 'nest-commander';
import packageJson from '../package.json';
import { AppModule } from './app.module';

async function bootstrap() {
  await CommandFactory.run(AppModule, {
    version: packageJson.version,
  });
}
bootstrap();
