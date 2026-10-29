package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) bindTicketFormRecovery(request TicketActionRequest) {
	m.feedback.Context = m.currentInteraction()
	if m.feedback.Result.Reload && request.Key != "" {
		m.feedback.Reload = func(m *Model) tea.Cmd {
			return m.confirmDraftChange("Reload this ticket and discard its unsaved input?", func(m *Model) tea.Cmd {
				m.reloadTicketForm(request)

				return nil
			})
		}
	}
}

func (m *Model) reloadTicketForm(request TicketActionRequest) {
	result, err := m.ticketExecutor.ExecuteTicketAction(TicketActionRequest{Action: TicketActionRead, Key: request.Key})
	m.recordOperation(result.Outcome)

	if err != nil || result.Editor == nil {
		m.statusMessage = "Reload failed; the complete ticket draft is retained."

		return
	}

	if result.Editor.CaseSlug != request.CaseSlug {
		m.statusMessage = "The ticket case binding changed. The draft is retained; reopen the ticket explicitly."

		return
	}

	m.openTicketForm(*result.Editor)
}
