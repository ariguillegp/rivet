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
	for i := range 20 {
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

func TestNewSessionTableStylesUsesAccentSelectedForegroundWithoutBackground(t *testing.T) {
	styles := NewStyles(Themes()[0])
	ts := newSessionTableStyles(styles)

	got := fmt.Sprint(ts.Selected.GetForeground())
	want := fmt.Sprint(styles.SelectedSuggestion.GetForeground())
	if got != want {
		t.Fatalf("expected selected foreground %q, got %q", want, got)
	}
	if got := fmt.Sprint(ts.Selected.GetBackground()); got != "{}" {
		t.Fatalf("expected no selected background, got %q", got)
	}
	if got := ts.Selected.GetPaddingLeft(); got != 0 {
		t.Fatalf("expected no selected left padding, got %d", got)
	}
	if got := ts.Selected.GetPaddingRight(); got != 0 {
		t.Fatalf("expected no selected right padding, got %d", got)
	}
}
