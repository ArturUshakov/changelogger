package changelogger

import (
	"fmt"
	"strings"
)

func (app App) runConfigCommand(args []string) error {
	if len(args) == 0 {
		return app.runInteractiveConfig()
	}

	switch args[0] {
	case "show":
		config, err := loadGlobalConfig()
		if err != nil {
			return err
		}

		app.printConfig(config)
		return nil
	case "task-link":
		if len(args) < 2 {
			return fmt.Errorf("укажите режим: ask, always или never")
		}

		return app.saveTaskLinkMode(args[1])
	case "commit":
		if len(args) < 2 {
			return fmt.Errorf("укажите режим: ask, always или never")
		}

		return app.saveCommitMode(args[1])
	case "changelog":
		if len(args) < 2 {
			return fmt.Errorf("укажите режим: ask, current или branch")
		}

		return app.saveChangelogMode(args[1])
	case "reset":
		return app.resetConfigPreferences()
	default:
		return fmt.Errorf("неизвестная команда config %q", args[0])
	}
}

func (app App) runInteractiveConfig() error {
	config, err := loadGlobalConfig()
	if err != nil {
		return err
	}

	app.printConfig(config)
	app.print("Что изменить?\n 1 - ссылки на заявку\n 2 - коммит\n 3 - запись changelog\n 4 - сбросить настройки\n Enter - выйти\n> ")

	answer, err := app.readLine()
	if err != nil {
		return err
	}

	switch strings.ToLower(answer) {
	case "":
		return nil
	case "1", "task-link":
		mode, err := app.askPreferenceMode("Режим ссылок на заявку:", "всегда добавлять", "никогда не добавлять")
		if err != nil {
			return err
		}

		return app.saveTaskLinkMode(mode)
	case "2", "commit":
		mode, err := app.askPreferenceMode("Режим коммита:", "всегда создавать", "никогда не создавать")
		if err != nil {
			return err
		}

		return app.saveCommitMode(mode)
	case "3", "changelog":
		mode, err := app.askChangelogPreferenceMode()
		if err != nil {
			return err
		}

		return app.saveChangelogMode(mode)
	case "4", "reset":
		return app.resetConfigPreferences()
	default:
		return fmt.Errorf("неизвестный пункт настройки %q", answer)
	}
}

func (app App) askChangelogPreferenceMode() (string, error) {
	app.print("Режим записи changelog:\n 1 - текущая ветка\n 2 - develop и новая ветка\n Enter - спрашивать каждый раз\n> ")

	answer, err := app.readLine()
	if err != nil {
		return "", err
	}

	switch strings.ToLower(answer) {
	case "1", changelogModeCurrentBranch, "current-branch":
		return changelogModeCurrentBranch, nil
	case "2", changelogModeNewBranch, "branch":
		return changelogModeNewBranch, nil
	default:
		return preferenceAsk, nil
	}
}

func (app App) printConfig(config Config) {
	app.print("Текущие настройки:\n\n")
	app.print("Ссылки на заявку: " + taskLinkModeLabel(config.TaskLinkMode) + "\n")
	app.print("Коммит: " + commitModeLabel(config.CommitMode) + "\n")
	app.print("Запись changelog: " + changelogModeLabel(config.ChangelogMode) + "\n")
}

func (app App) saveTaskLinkMode(value string) error {
	mode, err := parsePreferenceMode(value)
	if err != nil {
		return err
	}

	if err := saveGlobalConfig(func(config *Config) {
		config.TaskLinkMode = mode
	}); err != nil {
		return err
	}

	app.print("Сохранено: ссылки на заявку - " + taskLinkModeLabel(mode) + "\n")
	return nil
}

func (app App) saveCommitMode(value string) error {
	mode, err := parsePreferenceMode(value)
	if err != nil {
		return err
	}

	if err := saveGlobalConfig(func(config *Config) {
		config.CommitMode = mode
	}); err != nil {
		return err
	}

	app.print("Сохранено: коммит - " + commitModeLabel(mode) + "\n")
	return nil
}

func (app App) saveChangelogMode(value string) error {
	mode, err := parseChangelogMode(value)
	if err != nil {
		return err
	}

	if err := saveGlobalConfig(func(config *Config) {
		config.ChangelogMode = mode
	}); err != nil {
		return err
	}

	app.print("Сохранено: запись changelog - " + changelogModeLabel(mode) + "\n")
	return nil
}

func (app App) resetConfigPreferences() error {
	if err := saveGlobalConfig(func(config *Config) {
		config.TaskLinkMode = ""
		config.CommitMode = ""
		config.ChangelogMode = ""
	}); err != nil {
		return err
	}

	app.print("Настройки сброшены. Changelogger снова будет спрашивать каждый раз.\n")
	return nil
}

func (app App) runUpdateCommand() error {
	app.print("Обновляю changelogger до последнего релиза...\n")

	output, err := app.git.runner.Run("sh", "-c", updateInstallCommand)
	if output != "" {
		app.print(output)
	}
	if err != nil {
		return fmt.Errorf("обновить changelogger: %w", err)
	}

	return nil
}

func loadGlobalConfig() (Config, error) {
	path, err := globalConfigPath()
	if err != nil {
		return Config{}, err
	}

	return readGlobalConfig(path)
}

func parsePreferenceMode(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case preferenceAsk:
		return preferenceAsk, nil
	case preferenceAlways:
		return preferenceAlways, nil
	case preferenceNever:
		return preferenceNever, nil
	default:
		return "", fmt.Errorf("неизвестный режим %q, ожидается ask, always или never", value)
	}
}

func parseChangelogMode(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case preferenceAsk:
		return preferenceAsk, nil
	case changelogModeCurrentBranch, "current-branch":
		return changelogModeCurrentBranch, nil
	case changelogModeNewBranch, "branch":
		return changelogModeNewBranch, nil
	default:
		return "", fmt.Errorf("неизвестный режим %q, ожидается ask, current или branch", value)
	}
}

func taskLinkModeLabel(mode string) string {
	switch normalizePreferenceMode(mode) {
	case preferenceAlways:
		return "всегда добавлять"
	case preferenceNever:
		return "никогда не добавлять"
	default:
		return "спрашивать каждый раз"
	}
}

func commitModeLabel(mode string) string {
	switch normalizePreferenceMode(mode) {
	case preferenceAlways:
		return "всегда создавать"
	case preferenceNever:
		return "никогда не создавать"
	default:
		return "спрашивать каждый раз"
	}
}

func changelogModeLabel(mode string) string {
	switch normalizeChangelogMode(mode) {
	case changelogModeCurrentBranch:
		return "текущая ветка"
	case changelogModeNewBranch:
		return "develop и новая ветка"
	default:
		return "спрашивать каждый раз"
	}
}

const updateInstallCommand = `set -eu
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM
url="${CHANGELOGGER_INSTALL_URL:-https://github.com/SolasWyrd/changelogger/releases/latest/download/changelogger-install}"
installer="$tmp_dir/changelogger-install"

if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$installer"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$installer"
else
    printf '%s\n' "Ошибка: нужен curl или wget для скачивания installer" >&2
    exit 1
fi

sh "$installer"
`
