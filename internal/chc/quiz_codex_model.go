package chc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const QuizModelProviderCodex = AgentProviderCodex

type CodexQuizModelAdapter struct {
	ProjectDir string
	Command    string
	Model      string
	Timeout    time.Duration
	Runner     AgentRunner
}

type ActionRoutingQuizModelAdapter struct {
	ActionAdapters map[QuizModelAction]QuizModelAdapter
	DefaultAdapter QuizModelAdapter
}

func (a ActionRoutingQuizModelAdapter) InvokeQuizModel(
	ctx context.Context,
	request QuizModelRequest,
) (QuizModelResponse, error) {
	if adapter, ok := a.ActionAdapters[request.Action]; ok {
		if adapter == nil {
			return QuizModelResponse{}, fmt.Errorf("%s quiz model adapter is required", request.Action)
		}

		return adapter.InvokeQuizModel(ctx, request)
	}

	if a.DefaultAdapter == nil {
		return QuizModelResponse{}, fmt.Errorf("default quiz model adapter is required")
	}

	return a.DefaultAdapter.InvokeQuizModel(ctx, request)
}

func (a CodexQuizModelAdapter) InvokeQuizModel(
	ctx context.Context,
	request QuizModelRequest,
) (QuizModelResponse, error) {
	if err := ValidateQuizModelRequest(request); err != nil {
		return QuizModelResponse{}, err
	}

	schema, err := codexQuizModelOutputSchema(request.Action)
	if err != nil {
		return QuizModelResponse{}, err
	}

	prompt, err := codexQuizModelPrompt(request)
	if err != nil {
		return QuizModelResponse{}, err
	}

	run, err := runCodexAgent(ctx, a.agentRunner(), AgentRunRequest{
		ID:                 request.ID,
		Purpose:            string(request.Action),
		ProjectDir:         a.ProjectDir,
		Prompt:             prompt,
		OutputSchema:       schema,
		OutputName:         "output.json",
		Program:            a.Command,
		Model:              a.Model,
		Timeout:            a.Timeout,
		DeferResponseTrace: true,
	})
	if err != nil {
		return QuizModelResponse{}, err
	}

	response, err := quizModelResponseFromOutput(request, run.Output)
	if err != nil {
		recordAgentRunResponseEvent(ctx, a.agentRunRequest(request), run, "parse_error", err)

		return QuizModelResponse{}, err
	}

	response.Provider = QuizModelProviderCodex

	recordAgentRunResponseEvent(ctx, a.agentRunRequest(request), run, "ready", nil)

	return response, nil
}

func (a CodexQuizModelAdapter) agentRunner() AgentRunner {
	if a.Runner != nil {
		return a.Runner
	}

	return CodexAgentRunner{}
}

func (a CodexQuizModelAdapter) agentRunRequest(request QuizModelRequest) AgentRunRequest {
	return AgentRunRequest{
		ID:         request.ID,
		Purpose:    string(request.Action),
		ProjectDir: a.ProjectDir,
	}
}

func codexQuizModelOutputSchema(action QuizModelAction) ([]byte, error) {
	var schema map[string]any

	switch action {
	case QuizModelActionGenerateSpec:
		schema = quizModelDraftSpecJSONSchema()
	case QuizModelActionGeneratePlan:
		schema = quizModelPlanJSONSchema()
	default:
		return nil, fmt.Errorf("codex quiz model adapter does not support action %s", action)
	}

	content, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal Codex output schema: %w", err)
	}

	return content, nil
}

func codexQuizModelPrompt(request QuizModelRequest) (string, error) {
	payload, err := json.MarshalIndent(quizModelProviderPayload(request), "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal Codex quiz payload: %w", err)
	}

	var prompt strings.Builder
	prompt.WriteString("You are the ChrysoCoord ")
	prompt.WriteString(string(request.Action))
	prompt.WriteString(" backend.\n")
	prompt.WriteString("Generate the requested artifact and return only the final JSON object matching the provided output schema.\n")
	prompt.WriteString("Do not edit files, run commands, create commits, open pull requests, or ask follow-up questions.\n\n")
	prompt.WriteString("Instructions:\n")
	prompt.WriteString(quizModelInstructions(request))
	prompt.WriteString("\n\nRequest payload:\n```json\n")
	prompt.Write(payload)
	prompt.WriteString("\n```\n")

	return prompt.String(), nil
}
