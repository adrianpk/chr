package openai

import (
	"context"
	"encoding/json"

	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	"forge.adrianpk.com/chr/internal/harness"
)

func (a Adapter) invokeBound(ctx context.Context, request quizmodel.Request, trace callTrace) (invocationResult, error) {
	var result executionResult

	schema, err := json.Marshal(quizmodel.OutputSchema(request).Schema)
	if err != nil {
		return invocationResult{Execution: result}, err
	}

	call := a.runRequest(request, trace)
	call.OutputSchema = schema
	output, err := a.Host.Run(ctx, call, func(ctx context.Context) ([]byte, error) {
		call, err := a.invoke(ctx, request, trace)
		result = call.Execution

		return call.Output, err
	})

	return invocationResult{Output: output, Execution: result}, err
}

func (a Adapter) runRequest(request quizmodel.Request, trace callTrace) harness.AgentRunRequest {
	return harness.AgentRunRequest{ID: request.ID, Purpose: string(request.Action), ProjectDir: a.ProjectDir, Prompt: string(trace.RequestBody), OutputName: "output.json"}
}

type invocationResult struct {
	Output    []byte
	Execution executionResult
}
