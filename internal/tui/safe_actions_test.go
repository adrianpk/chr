package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// Dash must select its displayed queue independently of the Board, in both
// legacy execution and native read-only preview mode.
func TestDashTarget(t *testing.T) {
	for _, authority := range []workspaceconfig.WorkStateAuthority{"", workspaceconfig.WorkStateAuthorityTickets} {
		t.Run("authority="+string(authority), func(t *testing.T) {
			model := NewModel(Options{Observability: ObservabilityView{Authority: authority}, Tickets: []ticketcore.Indexed{
				{IndexedIdentity: ticketcore.IndexedIdentity{Key: "CHR-A"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusBacklog}},
				{IndexedIdentity: ticketcore.IndexedIdentity{Key: "CHR-B"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev}},
				{IndexedIdentity: ticketcore.IndexedIdentity{Key: "CHR-C"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev}},
			}})
			model.selectView(ViewDash)
			model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if model.workAction.Request.TicketKey != "CHR-B" {
				t.Fatal("Dash action targeted the Board ticket", model.workAction.Request)
			}

			model.Update(tea.KeyMsg{Type: tea.KeyEsc})
			model.Update(tea.KeyMsg{Type: tea.KeyDown})
			model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if model.workAction.Request.TicketKey != "CHR-C" || !strings.Contains(model.renderDash(), "> CHR-C") || model.selectedTicket != 0 {
				t.Fatal("queue navigation changed Board selection or targeted a hidden ticket")
			}
		})
	}
}

// Queue refresh, order and filters preserve a key, never a positional index;
// removal selects a visible fallback and an empty queue cannot plan execution.
func TestDashSelectionRefresh(t *testing.T) {
	tickets := []ticketcore.Indexed{
		{IndexedIdentity: ticketcore.IndexedIdentity{Key: "CHR-A"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusBacklog}},
		{IndexedIdentity: ticketcore.IndexedIdentity{Key: "CHR-B"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev, Priority: "high"}},
		{IndexedIdentity: ticketcore.IndexedIdentity{Key: "CHR-C"}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForDev, Priority: "normal"}},
	}
	cases := []struct {
		name string
		rows []ticketcore.Indexed
		want string
	}{
		{"reordered", []ticketcore.Indexed{tickets[2], tickets[0], tickets[1]}, "CHR-C"},
		{"removed", tickets[:2], "CHR-B"},
		{"empty", tickets[:1], ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			model := NewModel(Options{Tickets: tickets})
			model.selectView(ViewDash)
			model.Update(tea.KeyMsg{Type: tea.KeyDown})
			model.completeAgentWorkRefresh(agentWorkRefreshFinishedMsg{result: AgentWorkRefreshResult{Tickets: tc.rows}})
			model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if model.workAction.Request.TicketKey != tc.want || model.workAction.Pending != (tc.want != "") {
				t.Fatal("refresh planned a hidden queue item", model.workAction)
			}
		})
	}

	model := NewModel(Options{Tickets: tickets})
	model.selectView(ViewDash)
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model.cycleAgentWorkOrder()
	if selected, _ := model.selectedDashTicketView(); selected.Key != "CHR-C" {
		t.Fatal("sorting changed the selected key")
	}

	model.cycleAgentWorkPriorityFilter()
	model.planWorkAction(WorkActionStartTicket)
	if model.workAction.Pending {
		t.Fatal("empty priority filter reused a hidden selection")
	}
}
