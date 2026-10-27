package app

import quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"

import (
	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func quizGenerationOperationID(request core.QuizActionRequest, state quizcore.RuntimeState) string {
	if request.Action == "resume" && state.ContinuationPending && state.GenerationOperationID != "" {
		return state.GenerationOperationID
	}

	return "quiz-" + request.Binding.OperationID
}

func (s service) invokeQuizGeneration(state quizcore.RuntimeState, request quizmodel.QuizModelRequest) (quizmodel.QuizModelResponse, error) {
	bound, _, err := s.bindGeneration(s.ctx, state.GenerationOperationID, string(quizmodel.QuizModelActionNextQuestion))
	if err != nil {
		return quizmodel.QuizModelResponse{}, err
	}

	request.ID = core.HarnessGenerationRequestID(bound.ctx, request.ID)

	return quizmodel.InvokeQuizModel(bound.ctx, bound.quizModel, request)
}
