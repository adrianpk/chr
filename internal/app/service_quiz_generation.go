package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func quizGenerationOperationID(request quizcore.ActionRequest, state quizcore.RuntimeState) string {
	if request.Action == "resume" && state.ContinuationPending && state.GenerationOperationID != "" {
		return state.GenerationOperationID
	}

	return "quiz-" + request.Binding.OperationID
}

func (s service) invokeQuizGeneration(ctx context.Context, state quizcore.RuntimeState, request quizmodel.Request) (quizmodel.Response, error) {
	bound, err := s.bindGeneration(ctx, state.GenerationOperationID, string(quizmodel.ActionNextQuestion))
	if err != nil {
		return quizmodel.Response{}, err
	}

	ctx = bound.context(ctx)
	request.ID = core.HarnessGenerationRequestID(ctx, request.ID)

	return quizmodel.Invoke(ctx, bound.quizModel, request)
}
