package changelogger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigFilesMergesGlobalConfigWithEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "config.json")
	envPath := filepath.Join(dir, ".env")

	writeFile(t, globalPath, `{
  "taskSystemLink": "https://tracker.example/item/{code}",
  "changelogPath": "./CHANGELOG.md"
}
`)
	writeFile(t, envPath, `
REPOSITORY_LINK="https://git.example/group/app/-/tags/"
BRANCH_PREFIX='OPS'
CHANGELOG_PATH=./docs/CHANGELOG.md
TASK_SYSTEM_LINK=https://tasks.example/card?code=
IGNORED_LINE
`)

	config, err := LoadConfigFiles(globalPath, envPath)
	if err != nil {
		t.Fatalf("LoadConfigFiles() error = %v", err)
	}

	want := Config{
		RepositoryLink: "https://git.example/group/app/-/tags/",
		ChangelogPath:  "./docs/CHANGELOG.md",
		BranchPrefix:   "OPS",
		TaskSystemLink: "https://tasks.example/card?code=",
		TaskLinkMode:   "ask",
		CommitMode:     "ask",
	}
	if config != want {
		t.Fatalf("LoadConfigFiles() = %#v, want %#v", config, want)
	}
}

func TestLoadConfigFilesCreatesProjectEnvTemplateWhenMissing(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "config.json")
	envPath := filepath.Join(dir, ".env")

	writeFile(t, globalPath, `{
  "taskSystemLink": "https://tracker.example/item/{code}",
  "changelogPath": "./docs/CHANGELOG.md"
}
`)

	config, err := LoadConfigFiles(globalPath, envPath)
	if err != nil {
		t.Fatalf("LoadConfigFiles() error = %v", err)
	}

	if config.ChangelogPath != "./docs/CHANGELOG.md" {
		t.Fatalf("ChangelogPath = %q, want %q", config.ChangelogPath, "./docs/CHANGELOG.md")
	}
	if config.TaskSystemLink != "https://tracker.example/item/{code}" {
		t.Fatalf("TaskSystemLink = %q", config.TaskSystemLink)
	}

	envContent := readFile(t, envPath)
	for _, want := range []string{
		"# Changelogger project config",
		"# REPOSITORY_LINK=",
		"# BRANCH_PREFIX=",
		"# CHANGELOG_PATH=./docs/CHANGELOG.md",
		"# TASK_SYSTEM_LINK=https://tracker.example/item/{code}",
	} {
		if !strings.Contains(envContent, want) {
			t.Fatalf("created env template does not contain %q:\n%s", want, envContent)
		}
	}
}

func TestLoadConfigFilesUsesDefaultChangelogPathWhenGlobalPathEmpty(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "config.json")
	envPath := filepath.Join(dir, ".env")

	writeFile(t, globalPath, `{"taskSystemLink": "https://tracker.example/item/{code}"}`)
	writeFile(t, envPath, "")

	config, err := LoadConfigFiles(globalPath, envPath)
	if err != nil {
		t.Fatalf("LoadConfigFiles() error = %v", err)
	}

	if config.ChangelogPath != "./CHANGELOG.md" {
		t.Fatalf("ChangelogPath = %q, want %q", config.ChangelogPath, "./CHANGELOG.md")
	}
}

func TestLoadConfigFilesRequiresTaskSystemLink(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "config.json")
	envPath := filepath.Join(dir, ".env")

	writeFile(t, globalPath, `{"changelogPath": "./CHANGELOG.md"}`)
	writeFile(t, envPath, "")

	_, err := LoadConfigFiles(globalPath, envPath)
	if err == nil {
		t.Fatal("LoadConfigFiles() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "taskSystemLink") {
		t.Fatalf("error = %q, want taskSystemLink context", err)
	}
}

func TestLoadConfigFilesRejectsMalformedGlobalConfig(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "config.json")
	envPath := filepath.Join(dir, ".env")

	writeFile(t, globalPath, `{not-json`)
	writeFile(t, envPath, "")

	_, err := LoadConfigFiles(globalPath, envPath)
	if err == nil {
		t.Fatal("LoadConfigFiles() error = nil, want error")
	}
}

func TestReadEnvValuesTrimsQuotesAndIgnoresComments(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	writeFile(t, path, `
# comment
REPOSITORY_LINK = "https://git.example/group/app/-/tags/"
BRANCH_PREFIX='OPS'
EMPTY=
WITHOUT_SEPARATOR
`)

	values, err := readEnvValues(path)
	if err != nil {
		t.Fatalf("readEnvValues() error = %v", err)
	}

	if values["REPOSITORY_LINK"] != "https://git.example/group/app/-/tags/" {
		t.Fatalf("REPOSITORY_LINK = %q", values["REPOSITORY_LINK"])
	}
	if values["BRANCH_PREFIX"] != "OPS" {
		t.Fatalf("BRANCH_PREFIX = %q", values["BRANCH_PREFIX"])
	}
	if values["EMPTY"] != "" {
		t.Fatalf("EMPTY = %q, want empty string", values["EMPTY"])
	}
	if _, ok := values["WITHOUT_SEPARATOR"]; ok {
		t.Fatal("WITHOUT_SEPARATOR was parsed, want it ignored")
	}
}

func TestReadGlobalConfigCreatesDefaultWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")

	config, err := readGlobalConfig(path)
	if err != nil {
		t.Fatalf("readGlobalConfig() error = %v", err)
	}
	if config.TaskSystemLink == "" {
		t.Fatal("TaskSystemLink is empty, want generated default")
	}
	if config.ChangelogPath != "./CHANGELOG.md" {
		t.Fatalf("ChangelogPath = %q, want %q", config.ChangelogPath, "./CHANGELOG.md")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("generated config was not written: %v", err)
	}
}

func TestWriteGlobalConfigTemplateOmitsProjectOnlyFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	err := writeGlobalConfigTemplate(path, Config{
		RepositoryLink: "https://git.example/group/app/-/tags/",
		BranchPrefix:   "OPS",
		TaskSystemLink: "https://tracker.example/item/{code}",
		ChangelogPath:  "./CHANGELOG.md",
		TaskLinkMode:   "always",
		CommitMode:     "never",
	})
	if err != nil {
		t.Fatalf("writeGlobalConfigTemplate() error = %v", err)
	}

	content := readFile(t, path)
	for _, forbidden := range []string{"repositoryLink", "branchPrefix"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("global config contains project-only field %q:\n%s", forbidden, content)
		}
	}
	for _, want := range []string{"taskSystemLink", "changelogPath", "taskLinkMode", "commitMode"} {
		if !strings.Contains(content, want) {
			t.Fatalf("global config does not contain %q:\n%s", want, content)
		}
	}
}

func TestGlobalConfigPathUsesHomeDotConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg-config"))
	t.Setenv("APPDATA", filepath.Join(t.TempDir(), "appdata"))

	path, err := globalConfigPath()
	if err != nil {
		t.Fatalf("globalConfigPath() error = %v", err)
	}

	want := filepath.Join(home, ".config", "changelogger", "config.json")
	if path != want {
		t.Fatalf("globalConfigPath() = %q, want %q", path, want)
	}
}

func TestExampleGlobalConfigIsJSON(t *testing.T) {
	example := exampleGlobalConfig()

	if !strings.Contains(example, "taskSystemLink") || !strings.Contains(example, "changelogPath") {
		t.Fatalf("exampleGlobalConfig() = %q", example)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
