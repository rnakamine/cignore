package exclude

import "testing"

func TestIsComment(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"# comment", true},
		{"  # indented comment", true},
		{"node_modules/", false},
		{"", false},
		{"CLAUDE.md", false},
	}

	for _, tt := range tests {
		got := isComment(tt.line)
		if got != tt.want {
			t.Errorf("isComment(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestIsBlank(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"\t", true},
		{"foo", false},
		{" foo", false},
	}

	for _, tt := range tests {
		got := isBlank(tt.line)
		if got != tt.want {
			t.Errorf("isBlank(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestNormalizePattern(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"CLAUDE.md", "CLAUDE.md"},
		{"  CLAUDE.md  ", "CLAUDE.md"},
		{".claude/", ".claude/"},
	}

	for _, tt := range tests {
		got := normalizePattern(tt.input)
		if got != tt.want {
			t.Errorf("normalizePattern(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
