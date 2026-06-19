package changelogger

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAppRunWritesChangelogAndRunsGitWorkflow(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	setUserConfig(t, dir)

	writeFile(t, filepath.Join(dir, ".env"), strings.Join([]string{
		"REPOSITORY_LINK=https://git.example/group/app/-/tags/",
		"CHANGELOG_PATH=./CHANGELOG.md",
		"TASK_SYSTEM_LINK=https://tracker.example/item/{code}",
		"",
	}, "\n"))
	writeFile(t, filepath.Join(dir, "CHANGELOG.md"), "# History\n")

	branch := "feature/OPS-AB222222-CD222222-assign-to-changelog"
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "rev-list", "--tags", "--max-count=1"): {{output: "tagcommit\n"}},
		runnerKey("git", "describe", "--tags", "tagcommit"):     {{output: "1.2.3\n"}},
		runnerKey("git", "rev-parse", "origin/master"):          {{output: "master123\n"}},
		runnerKey("git", "log", "--pretty=format:%h|%an|%s|%cs", "--no-merges", "1.2.3..develop"): {{
			output: strings.Join([]string{
				"001|Dev One|[OPS-AB111111-CD111111] feat: добавлен экспорт|2026-01-01",
				"002|Dev Two|[OPS-AB222222-CD222222] fix: исправлена проверка|2026-01-02",
			}, "\n"),
		}},
		runnerKey("git", "cherry", "-v", "master123", "1.2.3"):               {{output: ""}},
		runnerKey("git", "branch", "--list", branch):                         {{output: ""}},
		runnerKey("git", "checkout", "develop"):                              {{}, {}},
		runnerKey("git", "checkout", "-b", branch):                           {{}},
		runnerKey("git", "add", "./CHANGELOG.md"):                            {{}},
		runnerKey("git", "commit", "-m", "wip: Отредактирован CHANGELOG.md"): {{}},
		runnerKey("git", "push", "origin", branch):                           {{}},
		runnerKey("git", "branch", "-D", branch):                             {{}},
	})

	var output bytes.Buffer
	input := strings.NewReader("\n\ny\ny\ny\n")
	app := NewApp(nil, input, &output, fixedTime, runner)

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v\noutput:\n%s", err, output.String())
	}

	changelog := readFile(t, filepath.Join(dir, "CHANGELOG.md"))
	wantChangelog := "# History\n\n" +
		"## [ [1.3.0](https://git.example/group/app/-/tags/1.3.0) ] - 15.06.2026\n\n" +
		"- Реализовано:\n" +
		"  - Добавлен экспорт [Заявка](https://tracker.example/item/AB111111)\n" +
		"- Исправлено:\n" +
		"  - Исправлена проверка [Заявка](https://tracker.example/item/AB222222)\n\n"
	if changelog != wantChangelog {
		t.Fatalf("CHANGELOG.md mismatch\nwant:\n%s\ngot:\n%s", wantChangelog, changelog)
	}

	wantCommands := [][]string{
		{"git", "rev-list", "--tags", "--max-count=1"},
		{"git", "describe", "--tags", "tagcommit"},
		{"git", "rev-parse", "origin/master"},
		{"git", "log", "--pretty=format:%h|%an|%s|%cs", "--no-merges", "1.2.3..develop"},
		{"git", "cherry", "-v", "master123", "1.2.3"},
		{"git", "branch", "--list", branch},
		{"git", "checkout", "develop"},
		{"git", "checkout", "-b", branch},
		{"git", "add", "./CHANGELOG.md"},
		{"git", "commit", "-m", "wip: Отредактирован CHANGELOG.md"},
		{"git", "push", "origin", branch},
		{"git", "checkout", "develop"},
		{"git", "branch", "-D", branch},
	}
	if !reflect.DeepEqual(runner.commands, wantCommands) {
		t.Fatalf("commands = %#v, want %#v", runner.commands, wantCommands)
	}

	if !strings.Contains(output.String(), "Следующая версия приложения: 1.3.0") {
		t.Fatalf("output does not contain next version:\n%s", output.String())
	}
}

func TestAppRunUsesTagLinkArgumentBeforeEnvAndGitRemote(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	setUserConfig(t, dir)

	writeFile(t, filepath.Join(dir, ".env"), strings.Join([]string{
		"REPOSITORY_LINK=https://git.example/group/env-app/-/tags/",
		"CHANGELOG_PATH=./CHANGELOG.md",
		"TASK_SYSTEM_LINK=https://tracker.example/item/{code}",
		"",
	}, "\n"))
	writeFile(t, filepath.Join(dir, "CHANGELOG.md"), "# History\n")

	branch := "feature/OPS-AB111111-CD111111-assign-to-changelog"
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "rev-list", "--tags", "--max-count=1"): {{output: "tagcommit\n"}},
		runnerKey("git", "describe", "--tags", "tagcommit"):     {{output: "1.2.3\n"}},
		runnerKey("git", "rev-parse", "origin/master"):          {{output: "master123\n"}},
		runnerKey("git", "log", "--pretty=format:%h|%an|%s|%cs", "--no-merges", "1.2.3..develop"): {{
			output: "001|Dev One|[OPS-AB111111-CD111111] fix: исправлена проверка|2026-01-01",
		}},
		runnerKey("git", "cherry", "-v", "master123", "1.2.3"):               {{output: ""}},
		runnerKey("git", "branch", "--list", branch):                         {{output: ""}},
		runnerKey("git", "checkout", "develop"):                              {{}, {}},
		runnerKey("git", "checkout", "-b", branch):                           {{}},
		runnerKey("git", "add", "./CHANGELOG.md"):                            {{}},
		runnerKey("git", "commit", "-m", "wip: Отредактирован CHANGELOG.md"): {{}},
		runnerKey("git", "push", "origin", branch):                           {{}},
		runnerKey("git", "branch", "-D", branch):                             {{}},
	})

	input := strings.NewReader("\n\ny\ny\ny\n")
	app := NewApp([]string{"https://git.example/group/arg-app/-/tags/"}, input, &bytes.Buffer{}, fixedTime, runner)

	if err := app.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	changelog := readFile(t, filepath.Join(dir, "CHANGELOG.md"))
	if !strings.Contains(changelog, "https://git.example/group/arg-app/-/tags/1.2.4") {
		t.Fatalf("changelog does not use tag link argument:\n%s", changelog)
	}

	for _, command := range runner.commands {
		if reflect.DeepEqual(command, []string{"git", "remote", "get-url", "origin"}) {
			t.Fatal("Run() called git remote even though repository link argument was provided")
		}
	}
}

func TestAppRunReturnsErrorWhenNoSupportedCommitsFound(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	setUserConfig(t, dir)

	writeFile(t, filepath.Join(dir, ".env"), strings.Join([]string{
		"REPOSITORY_LINK=https://git.example/group/app/-/tags/",
		"TASK_SYSTEM_LINK=https://tracker.example/item/{code}",
		"",
	}, "\n"))

	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "rev-list", "--tags", "--max-count=1"): {{output: "tagcommit\n"}},
		runnerKey("git", "describe", "--tags", "tagcommit"):     {{output: "1.2.3\n"}},
		runnerKey("git", "rev-parse", "origin/master"):          {{output: "master123\n"}},
		runnerKey("git", "log", "--pretty=format:%h|%an|%s|%cs", "--no-merges", "1.2.3..develop"): {{
			output: "001|Dev One|[OPS-AB111111-CD111111] docs: обновлена справка|2026-01-01",
		}},
		runnerKey("git", "cherry", "-v", "master123", "1.2.3"): {{output: ""}},
	})

	app := NewApp(nil, strings.NewReader(""), &bytes.Buffer{}, fixedTime, runner)
	err := app.Run()
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "отсутствуют коммиты") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestAskVersionLevelReturnsDefaultOrTypedValue(t *testing.T) {
	appWithDefault := NewApp(nil, strings.NewReader("\n"), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))
	level, err := appWithDefault.askVersionLevel(Version{Major: 1, Minor: 2, Patch: 3}, "minor", "есть feat")
	if err != nil {
		t.Fatalf("askVersionLevel() error = %v", err)
	}
	if level != "minor" {
		t.Fatalf("askVersionLevel() = %q, want %q", level, "minor")
	}

	appWithTypedValue := NewApp(nil, strings.NewReader("major\n"), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))
	level, err = appWithTypedValue.askVersionLevel(Version{Major: 1, Minor: 2, Patch: 3}, "fix", "нет feat")
	if err != nil {
		t.Fatalf("askVersionLevel() error = %v", err)
	}
	if level != "major" {
		t.Fatalf("askVersionLevel() = %q, want %q", level, "major")
	}
}

func TestAskBranchNameBuildsBranchWithOptionalPrefix(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		defaultTask string
		input       string
		want        string
	}{
		{
			name:        "uses default task and prefix",
			prefix:      "OPS",
			defaultTask: "AB123456-CD789000",
			input:       "\n",
			want:        "feature/OPS-AB123456-CD789000-assign-to-changelog",
		},
		{
			name:  "uses typed task without prefix",
			input: "AB123456-CD789000\n",
			want:  "feature/AB123456-CD789000-assign-to-changelog",
		},
		{
			name:   "normalizes lowercase task through validation only",
			prefix: "OPS",
			input:  "ab123456-cd789000\n",
			want:   "feature/OPS-ab123456-cd789000-assign-to-changelog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(nil, strings.NewReader(tt.input), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))
			got, err := app.askBranchName(tt.prefix, tt.defaultTask)
			if err != nil {
				t.Fatalf("askBranchName() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("askBranchName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAskBranchNameRejectsInvalidTask(t *testing.T) {
	app := NewApp(nil, strings.NewReader("invalid\n"), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))

	_, err := app.askBranchName("OPS", "")
	if err == nil {
		t.Fatal("askBranchName() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "неверный формат") {
		t.Fatalf("askBranchName() error = %v", err)
	}
}

func TestAskConfirmationAcceptsYesOnly(t *testing.T) {
	for _, input := range []string{"y\n", "yes\n", "Y\n", "YES\n"} {
		t.Run(strings.TrimSpace(input), func(t *testing.T) {
			app := NewApp(nil, strings.NewReader(input), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))
			if err := app.askConfirmation("Continue?"); err != nil {
				t.Fatalf("askConfirmation() error = %v", err)
			}
		})
	}

	app := NewApp(nil, strings.NewReader("n\n"), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))
	if err := app.askConfirmation("Continue?"); err == nil {
		t.Fatal("askConfirmation() error = nil, want cancellation")
	}
}

func TestReadLineTrimsWhitespaceAndAllowsEOF(t *testing.T) {
	app := NewApp(nil, strings.NewReader("  value  "), &bytes.Buffer{}, fixedTime, newScriptedRunner(nil))

	got, err := app.readLine()
	if err != nil {
		t.Fatalf("readLine() error = %v", err)
	}
	if got != "value" {
		t.Fatalf("readLine() = %q, want %q", got, "value")
	}
}

func setUserConfig(t *testing.T, dir string) {
	t.Helper()

	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg-config"))
	t.Setenv("APPDATA", filepath.Join(dir, "appdata"))

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(configDir, "changelogger", "config.json"), `{
  "taskSystemLink": "https://tracker.example/item/{code}",
  "changelogPath": "./CHANGELOG.md"
}
`)
}
