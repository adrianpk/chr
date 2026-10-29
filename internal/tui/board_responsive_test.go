package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// Wide Board renders every canonical column and attention without maximizing;
// resizing and pane focus retain the selected key and its canonical column.
func TestResponsiveBoard(t *testing.T) {
	tickets := []ticketcore.Indexed{}
	for index, status := range append(append([]string{}, ticketcore.NormalBoardColumns...), "blocked") {
		tickets = append(tickets, ticketcore.Indexed{IndexedIdentity: ticketcore.IndexedIdentity{Key: status}, IndexedSummary: ticketcore.IndexedSummary{Title: "Ticket " + status}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.Status(status)}, IndexedProjection: ticketcore.IndexedProjection{ContentHash: string(rune('a' + index))}})
	}

	model := NewModel(Options{Tickets: tickets})
	model.selectView(ViewBoard)
	model.selectBoardColumn("in_progress")

	before, _ := model.selectedTicketView()
	for _, size := range []tea.WindowSizeMsg{{Width: 160, Height: 44}, {Width: 80, Height: 24}, {Width: 180, Height: 50}} {
		model.Update(size)

		selected, _ := model.selectedTicketView()
		if selected.Key != before.Key || model.activeBoardColumn() != "in_progress" {
			t.Fatal("resize moved the selected target")
		}

		if size.Width >= 132 {
			checkWideBoard(t, model.View())
		}
	}

	model.Update(tea.KeyMsg{Type: tea.KeyTab})

	selected, _ := model.selectedTicketView()
	if selected.Key != before.Key || model.selectedPane != 1 {
		t.Fatal("focus changed selected key")
	}
}

func checkWideBoard(t *testing.T, view string) {
	t.Helper()

	for _, column := range ticketcore.NormalBoardColumns {
		if !strings.Contains(view, ticketcore.BoardColumnTitle(column)) {
			t.Fatal("wide Board hid column", column, view)
		}
	}

	if !strings.Contains(view, "Attention") {
		t.Fatal("wide Board hid attention")
	}
}
