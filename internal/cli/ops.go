package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/uxname/kodu/internal/registry"
	"github.com/uxname/kodu/internal/runbook"
	"github.com/uxname/kodu/internal/sortx"
)

func newOpsCommand(app *App) *cobra.Command {
	ops := &cobra.Command{
		Use:   "ops",
		Short: "Работа с проектами и стендами (local/dev/stage/prod) из любого места",
	}
	ops.AddCommand(
		newOpsInitCommand(app),
		newOpsListCommand(app),
		newOpsAddCommand(app),
		newOpsStatusCommand(app),
		newOpsUseCommand(app),
		newOpsPathCommand(app),
		newOpsRunbookCommand(app),
	)
	return ops
}

// completeProjectNames предлагает имена проектов из реестра для автодополнения.
func completeProjectNames(toComplete string) ([]string, cobra.ShellCompDirective) {
	projects, err := registry.New().List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var names []string
	for n := range projects {
		if strings.HasPrefix(n, toComplete) {
			names = append(names, n)
		}
	}
	sortx.LocaleStrings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeFirstArgProjects дополняет имя проекта только для первого аргумента.
func completeFirstArgProjects(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeProjectNames(toComplete)
}

// resolveProjectRoot возвращает путь репозитория проекта по имени.
func resolveProjectRoot(reg *registry.Service, name string) (string, error) {
	entry, ok, err := reg.Get(name)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("Проект %q не найден в реестре. Список проектов: kodu ops list", name)
	}
	return entry.Path, nil
}

func fail(app *App, msg string) {
	app.UI.Error(msg)
	os.Exit(1)
}

// --- ops init ---

func newOpsInitCommand(app *App) *cobra.Command {
	var name, active string
	var stand []string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Настроить стенды в текущем проекте: создать .runbook/, .gitignore и зарегистрировать проект",
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			reg, rb := registry.New(), runbook.New()

			projectName := name
			if projectName == "" {
				projectName = filepath.Base(cwd)
			}
			stands := stand
			if len(stands) == 0 {
				stands = registry.DefaultStands()
			}
			activeStand := active
			if activeStand == "" && len(stands) > 0 {
				activeStand = stands[0]
			}
			if activeStand == "" {
				activeStand = "local"
			}

			if err := rb.Scaffold(projectName, stands, activeStand, cwd); err != nil {
				fail(app, err.Error())
			}
			app.UI.Success(fmt.Sprintf("Создан .runbook/ для проекта %q.", projectName))
			app.UI.Info("Активный стенд: " + activeStand)
			app.UI.Info("Заполни шаги деплоя в " + rb.RunbookPath(cwd))

			res, err := rb.EnsureGitignore(cwd)
			if err != nil {
				fail(app, err.Error())
			}
			reportGitignore(app, res)

			registerProject(app, reg, projectName, cwd, stands)
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Уникальное имя проекта (по умолчанию — имя папки)")
	cmd.Flags().StringVarP(&active, "active", "a", "", "Активный стенд по умолчанию (по умолчанию local)")
	cmd.Flags().StringArrayVarP(&stand, "stand", "s", nil, "Стенд проекта (можно повторять)")
	return cmd
}

func reportGitignore(app *App, res runbook.GitignoreResult) {
	switch res {
	case runbook.GitignoreCreated:
		app.UI.Success("Создан .gitignore с записью /.runbook/")
	case runbook.GitignoreAdded:
		app.UI.Success("Добавил /.runbook/ в .gitignore")
	case runbook.GitignorePresent:
		app.UI.Info("/.runbook/ уже в .gitignore")
	case runbook.GitignoreNoGit:
		app.UI.Warn("Это не git-репозиторий — .gitignore не настроен. Не коммить .runbook/ вручную.")
	}
}

func registerProject(app *App, reg *registry.Service, name, root string, stands []string) {
	existing, ok, err := reg.Get(name)
	if err != nil {
		fail(app, err.Error())
	}
	if !ok {
		if err := reg.Add(name, registry.ProjectEntry{Path: root, Stands: stands}, false); err != nil {
			fail(app, err.Error())
		}
		app.UI.Success(fmt.Sprintf("Проект %q добавлен в реестр.", name))
		return
	}
	if existing.Path == root {
		app.UI.Info(fmt.Sprintf("Проект %q уже в реестре.", name))
		return
	}
	app.UI.Warn(fmt.Sprintf("Имя %q уже занято другим путём (%s). Запусти заново с другим именем: kodu ops init --name <другое-имя>", name, existing.Path))
}

// --- ops list ---

func newOpsListCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Показать все проекты из реестра и их активные стенды",
		RunE: func(_ *cobra.Command, _ []string) error {
			reg, rb := registry.New(), runbook.New()
			projects, err := reg.List()
			if err != nil {
				fail(app, err.Error())
			}
			names := make([]string, 0, len(projects))
			for n := range projects {
				names = append(names, n)
			}
			sortx.LocaleStrings(names)

			if len(names) == 0 {
				app.UI.Info("Реестр пуст. Добавь проект: kodu ops add <name> --path <dir>")
				app.UI.Info("Файл реестра: " + reg.FilePath())
				return nil
			}
			for _, n := range names {
				entry := projects[n]
				label := ""
				if active := readActiveStand(rb, entry.Path); active != "" {
					label = " [активный: " + active + "]"
				}
				app.UI.Info(n + label)
				app.UI.Info("  путь:   " + entry.Path)
				app.UI.Info("  стенды: " + strings.Join(entry.Stands, ", "))
			}
			return nil
		},
	}
}

func readActiveStand(rb *runbook.Service, root string) string {
	if !rb.Exists(root) {
		return ""
	}
	cfg, err := rb.ReadConfig(root)
	if err != nil {
		return ""
	}
	return cfg.ActiveStand
}

// --- ops add ---

func newOpsAddCommand(app *App) *cobra.Command {
	var path, repo string
	var stand []string
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Зарегистрировать проект в глобальном реестре по уникальному имени",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				fail(app, "Укажи имя проекта: kodu ops add <name> [--path <dir>]")
			}
			name := args[0]
			reg := registry.New()

			projectPath := path
			if projectPath == "" {
				projectPath, _ = os.Getwd()
			}
			projectPath, _ = filepath.Abs(projectPath)
			stands := stand
			if len(stands) == 0 {
				stands = registry.DefaultStands()
			}
			if err := reg.Add(name, registry.ProjectEntry{Path: projectPath, Repo: repo, Stands: stands}, false); err != nil {
				fail(app, err.Error())
			}
			app.UI.Success(fmt.Sprintf("Проект %q добавлен в реестр.", name))
			app.UI.Info("Путь: " + projectPath)
			app.UI.Info("Стенды: " + strings.Join(stands, ", "))
			return nil
		},
	}
	cmd.Flags().StringVarP(&path, "path", "p", "", "Путь к репозиторию (по умолчанию текущая директория)")
	cmd.Flags().StringVarP(&repo, "repo", "r", "", "URL репозитория (git clone)")
	cmd.Flags().StringArrayVarP(&stand, "stand", "s", nil, "Стенд проекта (можно повторять)")
	return cmd
}

// --- ops status ---

func newOpsStatusCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "status [name]",
		Short:             "Показать активный стенд и стенды проекта (по имени или в текущей папке)",
		ValidArgsFunction: completeFirstArgProjects,
		RunE: func(_ *cobra.Command, args []string) error {
			reg, rb := registry.New(), runbook.New()
			root, _ := os.Getwd()
			if len(args) > 0 {
				r, err := resolveProjectRoot(reg, args[0])
				if err != nil {
					fail(app, err.Error())
				}
				root = r
			}
			if !rb.Exists(root) {
				app.UI.Warn(fmt.Sprintf("В %s нет .runbook/. Инициализируй проект: kodu ops init", root))
				os.Exit(1)
			}
			cfg, err := rb.ReadConfig(root)
			if err != nil {
				fail(app, err.Error())
			}
			app.UI.Info("Проект:        " + cfg.Project)
			app.UI.Info("Активный стенд: " + cfg.ActiveStand)
			app.UI.Info("Стенды:        " + strings.Join(cfg.Stands, ", "))
			app.UI.Info("Путь:          " + root)
			return nil
		},
	}
}

// --- ops use ---

func newOpsUseCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "use <args...>",
		Aliases:           []string{"switch"},
		Short:             "Переключить активный стенд: kodu ops use <stand> или kodu ops use <name> <stand>",
		ValidArgsFunction: completeFirstArgProjects,
		RunE: func(_ *cobra.Command, args []string) error {
			reg, rb := registry.New(), runbook.New()
			root, _ := os.Getwd()
			var stand string
			if len(args) >= 2 {
				r, err := resolveProjectRoot(reg, args[0])
				if err != nil {
					fail(app, err.Error())
				}
				root, stand = r, args[1]
			} else if len(args) == 1 {
				stand = args[0]
			}

			if stand == "" {
				fail(app, "Укажи стенд: kodu ops use <stand> или kodu ops use <name> <stand>")
			}
			if !rb.Exists(root) {
				app.UI.Warn(fmt.Sprintf("В %s нет .runbook/. Инициализируй проект: kodu ops init", root))
				os.Exit(1)
			}
			cfg, err := rb.ReadConfig(root)
			if err != nil {
				fail(app, err.Error())
			}
			if !contains(cfg.Stands, stand) {
				cfg.Stands = append(cfg.Stands, stand)
				app.UI.Info(fmt.Sprintf("Стенд %q добавлен в список стендов проекта.", stand))
			}
			cfg.ActiveStand = stand
			if err := rb.WriteConfig(cfg, root); err != nil {
				fail(app, err.Error())
			}
			app.UI.Success(fmt.Sprintf("Активный стенд проекта %q → %s", cfg.Project, stand))
			return nil
		},
	}
}

// --- ops path ---

func newOpsPathCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "path <name>",
		Short:             "Напечатать путь к репозиторию проекта (удобно для cd $(kodu ops path <name>))",
		ValidArgsFunction: completeFirstArgProjects,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				fail(app, "Укажи имя проекта: kodu ops path <name>")
			}
			reg := registry.New()
			root, err := resolveProjectRoot(reg, args[0])
			if err != nil {
				fail(app, err.Error())
			}
			app.UI.Println(root) // чистый stdout для подстановки команд
			return nil
		},
	}
}

// --- ops runbook ---

func newOpsRunbookCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "runbook <name> [stand]",
		Short:             "Напечатать runbook проекта (или секцию конкретного стенда)",
		ValidArgsFunction: completeFirstArgProjects,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				fail(app, "Укажи имя проекта: kodu ops runbook <name> [stand]")
			}
			name := args[0]
			var stand string
			if len(args) > 1 {
				stand = args[1]
			}
			reg, rb := registry.New(), runbook.New()
			root, err := resolveProjectRoot(reg, name)
			if err != nil {
				fail(app, err.Error())
			}
			if !rb.Exists(root) {
				app.UI.Warn(fmt.Sprintf("В проекте %q нет .runbook/. Запусти: kodu ops init", name))
				os.Exit(1)
			}
			md, err := rb.ReadRunbook(root)
			if err != nil {
				fail(app, err.Error())
			}
			output := md
			if stand != "" {
				output = extractStand(md, stand)
				if output == "" {
					app.UI.Warn(fmt.Sprintf("Секция для стенда %q не найдена в runbook.", stand))
					os.Exit(1)
				}
			}
			app.UI.Println(strings.TrimRight(output, "\n \t"))
			return nil
		},
	}
}

// extractStand возвращает блок `## Стенд: <stand> ...` до следующего `## `.
func extractStand(md, stand string) string {
	lines := strings.Split(md, "\n")
	prefix := "## Стенд: " + stand
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			start = i
			break
		}
	}
	if start == -1 {
		return ""
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
