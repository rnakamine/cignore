package template

import (
	"os"
	"path/filepath"
)

// PatternInfo holds a gitignore pattern and its description.
type PatternInfo struct {
	Pattern     string
	Description string
}

// CollectPatterns returns fixed Claude Code patterns plus dynamically discovered
// files under .claude/ in the given repository path.
func CollectPatterns(repoPath string) []PatternInfo {
	patterns := []PatternInfo{
		{Pattern: "CLAUDE.md", Description: "Claude Code project information file"},
		{Pattern: ".claude/", Description: "Claude Code configuration directory"},
	}

	// Recursively scan .claude/ directory for individual files
	claudeDir := filepath.Join(repoPath, ".claude")
	filepath.WalkDir(claudeDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip entries that can't be read
		}
		if d.IsDir() {
			return nil // only list files, not directories
		}
		rel, err := filepath.Rel(repoPath, path)
		if err != nil {
			return nil
		}
		patterns = append(patterns, PatternInfo{
			Pattern:     rel,
			Description: "Claude Code file",
		})
		return nil
	})

	return patterns
}
