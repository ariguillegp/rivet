package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ariguillegp/solo/internal/core"
)

func hasExactMatch(worktrees []core.Worktree, query string) bool {
	query = strings.ToLower(query)
	for _, wt := range worktrees {
		if strings.ToLower(wt.Name) == query || strings.ToLower(wt.Branch) == query {
			return true
		}
	}
	return false
}

func (m Model) View() string {
	var content string

	switch m.core.Mode {
	case core.ModeLoading:
		content = m.spinner.View() + " Scanning..."

	case core.ModeBrowsing, core.ModeCreateDir:
		prompt := promptStyle.Render("Enter the project directory")
		input := prompt + "\n" + m.input.View()

		if dir, ok := m.core.SelectedDir(); ok {
			suggestion := suggestionStyle.Render(dir.Path)
			nav := ""
			if len(m.core.Filtered) > 1 {
				nav = navStyle.Render(fmt.Sprintf("  [%d/%d]", m.core.SelectedIdx+1, len(m.core.Filtered)))
			}
			content = input + "\n" + suggestion + nav
		} else if m.core.Query != "" {
			content = input + "\n" + suggestionStyle.Render("(create new)")
		} else {
			content = input
		}

	case core.ModeWorktree:
		prompt := promptStyle.Render("Select worktree or create new branch")
		input := prompt + "\n" + m.worktreeInput.View()

		if wt, ok := m.core.SelectedWorktree(); ok {
			suggestion := suggestionStyle.Render(wt.Path)
			nav := ""
			if len(m.core.FilteredWT) > 1 {
				nav = navStyle.Render(fmt.Sprintf("  [%d/%d]", m.core.WorktreeIdx+1, len(m.core.FilteredWT)))
			}
			content = input + "\n" + suggestion + nav
		} else if m.core.WorktreeQuery != "" {
			content = input + "\n" + suggestionStyle.Render("(create new: "+m.core.WorktreeQuery+")")
		} else {
			content = input
		}

	case core.ModeError:
		content = errorStyle.Render(fmt.Sprintf("Error: %v", m.core.Err))
	}

	box := boxStyle.Render(content)

	if m.height <= 0 || m.width <= 0 {
		return box
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}
