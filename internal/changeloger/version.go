package changeloger

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func ParseVersion(value string) (Version, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("неверный формат тэга %q, ожидается X.Y.Z", value)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("неверный major в тэге %q: %w", value, err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("неверный minor в тэге %q: %w", value, err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("неверный patch в тэге %q: %w", value, err)
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

func (version Version) Next(level string) (Version, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "1", "major":
		return Version{Major: version.Major + 1}, nil
	case "2", "minor":
		return Version{Major: version.Major, Minor: version.Minor + 1}, nil
	case "3", "fix", "patch":
		return Version{Major: version.Major, Minor: version.Minor, Patch: version.Patch + 1}, nil
	default:
		return Version{}, fmt.Errorf("неизвестное значение версии %q", level)
	}
}

func (version Version) String() string {
	return fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
}
