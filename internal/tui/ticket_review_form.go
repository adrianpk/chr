package tui

import (
	"crypto/rand"

	tea "github.com/charmbracelet/bubbletea"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

func (m *Model) openTicketReviewForm(editor TicketEditorView) {
	form := textFormState{Title: "Attach Review Reference " + editor.Key, SubmitLabel: "Attach", Submit: (*Model).submitTicketReview, TicketRequest: TicketActionRequest{Action: TicketActionAttachReviewRef, Key: editor.Key, Revision: editor.Revision, CaseSlug: editor.CaseSlug, OperationID: rand.Text()}}
	for _, label := range []string{"Type", "Provider", "URL", "ID (optional)", "Status (optional)", "Branch (optional)"} {
		form.Fields = append(form.Fields, newTextFormField(label, false, 8192))
	}

	form.Fields[0].Input.SetValue("pull_request")

	reason := editor.ReviewReadOnly
	if editor.ReadOnlyReason != "" {
		reason = editor.ReadOnlyReason
	}

	if reason != "" {
		form.Details = []string{reason}
		for i := range form.Fields {
			form.Fields[i].ReadOnly = reason
		}

		form.SubmitLabel, form.Submit = "Close", discardTextForm
	}

	m.openTextForm(form)
}

func (m *Model) submitTicketReview() tea.Cmd {
	request := m.textForm.TicketRequest
	fields := m.textForm.Fields
	request.ReviewRef = &ticketmetadata.ReviewRef{Type: fields[0].value(), Provider: fields[1].value(), URL: fields[2].value(), ID: fields[3].value(), Status: fields[4].value()}
	request.Branch = fields[5].value()
	m.executeTicketForm(request)

	return nil
}
