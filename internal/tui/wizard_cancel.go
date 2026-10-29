package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) confirmGuidedMutation(request GuidedRequest) tea.Cmd {
	if request.Command.Action != "cancel" {
		return m.startGuidedRequest(request)
	}

	return m.confirmDraftChange("Cancel workflow? Saved documents and completed changes will be kept. Any unsent text will be discarded.", func(m *Model) tea.Cmd { return m.startGuidedRequest(request) })
}

func (m *Model) completeWizardCancellation(request GuidedRequest) {
	if request.Action != "input" || request.Command.Action != "cancel" {
		return
	}

	if draft := m.quizAnswer.Guided; draft != nil && draft.Command.RunID == request.Command.RunID {
		m.quizAnswer = QuizAnswerState{}
	}

	if draft := m.dumpEdit.Guided; draft != nil && draft.Command.RunID == request.Command.RunID {
		m.dumpEdit = DumpEditState{}
	}

	if draft := m.textForm.GuidedRequest; draft != nil && draft.Command.RunID == request.Command.RunID {
		m.textForm = textFormState{}
	}
}
