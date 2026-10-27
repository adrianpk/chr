package app

import (
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// QuizListResult contains the current quiz list and a status message.
type QuizListResult struct {
	Message string
	States  []quizcore.RuntimeState
	Views   []QuizView
}

// QuizOperationResult is the TUI-facing result for quiz mutations.
type QuizOperationResult struct {
	Message string
	State   quizcore.RuntimeState
}

// QuizStartRequest selects the scope for a new quiz session.
type QuizStartRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
	Mode    quizcore.Mode
}

// QuizResumeRequest selects the scope for a resumed quiz session.
type QuizResumeRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
}

// QuizAnswerRequest submits an answer for the active quiz question.
type QuizAnswerRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
	Answer  string
}

// QuizSkipRequest skips the active quiz question for the selected scope.
type QuizSkipRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
}

// QuizCompleteRequest completes the quiz for the selected scope.
type QuizCompleteRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
}

// QuizDraftSpecRequest selects the quiz scope used to draft a spec.
type QuizDraftSpecRequest struct {
	Scope quizcore.Scope
}

// QuizCancelRequest cancels the quiz for the selected scope.
type QuizCancelRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
}

// QuizModeRequest changes the quiz interaction mode for the selected scope.
type QuizModeRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
	Mode    quizcore.Mode
}

func (s service) ListQuizzes() (QuizListResult, error) {
	states, err := s.listQuizRuntimeStates()
	if err != nil {
		return QuizListResult{}, err
	}

	views, err := s.quizViewsFromStates(states)
	if err != nil {
		return QuizListResult{}, err
	}

	return QuizListResult{Message: "Quizzes loaded.", States: states, Views: views}, nil
}

func (s service) StartQuiz(request QuizStartRequest) (QuizOperationResult, error) {
	mode := request.Mode
	if mode == "" {
		mode = quizcore.ModeDeep
	}

	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "start", Mode: mode})
}

func (s service) ResumeQuiz(request QuizResumeRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "resume"})
}

func (s service) AnswerQuiz(request QuizAnswerRequest) (QuizOperationResult, error) {
	answer := strings.TrimSpace(request.Answer)
	if answer == "" {
		return QuizOperationResult{}, fmt.Errorf("quiz answer is required")
	}

	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "answer", Answer: answer})
}

func (s service) SkipQuiz(request QuizSkipRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "skip"})
}

func (s service) CompleteQuiz(request QuizCompleteRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "done"})
}

func (s service) GenerateQuizDraftSpec(request QuizDraftSpecRequest) (QuizOperationResult, error) {
	return s.generateQuizDraftSpec(s.quizScope(request.Scope))
}

func (s service) CancelQuiz(request QuizCancelRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "cancel"})
}

func (s service) SwitchQuizMode(request QuizModeRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "mode", Mode: request.Mode})
}
