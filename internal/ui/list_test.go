package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ariguillegp/rivet/internal/core"
)

func TestSyncProjectListSetsAdaptiveHeight(t *testing.T) {
	m := newTestModel()
	m.height = 25

	m.core.Filtered = []core.DirEntry{{Path: "/one", Name: "one"}}
	m.syncProjectList()
	if got := m.projectList.Height(); got != 1 {
		t.Fatalf("expected list height 1 for single item, got %d", got)
	}

	m.core.Filtered = nil
	for i := 0; i < 20; i++ {
		m.core.Filtered = append(m.core.Filtered, core.DirEntry{Path: fmt.Sprintf("/p%d", i), Name: fmt.Sprintf("p%d", i)})
	}
	m.syncProjectList()
	if got := m.projectList.Height(); got != maxListSuggestions {
		t.Fatalf("expected list height capped at %d, got %d", maxListSuggestions, got)
	}
}

func TestViewShowsCountMessage(t *testing.T) {
	m := newTestModel()
	m.height = 25
	m.core.Mode = core.ModeBrowsing
	m.core.Filtered = []core.DirEntry{{Path: "/one", Name: "one"}, {Path: "/two", Name: "two"}}
	m.core.SelectedIdx = 1

	view := stripANSI(m.View())
	if !strings.Contains(view, "Showing 1-2 of 2") {
		t.Fatalf("expected count line in view, got %q", view)
	}
}
