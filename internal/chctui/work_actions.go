package chctui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chc/internal/chc"
)

type WorkActionID string

const (
	WorkActionStartTicket    WorkActionID = "start_ticket"
	WorkActionStartReadyLoop WorkActionID = "start_ready_loop"
	WorkActionRequestStop    WorkActionID = "request_stop"
)

type WorkActionExecutor interface {
	ExecuteWorkAction(action WorkActionRequest) (WorkActionResult, error)
}

type WorkActionRequest struct {
	Action    WorkActionID
	TicketKey string
}

type WorkActionResult struct {
	Message       string
	Tickets       []chc.IndexedTicket
	Observability ObservabilityView
}

type WorkActionState struct {
	Pending bool
	Request WorkActionRequest
	Title   string
	Details []string
}

func NewWorkActionState(action WorkActionID, ticket TicketView) WorkActionState {
	switch action {
	case WorkActionStartTicket:
		return WorkActionState{
			Pending: true,
			Request: WorkActionRequest{
				Action:    action,
				TicketKey: ticket.Key,
			},
			Title: "Start selected ticket",
			Details: []string{
				"Ticket: " + ticket.Key,
				"Runs the configured runtime harness.",
				"Execution evidence will be surfaced after completion.",
			},
		}
	case WorkActionStartReadyLoop:
		return WorkActionState{
			Pending: true,
			Request: WorkActionRequest{
				Action: action,
			},
			Title: "Start ready ticket loop",
			Details: []string{
				"Status filter: ready",
				"Max iterations: 1",
				"Stops on blocked or failed work.",
			},
		}
	case WorkActionRequestStop:
		return WorkActionState{
			Pending: true,
			Request: WorkActionRequest{
				Action: action,
			},
			Title: "Request agent stop",
			Details: []string{
				"Requests daemon stop through the daemon state repository.",
				"Running foreground loops observe this stop request.",
			},
		}
	default:
		return WorkActionState{}
	}
}

func (m Model) renderWorkAction() string {
	if !m.workAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Work Action\n")
	output.WriteString(fmt.Sprintf("  %s\n", m.workAction.Title))

	for _, detail := range m.workAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) updateWorkActionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelWorkAction()
	case tea.KeyEnter:
		m.confirmWorkAction()
	}

	return m, nil
}

func (m *Model) planWorkAction(action WorkActionID) {
	if action == WorkActionStartTicket {
		ticket, ok := m.selectedTicketView()
		if !ok {
			m.statusMessage = "No selected ticket is available."

			return
		}

		m.workAction = NewWorkActionState(action, ticket)
		m.statusMessage = ""

		return
	}

	m.workAction = NewWorkActionState(action, TicketView{})
	m.statusMessage = ""
}

func (m *Model) cancelWorkAction() {
	m.workAction = WorkActionState{}
	m.statusMessage = "Work action cancelled."
}

func (m *Model) confirmWorkAction() {
	request := m.workAction.Request
	m.workAction = WorkActionState{}

	if m.workExecutor == nil {
		m.statusMessage = "Work action executor is unavailable."

		return
	}

	result, err := m.workExecutor.ExecuteWorkAction(request)
	if err != nil {
		m.statusMessage = "Work action failed: " + err.Error()

		return
	}

	if result.Tickets != nil {
		m.tickets = NewTicketViews(result.Tickets)
		m.selectFirstFilteredTicket()
	}

	m.observability = result.Observability
	m.statusMessage = result.Message
}
