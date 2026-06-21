import { writeFile } from 'node:fs/promises';
import { execa } from 'execa';
import { configSchema } from '../src/core/config/config.schema';
import { registrySchema } from '../src/core/registry/registry.schema';

async function main(): Promise<void> {
  await writeFile(
    'kodu.schema.json',
    JSON.stringify(configSchema.toJSONSchema(), null, 2),
    'utf8',
  );
  await execa('biome', ['format', '--write', 'kodu.schema.json']);
  console.log('✅ JSON schema generated: kodu.schema.json');

  await writeFile(
    'registry.schema.json',
    JSON.stringify(registrySchema.toJSONSchema(), null, 2),
    'utf8',
  );
  await execa('biome', ['format', '--write', 'registry.schema.json']);
  console.log('✅ JSON schema generated: registry.schema.json');
}

main().catch((error: unknown) => {
  console.error(error);
  process.exitCode = 1;
});
