package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// Unsupported checklist writes must explain their guard before any confirmation,
// with or without a selected ticket; Enter must never reach the writer.
func TestPaletteChecklist(t *testing.T) {
	for _, test := range []struct {
		name, id string
		selected bool
	}{
		{"toggle selected", "ticket-checklist-toggle", true},
		{"add selected", "ticket-checklist-add", true},
		{"toggle empty", "ticket-checklist-toggle", false},
		{"add empty", "ticket-checklist-add", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			executor := &fakeTicketActionExecutor{}

			model := NewModel(Options{TicketExecutor: executor})
			if test.selected {
				model.replaceTickets([]ticketcore.Indexed{{IndexedIdentity: ticketcore.IndexedIdentity{Key: "TKT-1"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusBacklog}}})
			}

			model.selectView(ViewBoard)
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(test.id)})

			menu := model.View()
			model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			if model.ticketAction.Pending || model.feedback.Result.Status != operation.Blocked || !strings.Contains(model.View(), "safe checklist item writer") {
				t.Fatal("unsupported writer was offered for confirmation", model.View())
			}

			model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			if executor.request.Action != "" || !strings.Contains(menu, "unavailable") {
				t.Fatal("unavailable palette action dispatched a writer or hid its guard")
			}
		})
	}
}
