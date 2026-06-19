package changelogger

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

type App struct {
	args   []string
	input  *bufio.Reader
	output io.Writer
	now    func() time.Time
	git    Git
}

func NewApp(args []string, input io.Reader, output io.Writer, now func() time.Time, runner Runner) App {
	return App{
		args:   args,
		input:  bufio.NewReader(input),
		output: output,
		now:    now,
		git:    Git{runner: runner},
	}
}

func (app App) Run() error {
	config, err := LoadConfig(".env")
	if err != nil {
		return err
	}

	if len(app.args) > 0 {
		config.RepositoryLink = app.args[0]
	}

	if config.RepositoryLink == "" {
		repositoryLink, err := app.git.RepositoryLink()
		if err != nil {
			return err
		}

		config.RepositoryLink = repositoryLink
	}

	lastTag, err := app.git.LastTag()
	if err != nil {
		return err
	}

	version, err := ParseVersion(lastTag)
	if err != nil {
		return err
	}

	masterCommit, err := app.git.MasterCommit()
	if err != nil {
		return err
	}

	commitLines, err := app.git.ChangeLines(version.String(), masterCommit)
	if err != nil {
		return err
	}

	if config.BranchPrefix == "" {
		config.BranchPrefix = DetectBranchPrefix(commitLines)
	}
	branchTask := DetectBranchTask(commitLines)

	includeTaskLinks, err := app.taskLinkChoice(config.TaskLinkMode)
	if err != nil {
		return err
	}
	if !includeTaskLinks {
		config.TaskSystemLink = ""
	}

	changelog := NewChangelog(config, app.now)
	answerBody := changelog.Body(commitLines)
	if answerBody == "" {
		return fmt.Errorf("отсутствуют коммиты с нужными тэгами")
	}

	app.printColored("Изменения которые попадут в CHANGELOG.md: \n"+answerBody+" \n", yellow)

	recommendedLevel, recommendationReason := RecommendVersionLevel(commitLines)
	level, err := app.askVersionLevel(version, recommendedLevel, recommendationReason)
	if err != nil {
		return err
	}

	newVersion, err := version.Next(level)
	if err != nil {
		return err
	}

	app.printColored(fmt.Sprintf("Следующая версия приложения: %s \n", newVersion.String()), green)

	branchName, err := app.askBranchName(config.BranchPrefix, branchTask)
	if err != nil {
		return err
	}

	if err := app.git.CreateBranch(branchName); err != nil {
		return err
	}

	if err := app.askConfirmation("Все верно?"); err != nil {
		return err
	}

	if err := changelog.Write(config.ChangelogPath, newVersion.String(), answerBody); err != nil {
		return err
	}
	app.printColored("Файл CHANGELOG.md успешно отредактирован:  \n", green)

	shouldCommit, err := app.commitChoice(config.CommitMode)
	if err != nil {
		return err
	}
	if !shouldCommit {
		return nil
	}

	if err := app.git.Commit(config.ChangelogPath); err != nil {
		return err
	}
	app.printColored("Коммит успешно создан!  \n", green)

	if err := app.askConfirmation("Пушить ветку " + branchName + "?"); err != nil {
		return err
	}

	if err := app.git.Push(branchName); err != nil {
		return err
	}

	return app.git.DeleteBranch(branchName)
}

func (app App) askVersionLevel(version Version, recommendedLevel string, recommendationReason string) (string, error) {
	app.printColored(fmt.Sprintf("Текущая версия приложения: %s\n", version.String()), yellow)
	app.printColored(fmt.Sprintf("Рекомендация: %s (%s)\n", recommendedLevel, recommendationReason), green)
	app.print("Какую версию нужно поднять?\n 1 - major (*.0.0)\n 2 - minor (0.*.0)\n 3 - fix   (0.0.*)\n Enter - " + recommendedLevel + "\n> ")

	level, err := app.readLine()
	if err != nil {
		return "", err
	}
	if level == "" {
		return recommendedLevel, nil
	}

	return level, nil
}

func (app App) askBranchName(prefix string, defaultTask string) (string, error) {
	prompt := `Введите номер заявки и задачи (в формате Заявка-Задача, например "IU888000-W0999000")`
	if defaultTask != "" {
		prompt += "\nEnter - " + defaultTask
	}
	prompt += ": "
	app.printColored(prompt, yellow)

	input, err := app.readLine()
	if err != nil {
		return "", err
	}
	if input == "" && defaultTask != "" {
		input = defaultTask
	}

	if !validBranchTask(input) {
		return "", fmt.Errorf(`неверный формат. Ожидался формат Заявка-Задача, например "IU888000-W0999000"`)
	}

	if prefix != "" {
		prefix += "-"
	}

	return "feature/" + prefix + input + "-assign-to-changelog", nil
}

func (app App) askConfirmation(question string) error {
	confirmed, err := app.askYesNo(question)
	if err != nil {
		return err
	}

	if confirmed {
		return nil
	}

	return fmt.Errorf("выполнение команды отменено")
}

func (app App) taskLinkChoice(mode string) (bool, error) {
	switch normalizePreferenceMode(mode) {
	case preferenceAlways:
		return true, nil
	case preferenceNever:
		return false, nil
	}

	enabled, err := app.askYesNoWithDefaultNo("Добавлять ссылку на заявку в CHANGELOG.md?")
	if err != nil {
		return false, err
	}

	saveMode, err := app.askPreferenceMode(
		"Запомнить выбор для ссылок на заявку?",
		"всегда добавлять",
		"никогда не добавлять",
	)
	if err != nil {
		return false, err
	}
	if saveMode != preferenceAsk {
		if err := saveGlobalConfig(func(config *Config) {
			config.TaskLinkMode = saveMode
		}); err != nil {
			return false, err
		}
	}

	return enabled, nil
}

func (app App) commitChoice(mode string) (bool, error) {
	switch normalizePreferenceMode(mode) {
	case preferenceAlways:
		return true, nil
	case preferenceNever:
		return false, nil
	}

	enabled, err := app.askYesNo("Создать коммит?")
	if err != nil {
		return false, err
	}

	saveMode, err := app.askPreferenceMode(
		"Запомнить выбор для коммита?",
		"всегда создавать",
		"никогда не создавать",
	)
	if err != nil {
		return false, err
	}
	if saveMode != preferenceAsk {
		if err := saveGlobalConfig(func(config *Config) {
			config.CommitMode = saveMode
		}); err != nil {
			return false, err
		}
	}

	return enabled, nil
}

func (app App) askYesNo(question string) (bool, error) {
	app.print(question + " (y/n): ")

	return app.readYesNo()
}

func (app App) askYesNoWithDefaultNo(question string) (bool, error) {
	app.print(question + " (y/n, Enter - n): ")

	return app.readYesNo()
}

func (app App) readYesNo() (bool, error) {
	answer, err := app.readLine()
	if err != nil {
		return false, err
	}

	switch strings.ToLower(answer) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func (app App) askPreferenceMode(question string, alwaysLabel string, neverLabel string) (string, error) {
	app.print(question + "\n 1 - " + alwaysLabel + "\n 2 - " + neverLabel + "\n Enter - спрашивать каждый раз\n> ")

	answer, err := app.readLine()
	if err != nil {
		return "", err
	}

	switch strings.ToLower(answer) {
	case "1", preferenceAlways:
		return preferenceAlways, nil
	case "2", preferenceNever:
		return preferenceNever, nil
	default:
		return preferenceAsk, nil
	}
}

func (app App) readLine() (string, error) {
	line, err := app.input.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

func (app App) print(message string) {
	fmt.Fprint(app.output, message)
}

func (app App) printColored(message string, color string) {
	fmt.Fprintf(app.output, "\033[01;%sm%s\033[0m", color, message)
}

const (
	green  = "32"
	yellow = "33"
)

const (
	preferenceAsk    = "ask"
	preferenceAlways = "always"
	preferenceNever  = "never"
)
