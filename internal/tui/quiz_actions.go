package tui

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
)

// QuizActionID identifies a quiz action awaiting TUI confirmation.
type QuizActionID string

// Quiz action identifiers define supported quiz commands.
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

// QuizActionExecutor runs confirmed quiz actions.
type QuizActionExecutor interface {
	ExecuteQuizAction(request QuizActionRequest) (QuizActionResult, error)
}

// QuizActionBinding retains the question observed before confirmation or editing.
type QuizActionBinding struct {
	SessionID, QuestionID, ExpectedRevision, OperationID string
}

// QuizActionRequest carries selected quiz action input.
type QuizActionRequest struct {
	ReloadDraft bool
	Binding     QuizActionBinding
	Action      QuizActionID
	Scope       QuizScope
	Mode        string
	Answer      string
}

// QuizActionResult contains refreshed quiz state after an action.
type QuizActionResult struct {
	Outcome       operation.Result
	Message       string
	Quizzes       []QuizView
	SelectedScope QuizScope
}

// QuizActionState stores pending confirmation state for a quiz action.
type QuizActionState struct {
	Pending bool
	Request QuizActionRequest
	Title   string
	Details []string
}

// QuizAnswerState stores in-terminal quiz answer editing state.
type QuizAnswerState struct {
	History  []QuizAnswerRecord
	Question string
	Binding  QuizActionBinding
	Active   bool
	Scope    QuizScope
	Input    textarea.Model
}

type quizActionFinishedMsg struct {
	Binding operationBinding
	Request QuizActionRequest
	Result  QuizActionResult
	Err     error
}

func (m *Model) renderQuizAction() string {
	if !m.quizAction.Pending {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nConfirm Action\n")
	fmt.Fprintf(&output, "  %s\n", m.quizAction.Title)

	for _, detail := range m.quizAction.Details {
		output.WriteString("  ")
		output.WriteString(detail)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m *Model) updateQuizActionKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelQuizAction()
	case tea.KeyEnter:
		return m.confirmQuizAction()
	}

	return nil
}

func (m *Model) updateQuizAnswerKey(msg tea.KeyMsg) tea.Cmd {
	if command, handled := m.updateQuizAnswerControl(msg.String()); handled {
		return command
	}

	switch msg.String() {
	case "esc":
		if m.quizAnswer.Input.Value() != "" {
			return m.confirmDraftChange("Discard the unsent answer?", discardQuizDraft)
		}

		return discardQuizDraft(m)

	case "enter":
		return m.saveQuizAnswer()
	case "shift+enter", "ctrl+j":
		input, cmd := m.quizAnswer.Input.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m.quizAnswer.Input = input

		return cmd
	}

	previous := m.quizAnswer.Input.Value()

	input, cmd := m.quizAnswer.Input.Update(msg)

	m.quizAnswer.Input = input
	if previous != input.Value() {
		m.quizAnswer.Binding.OperationID = rand.Text()
	}

	return cmd
}

func (m *Model) updateQuizAnswerControl(key string) (tea.Cmd, bool) {
	if m.scrollQuizHistoryKey(key) {
		return nil, true
	}

	actions := map[string]QuizActionID{"alt+s": QuizActionSkip, "alt+d": QuizActionComplete, "alt+x": QuizActionCancel}
	if key == "alt+m" {
		actions[key] = QuizActionFastMode
		if quiz, ok := m.selectedQuizView(); ok && quiz.Mode == "fast" {
			actions[key] = QuizActionDeepMode
		}
	}

	if action, ok := actions[key]; ok {
		return m.requestQuizControl(action), true
	}

	return nil, false
}

func (m *Model) planQuizAction(action QuizActionID) {
	if mode := quizActionMode(action); mode != "" {
		if quiz, ok := m.selectedQuizView(); ok && quizCanStart(quiz) {
			m.setQuizStartMode(quiz, mode)

			return
		}
	}

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

func (m *Model) confirmQuizAction() tea.Cmd {
	request := m.quizAction.Request
	m.quizAction = QuizActionState{}

	cmd := m.startQuizAction(request)

	return cmd
}

func (m *Model) saveQuizAnswer() tea.Cmd {
	answer := m.quizAnswer.Input.Value()
	scope := m.quizAnswer.Scope
	binding := m.quizAnswer.Binding

	cmd := m.startQuizAction(QuizActionRequest{
		Action:  QuizActionAnswer,
		Binding: binding,
		Scope:   scope,
		Answer:  answer,
	})

	return cmd
}

func (m *Model) executeQuizActionCommand(request QuizActionRequest) tea.Cmd {
	binding := m.beginOperation("quiz")
	executor := m.quizExecutor

	return func() tea.Msg {
		if executor == nil {
			return quizActionFinishedMsg{
				Binding: binding,
				Request: request,
				Err:     fmt.Errorf("quiz action executor is unavailable"),
			}
		}

		result, err := executor.ExecuteQuizAction(request)

		return quizActionFinishedMsg{Binding: binding, Request: request, Result: result, Err: err}
	}
}

func (m *Model) startQuizAction(request QuizActionRequest) tea.Cmd {
	if m.quizExecutor == nil {
		m.statusMessage = "Quiz action executor is unavailable."

		return nil
	}

	m.quizBusy = true
	m.statusMessage = quizBusyMessage(request)

	return m.executeQuizActionCommand(request)
}

func (m *Model) completeQuizAction(message quizActionFinishedMsg) {
	owned, apply := m.finishOperation(message.Binding)
	if !owned {
		return
	}

	m.quizBusy = false

	if !apply {
		return
	}

	m.recordOperation(message.Result.Outcome)

	m.bindQuizRecovery(message.Request)

	if message.Err != nil {
		m.statusMessage = "Action failed: " + message.Err.Error()

		return
	}

	if message.Request.Action != QuizActionRefresh || message.Request.ReloadDraft {
		m.quizAnswer = QuizAnswerState{}
	}

	m.applyQuizActionResult(message.Result)
}

func (m *Model) applyQuizActionResult(result QuizActionResult) {
	if result.Quizzes != nil {
		m.quizzes = m.reconcileQuizModes(result.Quizzes)
		m.selectQuizByScope(result.SelectedScope)
	}

	m.statusMessage = result.Message
	m.startQuizAnswerIfActive()
}

func (m *Model) startQuizAnswer() {
	m.draftSequence++

	quiz, ok := m.selectedQuizView()
	if !ok {
		m.statusMessage = "No quiz selected."

		return
	}

	if strings.TrimSpace(quiz.ActiveQuestionSummary) == "" {
		m.statusMessage = "No active question."

		return
	}

	input := textarea.New()
	input.Focus()

	input.ShowLineNumbers = false
	input.Prompt = ""
	input.CharLimit = 0

	binding := quiz.Binding
	if binding.ExpectedRevision != "" {
		binding.OperationID = rand.Text()
	}

	m.quizAnswer = QuizAnswerState{
		History:  append([]QuizAnswerRecord(nil), quiz.History...),
		Question: quiz.ActiveQuestionSummary,
		Binding:  binding,
		Active:   true,
		Scope:    quiz.Scope,
		Input:    input,
	}
	m.quizHistoryScroll = 0
	m.selectedPane = 1
	m.maximizedPane = 1
	m.statusMessage = "Answering quiz."
}

func (m *Model) startQuizAnswerIfActive() {
	quiz, ok := m.selectedQuizView()
	if !ok || strings.TrimSpace(quiz.ActiveQuestionSummary) == "" {
		return
	}

	if m.quizAnswer.Active && quizScopeLabel(m.quizAnswer.Scope) == quizScopeLabel(quiz.Scope) {
		m.selectedPane = 1
		m.maximizedPane = 1

		return
	}

	m.startQuizAnswer()
}

func (m *Model) renderQuizAnswerInput(width, height int) string {
	input := m.quizAnswer.Input
	input.SetWidth(maxInt(width, 1))
	input.SetHeight(maxInt(height, 3))

	return input.View()
}

func quizBusyMessage(request QuizActionRequest) string {
	switch request.Action {
	case QuizActionAnswer:
		return "Generating next question..."
	case QuizActionStart, QuizActionResume:
		return "Loading question..."
	case QuizActionSkip:
		return "Skipping question..."
	case QuizActionComplete:
		return "Completing quiz..."
	case QuizActionRefresh:
		return "Refreshing quizzes..."
	default:
		return "Working..."
	}
}
