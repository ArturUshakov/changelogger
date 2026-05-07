package changeloger

import (
	"strings"
	"testing"
	"time"
)

func TestUppercaseFirst(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"исправлены опечатки": "Исправлены опечатки",
		"added english text":  "Added english text",
		"":                    "",
		"123 задачи":          "123 задачи",
	}

	for input, expected := range cases {
		if actual := UppercaseFirst(input); actual != expected {
			t.Fatalf("UppercaseFirst(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestParseCommitSubject(t *testing.T) {
	t.Parallel()

	change, ok := ParseCommitSubject("[IDPSPS-IU212256-W0835055] fix: исправлены опечатки")
	if !ok {
		t.Fatal("expected commit subject to match")
	}

	if change.TaskRaw != "IDPSPS-IU212256-W0835055" {
		t.Fatalf("TaskRaw = %q", change.TaskRaw)
	}

	if change.Kind != "fix" {
		t.Fatalf("Kind = %q", change.Kind)
	}

	if change.Description != "исправлены опечатки" {
		t.Fatalf("Description = %q", change.Description)
	}
}

func TestChangelogBody(t *testing.T) {
	t.Parallel()

	config := Config{
		TaskSystemName: "Tracker",
		TaskSystemLink: "https://tracker.example/tasks/view?code=",
	}
	changelog := NewChangelog(config, func() time.Time {
		return time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	})

	body := changelog.Body([]string{
		"jkl|Dmitry|[IDPSPS-IU212256-W0835055] remove: удален старый экран|2026-05-07",
		"mno|Dmitry|[IDPSPS-IU212256-W0835055] refactor: изменена форма|2026-05-07",
		"abc|Dmitry|[IDPSPS-IU212256-W0835055] fix: исправлены опечатки|2026-05-07",
		"def|Dmitry|[IDPSPS-IU212256-W0835055] feat: добавлен экспорт|2026-05-07",
		"ghi|Dmitry|wip: что-то промежуточное|2026-05-07",
	})

	assertContains(t, body, "- Реализовано:\n  - Добавлен экспорт")
	assertContains(t, body, "- Изменено:\n  - Изменена форма")
	assertContains(t, body, "- Исправлено:\n  - Исправлены опечатки [Заявка Tracker](https://tracker.example/tasks/view?code=IU212256)")
	assertContains(t, body, "- Удалено:\n  - Удален старый экран")
	assertOrder(t, body, "- Реализовано:", "- Изменено:", "- Исправлено:", "- Удалено:")

	if strings.Contains(body, "wip") {
		t.Fatalf("body should not contain ignored commit: %s", body)
	}
}

func TestVersionNext(t *testing.T) {
	t.Parallel()

	version, err := ParseVersion("1.2.3")
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]string{
		"1":     "2.0.0",
		"minor": "1.3.0",
		"fix":   "1.2.4",
	}

	for level, expected := range cases {
		next, err := version.Next(level)
		if err != nil {
			t.Fatalf("Next(%q) returned error: %v", level, err)
		}

		if next.String() != expected {
			t.Fatalf("Next(%q) = %s, want %s", level, next.String(), expected)
		}
	}
}

func assertContains(t *testing.T, value string, substring string) {
	t.Helper()

	if !strings.Contains(value, substring) {
		t.Fatalf("expected %q to contain %q", value, substring)
	}
}

func assertOrder(t *testing.T, value string, substrings ...string) {
	t.Helper()

	previous := -1
	for _, substring := range substrings {
		current := strings.Index(value, substring)
		if current < 0 {
			t.Fatalf("expected %q to contain %q", value, substring)
		}

		if current <= previous {
			t.Fatalf("expected %q to appear after previous section in %q", substring, value)
		}

		previous = current
	}
}
