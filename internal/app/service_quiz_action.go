package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func (s service) applyQuizAction(request core.QuizActionRequest) (QuizOperationResult, error) {
	binding, err := s.quizActionBinding(request.Scope, request.Binding)
	if err != nil {
		return QuizOperationResult{}, err
	}

	request.Binding = binding
	result, err := s.quizRepository().ApplyAction(s.ctx, request, func(ctx context.Context, observed core.QuizObservation) (core.QuizTransition, error) {
		bounded := s
		bounded.ctx = ctx

		return bounded.prepareQuizAction(request, observed)
	})

	return QuizOperationResult{Message: result.Message, State: result.State}, err
}

func (s service) quizActionBinding(scope quizcore.Scope, binding core.QuizActionBinding) (core.QuizActionBinding, error) {
	if binding.ExpectedRevision != "" {
		return binding, nil
	}

	if binding != (core.QuizActionBinding{}) {
		return binding, fmt.Errorf("complete quiz action binding is required")
	}
	// Legacy command callers observe before acquiring the writer. Concurrent
	// commands therefore conflict rather than answering the next question.
	observed, err := s.quizRepository().Observe(s.ctx, scope)
	if err != nil {
		return binding, err
	}

	binding = core.QuizActionBinding{SessionID: observed.State.SessionID, ExpectedRevision: observed.Revision, OperationID: rand.Text()}
	if observed.State.ActiveQuestion != nil {
		binding.QuestionID = observed.State.ActiveQuestion.ID
	}

	return binding, nil
}

func (s service) prepareQuizAction(request core.QuizActionRequest, observed core.QuizObservation) (core.QuizTransition, error) {
	observed.State.GenerationOperationID = quizGenerationOperationID(request, observed.State)

	switch request.Action {
	case "start":
		return s.prepareQuizStart(request, observed)
	case "answer", "skip":
		return s.prepareQuizOutcome(request, observed)
	case "mode":
		return s.prepareQuizMode(request.Mode, observed)
	case "resume":
		state := observed.State
		if state.ActiveQuestion == nil && state.Status != quizcore.RuntimeStatusDone {
			state = s.continueQuiz(state, observed.Entries)
		}

		return core.QuizTransition{State: state, Message: "Quiz resumed."}, nil
	case "done", "cancel":
		return s.prepareQuizClose(request.Action, observed)
	default:
		return core.QuizTransition{}, fmt.Errorf("unsupported quiz action")
	}
}

func (s service) prepareQuizStart(request core.QuizActionRequest, observed core.QuizObservation) (core.QuizTransition, error) {
	if !request.Mode.Valid() {
		return core.QuizTransition{}, fmt.Errorf("quiz mode is invalid: %s", request.Mode)
	}

	if observed.Exists && observed.State.Status != quizcore.RuntimeStatusDone && observed.State.Status != quizcore.RuntimeStatusCancelled {
		return core.QuizTransition{}, fmt.Errorf("quiz session already exists; resume it")
	}

	state, err := s.newQuizRuntimeState(request.Scope, request.Mode)
	if err != nil {
		return core.QuizTransition{}, err
	}

	state.GenerationOperationID = "quiz-" + request.Binding.OperationID

	state, err = s.nextQuizQuestion(state, observed.Entries)

	message := "Quiz started."
	if state.Status == quizcore.RuntimeStatusDone {
		message = "Quiz completed. The saved context is sufficient."
	}

	return core.QuizTransition{State: state, Message: message}, err
}

func (s service) prepareQuizOutcome(request core.QuizActionRequest, observed core.QuizObservation) (core.QuizTransition, error) {
	state := observed.State
	if state.Status != quizcore.RuntimeStatusActive || state.ActiveQuestion == nil {
		return core.QuizTransition{}, fmt.Errorf("quiz has no active question")
	}

	status := quizcore.EntryStatusAnswered
	if request.Action == "skip" {
		status = quizcore.EntryStatusSkipped
	}

	answer := request.Answer
	if status == quizcore.EntryStatusSkipped {
		answer = ""
	}

	entry := quizEntryFromState(state, status, answer, s.utcNow())
	entry.ID = fmt.Sprintf("%s-entry-%03d", state.SessionID, state.QuestionsAsked)
	state = updateQuizStateAfterEntry(state, entry)
	state = s.continueQuiz(state, append(append([]quizcore.Entry(nil), observed.Entries...), entry))

	message := "Quiz answer recorded."
	if status == quizcore.EntryStatusSkipped {
		message = "Quiz question skipped."
	}

	if state.ContinuationPending {
		message += " Continuation failed; reload and resume to retry."
	}

	if state.Status == quizcore.RuntimeStatusDone {
		message += " Quiz completed."
	}

	return core.QuizTransition{State: state, Entry: &entry, Message: message}, nil
}

func (s service) continueQuiz(state quizcore.RuntimeState, entries []quizcore.Entry) quizcore.RuntimeState {
	state.ContinuationPending, state.ContinuationError = false, ""
	if state.QuestionsAsked >= state.QuestionBudget {
		return s.finishQuiz(state)
	}

	next, err := s.nextQuizQuestion(state, entries)
	if err != nil {
		state.Status, state.ActiveQuestion = quizcore.RuntimeStatusBlocked, nil
		state.ContinuationPending, state.ContinuationError = true, "Question generation failed. Resume to retry."

		if errors.Is(err, errRepeatedQuizQuestion) {
			state.ContinuationError = "Question generation repeated an answered question. Resume to retry."
			state.GenerationOperationID = ""
		}

		state.UpdatedAt = s.utcNow()

		return state
	}

	return next
}

func (s service) prepareQuizMode(mode quizcore.Mode, observed core.QuizObservation) (core.QuizTransition, error) {
	if !mode.Valid() {
		return core.QuizTransition{}, fmt.Errorf("quiz mode is invalid: %s", mode)
	}

	state := observed.State
	if state.Status == quizcore.RuntimeStatusDone || state.Status == quizcore.RuntimeStatusCancelled {
		return core.QuizTransition{}, fmt.Errorf("quiz session is not open")
	}

	state.Mode = mode

	state.Candidates = nil
	if mode == quizcore.ModeFast && state.FastBatchID == "" {
		var err error

		state, err = s.generateFastQuizBatch(state, observed.Entries)
		if err != nil {
			return core.QuizTransition{}, err
		}
	}

	if mode == quizcore.ModeFast {
		state = retainUnansweredFastQuestions(state, observed.Entries)
		state = s.activateNextFastQuestion(state)
	}

	state.UpdatedAt = s.utcNow()

	return core.QuizTransition{State: state, Message: "Quiz mode switched."}, nil
}

func (s service) prepareQuizClose(action string, observed core.QuizObservation) (core.QuizTransition, error) {
	state := observed.State
	if action == "done" && s.guidedContext != nil {
		return s.prepareGuidedQuizDone(observed)
	}

	state.Status = quizcore.RuntimeStatusDone

	message := "Quiz completed."
	if action == "cancel" {
		state.Status, message = quizcore.RuntimeStatusCancelled, "Quiz cancelled."
	}

	state.ActiveQuestion, state.Candidates = nil, nil
	state.ContinuationPending, state.ContinuationError = false, ""
	state.UpdatedAt = s.utcNow()

	return core.QuizTransition{State: state, Message: message}, nil
}
