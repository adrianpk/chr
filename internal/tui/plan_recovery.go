package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) bindPlanRecovery(request PlanActionRequest) {
	m.feedback.Context = m.currentInteraction()
	if m.feedback.Result.Retry {
		m.feedback.Retry = func(m *Model) tea.Cmd {
			m.feedback.Open = false

			return m.startPlanAction(request)
		}
	}

	if m.feedback.Result.Reload {
		m.feedback.Reload = func(m *Model) tea.Cmd {
			m.feedback.Open = false

			return m.startPlanAction(PlanActionRequest{Action: PlanActionRefresh, Scope: request.Scope})
		}
	}
}
