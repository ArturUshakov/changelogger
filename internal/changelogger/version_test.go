package changelogger

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    Version
		wantErr bool
	}{
		{name: "valid", value: "1.2.3", want: Version{Major: 1, Minor: 2, Patch: 3}},
		{name: "missing patch", value: "1.2", wantErr: true},
		{name: "invalid major", value: "x.2.3", wantErr: true},
		{name: "invalid minor", value: "1.x.3", wantErr: true},
		{name: "invalid patch", value: "1.2.x", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ParseVersion() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestVersionNext(t *testing.T) {
	version := Version{Major: 1, Minor: 2, Patch: 3}

	tests := []struct {
		name    string
		level   string
		want    Version
		wantErr bool
	}{
		{name: "major number", level: "1", want: Version{Major: 2}},
		{name: "major word", level: " major ", want: Version{Major: 2}},
		{name: "minor number", level: "2", want: Version{Major: 1, Minor: 3}},
		{name: "minor word", level: "minor", want: Version{Major: 1, Minor: 3}},
		{name: "fix number", level: "3", want: Version{Major: 1, Minor: 2, Patch: 4}},
		{name: "fix word", level: "fix", want: Version{Major: 1, Minor: 2, Patch: 4}},
		{name: "patch alias", level: "patch", want: Version{Major: 1, Minor: 2, Patch: 4}},
		{name: "unknown", level: "release", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := version.Next(tt.level)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Next() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("Next() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	version := Version{Major: 1, Minor: 2, Patch: 3}

	if got := version.String(); got != "1.2.3" {
		t.Fatalf("String() = %q, want %q", got, "1.2.3")
	}
}
