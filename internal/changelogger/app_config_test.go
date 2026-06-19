package changelogger

import (
	"bytes"
	"strings"
	"testing"
)

func TestAppConfigShowPrintsCurrentPreferences(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	setUserConfigContent(t, dir, `{
  "taskSystemLink": "https://tracker.example/item/{code}",
  "changelogPath": "./CHANGELOG.md",
  "taskLinkMode": "always",
  "commitMode": "never",
  "changelogMode": "branch"
}
`)

	var output bytes.Buffer
	app := NewApp([]string{"config", "show"}, strings.NewReader(""), &output, fixedTime, newScriptedRunner(nil))

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	for _, want := range []string{
		"Ссылки на заявку: всегда добавлять",
		"Коммит: никогда не создавать",
		"Запись changelog: develop и новая ветка",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("output does not contain %q:\n%s", want, output.String())
		}
	}
}

func TestAppConfigSetChangelogMode(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	setUserConfig(t, dir)

	app := NewApp([]string{"config", "changelog", "current"}, strings.NewReader(""), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	config := readFile(t, globalConfigFile(t))
	if !strings.Contains(config, `"changelogMode": "current"`) {
		t.Fatalf("global config does not contain saved changelogMode:\n%s", config)
	}
}

func TestAppConfigSetTaskLinkMode(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	setUserConfig(t, dir)

	app := NewApp([]string{"config", "task-link", "never"}, strings.NewReader(""), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	config := readFile(t, globalConfigFile(t))
	if !strings.Contains(config, `"taskLinkMode": "never"`) {
		t.Fatalf("global config does not contain saved taskLinkMode:\n%s", config)
	}
}

func TestAppConfigSetCommitMode(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	setUserConfig(t, dir)

	app := NewApp([]string{"config", "commit", "always"}, strings.NewReader(""), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	config := readFile(t, globalConfigFile(t))
	if !strings.Contains(config, `"commitMode": "always"`) {
		t.Fatalf("global config does not contain saved commitMode:\n%s", config)
	}
}

func TestAppConfigInteractiveUpdatesPreference(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	setUserConfig(t, dir)

	var output bytes.Buffer
	app := NewApp([]string{"config"}, strings.NewReader("2\n2\n"), &output, fixedTime, newScriptedRunner(nil))

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v\noutput:\n%s", err, output.String())
	}

	config := readFile(t, globalConfigFile(t))
	if !strings.Contains(config, `"commitMode": "never"`) {
		t.Fatalf("global config does not contain saved commitMode:\n%s", config)
	}
}

func TestAppConfigResetPreferences(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	setUserConfigContent(t, dir, `{
  "taskSystemLink": "https://tracker.example/item/{code}",
  "changelogPath": "./CHANGELOG.md",
  "taskLinkMode": "always",
  "commitMode": "never",
  "changelogMode": "branch"
}
`)

	app := NewApp([]string{"config", "reset"}, strings.NewReader(""), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	config := readFile(t, globalConfigFile(t))
	if strings.Contains(config, "taskLinkMode") || strings.Contains(config, "commitMode") || strings.Contains(config, "changelogMode") {
		t.Fatalf("global config still contains preference modes after reset:\n%s", config)
	}
}

func TestAppUpdateRunsLatestInstaller(t *testing.T) {
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("sh", "-c", updateInstallCommand): {{}},
	})
	var output bytes.Buffer
	app := NewApp([]string{"update"}, strings.NewReader(""), &output, fixedTime, runner)

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v\noutput:\n%s", err, output.String())
	}

	if !strings.Contains(output.String(), "Обновляю changelogger") {
		t.Fatalf("output does not mention update:\n%s", output.String())
	}
}
