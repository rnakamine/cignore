package exclude

import "strings"

// Pattern represents a single line in a .git/info/exclude file.
type Pattern struct {
	Line      string
	IsIgnored bool // true if this line is an active ignore pattern (not comment/blank)
}

// isComment returns true if the line is a comment.
func isComment(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// isBlank returns true if the line is empty or whitespace only.
func isBlank(line string) bool {
	return strings.TrimSpace(line) == ""
}

// normalizePattern trims whitespace from a pattern for comparison.
func normalizePattern(pattern string) string {
	return strings.TrimSpace(pattern)
}
