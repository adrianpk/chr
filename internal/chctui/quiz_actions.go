package chctui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type QuizActionID string

const (
	QuizActionStart    QuizActionID = "start_quiz"
	QuizActionResume   QuizActionID = "resume_quiz"
	QuizActionAnswer   QuizActionID = "answer_quiz"
	QuizActionSkip     QuizActionID = "skip_quiz"
	QuizActionComplete QuizActionID = "complete_quiz"
	QuizActionCancel   QuizActionID = "cancel_quiz"
	QuizActionRefresh  QuizActionID = "refresh_quizzes"
	QuizActionDeepMode QuizActionID = "quiz_deep_mode"
	QuizActionFastMode QuizActionID = "quiz_fast_mode"
)

type QuizActionExecutor interface {
	ExecuteQuizAction(request QuizActionRequest) (QuizActionResult, error)
}

type QuizActionRequest struct {
	Action QuizActionID
	Scope  QuizScope
	Mode   string
	Answer string
}

type QuizActionResult struct {
	Message       string
	Quizzes       []QuizView
	SelectedScope QuizScope
}

type QuizActionState struct {
	Pending bool
	Request QuizActionRequest
	Title   string
	Details []string
}

type QuizAnswerState struct {
	Active bool
	Scope  QuizScope
	Input  textarea.Model
}

func (m Model) renderQuizAction() string {
	if !m.quizAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Action\n")
	output.WriteString(fmt.Sprintf("  %s\n", m.quizAction.Title))

	for _, detail := range m.quizAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m Model) updateQuizActionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelQuizAction()
	case tea.KeyEnter:
		m.confirmQuizAction()
	}

	return m, nil
}

func (m Model) updateQuizAnswerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.quizAnswer = QuizAnswerState{}
		m.statusMessage = ""

		return m, nil
	case tea.KeyCtrlS:
		m.saveQuizAnswer()

		return m, nil
	}

	input, cmd := m.quizAnswer.Input.Update(msg)
	m.quizAnswer.Input = input

	return m, cmd
}

func (m *Model) planQuizAction(action QuizActionID) {
	if action == QuizActionAnswer {
		m.startQuizAnswer()

		return
	}

	m.quizAction = m.newQuizActionState(action)
	m.statusMessage = ""
}

func (m *Model) cancelQuizAction() {
	m.quizAction = QuizActionState{}
	m.statusMessage = "Action cancelled."
}

func (m *Model) confirmQuizAction() {
	request := m.quizAction.Request
	m.quizAction = QuizActionState{}
	m.executeQuizAction(request)
}

func (m *Model) saveQuizAnswer() {
	answer := m.quizAnswer.Input.Value()
	scope := m.quizAnswer.Scope
	m.quizAnswer = QuizAnswerState{}

	m.executeQuizAction(QuizActionRequest{
		Action: QuizActionAnswer,
		Scope:  scope,
		Answer: answer,
	})
}

func (m *Model) executeQuizAction(request QuizActionRequest) {
	if m.quizExecutor == nil {
		m.statusMessage = "Quiz action executor is unavailable."

		return
	}

	result, err := m.quizExecutor.ExecuteQuizAction(request)
	if err != nil {
		m.statusMessage = "Action failed: " + err.Error()

		return
	}

	m.applyQuizActionResult(result)
}

func (m *Model) applyQuizActionResult(result QuizActionResult) {
	if result.Quizzes != nil {
		m.quizzes = normalizeQuizViews(result.Quizzes)
		m.selectQuizByScope(result.SelectedScope)
	}

	m.statusMessage = result.Message
}

func (m *Model) startQuizAnswer() {
	quiz, ok := m.selectedQuizView()
	if !ok {
		m.statusMessage = "No quiz selected."

		return
	}

	input := textarea.New()
	input.Focus()

	input.ShowLineNumbers = false
	input.Prompt = ""
	input.CharLimit = 0

	m.quizAnswer = QuizAnswerState{
		Active: true,
		Scope:  quiz.Scope,
		Input:  input,
	}
	m.statusMessage = "Editing quiz answer. Press ctrl+s to submit or esc to cancel."
}

func (m Model) renderQuizAnswerInput(width, height int) string {
	input := m.quizAnswer.Input
	input.SetWidth(maxInt(width, 1))
	input.SetHeight(maxInt(height, 3))

	return input.View()
}
