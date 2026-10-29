package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// TicketActionID identifies a ticket action awaiting TUI confirmation.
type TicketActionID string

// Ticket action identifiers define supported ticket commands.
const (
	TicketActionAddComment       TicketActionID = "add_comment"
	TicketActionRead             TicketActionID = "read_ticket"
	TicketActionCreate           TicketActionID = "create_ticket"
	TicketActionSave             TicketActionID = "save_ticket"
	TicketActionEdit             TicketActionID = "edit_ticket"
	TicketActionRefresh          TicketActionID = "refresh_tickets"
	TicketActionAttachReviewRef  TicketActionID = "attach_review_ref"
	TicketActionMoveStatus       TicketActionID = "move_status"
	TicketActionToggleChecklist  TicketActionID = "toggle_checklist"
	TicketActionAddChecklistItem TicketActionID = "add_checklist_item"
)

// TicketActionExecutor runs confirmed ticket actions.
type TicketActionExecutor interface {
	ExecuteTicketAction(action TicketActionRequest) (TicketActionResult, error)
}

// TicketActionRequest carries selected ticket action input.
type TicketActionRequest struct {
	OperationID, CaseSlug, Comment string
	Fields                         map[string]string
	Body                           *string
	Action                         TicketActionID
	Key                            string
	Revision                       string
	Status                         string
	Branch                         string
	ReviewRef                      *ticketmetadata.ReviewRef
}

// TicketActionResult contains refreshed ticket state after an action.
type TicketActionResult struct {
	Editor    *TicketEditorView
	Committed bool
	Outcome   operation.Result
	Message   string
	Tickets   []ticketcore.Indexed
}

// TicketActionState stores pending confirmation state for a ticket action.
type TicketActionState struct {
	Pending bool
	Request TicketActionRequest
	Title   string
	Details []string
}

// NewTicketActionState builds the confirmation details for a ticket action.
func NewTicketActionState(action TicketActionID, ticket TicketView, status string) TicketActionState {
	request := TicketActionRequest{
		Action:   action,
		Key:      ticket.Key,
		Revision: ticket.Revision,
		Status:   status,
	}

	switch action {
	case TicketActionMoveStatus:
		return TicketActionState{
			Pending: true,
			Request: request,
			Title:   "Move ticket status",
			Details: []string{
				"Ticket: " + ticket.Key,
				"Current status: " + ticket.Status,
				"New status: " + status,
			},
		}
	case TicketActionToggleChecklist:
		return TicketActionState{
			Pending: true,
			Request: request,
			Title:   "Toggle ticket checklist item",
			Details: []string{
				"Ticket: " + ticket.Key,
				fmt.Sprintf("Checklist: %d done / %d total", ticket.ChecklistDone, ticket.ChecklistTotal),
				"Requires a safe checklist item writer.",
			},
		}
	case TicketActionAddChecklistItem:
		return TicketActionState{
			Pending: true,
			Request: request,
			Title:   "Add ticket checklist item",
			Details: []string{
				"Ticket: " + ticket.Key,
				"Requires a safe checklist item writer.",
			},
		}
	default:
		return TicketActionState{}
	}
}

func (m *Model) renderTicketAction() string {
	if !m.ticketAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Ticket Action\n")
	fmt.Fprintf(&output, "  %s\n", m.ticketAction.Title)

	for _, detail := range m.ticketAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m *Model) updateTicketActionKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelTicketAction()
	case tea.KeyEnter:
		m.confirmTicketAction()
	}

	return nil
}

func (m *Model) planTicketAction(action TicketActionID, status string) {
	if action == TicketActionCreate {
		m.openTicketCreateForm()

		return
	}

	if action == TicketActionAddComment || action == TicketActionAttachReviewRef {
		m.openSelectedTicketForm(action)

		return
	}

	if action == TicketActionEdit {
		m.openTicketEditForm()

		return
	}

	ticket, ok := m.selectedTicketView()
	if !ok {
		m.statusMessage = "No selected ticket is available."

		return
	}

	m.ticketAction = NewTicketActionState(action, ticket, status)
	m.statusMessage = ""
}

func (m *Model) cancelTicketAction() {
	m.ticketAction = TicketActionState{}
	m.statusMessage = "Ticket action cancelled."
}

func (m *Model) confirmTicketAction() {
	request := m.ticketAction.Request
	m.ticketAction = TicketActionState{}

	if m.ticketExecutor == nil {
		m.statusMessage = "Ticket action executor is unavailable."

		return
	}

	result, err := m.ticketExecutor.ExecuteTicketAction(request)
	m.recordOperation(result.Outcome)

	if err != nil {
		m.statusMessage = "Ticket action failed: " + err.Error()

		return
	}

	if result.Tickets != nil {
		m.replaceTickets(result.Tickets)
	}

	m.statusMessage = result.Message
}
