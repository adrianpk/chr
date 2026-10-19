package core

import (
	"context"
	"encoding/json"

	"forge.adrianpk.com/chr/internal/harness"
)

func (a OpenAIQuizModelAdapter) invokeBoundQuizAPI(ctx context.Context, request QuizModelRequest, trace openAIQuizTrace) ([]byte, openAIQuizExecutionResult, error) {
	var result openAIQuizExecutionResult

	schema, err := json.Marshal(quizModelOpenAIFormat(request).Schema)
	if err != nil {
		return nil, result, err
	}

	call := a.quizAPIRunRequest(request, trace)
	call.OutputSchema = schema
	output, err := RunAPIGeneration(ctx, call, func(ctx context.Context) ([]byte, error) {
		body, execution, err := a.invokeOpenAI(ctx, request, trace)
		result = execution

		return body, err
	})

	return output, result, err
}

func (a OpenAIQuizModelAdapter) quizAPIRunRequest(request QuizModelRequest, trace openAIQuizTrace) harness.AgentRunRequest {
	return harness.AgentRunRequest{ID: request.ID, Purpose: string(request.Action), ProjectDir: a.ProjectDir, Prompt: string(trace.RequestBody), OutputName: "output.json"}
}

func validateParsedQuizModelResponse(response QuizModelResponse, parseErr error) error {
	if parseErr != nil {
		return parseErr
	}

	return ValidateQuizModelResponse(response)
}
