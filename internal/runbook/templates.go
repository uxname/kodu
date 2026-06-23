package runbook

import "strings"

// DetectedStack — что удалось определить про проект (для подсказок в runbook).
type DetectedStack struct {
	Compose    bool
	Dockerfile bool
	EnvExample bool
}

var standTitles = map[string]string{
	"local": "локальная разработка",
	"dev":   "dev-стенд",
	"stage": "stage-стенд",
	"prod":  "production (осторожно!)",
}

func renderStandSection(stand string, d DetectedStack) string {
	title, ok := standTitles[stand]
	if !ok {
		title = stand
	}

	startCmd := "<команда запуска, напр. docker compose up -d>"
	logsCmd := "<команда логов>"
	if d.Compose {
		startCmd = "docker compose up -d"
		logsCmd = "docker compose logs -f"
	}
	envNote := "<откуда взять переменные окружения / секреты>"
	if d.EnvExample {
		envNote = "Скопируй `.env.example` → `.env` и заполни значения."
	}

	return strings.Join([]string{
		"## Стенд: " + stand + " (" + title + ")",
		"",
		"- **Где живёт / доступ**: <ssh user@host или localhost>",
		"- **Рабочая директория**: <путь на сервере или локально>",
		"- **Получить код**: `git clone <repo>` (первый раз) / `git pull` (обновить)",
		"- **Запуск**: `" + startCmd + "`",
		"- **Логи**: `" + logsCmd + "`",
		"- **Деплой**: <пошаговые команды деплоя на этот стенд>",
		"- **Откат**: <как откатиться, если что-то пошло не так>",
		"- **Переменные окружения / секреты**: " + envNote,
		"",
	}, "\n")
}

// RenderRunbook возвращает стартовый runbook проекта.
func RenderRunbook(project string, stands []string, d DetectedStack) string {
	header := strings.Join([]string{
		"# Runbook: " + project,
		"",
		"> Этот файл описывает, как работать со стендами проекта.",
		"> Он лежит в `.gitignore` и не коммитится — здесь могут быть хосты и пути.",
		"> Заполни плейсхолдеры `<...>` под свою инфраструктуру.",
		"",
	}, "\n")

	sections := make([]string, len(stands))
	for i, s := range stands {
		sections[i] = renderStandSection(s, d)
	}
	return header + "\n" + strings.Join(sections, "\n")
}
