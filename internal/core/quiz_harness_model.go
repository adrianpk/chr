package core

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
	ActionAdapters map[QuizModelAction]QuizModelAdapter
	DefaultAdapter QuizModelAdapter
}

// InvokeQuizModel invokes the configured model adapter.
func (a ActionRoutingQuizModelAdapter) InvokeQuizModel(ctx context.Context, request QuizModelRequest) (QuizModelResponse, error) {
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

// InvokeQuizModel invokes the configured model adapter.
func (a HarnessQuizModelAdapter) InvokeQuizModel(ctx context.Context, request QuizModelRequest) (QuizModelResponse, error) {
	if err := ValidateQuizModelRequest(request); err != nil {
		return QuizModelResponse{}, err
	}

	schema, err := harnessQuizModelOutputSchema(request.Action)
	if err != nil {
		return QuizModelResponse{}, err
	}

	prompt, err := harnessQuizModelPrompt(request)
	if err != nil {
		return QuizModelResponse{}, err
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

	response.Provider = run.Provider

	recordAgentRunResponseEvent(ctx, a.agentRunRequest(request), run, "ready", nil)

	return response, nil
}

func (a HarnessQuizModelAdapter) agentRunRequest(request QuizModelRequest) harness.AgentRunRequest {
	return harness.AgentRunRequest{
		ID:         request.ID,
		Purpose:    string(request.Action),
		ProjectDir: a.ProjectDir,
	}
}

func harnessQuizModelOutputSchema(action QuizModelAction) ([]byte, error) {
	var schema map[string]any

	switch action {
	case QuizModelActionGenerateSpec:
		schema = quizModelDraftSpecJSONSchema()
	case QuizModelActionGenerateAcceptanceFeature:
		schema = quizModelFeatureJSONSchema()
	case QuizModelActionGeneratePlan:
		schema = quizModelPlanJSONSchema()
	case QuizModelActionMaterializePlanTracking:
		schema = quizModelPlanTrackingJSONSchema()
	default:
		return nil, fmt.Errorf("harness quiz model adapter does not support action %s", action)
	}

	content, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal harness output schema: %w", err)
	}

	return content, nil
}

func harnessQuizModelPrompt(request QuizModelRequest) (string, error) {
	payload, err := json.MarshalIndent(quizModelProviderPayload(request), "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal harness quiz payload: %w", err)
	}

	var prompt strings.Builder
	prompt.WriteString("You are the Chryso ")
	prompt.WriteString(string(request.Action))
	prompt.WriteString(" backend.\n")
	prompt.WriteString("Generate the requested artifact and return only the final JSON object matching the provided output schema.\n")
	prompt.WriteString("You may inspect workspace files with read-only operations when the instructions require additional project context.\n")
	prompt.WriteString("Do not edit files, create commits, open pull requests, or ask follow-up questions.\n\n")
	prompt.WriteString("Instructions:\n")
	prompt.WriteString(quizModelInstructions(request))

	if discovery := workspaceDiscoveryInstructions(request.Action); discovery != "" {
		prompt.WriteString("\n\nWorkspace discovery:\n")
		prompt.WriteString(discovery)
	}

	prompt.WriteString("\n\nRequest payload:\n```json\n")
	prompt.Write(payload)
	prompt.WriteString("\n```\n")

	return prompt.String(), nil
}

func workspaceDiscoveryInstructions(action QuizModelAction) string {
	switch action {
	case QuizModelActionGenerateSpec:
		return decisionDocumentDiscoveryInstructions("the generated spec")
	case QuizModelActionGenerateAcceptanceFeature:
		return decisionDocumentDiscoveryInstructions("the generated acceptance feature")
	case QuizModelActionGeneratePlan,
		QuizModelActionGeneratePlanOutline,
		QuizModelActionGeneratePlanSliceDetail:
		return decisionDocumentDiscoveryInstructions("the generated plan")
	case QuizModelActionMaterializePlanTracking:
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
