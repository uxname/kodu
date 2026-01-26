import { Injectable } from '@nestjs/common';
import { lilconfigSync } from 'lilconfig';
import pc from 'picocolors';
import { configSchema, type KoduConfig } from './config.schema';

@Injectable()
export class ConfigService {
  private config?: KoduConfig;

  getConfig(): KoduConfig {
    if (!this.config) {
      this.config = this.loadConfig();
    }

    return this.config;
  }

  private loadConfig(): KoduConfig {
    const explorer = lilconfigSync('kodu', { searchPlaces: ['kodu.json'] });
    const result = explorer.search(process.cwd());

    if (!result || result.isEmpty || !result.config) {
      this.terminate(
        'kodu.json config not found. Run `kodu init` to create the file.',
      );
    }

    const parsed = configSchema.safeParse(result.config);

    if (!parsed.success) {
      console.error(pc.red('kodu.json config is invalid:'));
      parsed.error.issues.forEach((issue) => {
        const path = issue.path.join('.') || '(root)';
        console.error(pc.red(`- ${path}: ${issue.message}`));
      });
      this.terminate('Fix the config and run the command again.');
    }

    return parsed.data;
  }

  private terminate(message: string): never {
    console.error(pc.red(message));
    process.exit(1);
  }
}
