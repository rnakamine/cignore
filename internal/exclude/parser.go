package exclude

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const sectionHeader = "# Claude Code files (managed by cexclude)"

// ExcludeFile manages a .git/info/exclude file.
type ExcludeFile struct {
	path     string
	patterns []Pattern
}

// Load reads and parses a .git/info/exclude file from the given repository path.
// If the file does not exist, an empty ExcludeFile is returned.
func Load(repoPath string) (*ExcludeFile, error) {
	excludePath := filepath.Join(repoPath, ".git", "info", "exclude")

	e := &ExcludeFile{
		path: excludePath,
	}

	f, err := os.Open(excludePath)
	if os.IsNotExist(err) {
		return e, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open .git/info/exclude: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		p := Pattern{
			Line:      line,
			IsIgnored: !isComment(line) && !isBlank(line),
		}
		e.patterns = append(e.patterns, p)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read .git/info/exclude: %w", err)
	}

	return e, nil
}

// HasPattern checks if the given pattern exists as an active ignore entry.
func (e *ExcludeFile) HasPattern(pattern string) bool {
	norm := normalizePattern(pattern)
	for _, p := range e.patterns {
		if p.IsIgnored && normalizePattern(p.Line) == norm {
			return true
		}
	}
	return false
}

// AddPattern adds a pattern to the cexclude-managed section.
// If the pattern already exists, it is a no-op.
func (e *ExcludeFile) AddPattern(pattern string) error {
	if e.HasPattern(pattern) {
		return nil
	}

	// Ensure the managed section exists and add the pattern there
	sectionIdx := e.findSectionHeader()
	if sectionIdx == -1 {
		// Add a blank line separator if file is not empty and doesn't end with blank
		if len(e.patterns) > 0 && !isBlank(e.patterns[len(e.patterns)-1].Line) {
			e.patterns = append(e.patterns, Pattern{Line: "", IsIgnored: false})
		}
		e.patterns = append(e.patterns, Pattern{Line: sectionHeader, IsIgnored: false})
		e.patterns = append(e.patterns, Pattern{Line: pattern, IsIgnored: true})
	} else {
		// Insert after the last pattern in the managed section
		insertIdx := e.findSectionEnd(sectionIdx)
		newPattern := Pattern{Line: pattern, IsIgnored: true}
		e.patterns = append(e.patterns[:insertIdx], append([]Pattern{newPattern}, e.patterns[insertIdx:]...)...)
	}

	return nil
}

// RemovePattern removes a pattern from the exclude file.
// If the pattern does not exist, it is a no-op.
func (e *ExcludeFile) RemovePattern(pattern string) error {
	norm := normalizePattern(pattern)
	newPatterns := make([]Pattern, 0, len(e.patterns))
	for _, p := range e.patterns {
		if p.IsIgnored && normalizePattern(p.Line) == norm {
			continue
		}
		newPatterns = append(newPatterns, p)
	}
	e.patterns = newPatterns

	// Clean up: remove section header if section is now empty
	e.cleanupEmptySection()

	return nil
}

// Save writes the current patterns back to the .git/info/exclude file.
// If no content remains (all lines removed), the file is deleted.
func (e *ExcludeFile) Save() error {
	// If all lines have been removed, delete the file
	if len(e.patterns) == 0 {
		if err := os.Remove(e.path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove .git/info/exclude: %w", err)
		}
		return nil
	}

	// Ensure .git/info/ directory exists
	dir := filepath.Dir(e.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.Create(e.path)
	if err != nil {
		return fmt.Errorf("failed to write .git/info/exclude: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for i, p := range e.patterns {
		if _, err := w.WriteString(p.Line); err != nil {
			return fmt.Errorf("failed to write pattern: %w", err)
		}
		if i < len(e.patterns)-1 {
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

// findSectionHeader returns the index of the cexclude section header, or -1.
func (e *ExcludeFile) findSectionHeader() int {
	for i, p := range e.patterns {
		if strings.TrimSpace(p.Line) == sectionHeader {
			return i
		}
	}
	return -1
}

// findSectionEnd returns the index after the last pattern in the managed section.
func (e *ExcludeFile) findSectionEnd(headerIdx int) int {
	i := headerIdx + 1
	for i < len(e.patterns) {
		line := e.patterns[i].Line
		// Stop at the next comment section or blank line followed by a comment
		if isBlank(line) {
			// Check if next non-blank line is a different comment section
			j := i + 1
			for j < len(e.patterns) && isBlank(e.patterns[j].Line) {
				j++
			}
			if j < len(e.patterns) && isComment(e.patterns[j].Line) && strings.TrimSpace(e.patterns[j].Line) != sectionHeader {
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
func (e *ExcludeFile) cleanupEmptySection() {
	headerIdx := e.findSectionHeader()
	if headerIdx == -1 {
		return
	}

	// Check if there are any active patterns in the section
	hasPatterns := false
	for i := headerIdx + 1; i < len(e.patterns); i++ {
		line := e.patterns[i].Line
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
		newPatterns := make([]Pattern, 0, len(e.patterns))
		// Remove blank line before header if present
		for i, p := range e.patterns {
			if i == headerIdx {
				continue
			}
			if i == headerIdx-1 && isBlank(p.Line) {
				continue
			}
			newPatterns = append(newPatterns, p)
		}
		e.patterns = newPatterns
	}
}
