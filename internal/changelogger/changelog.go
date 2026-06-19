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
	seen := make(map[string]map[string]bool)

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
			if taskLink := changelog.taskLink(taskCode); taskLink != "" {
				description += " [Заявка](" + taskLink + ")"
			}
		}

		if seen[title] == nil {
			seen[title] = make(map[string]bool)
		}
		if seen[title][description] {
			continue
		}
		seen[title][description] = true

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

func ExtractBranchPrefix(raw string) string {
	taskCodeIndex := taskCodePattern.FindStringIndex(raw)
	if taskCodeIndex == nil || taskCodeIndex[0] == 0 {
		return ""
	}

	return strings.Trim(raw[:taskCodeIndex[0]], "-_ /")
}

func ExtractBranchTask(raw string) string {
	taskCodeIndex := taskCodePattern.FindStringIndex(raw)
	if taskCodeIndex == nil {
		return ""
	}

	task := strings.Trim(raw[taskCodeIndex[0]:], "-_ /")
	if !validBranchTask(task) {
		return ""
	}

	return task
}

func DetectBranchTask(lines []string) string {
	for index := len(lines) - 1; index >= 0; index-- {
		change, ok := ParseCommitLine(lines[index])
		if !ok {
			continue
		}

		task := ExtractBranchTask(change.TaskRaw)
		if task != "" {
			return task
		}
	}

	return ""
}

func DetectBranchPrefix(lines []string) string {
	prefixes := make(map[string]bool)
	for _, line := range lines {
		change, ok := ParseCommitLine(line)
		if !ok {
			continue
		}

		prefix := ExtractBranchPrefix(change.TaskRaw)
		if prefix != "" {
			prefixes[prefix] = true
		}
	}

	if len(prefixes) != 1 {
		return ""
	}

	for prefix := range prefixes {
		return prefix
	}

	return ""
}

func RecommendVersionLevel(lines []string) (string, string) {
	for _, line := range lines {
		change, ok := ParseCommitLine(line)
		if !ok {
			continue
		}

		if change.Kind == "feat" {
			return "minor", "есть feat"
		}
	}

	return "fix", "нет feat"
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
	Keys  []string
	Title string
}

var changeTypes = []changeType{
	{Keys: []string{"feat"}, Title: "Реализовано"},
	{Keys: []string{"refactor", "change"}, Title: "Изменено"},
	{Keys: []string{"fix"}, Title: "Исправлено"},
	{Keys: []string{"remove"}, Title: "Удалено"},
}

func changeTypeTitle(key string) (string, bool) {
	for _, changeType := range changeTypes {
		for _, changeKey := range changeType.Keys {
			if changeKey == key {
				return changeType.Title, true
			}
		}
	}

	return "", false
}

var (
	commitSubjectPattern = regexp.MustCompile(`^\[(.*)\]\s+(\w+):\s*(.*)$`)
	taskCodePattern      = regexp.MustCompile(`[A-Z]{2,}\d{6,}`)
)
