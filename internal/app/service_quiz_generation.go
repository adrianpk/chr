package app

import (
	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func quizGenerationOperationID(request core.QuizActionRequest, state quizcore.RuntimeState) string {
	if request.Action == "resume" && state.ContinuationPending && state.GenerationOperationID != "" {
		return state.GenerationOperationID
	}

	return "quiz-" + request.Binding.OperationID
}

func (s service) invokeQuizGeneration(state quizcore.RuntimeState, request quizmodel.Request) (quizmodel.Response, error) {
	bound, _, err := s.bindGeneration(s.ctx, state.GenerationOperationID, string(quizmodel.ActionNextQuestion))
	if err != nil {
		return quizmodel.Response{}, err
	}

	request.ID = core.HarnessGenerationRequestID(bound.ctx, request.ID)

	return quizmodel.Invoke(bound.ctx, bound.quizModel, request)
}
