package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
)

// WorkActionID identifies an agent work action awaiting TUI confirmation.
type WorkActionID string

// Work action identifiers define supported agent work commands.
const (
	WorkActionStartTicket     WorkActionID = "start_ticket"
	WorkActionStartReadyLoop  WorkActionID = "start_ready_loop"
	WorkActionRunAllReady     WorkActionID = "run_all_ready"
	WorkActionRepairReadiness WorkActionID = "repair_readiness"
	WorkActionRequestStop     WorkActionID = "request_stop"
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
	Order          core.TicketLoopOrder
}

// WorkActionResult contains refreshed agent work state after an action.
type WorkActionResult struct {
	Message       string
	Tickets       []core.IndexedTicket
	Observability ObservabilityView
	Readiness     WorkReadinessView
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

const maxWorkActionEventRows = 8

// NewWorkActionState builds the confirmation details for an agent work action.
func NewWorkActionState(action WorkActionID, ticket TicketView) WorkActionState {
	switch action {
	case WorkActionStartTicket:
		return newStartTicketWorkActionState(ticket)
	case WorkActionStartReadyLoop:
		return newSimpleWorkActionState(action, "Start ready ticket loop", []string{
			"Status filter: ready_for_dev",
			"Max iterations: 1",
			"Stops on blocked or failed work.",
		})
	case WorkActionRunAllReady:
		return newSimpleWorkActionState(action, "Run all ready tickets", []string{
			"Status filter: ready_for_dev",
			"Order: oldest first",
			"Stops on blocked or failed work.",
		})
	case WorkActionRepairReadiness:
		return newSimpleWorkActionState(action, "Apply safe readiness repair", []string{
			"Runs safe readiness repair only.",
			"Does not start agent work automatically.",
			"Retry the work action after readiness is ready.",
		})
	case WorkActionRequestStop:
		return newSimpleWorkActionState(action, "Request agent stop", []string{
			"Requests daemon stop through the daemon state repository.",
			"Running foreground loops observe this stop request.",
		})
	default:
		return WorkActionState{}
	}
}

func newStartTicketWorkActionState(ticket TicketView) WorkActionState {
	return WorkActionState{
		Pending: true,
		Request: WorkActionRequest{
			Action:    WorkActionStartTicket,
			TicketKey: ticket.Key,
		},
		Title: "Start selected ticket",
		Details: []string{
			"Ticket: " + ticket.Key,
			"Runs the configured runtime harness.",
			"Execution evidence will be surfaced after completion.",
		},
	}
}

func newSimpleWorkActionState(action WorkActionID, title string, details []string) WorkActionState {
	return WorkActionState{
		Pending: true,
		Request: WorkActionRequest{
			Action: action,
		},
		Title:   title,
		Details: details,
	}
}

func (m Model) renderWorkAction() string {
	if m.workBusy {
		return m.renderRunningWorkAction()
	}

	if !m.workAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Work Action\n")
	fmt.Fprintf(&output, "  %s\n", m.workAction.Title)

	for _, detail := range m.workAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	output.WriteString("\n  Enter: confirm\n")
	output.WriteString("  Esc: cancel\n")

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) renderRunningWorkAction() string {
	var output strings.Builder

	output.WriteString("\n\nWork Action Running\n")

	if strings.TrimSpace(m.statusMessage) != "" {
		output.WriteString("  ")
		output.WriteString(m.statusMessage)
		output.WriteString("\n")
	}

	output.WriteString("  Navigation is locked until this action finishes.\n")

	output.WriteString("\nEvents\n")

	events := recentWorkActionEvents(m.observability.RecentEvents, maxWorkActionEventRows)
	if len(events) == 0 {
		output.WriteString("  No agent events recorded yet.\n")
	} else {
		for _, event := range events {
			output.WriteString("  ")
			output.WriteString(renderWorkActionEvent(event))
			output.WriteString("\n")
		}
	}

	output.WriteString(renderWorkActionDetails(events))

	return strings.TrimRight(output.String(), "\n")
}

func renderWorkActionDetails(events []AgentEventView) string {
	details := latestWorkActionDetails(events)
	if details.empty() {
		return ""
	}

	var output strings.Builder

	output.WriteString("\nDetails\n")
	writeWorkActionDetail(&output, "Ticket: ", details.TicketPath)
	writeWorkActionDetail(&output, "Execution report: ", details.ExecutionReport)
	writeWorkActionDetail(&output, "Runtime log: ", details.RuntimeLog)
	writeWorkActionDetail(&output, "Validation log: ", details.ValidationLog)
	writeWorkActionDetail(&output, "Lookup: chr work logs ", details.ExecutionID)

	return output.String()
}

func writeWorkActionDetail(output *strings.Builder, label, value string) {
	if value == "" {
		return
	}

	output.WriteString("  ")
	output.WriteString(label)
	output.WriteString(value)
	output.WriteString("\n")
}

func recentWorkActionEvents(events []AgentEventView, limit int) []AgentEventView {
	if limit <= 0 || len(events) <= limit {
		return events
	}

	return events[len(events)-limit:]
}

func renderWorkActionEvent(event AgentEventView) string {
	message := strings.TrimSpace(event.Message)
	if message == "" {
		message = "event recorded"
	}

	level := strings.TrimSpace(event.Level)
	if level != "" && level != string(core.AgentEventLevelInfo) {
		message = level + ": " + message
	}

	ticketKey := strings.TrimSpace(event.TicketKey)

	parts := make([]string, 0, 3)
	if strings.TrimSpace(event.Time) != "" {
		parts = append(parts, event.Time)
	}

	if ticketKey != "" && !strings.HasPrefix(message, ticketKey+" ") && message != ticketKey {
		parts = append(parts, ticketKey)
	}

	parts = append(parts, message)

	return strings.Join(parts, " ")
}

type workActionDetails struct {
	TicketPath      string
	ExecutionReport string
	RuntimeLog      string
	ValidationLog   string
	ExecutionID     string
}

func (d workActionDetails) empty() bool {
	return d.TicketPath == "" &&
		d.ExecutionReport == "" &&
		d.RuntimeLog == "" &&
		d.ValidationLog == "" &&
		d.ExecutionID == ""
}

func latestWorkActionDetails(events []AgentEventView) workActionDetails {
	details := workActionDetails{}

	for index := len(events) - 1; index >= 0; index-- {
		event := events[index]
		if details.TicketPath == "" {
			details.TicketPath = strings.TrimSpace(event.TicketPath)
		}

		if details.ExecutionReport == "" {
			details.ExecutionReport = strings.TrimSpace(event.ExecutionReport)
		}

		if details.RuntimeLog == "" {
			details.RuntimeLog = strings.TrimSpace(event.RuntimeLog)
		}

		if details.ValidationLog == "" {
			details.ValidationLog = strings.TrimSpace(event.ValidationLog)
		}

		if details.ExecutionID == "" {
			details.ExecutionID = strings.TrimSpace(event.ExecutionID)
		}

		if !details.empty() && workActionDetailsComplete(details) {
			break
		}
	}

	return details
}

func workActionDetailsComplete(details workActionDetails) bool {
	return details.TicketPath != "" &&
		details.ExecutionReport != "" &&
		details.RuntimeLog != "" &&
		details.ValidationLog != "" &&
		details.ExecutionID != ""
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

	if workActionRequiresReadiness(request.Action) && m.workReadiness.blocked() {
		m.statusMessage = m.blockedWorkReadinessStatus()

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
	m.workReadiness = result.Readiness
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
	case WorkActionRepairReadiness:
		return "Applying safe readiness repair..."
	case WorkActionRequestStop:
		return "Requesting agent stop..."
	default:
		return "Running work action..."
	}
}

func workActionRequiresReadiness(action WorkActionID) bool {
	switch action {
	case WorkActionStartTicket, WorkActionStartReadyLoop, WorkActionRunAllReady:
		return true
	default:
		return false
	}
}

func (m Model) blockedWorkReadinessStatus() string {
	if m.workReadiness.safeRepairAvailable() {
		return "Work readiness blocked. Press x to apply safe repair."
	}

	return "Work readiness blocked. Manual action required."
}
