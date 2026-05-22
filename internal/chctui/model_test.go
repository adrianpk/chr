package chctui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelViewRendersProjectAndNavigation(t *testing.T) {
	model := NewModel(Options{
		ProjectDir: "/workspace/project",
		ProjectID:  "project-1",
	})

	view := model.View()

	for _, want := range []string{
		"ChrysoCoord",
		"Project",
		"project-1",
		"/workspace/project",
		"> Dashboard",
		"Board",
		"Tickets",
		"Runtime",
		"Keys:",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelNavigatesViews(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	navigated, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.SelectedView() != ViewBoard {
		t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), ViewBoard)
	}

	updated, _ = navigated.Update(tea.KeyMsg{Type: tea.KeyUp})
	navigated, ok = updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.SelectedView() != ViewDashboard {
		t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), ViewDashboard)
	}
}

func TestModelRecordsWindowSize(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	resized, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if resized.width != 120 {
		t.Fatalf("width = %d, want 120", resized.width)
	}

	if resized.height != 40 {
		t.Fatalf("height = %d, want 40", resized.height)
	}
}
