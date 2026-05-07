package changeloger

import "regexp"

func validBranchTask(value string) bool {
	return branchTaskPattern.MatchString(value)
}

var branchTaskPattern = regexp.MustCompile(`(?i)^[A-Z]+\d+-[A-Z]+\d+$`)
