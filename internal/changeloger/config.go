package changeloger

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	RepositoryLink string
	ChangelogPath  string
	BranchPrefix   string
	TaskSystemName string
	TaskSystemLink string
}

func LoadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("файл .env не существует")
		}

		return Config{}, fmt.Errorf("прочитать .env: %w", err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		values[strings.TrimSpace(key)] = trimEnvValue(value)
	}

	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("прочитать .env: %w", err)
	}

	return Config{
		RepositoryLink: valueOrDefault(values["REPOSITORY_LINK"], "http://"),
		ChangelogPath:  valueOrDefault(values["CHANGELOG_PATH"], "./CHANGELOG.md"),
		BranchPrefix:   values["BRANCH_PREFIX"],
		TaskSystemName: values["TASK_SYSTEM_NAME"],
		TaskSystemLink: values["TASK_SYSTEM_LINK"],
	}, nil
}

func trimEnvValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, `'`)

	return value
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}
