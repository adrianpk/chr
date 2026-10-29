package tui

import (
	"testing"

	"forge.adrianpk.com/chr/internal/app/work"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// Queue refresh, reorder and removal keep the Board's separate selected key;
// removal falls back to its first visible row without dispatching work.
func TestDashBoardRefresh(t *testing.T) {
	model := NewModel(Options{Tickets: dashSelectionTickets("CHR-1", "CHR-2", "CHR-3")})
	model.selectView(ViewBoard)
	model.selectNextTicket()
	model.selectView(ViewDash)
	model.moveDashSelection(2)

	for _, item := range []struct {
		name        string
		keys        []string
		board, dash string
	}{
		{"reorder", []string{"CHR-3", "CHR-1", "CHR-2"}, "CHR-2", "CHR-3"},
		{"board removed", []string{"CHR-3", "CHR-1"}, "CHR-3", "CHR-3"},
		{"dash removed", []string{"CHR-1"}, "CHR-1", "CHR-1"},
		{"empty", []string{}, "", ""},
	} {
		t.Run(item.name, func(t *testing.T) {
			model.completeAgentWorkRefresh(agentWorkRefreshFinishedMsg{result: AgentWorkRefreshResult{Tickets: dashSelectionTickets(item.keys...)}})
			model.selectView(ViewBoard)

			board, _ := model.selectedTicketView()
			model.selectView(ViewDash)

			dash, _ := model.selectedDashTicketView()
			if board.Key != item.board || dash.Key != item.dash || model.workBusy || model.workAction.Pending {
				t.Fatal(board.Key, dash.Key, model.workAction)
			}
		})
	}
}

func dashSelectionTickets(keys ...string) []ticketcore.Indexed {
	tickets := make([]ticketcore.Indexed, 0, len(keys))
	for _, key := range keys {
		tickets = append(tickets, ticketcore.Indexed{
			IndexedIdentity:       ticketcore.IndexedIdentity{Key: key},
			IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev},
		})
	}

	return tickets
}

// A response for the old filter/page cannot install rows or restore an obsolete
// executable choice after the user changes the display query.
func TestDashRefreshQuery(t *testing.T) {
	executor := &fakeAgentWorkRefreshExecutor{}
	model := NewModel(Options{AgentWorkExecutor: executor, WorkObservation: &work.Observation{QueueNext: 128}})
	model.selectView(ViewDash)

	command := model.handleAgentWorkRefreshTick()
	message := command()

	model.cycleAgentWorkPriorityFilter()

	reply, ok := message.(agentWorkRefreshFinishedMsg)
	if !ok {
		t.Fatal(message)
	}

	model.completeAgentWorkRefresh(reply)

	if model.workObservation == nil || model.agentWorkPriorityFilter != "critical" {
		t.Fatal("old observation installed after query changed")
	}
}
