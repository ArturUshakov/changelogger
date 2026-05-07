package changelogger

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Changelog struct {
	config Config
	now    func() time.Time
}

type CommitChange struct {
	Kind        string
	Description string
	TaskRaw     string
}

func NewChangelog(config Config, now func() time.Time) Changelog {
	return Changelog{config: config, now: now}
}

func (changelog Changelog) Body(lines []string) string {
	grouped := make(map[string][]string)

	for _, line := range lines {
		change, ok := ParseCommitLine(line)
		if !ok {
			continue
		}

		title, ok := changeTypeTitle(change.Kind)
		if !ok {
			continue
		}

		description := UppercaseFirst(strings.TrimSpace(change.Description))
		if taskCode := ExtractTaskCode(change.TaskRaw); taskCode != "" {
			if taskLink := changelog.taskLink(taskCode); taskLink != "" && changelog.config.TaskSystemName != "" {
				description += " [Заявка " + changelog.config.TaskSystemName + "](" + taskLink + ")"
			}
		}

		grouped[title] = append(grouped[title], description)
	}

	var builder strings.Builder
	for _, changeType := range changeTypes {
		items := grouped[changeType.Title]
		if len(items) == 0 {
			continue
		}

		builder.WriteString("- ")
		builder.WriteString(changeType.Title)
		builder.WriteString(":\n")

		for _, item := range items {
			builder.WriteString("  - ")
			builder.WriteString(item)
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

func (changelog Changelog) Write(path string, tag string, body string) error {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("прочитать changelog: %w", err)
	}

	title := fmt.Sprintf(
		"## [ [%s](%s%s) ] - %s\n\n",
		tag,
		changelog.config.RepositoryLink,
		tag,
		changelog.now().Format("02.01.2006"),
	)

	insert := title + body + "\n"
	content := string(contentBytes)

	pos := strings.Index(content, "## [")
	if pos < 0 {
		if strings.TrimSpace(content) != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + insert
	} else {
		content = content[:pos] + insert + content[pos:]
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("записать changelog: %w", err)
	}

	return nil
}

func ParseCommitLine(line string) (CommitChange, bool) {
	parts := strings.SplitN(line, "|", 4)
	if len(parts) < 4 {
		return CommitChange{}, false
	}

	return ParseCommitSubject(parts[2])
}

func ParseCommitSubject(subject string) (CommitChange, bool) {
	matches := commitSubjectPattern.FindStringSubmatch(subject)
	if len(matches) == 0 {
		return CommitChange{}, false
	}

	return CommitChange{
		TaskRaw:     matches[1],
		Kind:        matches[2],
		Description: matches[3],
	}, true
}

func ExtractTaskCode(raw string) string {
	return taskCodePattern.FindString(raw)
}

func UppercaseFirst(value string) string {
	if value == "" {
		return ""
	}

	first, size := utf8.DecodeRuneInString(value)
	if first == utf8.RuneError && size == 0 {
		return value
	}

	return string(unicode.ToUpper(first)) + value[size:]
}

func (changelog Changelog) taskLink(code string) string {
	if changelog.config.TaskSystemLink == "" {
		return ""
	}

	link := strings.TrimSpace(changelog.config.TaskSystemLink)
	escapedCode := url.QueryEscape(code)

	if strings.Contains(link, "{code}") {
		return strings.ReplaceAll(link, "{code}", escapedCode)
	}

	if strings.HasSuffix(link, "=") {
		return link + escapedCode
	}

	return strings.TrimRight(link, "/") + "/tasks/view?code=" + escapedCode
}

type changeType struct {
	Key   string
	Title string
}

var changeTypes = []changeType{
	{Key: "feat", Title: "Реализовано"},
	{Key: "refactor", Title: "Изменено"},
	{Key: "fix", Title: "Исправлено"},
	{Key: "remove", Title: "Удалено"},
}

func changeTypeTitle(key string) (string, bool) {
	for _, changeType := range changeTypes {
		if changeType.Key == key {
			return changeType.Title, true
		}
	}

	return "", false
}

var (
	commitSubjectPattern = regexp.MustCompile(`^\[(.*)\]\s+(\w+):\s*(.*)$`)
	taskCodePattern      = regexp.MustCompile(`[A-Z]{2,}\d{6,}`)
)
