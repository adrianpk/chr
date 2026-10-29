package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Existing tab shortcuts preserve the run binding and show stage progress without
// opening a detached wizard menu or resubmitting the retained authoring input.
func TestWizardTabs(t *testing.T) {
	model := guidedTerminalFixture()
	model.guided.View.Namespace = "private-namespace"
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})

	keys := []tea.KeyType{tea.KeyF2, tea.KeyF3, tea.KeyF4, tea.KeyF5, tea.KeyF6}
	for _, key := range keys {
		_, command := model.Update(tea.KeyMsg{Type: key})
		if command != nil || !model.wizardBound() || model.caseContext.RunID != "guided-one" {
			t.Fatal("navigation changed run or dispatched work")
		}

		frame := model.View()
		for _, text := range []string{"F1", "Cases", "Dumps", "Quizzes", "Specs", "Plans", "Board", "Wizard: 3/7 quiz"} {
			if !strings.Contains(frame, text) {
				t.Fatal("shared shell or stage missing", frame)
			}
		}

		if strings.Contains(frame, "Guided use cases") || strings.Contains(frame, "private-namespace") || strings.Contains(frame, "guided-one") {
			t.Fatal("detached or technical identity foregrounded", frame)
		}
	}
}
