package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) guidedReplyCurrent(message guidedFinishedMsg) bool {
	before, after := m.guided.View.Session, message.View.Session
	if before == nil || after == nil {
		return true
	}

	if message.Request.Action == "open" || message.Request.Action == "start" {
		return true
	}

	if !m.guidedNativeReplyCurrent(message.View) {
		return false
	}

	if !sameGuidedTerminalRun(m.guided.View, message.View) || after.Revision < before.Revision {
		m.guided.Problem = "A newer or different run is already displayed. Refresh its current state; the captured draft remains unchanged."

		return false
	}

	return true
}

func (m *Model) failGuidedRequest(message guidedFinishedMsg) {
	if sameGuidedTerminalRun(m.guided.View, message.View) {
		if retainGuidedOriginal(GuidedRequest{Action: "inspect"}, m.guided.View, message.View) {
			message.View.Document = m.guided.View.Document
		}

		m.guided.View = message.View
	}

	m.guided.Problem = message.Err.Error()
	m.statusMessage = "Wizard: " + message.Err.Error() + ". The original input and operation are retained."
}

func (m *Model) finishGuidedRequest(request GuidedRequest) {
	if m.guided.Pending != nil && *m.guided.Pending == request {
		m.guided.Pending = nil
	}

	if request.ReloadQuiz {
		m.guided.Pending = nil
	}

	if m.guided.Pending == nil {
		m.guided.Problem = ""
	}
}

func (m *Model) wizardQuizRecoveryKey(key string) (tea.Cmd, bool) {
	if m.quizAnswer.Guided == nil {
		return nil, false
	}

	switch key {
	case "alt+y":
		return m.retryWizardRequest(), true
	case "alt+r":
		request := QuizActionRequest{Action: QuizActionRefresh, Guided: m.quizAnswer.Guided, Scope: m.quizAnswer.Scope, ReloadDraft: true}

		return m.confirmDraftChange("Reload the current quiz and discard the unsent answer?", func(m *Model) tea.Cmd { return m.startQuizAction(request) }), true
	default:
		return nil, false
	}
}

func (m *Model) wizardQuizDraftView() (QuizView, bool) {
	draft := m.quizAnswer
	if !draft.Active || draft.Guided == nil || !m.wizardBound() || m.SelectedView() != ViewQuizzes {
		return QuizView{}, false
	}

	return QuizView{Scope: draft.Scope, Mode: draft.Mode, Binding: draft.Binding, History: draft.History, ActiveQuestionSummary: draft.Question}, true
}

func (m *Model) wizardStagePrompt() string {
	if m.quizAnswer.Active && m.quizAnswer.Guided != nil {
		return "This answer draft retains its original question and revision."
	}

	return m.guided.View.Prompt
}

func (m *Model) guidedNativeReplyCurrent(view GuidedView) bool {
	if view.Native == nil || view.Native.ID != m.workflow.ID {
		return true
	}

	return view.Native.Revision >= m.workflow.Revision
}
