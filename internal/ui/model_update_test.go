package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ariguillegp/rivet/internal/core"
)

func TestUpdateKeyNavigationMovesOnceInBrowsing(t *testing.T) {
	m := newTestModel()
	m.core.Mode = core.ModeBrowsing
	m.core.Filtered = []core.DirEntry{
		{Path: "/one", Name: "one"},
		{Path: "/two", Name: "two"},
		{Path: "/three", Name: "three"},
	}
	m.core.SelectedIdx = 0
	m.syncLists()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	next := updated.(Model)
	if next.core.SelectedIdx != 1 {
		t.Fatalf("expected single-step move to index 1, got %d", next.core.SelectedIdx)
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyDown})
	next = updated.(Model)
	if next.core.SelectedIdx != 2 {
		t.Fatalf("expected single-step move to index 2, got %d", next.core.SelectedIdx)
	}
}
