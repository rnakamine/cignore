package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rnakamine/cignore/internal/exclude"
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

	// Load .git/info/exclude
	ef, err := exclude.Load(repoPath)
	if err != nil {
		return err
	}

	// Get git-tracked files to exclude from selection
	tracked, err := gitTrackedFiles(repoPath)
	if err != nil {
		return err
	}

	// Build items from template patterns, filtering out tracked files
	patterns := template.CollectPatterns(repoPath)
	var items []selector.Item
	for _, p := range patterns {
		if tracked[p.Pattern] {
			continue
		}
		items = append(items, selector.Item{
			Pattern:     p.Pattern,
			IsIgnored:   ef.HasPattern(p.Pattern),
			Description: p.Description,
		})
	}

	if len(items) == 0 {
		fmt.Println("No patterns to manage. All Claude Code files are already tracked by git.")
		return nil
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
			if err := ef.AddPattern(item.Pattern); err != nil {
				return fmt.Errorf("failed to add pattern %q: %w", item.Pattern, err)
			}
			added = append(added, item.Pattern)
		} else if wasIgnored && !nowIgnored {
			if err := ef.RemovePattern(item.Pattern); err != nil {
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
	if err := ef.Save(); err != nil {
		return err
	}

	// Print summary
	fmt.Println("Applied changes to .git/info/exclude:")
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

// gitTrackedFiles returns a set of files currently tracked by git.
func gitTrackedFiles(repoPath string) (map[string]bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "ls-files")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked files: %w", err)
	}

	tracked := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			tracked[line] = true
		}
	}
	return tracked, nil
}
