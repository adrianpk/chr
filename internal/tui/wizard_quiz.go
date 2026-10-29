package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

func (m *Model) wizardQuizCurrent() bool {
	return m.wizardQuizVisible() && authoringcore.GuidedDispatchAllowed(*m.guided.View.Session) && m.guided.View.Session.Guided.PendingOperation == nil
}

func wizardQuizKey(key string) bool {
	switch key {
	case "enter", "m", "u", "x", "pgup", "pgdown":
		return true
	default:
		return false
	}
}

func (m *Model) capturedWizardQuiz() *GuidedRequest {
	if !m.wizardQuizCurrent() {
		return nil
	}

	request := m.guidedRequest("input")

	return &request
}

// Shared quiz editing and controls retain the guided owner's observed command.
// They never invoke the manual quiz executor for a guided case.
func (m *Model) startWizardQuizAction(request QuizActionRequest) tea.Cmd {
	guided := *request.Guided
	guided.QuizAction = request.Action
	guided.Command.OperationID = request.Binding.OperationID
	actions := map[QuizActionID]string{QuizActionAnswer: "answer", QuizActionSkip: "quiz-skip", QuizActionComplete: "quiz-done", QuizActionCancel: "cancel", QuizActionResume: "quiz", QuizActionDeepMode: "quiz-mode", QuizActionFastMode: "quiz-mode"}

	action, ok := actions[request.Action]
	if !ok {
		return nil
	}

	guided.Command.Action = action

	guided.Command.Content = request.Answer
	if action == "quiz-mode" {
		guided.Command.Content = request.Mode
	}

	return m.startGuidedRequest(guided)
}

func (m *Model) updateWizardQuizKey(message tea.KeyMsg) (tea.Cmd, bool) {
	if !m.wizardQuizCurrent() || !wizardQuizKey(message.String()) {
		return nil, false
	}

	return m.updateQuizzesKey(message), true
}

func (m *Model) wizardQuizLines() []string {
	if !m.wizardQuizVisible() {
		return nil
	}

	width := 72
	if m.width != 0 {
		width = max(1, m.width-6)
	}

	return m.renderSelectedQuizDetailLines(width)
}

func (m *Model) wizardQuizVisible() bool {
	return m.wizardBound() && m.SelectedView() == ViewQuizzes && m.wizardStageView() == ViewQuizzes && m.guided.View.Quiz != nil
}

func (m *Model) wizardQuizHint(quiz QuizView) string {
	if m.guided.Busy {
		return "Working"
	}

	if !m.wizardQuizCurrent() {
		return "Saved clarification is read-only. Use the eligible workflow controls."
	}

	if m.quizAnswer.Active {
		return "enter submits  shift+enter inserts line  alt+s skip  alt+d done  esc cancel"
	}

	if quiz.ActiveQuestionSummary != "" {
		return "enter answers  m mode  x cancel  r refresh"
	}

	return "u resume clarification  m Deep/Fast  x cancel  r refresh"
}
