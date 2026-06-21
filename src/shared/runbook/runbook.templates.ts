/** Что удалось определить про проект (для подсказок в runbook). */
export type DetectedStack = {
  compose: boolean;
  dockerfile: boolean;
  envExample: boolean;
};

const STAND_TITLES: Record<string, string> = {
  local: 'локальная разработка',
  dev: 'dev-стенд',
  stage: 'stage-стенд',
  prod: 'production (осторожно!)',
};

function renderStandSection(stand: string, detected: DetectedStack): string {
  const title = STAND_TITLES[stand] ?? stand;

  const startCmd = detected.compose
    ? 'docker compose up -d'
    : '<команда запуска, напр. docker compose up -d>';
  const logsCmd = detected.compose
    ? 'docker compose logs -f'
    : '<команда логов>';
  const envNote = detected.envExample
    ? 'Скопируй `.env.example` → `.env` и заполни значения.'
    : '<откуда взять переменные окружения / секреты>';

  return [
    `## Стенд: ${stand} (${title})`,
    '',
    '- **Где живёт / доступ**: <ssh user@host или localhost>',
    '- **Рабочая директория**: <путь на сервере или локально>',
    '- **Получить код**: `git clone <repo>` (первый раз) / `git pull` (обновить)',
    `- **Запуск**: \`${startCmd}\``,
    `- **Логи**: \`${logsCmd}\``,
    '- **Деплой**: <пошаговые команды деплоя на этот стенд>',
    '- **Откат**: <как откатиться, если что-то пошло не так>',
    `- **Переменные окружения / секреты**: ${envNote}`,
    '',
  ].join('\n');
}

/**
 * Стартовый runbook для проекта. Заполняется человеком/агентом — конкретные
 * команды и пути подставляются под реальную инфраструктуру.
 */
export function renderRunbook(
  project: string,
  stands: string[],
  detected: DetectedStack,
): string {
  const header = [
    `# Runbook: ${project}`,
    '',
    '> Этот файл описывает, как работать со стендами проекта.',
    '> Он лежит в `.gitignore` и не коммитится — здесь могут быть хосты и пути.',
    '> Заполни плейсхолдеры `<...>` под свою инфраструктуру.',
    '',
  ].join('\n');

  const sections = stands
    .map((stand) => renderStandSection(stand, detected))
    .join('\n');

  return `${header}\n${sections}`;
}
