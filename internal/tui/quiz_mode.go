package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func quizCanStart(quiz QuizView) bool {
	return quiz.Status == "not_started" || quiz.Status == "done" || quiz.Status == "cancelled"
}

func quizModeLine(quiz QuizView) string {
	mode := "Deep"
	if quiz.Mode == "fast" {
		mode = "Fast"
	}

	return "Mode: " + mode
}

func quizListLabel(quiz QuizView) string {
	if quiz.Status == "active" && strings.TrimSpace(quiz.ActiveQuestionSummary) != "" {
		return quiz.Title + " [answer pending]"
	}

	return quiz.Title
}

func (m *Model) setQuizStartMode(quiz QuizView, mode string) {
	for index := range m.quizzes {
		if m.quizzes[index].Index == quiz.Index {
			m.quizzes[index].Mode = mode
		}
	}

	quiz.Mode = mode
	m.statusMessage = quizModeLine(quiz) + ". Press s to start."
}

func (m *Model) cycleQuizMode() tea.Cmd {
	quiz, ok := m.selectedQuizView()
	if !ok {
		m.statusMessage = "No quiz selected."

		return nil
	}

	mode, action := "fast", QuizActionFastMode
	if quiz.Mode == "fast" {
		mode, action = "deep", QuizActionDeepMode
	}

	if quizCanStart(quiz) {
		m.setQuizStartMode(quiz, mode)

		return nil
	}

	return m.startQuizAction(m.newQuizActionState(action).Request)
}

func (m *Model) reconcileQuizModes(quizzes []QuizView) []QuizView {
	previous := make(map[QuizScope]QuizView, len(m.quizzes))
	for _, quiz := range m.quizzes {
		previous[quiz.Scope] = quiz
	}

	next := normalizeQuizViews(quizzes)
	for index := range next {
		if before, ok := previous[next[index].Scope]; ok && quizCanStart(next[index]) && before.Binding == next[index].Binding {
			next[index].Mode = before.Mode
		}
	}

	return next
}
