package tui

import (
	"crypto/rand"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateTicketKey(message tea.KeyMsg) tea.Cmd {
	switch message.String() {
	case "c":
		m.openTicketCreateForm()
	case "e":
		if m.selectedPane == 1 {
			return m.updateNavigationKey(message)
		}

		m.openTicketEditForm()
	case "v":
		m.openSelectedTicketForm(TicketActionAttachReviewRef)
	case "a":
		m.openSelectedTicketForm(TicketActionAddComment)
	case "r":
		m.refreshTicketFormSource()
	default:
		return m.updateNavigationKey(message)
	}

	return nil
}

func (m *Model) openTicketCreateForm() {
	if !m.caseContextAvailable() {
		m.explainCaseContext()

		return
	}

	m.readTicketForm(TicketActionRequest{Action: TicketActionCreate, CaseSlug: m.caseContext.CaseSlug})
}

func (m *Model) openTicketEditForm() { m.openSelectedTicketForm(TicketActionSave) }

func (m *Model) openSelectedTicketForm(action TicketActionID) {
	ticket, ok := m.selectedTicketView()
	if !ok {
		m.statusMessage = "No ticket selected."

		return
	}

	m.readTicketForm(TicketActionRequest{Action: action, Key: ticket.Key})
}

func (m *Model) readTicketForm(request TicketActionRequest) {
	if m.ticketExecutor == nil {
		m.statusMessage = "Ticket editor is unavailable."

		return
	}

	result, err := m.ticketExecutor.ExecuteTicketAction(TicketActionRequest{Action: TicketActionRead, Key: request.Key, CaseSlug: request.CaseSlug})
	m.recordOperation(result.Outcome)

	if err != nil || result.Editor == nil {
		m.statusMessage = "Ticket could not be opened. Refresh and try again."

		return
	}

	m.openTicketInput(*result.Editor, request.Action)
}

func (m *Model) openTicketForm(editor TicketEditorView) {
	title, action, submit := "Create Ticket", TicketActionCreate, "Create"
	if !editor.New {
		title, action, submit = "Edit Ticket "+editor.Key, TicketActionSave, "Save"
	}

	form := textFormState{Title: title, SubmitLabel: submit, Submit: (*Model).submitTicketForm, TicketRequest: TicketActionRequest{Action: action, Key: editor.Key, CaseSlug: editor.CaseSlug, Revision: editor.Revision, OperationID: rand.Text()}}
	if !editor.New {
		form.Details = []string{"Phase: " + editor.Phase + "  Slice: " + editor.Slice}
	}

	for _, value := range editor.Fields {
		field := ticketTextFormField(value)
		if editor.ReadOnlyReason != "" {
			field.ReadOnly = editor.ReadOnlyReason
		}

		form.Fields = append(form.Fields, field)
	}

	if editor.ReadOnlyReason != "" {
		form.Details = append(form.Details, editor.ReadOnlyReason)
		form.SubmitLabel, form.Submit = "Close", discardTextForm
	}

	m.openTextForm(form)
}

func ticketTextFormField(value TicketEditorField) textFormField {
	field := newTextFormField(value.Label, value.Multiline, 1<<20)
	field.Name, field.ReadOnly = value.Name, value.ReadOnly
	field.Input.SetValue(value.Value)

	for _, option := range value.Options {
		field.Choices = append(field.Choices, textFormChoice{Value: option.Value, Label: option.Label})
		if option.Value == value.Value {
			field.Choice = len(field.Choices) - 1
		}
	}

	return field
}

func (m *Model) submitTicketForm() tea.Cmd {
	request := m.textForm.TicketRequest

	request.Fields = map[string]string{}
	for _, field := range m.textForm.Fields {
		if field.ReadOnly != "" {
			continue
		}

		if field.Name == "description" {
			body := field.value()
			request.Body = &body
		} else {
			request.Fields[field.Name] = field.value()
		}
	}

	m.executeTicketForm(request)

	return nil
}

func (m *Model) executeTicketForm(request TicketActionRequest) {
	result, err := m.ticketExecutor.ExecuteTicketAction(request)
	m.recordOperation(result.Outcome)

	if err != nil {
		m.statusMessage = "Ticket submission failed: " + err.Error()
		if result.Committed {
			m.textForm.SubmitLabel, m.textForm.Submit = "Close", discardTextForm
		}

		m.bindTicketFormRecovery(request)

		return
	}

	m.textForm = textFormState{}
	if result.Editor != nil {
		m.bindCaseContext(result.Editor.CaseSlug, "")
	}

	if result.Tickets != nil {
		m.replaceTickets(result.Tickets)
	}

	if result.Editor != nil {
		m.focusSavedTicket(result.Editor.Key)
	}

	m.statusMessage = result.Message
}

func (m *Model) focusSavedTicket(key string) {
	for _, ticket := range m.caseTickets() {
		if ticket.Key == key {
			m.boardColumn, m.selectedTicket = ticket.BoardColumn, ticket.Index

			return
		}
	}
}

func (m *Model) refreshTicketFormSource() {
	if m.ticketExecutor == nil {
		return
	}

	result, err := m.ticketExecutor.ExecuteTicketAction(TicketActionRequest{Action: TicketActionRefresh})
	m.recordOperation(result.Outcome)

	if err == nil && result.Tickets != nil {
		m.replaceTickets(result.Tickets)
	}

	m.statusMessage = result.Message
}

func (m *Model) openTicketInput(editor TicketEditorView, action TicketActionID) {
	if action == TicketActionAttachReviewRef {
		m.openTicketReviewForm(editor)

		return
	}

	if action == TicketActionAddComment {
		m.openTicketCommentForm(editor)

		return
	}

	m.openTicketForm(editor)
}
