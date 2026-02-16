package gitignore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const sectionHeader = "# Claude Code files (managed by cignore)"

// GitIgnore manages a .gitignore file.
type GitIgnore struct {
	path     string
	patterns []Pattern
}

// Load reads and parses a .gitignore file from the given repository path.
// If the file does not exist, an empty GitIgnore is returned.
func Load(repoPath string) (*GitIgnore, error) {
	gitignorePath := filepath.Join(repoPath, ".gitignore")

	g := &GitIgnore{
		path: gitignorePath,
	}

	f, err := os.Open(gitignorePath)
	if os.IsNotExist(err) {
		return g, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		p := Pattern{
			Line:      line,
			IsIgnored: !isComment(line) && !isBlank(line),
		}
		g.patterns = append(g.patterns, p)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read .gitignore: %w", err)
	}

	return g, nil
}

// HasPattern checks if the given pattern exists as an active ignore entry.
func (g *GitIgnore) HasPattern(pattern string) bool {
	norm := normalizePattern(pattern)
	for _, p := range g.patterns {
		if p.IsIgnored && normalizePattern(p.Line) == norm {
			return true
		}
	}
	return false
}

// AddPattern adds a pattern to the cignore-managed section.
// If the pattern already exists, it is a no-op.
func (g *GitIgnore) AddPattern(pattern string) error {
	if g.HasPattern(pattern) {
		return nil
	}

	// Ensure the managed section exists and add the pattern there
	sectionIdx := g.findSectionHeader()
	if sectionIdx == -1 {
		// Add a blank line separator if file is not empty and doesn't end with blank
		if len(g.patterns) > 0 && !isBlank(g.patterns[len(g.patterns)-1].Line) {
			g.patterns = append(g.patterns, Pattern{Line: "", IsIgnored: false})
		}
		g.patterns = append(g.patterns, Pattern{Line: sectionHeader, IsIgnored: false})
		g.patterns = append(g.patterns, Pattern{Line: pattern, IsIgnored: true})
	} else {
		// Insert after the last pattern in the managed section
		insertIdx := g.findSectionEnd(sectionIdx)
		newPattern := Pattern{Line: pattern, IsIgnored: true}
		g.patterns = append(g.patterns[:insertIdx], append([]Pattern{newPattern}, g.patterns[insertIdx:]...)...)
	}

	return nil
}

// RemovePattern removes a pattern from the .gitignore file.
// If the pattern does not exist, it is a no-op.
func (g *GitIgnore) RemovePattern(pattern string) error {
	norm := normalizePattern(pattern)
	newPatterns := make([]Pattern, 0, len(g.patterns))
	for _, p := range g.patterns {
		if p.IsIgnored && normalizePattern(p.Line) == norm {
			continue
		}
		newPatterns = append(newPatterns, p)
	}
	g.patterns = newPatterns

	// Clean up: remove section header if section is now empty
	g.cleanupEmptySection()

	return nil
}

// Save writes the current patterns back to the .gitignore file.
func (g *GitIgnore) Save() error {
	f, err := os.Create(g.path)
	if err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for i, p := range g.patterns {
		if _, err := w.WriteString(p.Line); err != nil {
			return fmt.Errorf("failed to write pattern: %w", err)
		}
		if i < len(g.patterns)-1 {
			if _, err := w.WriteString("\n"); err != nil {
				return fmt.Errorf("failed to write newline: %w", err)
			}
		}
	}
	// Always end with a newline
	if _, err := w.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write final newline: %w", err)
	}
	return w.Flush()
}

// findSectionHeader returns the index of the cignore section header, or -1.
func (g *GitIgnore) findSectionHeader() int {
	for i, p := range g.patterns {
		if strings.TrimSpace(p.Line) == sectionHeader {
			return i
		}
	}
	return -1
}

// findSectionEnd returns the index after the last pattern in the managed section.
func (g *GitIgnore) findSectionEnd(headerIdx int) int {
	i := headerIdx + 1
	for i < len(g.patterns) {
		line := g.patterns[i].Line
		// Stop at the next comment section or blank line followed by a comment
		if isBlank(line) {
			// Check if next non-blank line is a different comment section
			j := i + 1
			for j < len(g.patterns) && isBlank(g.patterns[j].Line) {
				j++
			}
			if j < len(g.patterns) && isComment(g.patterns[j].Line) && strings.TrimSpace(g.patterns[j].Line) != sectionHeader {
				break
			}
		}
		if isComment(line) && strings.TrimSpace(line) != sectionHeader {
			break
		}
		i++
	}
	return i
}

// cleanupEmptySection removes the section header if no patterns remain in the section.
func (g *GitIgnore) cleanupEmptySection() {
	headerIdx := g.findSectionHeader()
	if headerIdx == -1 {
		return
	}

	// Check if there are any active patterns in the section
	hasPatterns := false
	for i := headerIdx + 1; i < len(g.patterns); i++ {
		line := g.patterns[i].Line
		if isComment(line) && strings.TrimSpace(line) != sectionHeader {
			break
		}
		if !isBlank(line) && !isComment(line) {
			hasPatterns = true
			break
		}
	}

	if !hasPatterns {
		// Remove the section header and any trailing blank line before it
		newPatterns := make([]Pattern, 0, len(g.patterns))
		// Remove blank line before header if present
		for i, p := range g.patterns {
			if i == headerIdx {
				continue
			}
			if i == headerIdx-1 && isBlank(p.Line) {
				continue
			}
			newPatterns = append(newPatterns, p)
		}
		g.patterns = newPatterns
	}
}
