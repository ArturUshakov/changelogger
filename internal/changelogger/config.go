package changelogger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	RepositoryLink string `json:"repositoryLink"`
	ChangelogPath  string `json:"changelogPath"`
	BranchPrefix   string `json:"branchPrefix"`
	TaskSystemLink string `json:"taskSystemLink"`
	TaskLinkMode   string `json:"taskLinkMode,omitempty"`
	CommitMode     string `json:"commitMode,omitempty"`
}

func LoadConfig(path string) (Config, error) {
	globalPath, err := globalConfigPath()
	if err != nil {
		return Config{}, err
	}

	return LoadConfigFiles(globalPath, path)
}

func LoadConfigFiles(globalPath string, envPath string) (Config, error) {
	global, err := readGlobalConfig(globalPath)
	if err != nil {
		return Config{}, err
	}

	if _, err := os.Stat(envPath); err != nil {
		if !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("проверить .env: %w", err)
		}

		if err := writeProjectEnvTemplate(envPath, global); err != nil {
			return Config{}, err
		}
	}

	values, err := readEnvValues(envPath)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		RepositoryLink: values["REPOSITORY_LINK"],
		ChangelogPath:  valueOrDefault(global.ChangelogPath, "./CHANGELOG.md"),
		BranchPrefix:   values["BRANCH_PREFIX"],
		TaskSystemLink: global.TaskSystemLink,
		TaskLinkMode:   normalizePreferenceMode(global.TaskLinkMode),
		CommitMode:     normalizePreferenceMode(global.CommitMode),
	}

	if values["CHANGELOG_PATH"] != "" {
		config.ChangelogPath = values["CHANGELOG_PATH"]
	}
	if values["TASK_SYSTEM_LINK"] != "" {
		config.TaskSystemLink = values["TASK_SYSTEM_LINK"]
	}

	if err := validateConfig(config, envPath, globalPath); err != nil {
		return Config{}, err
	}

	return config, nil
}

func globalConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("получить каталог конфигурации пользователя: %w", err)
	}

	return filepath.Join(configDir, "changelogger", "config.json"), nil
}

func readGlobalConfig(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			config := defaultGlobalConfig()
			if err := writeGlobalConfigTemplate(path, config); err != nil {
				return Config{}, err
			}

			return config, nil
		}

		return Config{}, fmt.Errorf("прочитать глобальный config %s: %w", path, err)
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return Config{}, fmt.Errorf("разобрать глобальный config %s: %w", path, err)
	}

	return config, nil
}

func writeGlobalConfigTemplate(path string, config Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("создать каталог глобального config %s: %w", filepath.Dir(path), err)
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("подготовить глобальный config: %w", err)
	}
	content = append(content, '\n')

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("создать глобальный config %s: %w", path, err)
	}

	return nil
}

func saveGlobalConfig(update func(*Config)) error {
	path, err := globalConfigPath()
	if err != nil {
		return err
	}

	config, err := readGlobalConfig(path)
	if err != nil {
		return err
	}

	update(&config)

	return writeGlobalConfigTemplate(path, config)
}

func normalizePreferenceMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case preferenceAlways:
		return preferenceAlways
	case preferenceNever:
		return preferenceNever
	default:
		return preferenceAsk
	}
}

func writeProjectEnvTemplate(path string, global Config) error {
	content := strings.Builder{}
	content.WriteString("# Changelogger project config\n")
	content.WriteString("# Optional overrides. Changelogger detects repository link and branch prefix automatically.\n")
	content.WriteString("# REPOSITORY_LINK=\n")
	content.WriteString("# BRANCH_PREFIX=\n")
	content.WriteString("# CHANGELOG_PATH=")
	content.WriteString(valueOrDefault(global.ChangelogPath, "./CHANGELOG.md"))
	content.WriteString("\n")
	if global.TaskSystemLink != "" {
		content.WriteString("# TASK_SYSTEM_LINK=")
		content.WriteString(global.TaskSystemLink)
		content.WriteString("\n")
	}

	if err := os.WriteFile(path, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("создать шаблон .env: %w", err)
	}

	return nil
}

func validateConfig(config Config, envPath string, globalPath string) error {
	if config.TaskSystemLink == "" {
		return fmt.Errorf("заполните %s: taskSystemLink", globalPath)
	}

	return nil
}

func readEnvValues(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("файл .env не существует")
		}

		return nil, fmt.Errorf("прочитать .env: %w", err)
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
		return nil, fmt.Errorf("прочитать .env: %w", err)
	}

	return values, nil
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

func exampleGlobalConfig() string {
	content, err := json.MarshalIndent(defaultGlobalConfig(), "", "  ")
	if err != nil {
		return ""
	}

	return string(content)
}

func defaultGlobalConfig() Config {
	return Config{
		TaskSystemLink: "https://helpdesk.efko.ru/tasks/view?code={code}",
		ChangelogPath:  "./CHANGELOG.md",
	}
}
