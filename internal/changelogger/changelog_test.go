package changelogger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestChangelogBodyGroupsSupportedCommitTypesInFixedOrder(t *testing.T) {
	changelog := NewChangelog(Config{
		TaskSystemLink: "https://tracker.example/item/{code}",
	}, fixedTime)

	body := changelog.Body([]string{
		"001|Dev One|[OPS-AB123456-CD789000] fix: исправлена проверка статуса|2026-01-01",
		"002|Dev Two|[OPS-AB123457-CD789001] feat: добавлен экспорт отчета|2026-01-02",
		"003|Dev One|[OPS-AB123458-CD789002] remove: удален устаревший фильтр|2026-01-03",
		"004|Dev Two|[OPS-AB123459-CD789003] change: обновлен расчет суммы|2026-01-04",
		"005|Dev Two|[OPS-AB123460-CD789004] refactor: упрощена подготовка данных|2026-01-05",
		"006|Dev Two|[OPS-AB123461-CD789005] docs: обновлена документация|2026-01-06",
	})

	want := strings.Join([]string{
		"- Реализовано:",
		"  - Добавлен экспорт отчета [Заявка](https://tracker.example/item/AB123457)",
		"- Изменено:",
		"  - Обновлен расчет суммы [Заявка](https://tracker.example/item/AB123459)",
		"  - Упрощена подготовка данных [Заявка](https://tracker.example/item/AB123460)",
		"- Исправлено:",
		"  - Исправлена проверка статуса [Заявка](https://tracker.example/item/AB123456)",
		"- Удалено:",
		"  - Удален устаревший фильтр [Заявка](https://tracker.example/item/AB123458)",
		"",
	}, "\n")

	if body != want {
		t.Fatalf("Body() mismatch\nwant:\n%s\ngot:\n%s", want, body)
	}
}

func TestChangelogBodySkipsInvalidLinesAndDeduplicatesWithinSection(t *testing.T) {
	changelog := NewChangelog(Config{
		TaskSystemLink: "https://tracker.example/item?code=",
	}, fixedTime)

	body := changelog.Body([]string{
		"not a formatted git log line",
		"001|Dev One|[OPS-AB123456-CD789000] fix: исправлена проверка статуса|2026-01-01",
		"002|Dev Two|[OPS-AB123456-CD789000] fix: исправлена проверка статуса|2026-01-02",
		"003|Dev Two|[OPS-AB123456-CD789000] feat: исправлена проверка статуса|2026-01-03",
	})

	want := strings.Join([]string{
		"- Реализовано:",
		"  - Исправлена проверка статуса [Заявка](https://tracker.example/item?code=AB123456)",
		"- Исправлено:",
		"  - Исправлена проверка статуса [Заявка](https://tracker.example/item?code=AB123456)",
		"",
	}, "\n")

	if body != want {
		t.Fatalf("Body() mismatch\nwant:\n%s\ngot:\n%s", want, body)
	}
}

func TestChangelogBodyOmitsTaskLinkWhenConfigHasNoTaskSystemLink(t *testing.T) {
	changelog := NewChangelog(Config{}, fixedTime)

	body := changelog.Body([]string{
		"001|Dev One|[OPS-AB123456-CD789000] feat: добавлен экспорт отчета|2026-01-01",
	})

	want := "- Реализовано:\n  - Добавлен экспорт отчета\n"
	if body != want {
		t.Fatalf("Body() = %q, want %q", body, want)
	}
}

func TestChangelogWritePrependsNewReleaseBeforeExistingReleases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	initial := "# History\n\n## [ [1.2.2](https://git.example/app/-/tags/1.2.2) ] - 01.01.2026\n\nold\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	changelog := NewChangelog(Config{
		RepositoryLink: "https://git.example/app/-/tags/",
	}, fixedTime)

	err := changelog.Write(path, "1.2.3", "- Реализовано:\n  - Добавлен экспорт\n")
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content := readFile(t, path)
	want := "# History\n\n" +
		"## [ [1.2.3](https://git.example/app/-/tags/1.2.3) ] - 15.06.2026\n\n" +
		"- Реализовано:\n  - Добавлен экспорт\n\n" +
		"## [ [1.2.2](https://git.example/app/-/tags/1.2.2) ] - 01.01.2026\n\nold\n"

	if content != want {
		t.Fatalf("written changelog mismatch\nwant:\n%s\ngot:\n%s", want, content)
	}
}

func TestChangelogWriteAppendsReleaseWhenNoExistingReleaseHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := os.WriteFile(path, []byte("# History"), 0644); err != nil {
		t.Fatal(err)
	}

	changelog := NewChangelog(Config{
		RepositoryLink: "https://git.example/app/-/tags/",
	}, fixedTime)

	err := changelog.Write(path, "1.2.3", "- Исправлено:\n  - Устранена ошибка\n")
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	want := "# History\n\n## [ [1.2.3](https://git.example/app/-/tags/1.2.3) ] - 15.06.2026\n\n" +
		"- Исправлено:\n  - Устранена ошибка\n\n"
	if got := readFile(t, path); got != want {
		t.Fatalf("written changelog mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestParseCommitLine(t *testing.T) {
	change, ok := ParseCommitLine("abc123|Dev One|[OPS-AB123456-CD789000] feat: добавлен экспорт|2026-01-01")
	if !ok {
		t.Fatal("ParseCommitLine() ok = false, want true")
	}

	if change.TaskRaw != "OPS-AB123456-CD789000" || change.Kind != "feat" || change.Description != "добавлен экспорт" {
		t.Fatalf("ParseCommitLine() = %#v", change)
	}

	if _, ok := ParseCommitLine("abc123|Dev One|[OPS-AB123456-CD789000] feat: добавлен экспорт"); ok {
		t.Fatal("ParseCommitLine() ok = true for malformed log line, want false")
	}
}

func TestParseCommitSubject(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		want    CommitChange
		ok      bool
	}{
		{
			name:    "supported subject",
			subject: "[OPS-AB123456-CD789000] fix: исправлена проверка",
			want: CommitChange{
				TaskRaw:     "OPS-AB123456-CD789000",
				Kind:        "fix",
				Description: "исправлена проверка",
			},
			ok: true,
		},
		{
			name:    "missing task marker",
			subject: "fix: исправлена проверка",
			ok:      false,
		},
		{
			name:    "missing separator",
			subject: "[OPS-AB123456-CD789000] fix исправлена проверка",
			ok:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseCommitSubject(tt.subject)
			if ok != tt.ok {
				t.Fatalf("ParseCommitSubject() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("ParseCommitSubject() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestTaskAndBranchExtraction(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		taskCode   string
		prefix     string
		branchTask string
	}{
		{
			name:       "prefix with task",
			raw:        "OPS-AB123456-CD789000",
			taskCode:   "AB123456",
			prefix:     "OPS",
			branchTask: "AB123456-CD789000",
		},
		{
			name:       "task without prefix",
			raw:        "AB123456-CD789000",
			taskCode:   "AB123456",
			prefix:     "",
			branchTask: "AB123456-CD789000",
		},
		{
			name:       "task with separators before code",
			raw:        "OPS / AB123456-CD789000",
			taskCode:   "AB123456",
			prefix:     "OPS",
			branchTask: "AB123456-CD789000",
		},
		{
			name:       "invalid branch task suffix",
			raw:        "OPS-AB123456",
			taskCode:   "AB123456",
			prefix:     "OPS",
			branchTask: "",
		},
		{
			name:       "no task code",
			raw:        "OPS-ticket",
			taskCode:   "",
			prefix:     "",
			branchTask: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractTaskCode(tt.raw); got != tt.taskCode {
				t.Fatalf("ExtractTaskCode() = %q, want %q", got, tt.taskCode)
			}
			if got := ExtractBranchPrefix(tt.raw); got != tt.prefix {
				t.Fatalf("ExtractBranchPrefix() = %q, want %q", got, tt.prefix)
			}
			if got := ExtractBranchTask(tt.raw); got != tt.branchTask {
				t.Fatalf("ExtractBranchTask() = %q, want %q", got, tt.branchTask)
			}
		})
	}
}

func TestDetectBranchTaskUsesLatestValidCommitTask(t *testing.T) {
	lines := []string{
		"001|Dev One|[OPS-AB111111-CD111111] feat: добавлен экспорт|2026-01-01",
		"002|Dev Two|[OPS-AB222222] fix: исправлена проверка|2026-01-02",
		"003|Dev Two|[OPS-AB333333-CD333333] fix: исправлена проверка|2026-01-03",
	}

	if got := DetectBranchTask(lines); got != "AB333333-CD333333" {
		t.Fatalf("DetectBranchTask() = %q, want %q", got, "AB333333-CD333333")
	}
}

func TestDetectBranchPrefixReturnsSingleCommonPrefixOnly(t *testing.T) {
	t.Run("single prefix", func(t *testing.T) {
		lines := []string{
			"001|Dev One|[OPS-AB111111-CD111111] feat: добавлен экспорт|2026-01-01",
			"002|Dev Two|[OPS-AB222222-CD222222] fix: исправлена проверка|2026-01-02",
		}

		if got := DetectBranchPrefix(lines); got != "OPS" {
			t.Fatalf("DetectBranchPrefix() = %q, want %q", got, "OPS")
		}
	})

	t.Run("mixed prefixes", func(t *testing.T) {
		lines := []string{
			"001|Dev One|[OPS-AB111111-CD111111] feat: добавлен экспорт|2026-01-01",
			"002|Dev Two|[APP-AB222222-CD222222] fix: исправлена проверка|2026-01-02",
		}

		if got := DetectBranchPrefix(lines); got != "" {
			t.Fatalf("DetectBranchPrefix() = %q, want empty string", got)
		}
	})

	t.Run("no prefix", func(t *testing.T) {
		lines := []string{
			"001|Dev One|[AB111111-CD111111] feat: добавлен экспорт|2026-01-01",
		}

		if got := DetectBranchPrefix(lines); got != "" {
			t.Fatalf("DetectBranchPrefix() = %q, want empty string", got)
		}
	})
}

func TestRecommendVersionLevel(t *testing.T) {
	tests := []struct {
		name       string
		lines      []string
		wantLevel  string
		wantReason string
	}{
		{
			name: "feat recommends minor",
			lines: []string{
				"001|Dev One|[OPS-AB111111-CD111111] fix: исправлена проверка|2026-01-01",
				"002|Dev Two|[OPS-AB222222-CD222222] feat: добавлен экспорт|2026-01-02",
			},
			wantLevel:  "minor",
			wantReason: "есть feat",
		},
		{
			name: "no feat recommends fix",
			lines: []string{
				"001|Dev One|[OPS-AB111111-CD111111] fix: исправлена проверка|2026-01-01",
				"002|Dev Two|[OPS-AB222222-CD222222] change: обновлен расчет|2026-01-02",
			},
			wantLevel:  "fix",
			wantReason: "нет feat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, reason := RecommendVersionLevel(tt.lines)
			if level != tt.wantLevel || reason != tt.wantReason {
				t.Fatalf("RecommendVersionLevel() = %q, %q; want %q, %q", level, reason, tt.wantLevel, tt.wantReason)
			}
		})
	}
}

func TestUppercaseFirst(t *testing.T) {
	tests := map[string]string{
		"":               "",
		"добавлен отчет": "Добавлен отчет",
		"export ready":   "Export ready",
		"1 item":         "1 item",
	}

	for input, want := range tests {
		if got := UppercaseFirst(input); got != want {
			t.Fatalf("UppercaseFirst(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestTaskLinkFormatsSupportedConfigValues(t *testing.T) {
	tests := []struct {
		name string
		link string
		want string
	}{
		{
			name: "placeholder",
			link: "https://tracker.example/item/{code}",
			want: "https://tracker.example/item/AB123456",
		},
		{
			name: "query prefix",
			link: "https://tracker.example/item?code=",
			want: "https://tracker.example/item?code=AB123456",
		},
		{
			name: "base url",
			link: "https://tracker.example",
			want: "https://tracker.example/tasks/view?code=AB123456",
		},
		{
			name: "base url with trailing slash",
			link: "https://tracker.example/",
			want: "https://tracker.example/tasks/view?code=AB123456",
		},
		{
			name: "empty link",
			link: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changelog := NewChangelog(Config{TaskSystemLink: tt.link}, fixedTime)
			if got := changelog.taskLink("AB123456"); got != tt.want {
				t.Fatalf("taskLink() = %q, want %q", got, tt.want)
			}
		})
	}
}

func fixedTime() time.Time {
	return time.Date(2026, 6, 15, 10, 20, 30, 0, time.UTC)
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}
