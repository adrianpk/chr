package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

// Ctrl+C must reach Bubble Tea shutdown before a busy mode can consume it.
func TestBusyQuit(t *testing.T) {
	cases := []struct {
		name string
		busy func(*Model)
	}{
		{"spec", func(m *Model) { m.specBusy = true }},
		{"plan", func(m *Model) { m.planBusy = true }},
		{"quiz", func(m *Model) { m.quizBusy = true }},
		{"work", func(m *Model) { m.workBusy = true }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := NewModel(Options{})
			tc.busy(model)
			_, command := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
			if command == nil {
				t.Fatal("busy mode swallowed Ctrl+C")
			}

			if _, ok := command().(tea.QuitMsg); !ok {
				t.Fatal("Ctrl+C did not request graceful shutdown")
			}
		})
	}
}
