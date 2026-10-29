package tui

import (
	"crypto/rand"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) bindSpecFormRecovery(request SpecActionRequest) {
	context := m.currentInteraction()

	m.feedback.Context = context
	if m.feedback.Result.Reload && request.Action == SpecActionSave {
		m.feedback.Reload = func(m *Model) tea.Cmd {
			if context != m.currentInteraction() {
				return nil
			}

			return m.confirmDraftChange("Reload this spec and discard its unsaved fields?", func(m *Model) tea.Cmd {
				m.reloadSpecForm(request)

				return nil
			})
		}
	}

	if m.feedback.Result.Retry {
		m.feedback.Retry = func(m *Model) tea.Cmd {
			if context != m.currentInteraction() {
				return nil
			}

			m.feedback.Open = false
			m.executeSpecForm(request)

			return nil
		}
	}
}

func (m *Model) reloadSpecForm(request SpecActionRequest) {
	result, err := m.specExecutor.ExecuteSpecAction(SpecActionRequest{Action: SpecActionEdit, Path: request.Path})
	m.recordOperation(result.Outcome)

	if err != nil || result.Document == nil {
		m.statusMessage = "Reload failed; the unsaved spec fields are retained."

		return
	}

	if result.Document.Scope != request.Scope {
		m.statusMessage = "The spec case binding changed. The draft is retained; reopen the document explicitly."

		return
	}

	request.Revision, request.OperationID = result.Document.Revision, rand.Text()
	m.openSpecForm(request, *result.Document)
}
