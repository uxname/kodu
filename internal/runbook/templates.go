package runbook

import "strings"

// DetectedStack is what could be determined about the project (for hints in the runbook).
type DetectedStack struct {
	Compose    bool
	Dockerfile bool
	EnvExample bool
}

var standTitles = map[string]string{
	"local": "local development",
	"dev":   "dev stand",
	"stage": "stage stand",
	"prod":  "production (be careful!)",
}

func renderStandSection(stand string, d DetectedStack) string {
	title, ok := standTitles[stand]
	if !ok {
		title = stand
	}

	startCmd := "<start command, e.g. docker compose up -d>"
	logsCmd := "<logs command>"
	if d.Compose {
		startCmd = "docker compose up -d"
		logsCmd = "docker compose logs -f"
	}
	envNote := "<where to get environment variables / secrets>"
	if d.EnvExample {
		envNote = "Copy `.env.example` → `.env` and fill in the values."
	}

	return strings.Join([]string{
		"## Stand: " + stand + " (" + title + ")",
		"",
		"- **Where it lives / access**: <ssh user@host or localhost>",
		"- **Working directory**: <path on the server or locally>",
		"- **Get the code**: `git clone <repo>` (first time) / `git pull` (to update)",
		"- **Start**: `" + startCmd + "`",
		"- **Logs**: `" + logsCmd + "`",
		"- **Deploy**: <step-by-step deploy commands for this stand>",
		"- **Rollback**: <how to roll back if something goes wrong>",
		"- **Environment variables / secrets**: " + envNote,
		"",
	}, "\n")
}

// RenderRunbook returns the project's starter runbook.
func RenderRunbook(project string, stands []string, d DetectedStack) string {
	header := strings.Join([]string{
		"# Runbook: " + project,
		"",
		"> This file describes how to work with the project's stands.",
		"> It is listed in `.gitignore` and is not committed — it may contain hosts and paths.",
		"> Fill in the `<...>` placeholders for your own infrastructure.",
		"",
	}, "\n")

	sections := make([]string, len(stands))
	for i, s := range stands {
		sections[i] = renderStandSection(s, d)
	}
	return header + "\n" + strings.Join(sections, "\n")
}
