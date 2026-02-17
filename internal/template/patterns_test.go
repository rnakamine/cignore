package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectPatternsNothingExists(t *testing.T) {
	dir := t.TempDir()
	patterns := CollectPatterns(dir)

	// Neither CLAUDE.md nor .claude/ exists → empty
	if len(patterns) != 0 {
		t.Fatalf("expected 0 patterns when nothing exists, got %d", len(patterns))
	}
}

func TestCollectPatternsOnlyCLAUDEmd(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}

	patterns := CollectPatterns(dir)
	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}
	if patterns[0].Pattern != "CLAUDE.md" {
		t.Errorf("expected CLAUDE.md, got %q", patterns[0].Pattern)
	}
}

func TestCollectPatternsWithFiles(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .claude/ directory with files
	claudeDir := filepath.Join(dir, ".claude")
	commandsDir := filepath.Join(claudeDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandsDir, "foo.md"), []byte("# foo"), 0644); err != nil {
		t.Fatal(err)
	}

	patterns := CollectPatterns(dir)

	// 1 CLAUDE.md + 1 .claude/ + 2 dynamic files = 4
	if len(patterns) != 4 {
		t.Fatalf("expected 4 patterns, got %d", len(patterns))
	}

	if patterns[0].Pattern != "CLAUDE.md" {
		t.Errorf("expected CLAUDE.md, got %q", patterns[0].Pattern)
	}
	if patterns[1].Pattern != ".claude/" {
		t.Errorf("expected .claude/, got %q", patterns[1].Pattern)
	}

	// Check dynamic patterns
	dynamicPatterns := make(map[string]bool)
	for _, p := range patterns[2:] {
		dynamicPatterns[p.Pattern] = true
	}
	expectedFiles := []string{
		filepath.Join(".claude", "commands", "foo.md"),
		filepath.Join(".claude", "settings.json"),
	}
	for _, f := range expectedFiles {
		if !dynamicPatterns[f] {
			t.Errorf("expected dynamic pattern %q to be present", f)
		}
	}
}

func TestCollectPatternsDescription(t *testing.T) {
	dir := t.TempDir()

	// Create files so patterns are returned
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}

	patterns := CollectPatterns(dir)
	for _, p := range patterns {
		if p.Description == "" {
			t.Errorf("pattern %q: description should not be empty", p.Pattern)
		}
	}
}
