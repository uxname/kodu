import { RegistryService } from '../../core/registry/registry.service';

/**
 * Возвращает путь к репозиторию проекта по его имени из реестра.
 * Бросает понятную ошибку, если проект не найден.
 */
export async function resolveProjectRoot(
  registry: RegistryService,
  name: string,
): Promise<string> {
  const entry = await registry.get(name);

  if (!entry) {
    throw new Error(
      `Проект "${name}" не найден в реестре. Список проектов: kodu ops list`,
    );
  }

  return entry.path;
}
