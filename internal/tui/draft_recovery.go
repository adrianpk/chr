package tui

import tea "github.com/charmbracelet/bubbletea"

func discardDumpDraft(m *Model) tea.Cmd {
	m.dumpEdit = DumpEditState{}
	m.feedback.Open = false
	m.statusMessage = "Draft discarded."

	return nil
}

func discardQuizDraft(m *Model) tea.Cmd {
	m.quizAnswer = QuizAnswerState{}
	m.feedback.Open = false
	m.maximizedPane = -1
	m.selectPrimaryPane()

	m.statusMessage = "Draft discarded."

	return nil
}

func (m *Model) bindDumpRecovery(request DumpActionRequest) {
	context := m.currentInteraction()

	m.feedback.Context = context
	if m.feedback.Result.Reload {
		m.feedback.Reload = func(m *Model) tea.Cmd {
			if context != m.currentInteraction() {
				m.statusMessage = "The active draft changed. Reopen its current source before reloading."

				return nil
			}

			return m.confirmDraftChange("Reload this dump and discard its unsaved text?", func(m *Model) tea.Cmd {
				result, err := m.dumpExecutor.ExecuteDumpAction(DumpActionRequest{Action: DumpActionEdit, Path: request.Path})
				m.recordOperation(result.Outcome)

				if err != nil {
					m.statusMessage = "Reload failed: " + err.Error()

					return nil
				}

				m.startDumpEdit(result.SelectedPath, result.Content, result.Revision)

				return nil
			})
		}
	}

	if m.feedback.Result.Retry {
		m.feedback.Retry = func(m *Model) tea.Cmd {
			if context != m.currentInteraction() {
				m.statusMessage = "The active draft changed. The original retry was not submitted."

				return nil
			}

			m.feedback.Open = false
			m.executeDumpSave(request)

			return nil
		}
	}
}

func (m *Model) requestQuizControl(action QuizActionID) tea.Cmd {
	request := QuizActionRequest{Guided: m.quizAnswer.Guided, Action: action, Scope: m.quizAnswer.Scope, Binding: m.quizAnswer.Binding, Mode: quizActionMode(action)}

	apply := func(m *Model) tea.Cmd { return m.startQuizAction(request) }
	if action == QuizActionCancel {
		return m.confirmDraftChange("Cancel this quiz? Saved answers are retained; unsent text will be discarded.", apply)
	}

	if m.quizAnswer.Input.Value() != "" {
		return m.confirmDraftChange("Continue without submitting the current answer?", apply)
	}

	return apply(m)
}

func discardProjectField(m *Model) tea.Cmd {
	m.projectEdit = ProjectEditState{}
	m.statusMessage = "Field edit discarded."

	return nil
}

func discardCaseField(m *Model) tea.Cmd {
	m.caseEdit = ProjectEditState{}
	m.statusMessage = "Field edit discarded."

	return nil
}

func (m *Model) bindQuizRecovery(request QuizActionRequest) {
	context := m.currentInteraction()

	m.feedback.Context = context
	if m.feedback.Result.Reload {
		m.feedback.Reload = func(m *Model) tea.Cmd {
			if context != m.currentInteraction() {
				return nil
			}

			return m.confirmDraftChange("Reload the current quiz and discard the unsent answer?", func(m *Model) tea.Cmd {
				m.feedback.Open = false

				return m.startQuizAction(QuizActionRequest{Action: QuizActionRefresh, Scope: request.Scope, ReloadDraft: true})
			})
		}
	}

	if m.feedback.Result.Retry {
		m.feedback.Retry = func(m *Model) tea.Cmd {
			if context != m.currentInteraction() {
				return nil
			}

			m.feedback.Open = false

			return m.startQuizAction(request)
		}
	}
}
