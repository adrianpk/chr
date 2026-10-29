package tui

import (
	"slices"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// All six destinations are offered once, and confirmation freezes the observed
// ticket image rather than rereading a later Board selection.
func TestBoardMoveCommands(t *testing.T) {
	var destinations []string
	for _, command := range DefaultCommandRegistry() {
		if command.Ticket == TicketActionMoveStatus {
			destinations = append(destinations, command.Status)
		}
	}

	if !slices.Equal(destinations, ticketcore.NormalBoardColumns) {
		t.Fatal("move commands differ from canonical Board destinations", destinations)
	}

	ticket := TicketView{Key: "CHR-1", Revision: "observed", Status: "backlog"}
	state := NewTicketActionState(TicketActionMoveStatus, ticket, "ready_for_dev")
	ticket.Key, ticket.Revision = "CHR-2", "newer"
	if state.Request.Key != "CHR-1" || state.Request.Revision != "observed" {
		t.Fatal("confirmation did not freeze the displayed ticket")
	}
}
