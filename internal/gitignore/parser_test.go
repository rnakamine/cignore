package gitignore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestRepo(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	infoDir := filepath.Join(dir, ".git", "info")
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if content != "" {
		err := os.WriteFile(filepath.Join(infoDir, "exclude"), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func readExclude(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	gi, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(gi.patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(gi.patterns))
	}
}

func TestLoadExisting(t *testing.T) {
	content := "node_modules/\n*.log\n# comment\n\nCLAUDE.md\n"
	dir := setupTestRepo(t, content)

	gi, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(gi.patterns) != 5 {
		t.Fatalf("expected 5 patterns, got %d", len(gi.patterns))
	}

	// Check active patterns
	if !gi.HasPattern("node_modules/") {
		t.Error("expected HasPattern(node_modules/) to be true")
	}
	if !gi.HasPattern("CLAUDE.md") {
		t.Error("expected HasPattern(CLAUDE.md) to be true")
	}
	if gi.HasPattern("# comment") {
		t.Error("expected HasPattern(# comment) to be false")
	}
}

func TestHasPattern(t *testing.T) {
	content := "CLAUDE.md\n.claude/\n"
	dir := setupTestRepo(t, content)

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		pattern string
		want    bool
	}{
		{"CLAUDE.md", true},
		{".claude/", true},
		{"claude.json", false},
		{"", false},
	}

	for _, tt := range tests {
		got := gi.HasPattern(tt.pattern)
		if got != tt.want {
			t.Errorf("HasPattern(%q) = %v, want %v", tt.pattern, got, tt.want)
		}
	}
}

func TestAddPattern(t *testing.T) {
	dir := setupTestRepo(t, "node_modules/\n")

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := gi.AddPattern("CLAUDE.md"); err != nil {
		t.Fatal(err)
	}
	if err := gi.AddPattern(".claude/"); err != nil {
		t.Fatal(err)
	}

	if !gi.HasPattern("CLAUDE.md") {
		t.Error("CLAUDE.md should be present after adding")
	}
	if !gi.HasPattern(".claude/") {
		t.Error(".claude/ should be present after adding")
	}

	if err := gi.Save(); err != nil {
		t.Fatal(err)
	}

	content := readExclude(t, dir)
	if !strings.Contains(content, sectionHeader) {
		t.Error("saved file should contain section header")
	}
	if !strings.Contains(content, "CLAUDE.md") {
		t.Error("saved file should contain CLAUDE.md")
	}
	if !strings.Contains(content, ".claude/") {
		t.Error("saved file should contain .claude/")
	}
}

func TestAddPatternDuplicate(t *testing.T) {
	content := "CLAUDE.md\n"
	dir := setupTestRepo(t, content)

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := gi.AddPattern("CLAUDE.md"); err != nil {
		t.Fatal(err)
	}

	// Pattern count should not change
	count := 0
	for _, p := range gi.patterns {
		if p.IsIgnored && normalizePattern(p.Line) == "CLAUDE.md" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 occurrence of CLAUDE.md, got %d", count)
	}
}

func TestRemovePattern(t *testing.T) {
	content := "node_modules/\nCLAUDE.md\n.claude/\n"
	dir := setupTestRepo(t, content)

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := gi.RemovePattern("CLAUDE.md"); err != nil {
		t.Fatal(err)
	}

	if gi.HasPattern("CLAUDE.md") {
		t.Error("CLAUDE.md should not be present after removal")
	}
	if !gi.HasPattern("node_modules/") {
		t.Error("node_modules/ should still be present")
	}

	if err := gi.Save(); err != nil {
		t.Fatal(err)
	}

	result := readExclude(t, dir)
	if strings.Contains(result, "CLAUDE.md") {
		t.Error("saved file should not contain CLAUDE.md")
	}
	if !strings.Contains(result, "node_modules/") {
		t.Error("saved file should still contain node_modules/")
	}
}

func TestRemovePatternNonexistent(t *testing.T) {
	content := "node_modules/\n"
	dir := setupTestRepo(t, content)

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Should not error
	if err := gi.RemovePattern("CLAUDE.md"); err != nil {
		t.Fatal(err)
	}

	if len(gi.patterns) != 1 {
		t.Errorf("expected 1 pattern, got %d", len(gi.patterns))
	}
}

func TestAddToEmptyFile(t *testing.T) {
	dir := t.TempDir()

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := gi.AddPattern("CLAUDE.md"); err != nil {
		t.Fatal(err)
	}
	if err := gi.Save(); err != nil {
		t.Fatal(err)
	}

	content := readExclude(t, dir)
	if !strings.Contains(content, sectionHeader) {
		t.Error("should contain section header")
	}
	if !strings.Contains(content, "CLAUDE.md") {
		t.Error("should contain CLAUDE.md")
	}
}

func TestSectionCleanup(t *testing.T) {
	content := "node_modules/\n\n# Claude Code files (managed by cignore)\nCLAUDE.md\n"
	dir := setupTestRepo(t, content)

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := gi.RemovePattern("CLAUDE.md"); err != nil {
		t.Fatal(err)
	}
	if err := gi.Save(); err != nil {
		t.Fatal(err)
	}

	result := readExclude(t, dir)
	if strings.Contains(result, sectionHeader) {
		t.Error("section header should be removed when section is empty")
	}
}

func TestSaveRemovesEmptyFile(t *testing.T) {
	content := "# Claude Code files (managed by cignore)\nCLAUDE.md\n"
	dir := setupTestRepo(t, content)

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := gi.RemovePattern("CLAUDE.md"); err != nil {
		t.Fatal(err)
	}
	if err := gi.Save(); err != nil {
		t.Fatal(err)
	}

	excludePath := filepath.Join(dir, ".git", "info", "exclude")
	if _, err := os.Stat(excludePath); !os.IsNotExist(err) {
		t.Error("exclude file should be removed when empty")
	}
}

func TestSavePreservesFileWithExistingComments(t *testing.T) {
	// Simulate git's default exclude file with only comments
	content := "# git ls-files --others --exclude-from=.git/info/exclude\n# Lines that start with '#' are comments.\n\n# Claude Code files (managed by cignore)\nCLAUDE.md\n"
	dir := setupTestRepo(t, content)

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := gi.RemovePattern("CLAUDE.md"); err != nil {
		t.Fatal(err)
	}
	if err := gi.Save(); err != nil {
		t.Fatal(err)
	}

	// File should still exist because it has pre-existing comment lines
	excludePath := filepath.Join(dir, ".git", "info", "exclude")
	if _, err := os.Stat(excludePath); os.IsNotExist(err) {
		t.Error("exclude file should be preserved when it has pre-existing comments")
	}

	result := readExclude(t, dir)
	if !strings.Contains(result, "# git ls-files") {
		t.Error("should preserve original comments")
	}
	if strings.Contains(result, "CLAUDE.md") {
		t.Error("should not contain removed pattern")
	}
}

func TestSavePreservesExistingContent(t *testing.T) {
	content := "# My project\nnode_modules/\n*.log\ndist/\n"
	dir := setupTestRepo(t, content)

	gi, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := gi.AddPattern("CLAUDE.md"); err != nil {
		t.Fatal(err)
	}
	if err := gi.Save(); err != nil {
		t.Fatal(err)
	}

	result := readExclude(t, dir)
	if !strings.Contains(result, "# My project") {
		t.Error("should preserve existing comments")
	}
	if !strings.Contains(result, "node_modules/") {
		t.Error("should preserve existing patterns")
	}
	if !strings.Contains(result, "dist/") {
		t.Error("should preserve existing patterns")
	}
}
