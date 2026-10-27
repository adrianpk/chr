package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func (s service) applyQuizAction(ctx context.Context, request quizcore.ActionRequest) (QuizOperationResult, error) {
	binding, err := s.quizActionBinding(ctx, request.Scope, request.Binding)
	if err != nil {
		return QuizOperationResult{}, err
	}

	request.Binding = binding
	result, err := s.quizRepository().ApplyAction(ctx, request, func(ctx context.Context, observed quizcore.Observation) (quizcore.Transition, error) {
		return s.prepareQuizAction(ctx, request, observed)
	})

	return QuizOperationResult{Message: result.Message, State: result.State}, err
}

func (s service) quizActionBinding(ctx context.Context, scope quizcore.Scope, binding quizcore.ActionBinding) (quizcore.ActionBinding, error) {
	if binding.ExpectedRevision != "" {
		return binding, nil
	}

	if binding != (quizcore.ActionBinding{}) {
		return binding, fmt.Errorf("complete quiz action binding is required")
	}
	// Legacy command callers observe before acquiring the writer. Concurrent
	// commands therefore conflict rather than answering the next question.
	observed, err := s.quizRepository().Observe(ctx, scope)
	if err != nil {
		return binding, err
	}

	binding = quizcore.ActionBinding{SessionID: observed.State.SessionID, ExpectedRevision: observed.Revision, OperationID: rand.Text()}
	if observed.State.ActiveQuestion != nil {
		binding.QuestionID = observed.State.ActiveQuestion.ID
	}

	return binding, nil
}

func (s service) prepareQuizAction(ctx context.Context, request quizcore.ActionRequest, observed quizcore.Observation) (quizcore.Transition, error) {
	observed.State.GenerationOperationID = quizGenerationOperationID(request, observed.State)

	switch request.Action {
	case "start":
		return s.prepareQuizStart(ctx, request, observed)
	case "answer", "skip":
		return s.prepareQuizOutcome(ctx, request, observed)
	case "mode":
		return s.prepareQuizMode(ctx, request.Mode, observed)
	case "resume":
		state := observed.State
		if state.ActiveQuestion == nil && state.Status != quizcore.RuntimeStatusDone {
			state = s.continueQuiz(ctx, state, observed.Entries)
		}

		return quizcore.Transition{State: state, Message: "Quiz resumed."}, nil
	case "done", "cancel":
		return s.prepareQuizClose(ctx, request.Action, observed)
	default:
		return quizcore.Transition{}, fmt.Errorf("unsupported quiz action")
	}
}

func (s service) prepareQuizStart(ctx context.Context, request quizcore.ActionRequest, observed quizcore.Observation) (quizcore.Transition, error) {
	if !request.Mode.Valid() {
		return quizcore.Transition{}, fmt.Errorf("quiz mode is invalid: %s", request.Mode)
	}

	if observed.Exists && observed.State.Status != quizcore.RuntimeStatusDone && observed.State.Status != quizcore.RuntimeStatusCancelled {
		return quizcore.Transition{}, fmt.Errorf("quiz session already exists; resume it")
	}

	state, err := s.newQuizRuntimeState(request.Scope, request.Mode)
	if err != nil {
		return quizcore.Transition{}, err
	}

	state.GenerationOperationID = "quiz-" + request.Binding.OperationID

	state, err = s.nextQuizQuestion(ctx, state, observed.Entries)

	message := "Quiz started."
	if state.Status == quizcore.RuntimeStatusDone {
		message = "Quiz completed. The saved context is sufficient."
	}

	return quizcore.Transition{State: state, Message: message}, err
}

func (s service) prepareQuizOutcome(ctx context.Context, request quizcore.ActionRequest, observed quizcore.Observation) (quizcore.Transition, error) {
	state := observed.State
	if state.Status != quizcore.RuntimeStatusActive || state.ActiveQuestion == nil {
		return quizcore.Transition{}, fmt.Errorf("quiz has no active question")
	}

	status := quizcore.EntryStatusAnswered
	if request.Action == "skip" {
		status = quizcore.EntryStatusSkipped
	}

	answer := request.Answer
	if status == quizcore.EntryStatusSkipped {
		answer = ""
	}

	entry := quizcore.EntryFromState(state, status, answer, s.utcNow())
	entry.ID = fmt.Sprintf("%s-entry-%03d", state.SessionID, state.QuestionsAsked)
	state = quizcore.StateAfterEntry(state, entry)
	state = s.continueQuiz(ctx, state, append(append([]quizcore.Entry(nil), observed.Entries...), entry))

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

	return quizcore.Transition{State: state, Entry: &entry, Message: message}, nil
}

func (s service) continueQuiz(ctx context.Context, state quizcore.RuntimeState, entries []quizcore.Entry) quizcore.RuntimeState {
	state.ContinuationPending, state.ContinuationError = false, ""
	if state.QuestionsAsked >= state.QuestionBudget {
		return s.finishQuiz(state)
	}

	next, err := s.nextQuizQuestion(ctx, state, entries)
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

func (s service) prepareQuizMode(ctx context.Context, mode quizcore.Mode, observed quizcore.Observation) (quizcore.Transition, error) {
	if !mode.Valid() {
		return quizcore.Transition{}, fmt.Errorf("quiz mode is invalid: %s", mode)
	}

	state := observed.State
	if state.Status == quizcore.RuntimeStatusDone || state.Status == quizcore.RuntimeStatusCancelled {
		return quizcore.Transition{}, fmt.Errorf("quiz session is not open")
	}

	state.Mode = mode

	state.Candidates = nil
	if mode == quizcore.ModeFast && state.FastBatchID == "" {
		var err error

		state, err = s.generateFastQuizBatch(ctx, state, observed.Entries)
		if err != nil {
			return quizcore.Transition{}, err
		}
	}

	if mode == quizcore.ModeFast {
		state = retainUnansweredFastQuestions(state, observed.Entries)
		state = s.activateNextFastQuestion(state)
	}

	state.UpdatedAt = s.utcNow()

	return quizcore.Transition{State: state, Message: "Quiz mode switched."}, nil
}

func (s service) prepareQuizClose(ctx context.Context, action string, observed quizcore.Observation) (quizcore.Transition, error) {
	state := observed.State
	if action == "done" && s.guidedContext != nil {
		return s.prepareGuidedQuizDone(ctx, observed)
	}

	state.Status = quizcore.RuntimeStatusDone

	message := "Quiz completed."
	if action == "cancel" {
		state.Status, message = quizcore.RuntimeStatusCancelled, "Quiz cancelled."
	}

	state.ActiveQuestion, state.Candidates = nil, nil
	state.ContinuationPending, state.ContinuationError = false, ""
	state.UpdatedAt = s.utcNow()

	return quizcore.Transition{State: state, Message: message}, nil
}
