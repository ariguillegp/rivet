package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ariguillegp/solo/internal/core"
)

const maxSuggestions = 5

func renderSuggestionList(lines []string, selectedIdx int) string {
	if len(lines) == 0 {
		return ""
	}

	start := 0
	if selectedIdx >= maxSuggestions {
		start = selectedIdx - maxSuggestions + 1
	}
	end := start + maxSuggestions
	if end > len(lines) {
		end = len(lines)
		if end-start < maxSuggestions && start > 0 {
			start = end - maxSuggestions
			if start < 0 {
				start = 0
			}
		}
	}

	var out strings.Builder
	for i := start; i < end; i++ {
		if i > start {
			out.WriteString("\n")
		}

		prefix := "  "
		if i == selectedIdx {
			prefix = "> "
		}
		row := suggestionStyle.Render(prefix + lines[i])
		if i == selectedIdx && len(lines) > 1 {
			row += navStyle.Render(fmt.Sprintf("  [%d/%d]", selectedIdx+1, len(lines)))
		}
		out.WriteString(row)
	}

	return out.String()
}

func (m Model) View() string {
	var content string
	var helpLine string

	switch m.core.Mode {
	case core.ModeLoading:
		content = m.spinner.View() + " Scanning..."
		helpLine = "esc: quit"

	case core.ModeBrowsing:
		prompt := promptStyle.Render("Enter the project directory:")
		input := prompt + " " + m.input.View()

		if len(m.core.Filtered) > 0 {
			lines := make([]string, 0, len(m.core.Filtered))
			for _, dir := range m.core.Filtered {
				lines = append(lines, dir.Path)
			}
			content = input + "\n" + renderSuggestionList(lines, m.core.SelectedIdx)
		} else if m.core.Query != "" {
			content = input + "\n" + suggestionStyle.Render("(create new)")
		} else {
			content = input
		}
		helpLine = "up/down: navigate  enter: select  ctrl+n: create  esc: quit"

	case core.ModeWorktree:
		prompt := promptStyle.Render("Select worktree or create new branch:")
		input := prompt + " " + m.worktreeInput.View()

		if len(m.core.FilteredWT) > 0 {
			lines := make([]string, 0, len(m.core.FilteredWT))
			for _, wt := range m.core.FilteredWT {
				lines = append(lines, wt.Path)
			}
			content = input + "\n" + renderSuggestionList(lines, m.core.WorktreeIdx)
		} else if m.core.WorktreeQuery != "" {
			content = input + "\n" + suggestionStyle.Render("(create new: "+m.core.WorktreeQuery+")")
		} else {
			content = input
		}

		if m.core.ProjectWarning != "" {
			content += "\n" + warningStyle.Render(m.core.ProjectWarning)
		}
		helpLine = "up/down: navigate  enter: select  ctrl+n: create  ctrl+d: delete  esc: back"

	case core.ModeWorktreeDeleteConfirm:
		prompt := promptStyle.Render("Delete worktree?")
		content = prompt + "\n" + suggestionStyle.Render(m.core.WorktreeDeletePath) + "\n" + suggestionStyle.Render("(enter to confirm, esc to cancel)")
		helpLine = "enter: confirm  esc: cancel"

	case core.ModeTool:
		prompt := promptStyle.Render("Select tool:")
		input := prompt + " " + m.toolInput.View()

		if len(m.core.FilteredTools) > 0 {
			content = input + "\n" + renderSuggestionList(m.core.FilteredTools, m.core.ToolIdx)
		} else {
			content = input
		}
		helpLine = "up/down: navigate  enter: open  esc: back"

	case core.ModeError:
		content = errorStyle.Render(fmt.Sprintf("Error: %v", m.core.Err))
		helpLine = "esc: quit"
	}

	if helpLine != "" {
		content += "\n\n" + helpStyle.Render(helpLine)
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
