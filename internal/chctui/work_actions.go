package chctui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chc/internal/chc"
)

// WorkActionID identifies an agent work action awaiting TUI confirmation.
type WorkActionID string

// Work action identifiers define supported agent work commands.
const (
	WorkActionStartTicket    WorkActionID = "start_ticket"
	WorkActionStartReadyLoop WorkActionID = "start_ready_loop"
	WorkActionRunAllReady    WorkActionID = "run_all_ready"
	WorkActionRequestStop    WorkActionID = "request_stop"
)

// WorkActionExecutor runs confirmed agent work actions.
type WorkActionExecutor interface {
	ExecuteWorkAction(action WorkActionRequest) (WorkActionResult, error)
}

// WorkActionRequest carries selected agent work action input.
type WorkActionRequest struct {
	Action         WorkActionID
	TicketKey      string
	PriorityFilter string
	Order          chc.TicketLoopOrder
}

// WorkActionResult contains refreshed agent work state after an action.
type WorkActionResult struct {
	Message       string
	Tickets       []chc.IndexedTicket
	Observability ObservabilityView
}

// WorkActionState stores pending confirmation state for an agent work action.
type WorkActionState struct {
	Pending bool
	Request WorkActionRequest
	Title   string
	Details []string
}

type workActionFinishedMsg struct {
	Request WorkActionRequest
	Result  WorkActionResult
	Err     error
}

// NewWorkActionState builds the confirmation details for an agent work action.
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
				"Status filter: ready_for_dev",
				"Max iterations: 1",
				"Stops on blocked or failed work.",
			},
		}
	case WorkActionRunAllReady:
		return WorkActionState{
			Pending: true,
			Request: WorkActionRequest{
				Action: action,
			},
			Title: "Run all ready tickets",
			Details: []string{
				"Status filter: ready_for_dev",
				"Order: oldest first",
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
	if m.workBusy {
		return strings.TrimRight("\n\nWork Action Running\n  "+m.statusMessage+"\n\n  Please wait for this action to finish.", "\n")
	}

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

	output.WriteString("\n  Enter: confirm\n")
	output.WriteString("  Esc: cancel\n")

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) updateWorkActionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelWorkAction()
	case tea.KeyEnter:
		return m, m.confirmWorkAction()
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
	if action == WorkActionRunAllReady {
		m.applyDashAgentWorkSelectionToAction()
	}

	m.statusMessage = ""
}

func (m *Model) applyDashAgentWorkSelectionToAction() {
	m.workAction.Request.PriorityFilter = m.agentWorkPriorityFilter
	m.workAction.Request.Order = m.agentWorkOrder
	m.workAction.Details = []string{
		"Status filter: ready_for_dev",
		"Priority filter: " + displayDashPriorityFilter(m.agentWorkPriorityFilter),
		"Order: " + displayDashAgentWorkOrder(m.agentWorkOrder),
		"Stops on blocked or failed work.",
	}
}

func (m *Model) cancelWorkAction() {
	m.workAction = WorkActionState{}
	m.statusMessage = "Work action cancelled."
}

func (m *Model) confirmWorkAction() tea.Cmd {
	request := m.workAction.Request
	m.workAction = WorkActionState{}

	return m.startWorkAction(request)
}

func (m Model) executeWorkActionCommand(request WorkActionRequest) tea.Cmd {
	executor := m.workExecutor

	return func() tea.Msg {
		if executor == nil {
			return workActionFinishedMsg{
				Request: request,
				Err:     fmt.Errorf("work action executor is unavailable"),
			}
		}

		result, err := executor.ExecuteWorkAction(request)

		return workActionFinishedMsg{Request: request, Result: result, Err: err}
	}
}

func (m *Model) startWorkAction(request WorkActionRequest) tea.Cmd {
	if m.workBusy {
		m.statusMessage = "Work action already running."

		return nil
	}

	if m.workExecutor == nil {
		m.statusMessage = "Work action executor is unavailable."

		return nil
	}

	m.workBusy = true
	m.statusMessage = workBusyMessage(request)

	return m.executeWorkActionCommand(request)
}

func (m *Model) completeWorkAction(message workActionFinishedMsg) {
	m.workBusy = false

	if message.Err != nil {
		m.statusMessage = "Work action failed: " + message.Err.Error()

		return
	}

	m.applyWorkActionResult(message.Result)
}

func (m *Model) applyWorkActionResult(result WorkActionResult) {
	if result.Tickets != nil {
		m.tickets = NewTicketViews(result.Tickets)
		m.selectFirstFilteredTicket()
	}

	m.observability = result.Observability
	m.statusMessage = result.Message
}

func workBusyMessage(request WorkActionRequest) string {
	switch request.Action {
	case WorkActionStartTicket:
		if strings.TrimSpace(request.TicketKey) != "" {
			return "Starting agent work for " + request.TicketKey + "..."
		}

		return "Starting selected ticket..."
	case WorkActionStartReadyLoop:
		return "Starting ready ticket loop..."
	case WorkActionRunAllReady:
		return "Running all ready tickets..."
	case WorkActionRequestStop:
		return "Requesting agent stop..."
	default:
		return "Running work action..."
	}
}
