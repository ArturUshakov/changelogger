package changelogger

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestRepositoryLinkFromRemote(t *testing.T) {
	tests := []struct {
		name    string
		remote  string
		want    string
		wantErr bool
	}{
		{
			name:   "https remote",
			remote: "https://git.example/group/app.git",
			want:   "https://git.example/group/app/-/tags/",
		},
		{
			name:   "http remote",
			remote: "http://git.example/group/app",
			want:   "http://git.example/group/app/-/tags/",
		},
		{
			name:   "scp style remote",
			remote: "git@git.example:group/app.git",
			want:   "https://git.example/group/app/-/tags/",
		},
		{
			name:   "ssh url remote",
			remote: "ssh://git.example/group/app.git",
			want:   "https://git.example/group/app/-/tags/",
		},
		{
			name:    "empty",
			remote:  "",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			remote:  "file:///tmp/app.git",
			wantErr: true,
		},
		{
			name:    "malformed scp style",
			remote:  "git@git.example/group/app.git",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RepositoryLinkFromRemote(tt.remote)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RepositoryLinkFromRemote() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("RepositoryLinkFromRemote() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGitLastTag(t *testing.T) {
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "rev-list", "--tags", "--max-count=1"): {{output: "abc123\n"}},
		runnerKey("git", "describe", "--tags", "abc123"):        {{output: "1.2.3\n"}},
	})

	tag, err := (Git{runner: runner}).LastTag()
	if err != nil {
		t.Fatalf("LastTag() error = %v", err)
	}
	if tag != "1.2.3" {
		t.Fatalf("LastTag() = %q, want %q", tag, "1.2.3")
	}
}

func TestGitLastTagReturnsErrorsForMissingValues(t *testing.T) {
	t.Run("empty commit", func(t *testing.T) {
		runner := newScriptedRunner(map[string][]runnerResult{
			runnerKey("git", "rev-list", "--tags", "--max-count=1"): {{output: "\n"}},
		})

		_, err := (Git{runner: runner}).LastTag()
		if err == nil || !strings.Contains(err.Error(), "отсутствует последний тэг") {
			t.Fatalf("LastTag() error = %v, want missing tag error", err)
		}
	})

	t.Run("empty tag", func(t *testing.T) {
		runner := newScriptedRunner(map[string][]runnerResult{
			runnerKey("git", "rev-list", "--tags", "--max-count=1"): {{output: "abc123\n"}},
			runnerKey("git", "describe", "--tags", "abc123"):        {{output: "\n"}},
		})

		_, err := (Git{runner: runner}).LastTag()
		if err == nil || !strings.Contains(err.Error(), "отсутствует последний тэг") {
			t.Fatalf("LastTag() error = %v, want missing tag error", err)
		}
	})
}

func TestGitRepositoryLinkUsesOriginRemote(t *testing.T) {
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "remote", "get-url", "origin"): {{output: "https://git.example/group/app.git\n"}},
	})

	link, err := (Git{runner: runner}).RepositoryLink()
	if err != nil {
		t.Fatalf("RepositoryLink() error = %v", err)
	}
	if link != "https://git.example/group/app/-/tags/" {
		t.Fatalf("RepositoryLink() = %q", link)
	}
}

func TestGitMasterCommitTrimsOutput(t *testing.T) {
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "rev-parse", "origin/master"): {{output: "def456\n"}},
	})

	commit, err := (Git{runner: runner}).MasterCommit()
	if err != nil {
		t.Fatalf("MasterCommit() error = %v", err)
	}
	if commit != "def456" {
		t.Fatalf("MasterCommit() = %q, want %q", commit, "def456")
	}
}

func TestGitChangeLinesIncludesCherryLinesWhenAvailable(t *testing.T) {
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "log", "--pretty=format:%h|%an|%s|%cs", "--no-merges", "1.2.3..develop"): {{
			output: "001|Dev One|[OPS-AB111111-CD111111] feat: добавлен экспорт|2026-01-01\n",
		}},
		runnerKey("git", "cherry", "-v", "master123", "1.2.3"): {{
			output: "+ 002 [OPS-AB222222-CD222222] fix: исправлена проверка\n",
		}},
	})

	lines, err := (Git{runner: runner}).ChangeLines("1.2.3", "master123")
	if err != nil {
		t.Fatalf("ChangeLines() error = %v", err)
	}

	want := []string{
		"001|Dev One|[OPS-AB111111-CD111111] feat: добавлен экспорт|2026-01-01",
		"+ 002 [OPS-AB222222-CD222222] fix: исправлена проверка",
	}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("ChangeLines() = %#v, want %#v", lines, want)
	}
}

func TestGitChangeLinesIgnoresCherryError(t *testing.T) {
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "log", "--pretty=format:%h|%an|%s|%cs", "--no-merges", "1.2.3..develop"): {{
			output: "001|Dev One|[OPS-AB111111-CD111111] feat: добавлен экспорт|2026-01-01\n",
		}},
		runnerKey("git", "cherry", "-v", "master123", "1.2.3"): {{
			err: errors.New("not available"),
		}},
	})

	lines, err := (Git{runner: runner}).ChangeLines("1.2.3", "master123")
	if err != nil {
		t.Fatalf("ChangeLines() error = %v", err)
	}

	want := []string{"001|Dev One|[OPS-AB111111-CD111111] feat: добавлен экспорт|2026-01-01"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("ChangeLines() = %#v, want %#v", lines, want)
	}
}

func TestGitCreateBranchChecksOutExistingBranch(t *testing.T) {
	branch := "feature/OPS-AB123456-CD789000-assign-to-changelog"
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "branch", "--list", branch): {{output: branch + "\n"}},
		runnerKey("git", "checkout", branch):         {{}},
	})

	if err := (Git{runner: runner}).CreateBranch(branch); err != nil {
		t.Fatalf("CreateBranch() error = %v", err)
	}

	wantCommands := [][]string{
		{"git", "branch", "--list", branch},
		{"git", "checkout", branch},
	}
	if !reflect.DeepEqual(runner.commands, wantCommands) {
		t.Fatalf("commands = %#v, want %#v", runner.commands, wantCommands)
	}
}

func TestGitCreateBranchCreatesNewBranchFromDevelop(t *testing.T) {
	branch := "feature/OPS-AB123456-CD789000-assign-to-changelog"
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "branch", "--list", branch): {{output: ""}},
		runnerKey("git", "checkout", "develop"):      {{}},
		runnerKey("git", "checkout", "-b", branch):   {{}},
	})

	if err := (Git{runner: runner}).CreateBranch(branch); err != nil {
		t.Fatalf("CreateBranch() error = %v", err)
	}

	wantCommands := [][]string{
		{"git", "branch", "--list", branch},
		{"git", "checkout", "develop"},
		{"git", "checkout", "-b", branch},
	}
	if !reflect.DeepEqual(runner.commands, wantCommands) {
		t.Fatalf("commands = %#v, want %#v", runner.commands, wantCommands)
	}
}

func TestGitCommitPushAndDeleteBranch(t *testing.T) {
	branch := "feature/OPS-AB123456-CD789000-assign-to-changelog"
	runner := newScriptedRunner(map[string][]runnerResult{
		runnerKey("git", "add", "./CHANGELOG.md"):                            {{}},
		runnerKey("git", "commit", "-m", "wip: Отредактирован CHANGELOG.md"): {{}},
		runnerKey("git", "push", "origin", branch):                           {{}},
		runnerKey("git", "checkout", "develop"):                              {{}},
		runnerKey("git", "branch", "-D", branch):                             {{}},
	})
	git := Git{runner: runner}

	if err := git.Commit("./CHANGELOG.md"); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}
	if err := git.Push(branch); err != nil {
		t.Fatalf("Push() error = %v", err)
	}
	if err := git.DeleteBranch(branch); err != nil {
		t.Fatalf("DeleteBranch() error = %v", err)
	}

	wantCommands := [][]string{
		{"git", "add", "./CHANGELOG.md"},
		{"git", "commit", "-m", "wip: Отредактирован CHANGELOG.md"},
		{"git", "push", "origin", branch},
		{"git", "checkout", "develop"},
		{"git", "branch", "-D", branch},
	}
	if !reflect.DeepEqual(runner.commands, wantCommands) {
		t.Fatalf("commands = %#v, want %#v", runner.commands, wantCommands)
	}
}

func TestSplitLines(t *testing.T) {
	if got := splitLines(""); got != nil {
		t.Fatalf("splitLines(\"\") = %#v, want nil", got)
	}

	want := []string{"one", "two"}
	if got := splitLines("\none\ntwo\n"); !reflect.DeepEqual(got, want) {
		t.Fatalf("splitLines() = %#v, want %#v", got, want)
	}
}

type scriptedRunner struct {
	t         *testing.T
	results   map[string][]runnerResult
	commands  [][]string
	callIndex map[string]int
}

type runnerResult struct {
	output string
	err    error
}

func newScriptedRunner(results map[string][]runnerResult) *scriptedRunner {
	return &scriptedRunner{
		results:   results,
		callIndex: make(map[string]int),
	}
}

func (runner *scriptedRunner) Run(name string, args ...string) (string, error) {
	command := append([]string{name}, args...)
	runner.commands = append(runner.commands, command)

	key := runnerKey(name, args...)
	index := runner.callIndex[key]
	runner.callIndex[key]++

	results, ok := runner.results[key]
	if !ok || index >= len(results) {
		return "", fmt.Errorf("unexpected command: %s", strings.Join(command, " "))
	}

	result := results[index]
	return result.output, result.err
}

func runnerKey(name string, args ...string) string {
	parts := append([]string{name}, args...)
	return strings.Join(parts, "\x00")
}
