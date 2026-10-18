package app

import (
	"context"
	"crypto/rand"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
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

func (s service) quizActionBinding(scope core.QuizScope, binding core.QuizActionBinding) (core.QuizActionBinding, error) {
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
	state := observed.State

	switch request.Action {
	case "start":
		return s.prepareQuizStart(request, observed)
	case "answer", "skip":
		return s.prepareQuizOutcome(request, observed)
	case "mode":
		return s.prepareQuizMode(request.Mode, observed)
	case "resume":
		if state.ActiveQuestion == nil && state.Status != core.QuizRuntimeStatusDone {
			state = s.continueQuiz(state, observed.Entries)
		}

		return core.QuizTransition{State: state, Message: "Quiz resumed."}, nil
	case "done", "cancel":
		state.Status = core.QuizRuntimeStatusDone

		message := "Quiz completed."
		if request.Action == "cancel" {
			state.Status, message = core.QuizRuntimeStatusCancelled, "Quiz cancelled."
		}

		state.ActiveQuestion, state.Candidates = nil, nil
		state.ContinuationPending, state.ContinuationError = false, ""
		state.UpdatedAt = s.utcNow()

		return core.QuizTransition{State: state, Message: message}, nil
	default:
		return core.QuizTransition{}, fmt.Errorf("unsupported quiz action")
	}
}

func (s service) prepareQuizStart(request core.QuizActionRequest, observed core.QuizObservation) (core.QuizTransition, error) {
	if !request.Mode.Valid() {
		return core.QuizTransition{}, fmt.Errorf("quiz mode is invalid: %s", request.Mode)
	}

	if observed.Exists && observed.State.Status != core.QuizRuntimeStatusDone && observed.State.Status != core.QuizRuntimeStatusCancelled {
		return core.QuizTransition{}, fmt.Errorf("quiz session already exists; resume it")
	}

	state, err := s.newQuizRuntimeState(request.Scope, request.Mode)
	if err != nil {
		return core.QuizTransition{}, err
	}

	state, err = s.nextQuizQuestion(state, observed.Entries)

	return core.QuizTransition{State: state, Message: "Quiz started."}, err
}

func (s service) prepareQuizOutcome(request core.QuizActionRequest, observed core.QuizObservation) (core.QuizTransition, error) {
	state := observed.State
	if state.Status != core.QuizRuntimeStatusActive || state.ActiveQuestion == nil {
		return core.QuizTransition{}, fmt.Errorf("quiz has no active question")
	}

	status := core.QuizEntryStatusAnswered
	if request.Action == "skip" {
		status = core.QuizEntryStatusSkipped
	}

	answer := request.Answer
	if status == core.QuizEntryStatusSkipped {
		answer = ""
	}

	entry := quizEntryFromState(state, status, answer, s.utcNow())
	entry.ID = fmt.Sprintf("%s-entry-%03d", state.SessionID, state.QuestionsAsked)
	state = updateQuizStateAfterEntry(state, entry)
	state = s.continueQuiz(state, append(append([]core.QuizEntry(nil), observed.Entries...), entry))

	message := "Quiz answer recorded."
	if status == core.QuizEntryStatusSkipped {
		message = "Quiz question skipped."
	}

	if state.ContinuationPending {
		message += " Continuation failed; reload and resume to retry."
	}

	return core.QuizTransition{State: state, Entry: &entry, Message: message}, nil
}

func (s service) continueQuiz(state core.QuizRuntimeState, entries []core.QuizEntry) core.QuizRuntimeState {
	state.ContinuationPending, state.ContinuationError = false, ""
	if state.QuestionsAsked >= state.QuestionBudget {
		state.Status, state.ActiveQuestion = core.QuizRuntimeStatusIdle, nil

		return state
	}

	next, err := s.nextQuizQuestion(state, entries)
	if err != nil {
		state.Status, state.ActiveQuestion = core.QuizRuntimeStatusBlocked, nil
		state.ContinuationPending, state.ContinuationError = true, "Question generation failed. Resume to retry."
		state.UpdatedAt = s.utcNow()

		return state
	}

	return next
}

func (s service) prepareQuizMode(mode core.QuizMode, observed core.QuizObservation) (core.QuizTransition, error) {
	if !mode.Valid() {
		return core.QuizTransition{}, fmt.Errorf("quiz mode is invalid: %s", mode)
	}

	state := observed.State
	if state.Status == core.QuizRuntimeStatusDone || state.Status == core.QuizRuntimeStatusCancelled {
		return core.QuizTransition{}, fmt.Errorf("quiz session is not open")
	}

	state.Mode = mode

	state.Candidates = nil
	if mode == core.QuizModeFast && state.FastBatchID == "" {
		var err error

		state, err = s.generateFastQuizBatch(state, observed.Entries)
		if err != nil {
			return core.QuizTransition{}, err
		}
	}

	if mode == core.QuizModeFast {
		state = retainUnansweredFastQuestions(state, observed.Entries)
		state = s.activateNextFastQuestion(state)
	}

	state.UpdatedAt = s.utcNow()

	return core.QuizTransition{State: state, Message: "Quiz mode switched."}, nil
}
