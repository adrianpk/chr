package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) offerGenerationResult(view View, path string, scope QuizScope, tickets bool) {
	if path == "" || m.SelectedView() != view || m.selectedSourcePath() != path {
		return
	}

	m.feedback.Context = m.currentInteraction()
	m.feedback.Open = true

	m.feedback.OpenDocument = func(m *Model) tea.Cmd {
		m.openGenerationDocument(view, path, scope)

		return nil
	}
	if tickets {
		m.feedback.ViewTickets = func(m *Model) tea.Cmd {
			m.bindCaseContext(scope.CaseSlug, "")
			m.selectView(ViewBoard)

			m.boardColumn = defaultBoardColumn(m.caseTickets())
			m.selectFirstFilteredTicket()

			m.feedback.Open = false
			m.statusMessage = "Prepared tickets. Work has not started."

			return nil
		}
	}
}

func (m *Model) openGenerationDocument(view View, path string, scope QuizScope) {
	m.bindCaseContext(scope.CaseSlug, "")
	m.selectView(view)

	if view == ViewSpecs {
		m.selectSpecByPath(path)
	} else {
		m.selectPlanByScope(scope)
	}

	if m.selectedSourcePath() != path {
		m.statusMessage = "The generated document is unavailable. Refresh this case."

		return
	}

	m.selectedPane, m.maximizedPane = 1, 1
	m.detailScroll = 0
	m.feedback.Open = false
	m.statusMessage = "Opened generated document."
}

func (m *Model) generationFeedbackActions() []string {
	if !m.feedbackCurrent() {
		return nil
	}

	var lines []string
	if m.feedback.Reload != nil && m.feedback.Result.Reload {
		lines = append(lines, "r Reload the captured source")
	}

	if m.feedback.Retry != nil && m.feedback.Result.Retry {
		lines = append(lines, "y Retry the original operation")
	}

	if m.feedback.OpenDocument != nil {
		lines = append(lines, "o Open document")
	}

	if m.feedback.ViewTickets != nil {
		lines = append(lines, "t View tickets")
	}

	return lines
}
