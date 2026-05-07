package changelogger

import (
	"fmt"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(name string, args ...string) (string, error)
}

type OSRunner struct{}

func (OSRunner) Run(name string, args ...string) (string, error) {
	command := exec.Command(name, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}

	return string(output), nil
}

type Git struct {
	runner Runner
}

func (git Git) LastTag() (string, error) {
	commit, err := git.runner.Run("git", "rev-list", "--tags", "--max-count=1")
	if err != nil {
		return "", fmt.Errorf("получить последний commit тэга: %w", err)
	}

	commit = strings.TrimSpace(commit)
	if commit == "" {
		return "", fmt.Errorf("отсутствует последний тэг")
	}

	tag, err := git.runner.Run("git", "describe", "--tags", commit)
	if err != nil {
		return "", fmt.Errorf("получить последний тэг: %w", err)
	}

	tag = strings.TrimSpace(tag)
	if tag == "" {
		return "", fmt.Errorf("отсутствует последний тэг")
	}

	return tag, nil
}

func (git Git) MasterCommit() (string, error) {
	commit, err := git.runner.Run("git", "rev-parse", "origin/master")
	if err != nil {
		return "", fmt.Errorf("получить commit origin/master: %w", err)
	}

	return strings.TrimSpace(commit), nil
}

func (git Git) ChangeLines(lastTag string, masterCommit string) ([]string, error) {
	logOutput, err := git.runner.Run(
		"git",
		"log",
		"--pretty=format:%h|%an|%s|%cs",
		"--no-merges",
		lastTag+"..develop",
	)
	if err != nil {
		return nil, fmt.Errorf("получить список коммитов: %w", err)
	}

	lines := splitLines(logOutput)

	cherryOutput, err := git.runner.Run("git", "cherry", "-v", masterCommit, lastTag)
	if err == nil {
		lines = append(lines, splitLines(cherryOutput)...)
	}

	return lines, nil
}

func (git Git) CreateBranch(branchName string) error {
	output, err := git.runner.Run("git", "branch", "--list", branchName)
	if err != nil {
		return fmt.Errorf("проверить существование ветки: %w", err)
	}

	if strings.TrimSpace(output) != "" {
		_, err := git.runner.Run("git", "checkout", branchName)
		if err != nil {
			return fmt.Errorf("переключиться на существующую ветку: %w", err)
		}

		return nil
	}

	if _, err := git.runner.Run("git", "checkout", "develop"); err != nil {
		return fmt.Errorf("переключиться на develop: %w", err)
	}

	if _, err := git.runner.Run("git", "checkout", "-b", branchName); err != nil {
		return fmt.Errorf("создать ветку %s: %w", branchName, err)
	}

	return nil
}

func (git Git) Commit(changelogPath string) error {
	if _, err := git.runner.Run("git", "add", changelogPath); err != nil {
		return fmt.Errorf("добавить CHANGELOG.md в индекс: %w", err)
	}

	if _, err := git.runner.Run("git", "commit", "-m", "wip: Отредактирован CHANGELOG.md"); err != nil {
		return fmt.Errorf("создать commit: %w", err)
	}

	return nil
}

func (git Git) Push(branchName string) error {
	if _, err := git.runner.Run("git", "push", "origin", branchName); err != nil {
		return fmt.Errorf("push ветки %s: %w", branchName, err)
	}

	return nil
}

func (git Git) DeleteBranch(branchName string) error {
	if _, err := git.runner.Run("git", "checkout", "develop"); err != nil {
		return fmt.Errorf("переключиться на develop: %w", err)
	}

	if _, err := git.runner.Run("git", "branch", "-D", branchName); err != nil {
		return fmt.Errorf("удалить локальную ветку %s: %w", branchName, err)
	}

	return nil
}

func splitLines(output string) []string {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}

	return strings.Split(output, "\n")
}
