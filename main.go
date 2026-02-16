package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rnakamine/cignore/internal/gitignore"
	"github.com/rnakamine/cignore/internal/selector"
	"github.com/rnakamine/cignore/internal/template"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Detect git repository root
	repoPath, err := findGitRoot()
	if err != nil {
		return fmt.Errorf("not a git repository (or any of the parent directories)\n  Run this command inside a git repository")
	}

	// Load .gitignore
	gi, err := gitignore.Load(repoPath)
	if err != nil {
		return err
	}

	// Build items from template patterns
	patterns := template.DefaultPatterns()
	items := make([]selector.Item, len(patterns))
	for i, p := range patterns {
		items[i] = selector.Item{
			Pattern:     p.Pattern,
			IsIgnored:   gi.HasPattern(p.Pattern),
			Description: p.Description,
		}
	}

	// Show fuzzy finder
	updated, err := selector.SelectPatterns(items)
	if err != nil {
		// User cancelled (Esc/Ctrl-C)
		if strings.Contains(err.Error(), "abort") || strings.Contains(err.Error(), "cancel") {
			fmt.Println("No changes made.")
			return nil
		}
		return err
	}

	// Apply changes
	var added, removed []string
	for i, item := range updated {
		wasIgnored := items[i].IsIgnored
		nowIgnored := item.IsIgnored

		if !wasIgnored && nowIgnored {
			if err := gi.AddPattern(item.Pattern); err != nil {
				return fmt.Errorf("failed to add pattern %q: %w", item.Pattern, err)
			}
			added = append(added, item.Pattern)
		} else if wasIgnored && !nowIgnored {
			if err := gi.RemovePattern(item.Pattern); err != nil {
				return fmt.Errorf("failed to remove pattern %q: %w", item.Pattern, err)
			}
			removed = append(removed, item.Pattern)
		}
	}

	if len(added) == 0 && len(removed) == 0 {
		fmt.Println("No changes made.")
		return nil
	}

	// Save changes
	if err := gi.Save(); err != nil {
		return err
	}

	// Print summary
	fmt.Println("Applied changes to .gitignore:")
	for _, p := range added {
		fmt.Printf("  + Added: %s\n", p)
	}
	for _, p := range removed {
		fmt.Printf("  - Removed: %s\n", p)
	}
	fmt.Printf("\n%d pattern(s) updated.\n", len(added)+len(removed))

	return nil
}

// findGitRoot returns the root directory of the current git repository.
func findGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
