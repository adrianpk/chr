package app

import (
	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func quizGenerationOperationID(request core.QuizActionRequest, state quizcore.QuizRuntimeState) string {
	if request.Action == "resume" && state.ContinuationPending && state.GenerationOperationID != "" {
		return state.GenerationOperationID
	}

	return "quiz-" + request.Binding.OperationID
}

func (s service) invokeQuizGeneration(state quizcore.QuizRuntimeState, request core.QuizModelRequest) (core.QuizModelResponse, error) {
	bound, _, err := s.bindGeneration(s.ctx, state.GenerationOperationID, string(core.QuizModelActionNextQuestion))
	if err != nil {
		return core.QuizModelResponse{}, err
	}

	request.ID = core.HarnessGenerationRequestID(bound.ctx, request.ID)

	return core.InvokeQuizModel(bound.ctx, bound.quizModel, request)
}
