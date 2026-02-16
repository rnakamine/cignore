package template

import "testing"

func TestDefaultPatterns(t *testing.T) {
	patterns := DefaultPatterns()
	if len(patterns) == 0 {
		t.Fatal("DefaultPatterns() returned empty slice")
	}

	expectedPatterns := []string{"CLAUDE.md", ".claude/", ".claude/*", "claude.json"}
	if len(patterns) != len(expectedPatterns) {
		t.Fatalf("expected %d patterns, got %d", len(expectedPatterns), len(patterns))
	}

	for i, expected := range expectedPatterns {
		if patterns[i].Pattern != expected {
			t.Errorf("pattern[%d]: expected %q, got %q", i, expected, patterns[i].Pattern)
		}
		if patterns[i].Description == "" {
			t.Errorf("pattern[%d] (%s): description should not be empty", i, expected)
		}
	}
}

func TestGetPatternDescription(t *testing.T) {
	tests := []struct {
		pattern  string
		wantDesc bool
	}{
		{"CLAUDE.md", true},
		{".claude/", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		desc := GetPatternDescription(tt.pattern)
		if tt.wantDesc && desc == "" {
			t.Errorf("GetPatternDescription(%q): expected non-empty description", tt.pattern)
		}
		if !tt.wantDesc && desc != "" {
			t.Errorf("GetPatternDescription(%q): expected empty description, got %q", tt.pattern, desc)
		}
	}
}
