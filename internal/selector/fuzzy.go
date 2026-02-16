package selector

import (
	"fmt"

	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

// Item represents a selectable pattern in the fuzzy finder.
type Item struct {
	Pattern     string
	IsIgnored   bool
	Description string
}

// SelectPatterns displays a fuzzy finder with multi-select support.
// Users select patterns they want ignored. The returned items reflect the new state.
// Patterns selected via Tab will be added; unselected ones will be removed.
func SelectPatterns(items []Item) ([]Item, error) {
	selectedIdxs, err := fuzzyfinder.FindMulti(
		items,
		func(i int) string {
			status := "[ ] "
			if items[i].IsIgnored {
				status = "[✓] "
			}
			return fmt.Sprintf("%s%s", status, items[i].Pattern)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			item := items[i]
			status := "not ignored"
			if item.IsIgnored {
				status = "currently ignored"
			}
			return fmt.Sprintf(
				"Pattern:     %s\nStatus:      %s\nDescription: %s\n\nUse Tab to toggle selection, Enter to apply.",
				item.Pattern,
				status,
				item.Description,
			)
		}),
		fuzzyfinder.WithHeader("Select Claude Code patterns to add to .gitignore (Tab: toggle, Enter: apply)"),
	)
	if err != nil {
		return nil, err
	}

	// Build a set of selected indices
	selected := make(map[int]bool)
	for _, idx := range selectedIdxs {
		selected[idx] = true
	}

	// Update items based on selection
	result := make([]Item, len(items))
	copy(result, items)
	for i := range result {
		result[i].IsIgnored = selected[i]
	}

	return result, nil
}
