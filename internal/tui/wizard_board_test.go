package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// Wizard execution uses the shared responsive Board, case filter, column focus
// and read-only details. Manual ticket writers cannot fork the retained run.
func TestWizardBoard(t *testing.T) {
	model := guidedTerminalFixture()
	model.guided.View.Session.Guided.Stage = "execution"
	model.guided.View.Native = &WorkflowView{ID: "native-run", Revision: 3}
	model.guided.View.Tickets = []ticketcore.Indexed{
		{IndexedIdentity: ticketcore.IndexedIdentity{Key: "BOUND"}, IndexedSummary: ticketcore.IndexedSummary{Title: "Bound ticket"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForReview, UseCase: "bounded-audit"}},
		{IndexedIdentity: ticketcore.IndexedIdentity{Key: "FOREIGN"}, IndexedSummary: ticketcore.IndexedSummary{Title: "Foreign ticket"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForReview, UseCase: "another-case"}},
	}
	model.observeGuidedBoard(model.guided.View)
	model.selectView(ViewBoard)

	for _, size := range []tea.WindowSizeMsg{{Width: 160, Height: 44}, {Width: 80, Height: 24}} {
		model.Update(size)

		if strings.Contains(model.View(), "FOREIGN") || !strings.Contains(model.View(), "BOUND") {
			t.Fatal("Board lost case filter or selected ticket", model.View())
		}

		if size.Width >= 132 {
			checkWideBoard(t, model.View())
		}
	}

	before, _ := model.selectedTicketView()
	model.Update(tea.KeyMsg{Type: tea.KeyTab})

	after, _ := model.selectedTicketView()
	if model.selectedPane != 1 || before.Key != after.Key {
		t.Fatal("shared Board focus changed the target")
	}

	for _, key := range []string{"e", "s"} {
		_, command := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		if command != nil || model.textForm.Active {
			t.Fatal("manual ticket writer escaped wizard mode")
		}
	}
}
