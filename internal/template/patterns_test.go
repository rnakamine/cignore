package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectPatternsNoClaudeDir(t *testing.T) {
	dir := t.TempDir()
	patterns := CollectPatterns(dir)

	// Should return only fixed patterns when .claude/ doesn't exist
	if len(patterns) != 2 {
		t.Fatalf("expected 2 fixed patterns, got %d", len(patterns))
	}
	if patterns[0].Pattern != "CLAUDE.md" {
		t.Errorf("expected first pattern to be CLAUDE.md, got %q", patterns[0].Pattern)
	}
	if patterns[1].Pattern != ".claude/" {
		t.Errorf("expected second pattern to be .claude/, got %q", patterns[1].Pattern)
	}
}

func TestCollectPatternsWithFiles(t *testing.T) {
	dir := t.TempDir()

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

	// 2 fixed + 2 dynamic files
	if len(patterns) != 4 {
		t.Fatalf("expected 4 patterns, got %d", len(patterns))
	}

	// Check fixed patterns
	if patterns[0].Pattern != "CLAUDE.md" {
		t.Errorf("expected CLAUDE.md, got %q", patterns[0].Pattern)
	}
	if patterns[1].Pattern != ".claude/" {
		t.Errorf("expected .claude/, got %q", patterns[1].Pattern)
	}

	// Check dynamic patterns (order may vary based on WalkDir, but typically alphabetical)
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
	patterns := CollectPatterns(dir)

	for _, p := range patterns {
		if p.Description == "" {
			t.Errorf("pattern %q: description should not be empty", p.Pattern)
		}
	}
}
