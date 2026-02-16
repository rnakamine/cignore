package template

// PatternInfo holds a gitignore pattern and its description.
type PatternInfo struct {
	Pattern     string
	Description string
}

// DefaultPatterns returns the list of Claude Code related gitignore patterns.
func DefaultPatterns() []PatternInfo {
	return []PatternInfo{
		{Pattern: "CLAUDE.md", Description: "Claude Code project information file"},
		{Pattern: ".claude/", Description: "Claude Code configuration directory"},
		{Pattern: ".claude/*", Description: "All files in Claude Code config directory"},
		{Pattern: "claude.json", Description: "Claude Code configuration file"},
	}
}

// GetPatternDescription returns the description for a given pattern.
// Returns an empty string if the pattern is not found.
func GetPatternDescription(pattern string) string {
	for _, p := range DefaultPatterns() {
		if p.Pattern == pattern {
			return p.Description
		}
	}
	return ""
}
