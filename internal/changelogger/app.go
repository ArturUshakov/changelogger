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

	branchName, err := app.askBranchName(config.BranchPrefix)
	if err != nil {
		return err
	}

	if err := app.git.CreateBranch(branchName); err != nil {
		return err
	}

	app.printColored(fmt.Sprintf("Текущая версия приложения: %s\n", version.String()), yellow)
	app.print("Какую версию нужно поднять?\n 1 - major (*.0.0)\n 2 - minor (0.*.0)\n 3 - fix   (0.0.*)  - ")

	level, err := app.readLine()
	if err != nil {
		return err
	}

	newVersion, err := version.Next(level)
	if err != nil {
		return err
	}

	app.printColored(fmt.Sprintf("Следующая версия приложения: %s \n", newVersion.String()), green)

	commitLines, err := app.git.ChangeLines(version.String(), masterCommit)
	if err != nil {
		return err
	}

	changelog := NewChangelog(config, app.now)
	answerBody := changelog.Body(commitLines)
	if answerBody == "" {
		return fmt.Errorf("отсутствуют коммиты с нужными тэгами")
	}

	app.printColored("Изменения которые попадут в CHANGELOG.md: \n"+answerBody+" \n", yellow)

	if err := app.askConfirmation("Все верно?"); err != nil {
		return err
	}

	if err := changelog.Write(config.ChangelogPath, newVersion.String(), answerBody); err != nil {
		return err
	}
	app.printColored("Файл CHANGELOG.md успешно отредактирован:  \n", green)

	if err := app.askConfirmation("Создать коммит?"); err != nil {
		return err
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

func (app App) askBranchName(prefix string) (string, error) {
	app.printColored(`Введите номер заявки и задачи (в формате Заявка-Задача, например "IU888000-W0999000"): `, yellow)

	input, err := app.readLine()
	if err != nil {
		return "", err
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
	app.print(question + " (y/n): ")

	answer, err := app.readLine()
	if err != nil {
		return err
	}

	switch strings.ToLower(answer) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("выполнение команды отменено")
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
