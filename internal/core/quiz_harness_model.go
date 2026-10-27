package core

import quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
)

// HarnessQuizModelAdapter describes a package boundary value.
type HarnessQuizModelAdapter struct {
	ProjectDir string
	Command    string
	Model      string
	Timeout    time.Duration
	Runner     harness.AgentRunner
}

// ActionRoutingQuizModelAdapter describes a package boundary value.
type ActionRoutingQuizModelAdapter struct {
	ActionAdapters map[quizmodel.QuizModelAction]quizmodel.QuizModelAdapter
	DefaultAdapter quizmodel.QuizModelAdapter
}

// InvokeQuizModel invokes the configured model adapter.
func (a ActionRoutingQuizModelAdapter) InvokeQuizModel(ctx context.Context, request quizmodel.QuizModelRequest) (quizmodel.QuizModelResponse, error) {
	if adapter, ok := a.ActionAdapters[request.Action]; ok {
		if adapter == nil {
			return quizmodel.QuizModelResponse{}, fmt.Errorf("%s quiz model adapter is required", request.Action)
		}

		return adapter.InvokeQuizModel(ctx, request)
	}

	if a.DefaultAdapter == nil {
		return quizmodel.QuizModelResponse{}, fmt.Errorf("default quiz model adapter is required")
	}

	return a.DefaultAdapter.InvokeQuizModel(ctx, request)
}

// InvokeQuizModel invokes the configured model adapter.
func (a HarnessQuizModelAdapter) InvokeQuizModel(ctx context.Context, request quizmodel.QuizModelRequest) (quizmodel.QuizModelResponse, error) {
	request.ID = HarnessGenerationRequestID(ctx, request.ID)
	if err := quizmodel.ValidateQuizModelRequest(request); err != nil {
		return quizmodel.QuizModelResponse{}, err
	}

	schema, err := harnessQuizModelOutputSchema(request)
	if err != nil {
		return quizmodel.QuizModelResponse{}, err
	}

	prompt, err := harnessQuizModelPrompt(request)
	if err != nil {
		return quizmodel.QuizModelResponse{}, err
	}

	run, err := RunAgent(ctx, a.Runner, harness.AgentRunRequest{
		ID:                 request.ID,
		Purpose:            string(request.Action),
		ProjectDir:         a.ProjectDir,
		Prompt:             prompt,
		OutputSchema:       schema,
		OutputName:         "output.json",
		Program:            a.Command,
		Model:              a.Model,
		Timeout:            a.Timeout,
		ReadOnlyFiles:      workspaceDiscoveryInstructions(request.Action) != "",
		DeferResponseTrace: true,
	})
	if err != nil {
		return quizmodel.QuizModelResponse{}, err
	}

	response, err := quizmodel.ResponseFromOutput(request, run.Output)
	response.Provider = run.Provider
	err = quizmodel.ValidateParsedResponse(response, err)

	err = ValidateHarnessGenerationOutput(ctx, a.agentRunRequest(request), err)
	if err != nil {
		recordAgentRunResponseEvent(ctx, a.agentRunRequest(request), run, "parse_error", err)

		return quizmodel.QuizModelResponse{}, err
	}

	recordAgentRunResponseEvent(ctx, a.agentRunRequest(request), run, "ready", nil)

	return response, nil
}

func (a HarnessQuizModelAdapter) agentRunRequest(request quizmodel.QuizModelRequest) harness.AgentRunRequest {
	return harness.AgentRunRequest{
		ID:         request.ID,
		Purpose:    string(request.Action),
		ProjectDir: a.ProjectDir,
	}
}

func harnessQuizModelOutputSchema(request quizmodel.QuizModelRequest) ([]byte, error) {
	schema := quizmodel.OutputFormat(request).Schema

	content, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal harness output schema: %w", err)
	}

	return content, nil
}

func harnessQuizModelPrompt(request quizmodel.QuizModelRequest) (string, error) {
	payload, err := json.MarshalIndent(quizmodel.ProviderPayload(request), "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal harness quiz payload: %w", err)
	}

	var prompt strings.Builder
	prompt.WriteString("You are the chr ")
	prompt.WriteString(string(request.Action))
	prompt.WriteString(" backend.\n")
	prompt.WriteString("Generate the requested artifact and return only the final JSON object matching the provided output schema.\n")
	prompt.WriteString("You may inspect workspace files with read-only operations when the instructions require additional project context.\n")
	prompt.WriteString("Do not edit files, create commits, open pull requests, or ask follow-up questions.\n\n")
	prompt.WriteString("Instructions:\n")
	prompt.WriteString(quizmodel.Instructions(request))

	if discovery := workspaceDiscoveryInstructions(request.Action); discovery != "" {
		prompt.WriteString("\n\nWorkspace discovery:\n")
		prompt.WriteString(discovery)
	}

	prompt.WriteString("\n\nRequest payload:\n```json\n")
	prompt.Write(payload)
	prompt.WriteString("\n```\n")

	return prompt.String(), nil
}

func workspaceDiscoveryInstructions(action quizmodel.QuizModelAction) string {
	switch action {
	case quizmodel.QuizModelActionGenerateSpec:
		return decisionDocumentDiscoveryInstructions("the generated spec")
	case quizmodel.QuizModelActionGenerateAcceptanceFeature:
		return decisionDocumentDiscoveryInstructions("the generated acceptance feature")
	case quizmodel.QuizModelActionGeneratePlan,
		quizmodel.QuizModelActionGeneratePlanOutline,
		quizmodel.QuizModelActionGeneratePlanSliceDetail:
		return decisionDocumentDiscoveryInstructions("the generated plan")
	case quizmodel.QuizModelActionMaterializePlanTracking:
		return decisionDocumentDiscoveryInstructions("the generated tracking and tickets")
	default:
		return ""
	}
}

func decisionDocumentDiscoveryInstructions(target string) string {
	return strings.TrimSpace(`
Before generating the artifact, search the workspace for architecture decision records or equivalent decision documents.
Look for ADRs, decision records, architecture decisions, RFCs, engineering principles, decision logs, or similarly named files and directories.
Use accepted or active decisions as high-authority constraints for ` + target + `.
Incorporate relevant decisions into requirements, constraints, non-goals, acceptance criteria, slice definitions, ticket metadata, and validation expectations when applicable.
If a decision conflicts with dumps, quizzes, existing docs, or inferred requirements, prefer the decision and explicitly call out the conflict.
Do not require configured decision-document paths; discover them from the workspace structure.
`)
}
