package tui

import (
	"crypto/rand"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openTicketCommentForm(editor TicketEditorView) {
	history := ticketTextFormField(TicketEditorField{Name: "history", Label: "Retained comments", Value: editor.Discussion, Multiline: true, ReadOnly: "Retained history; use the separate comment field to add a comment."})
	comment := newTextFormField("Comment", true, 64*1024)

	form := textFormState{Title: "Ticket Discussion " + editor.Key, SubmitLabel: "Add comment", Submit: (*Model).submitTicketComment, Fields: []textFormField{history, comment}, TicketRequest: TicketActionRequest{Action: TicketActionAddComment, Key: editor.Key, Revision: editor.Revision, CaseSlug: editor.CaseSlug, OperationID: rand.Text()}}
	if editor.ReadOnlyReason != "" {
		form.Details = []string{editor.ReadOnlyReason}
		form.Fields[1].ReadOnly = editor.ReadOnlyReason
		form.SubmitLabel, form.Submit = "Close", discardTextForm
	}

	m.openTextForm(form)
	m.focusTextForm(1)
}

func (m *Model) submitTicketComment() tea.Cmd {
	request := m.textForm.TicketRequest
	request.Comment = m.textForm.Fields[1].value()
	m.executeTicketForm(request)

	return nil
}
