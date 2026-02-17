package template

import (
	"os"
	"path/filepath"
)

// PatternInfo holds an exclude pattern and its description.
type PatternInfo struct {
	Pattern     string
	Description string
}

// CollectPatterns returns fixed Claude Code patterns plus dynamically discovered
// files under .claude/ in the given repository path.
// Only patterns for existing files/directories are included.
func CollectPatterns(repoPath string) []PatternInfo {
	var patterns []PatternInfo

	if _, err := os.Stat(filepath.Join(repoPath, "CLAUDE.md")); err == nil {
		patterns = append(patterns, PatternInfo{Pattern: "CLAUDE.md", Description: "Claude Code project information file"})
	}

	claudeDir := filepath.Join(repoPath, ".claude")
	if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
		patterns = append(patterns, PatternInfo{Pattern: ".claude/", Description: "Claude Code configuration directory"})
	}

	// Recursively scan .claude/ directory for individual files
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
