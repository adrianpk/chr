package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
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
	ActionAdapters map[quizmodel.Action]quizmodel.Adapter
	DefaultAdapter quizmodel.Adapter
}

// Invoke invokes the configured model adapter.
func (a ActionRoutingQuizModelAdapter) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	if adapter, ok := a.ActionAdapters[request.Action]; ok {
		if adapter == nil {
			return quizmodel.Response{}, fmt.Errorf("%s quiz model adapter is required", request.Action)
		}

		return adapter.Invoke(ctx, request)
	}

	if a.DefaultAdapter == nil {
		return quizmodel.Response{}, fmt.Errorf("default quiz model adapter is required")
	}

	return a.DefaultAdapter.Invoke(ctx, request)
}

// Invoke invokes the configured model adapter.
func (a HarnessQuizModelAdapter) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	request.ID = harnesshost.HarnessGenerationRequestID(ctx, request.ID)
	if err := quizmodel.ValidateRequest(request); err != nil {
		return quizmodel.Response{}, err
	}

	schema, err := harnessQuizModelOutputSchema(request)
	if err != nil {
		return quizmodel.Response{}, err
	}

	prompt, err := harnessQuizModelPrompt(request)
	if err != nil {
		return quizmodel.Response{}, err
	}

	run, err := harnesshost.RunAgent(ctx, a.Runner, harness.AgentRunRequest{
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
		return quizmodel.Response{}, err
	}

	response, err := quizmodel.ResponseFromOutput(request, run.Output, run.Provider)
	err = quizmodel.ValidateParsedResponse(response, err)

	err = harnesshost.ValidateHarnessGenerationOutput(ctx, a.agentRunRequest(request), err)
	if err != nil {
		harnesshost.RecordAgentRunResponseEvent(ctx, a.agentRunRequest(request), run, "parse_error", err)

		return quizmodel.Response{}, err
	}

	harnesshost.RecordAgentRunResponseEvent(ctx, a.agentRunRequest(request), run, "ready", nil)

	return response, nil
}

func (a HarnessQuizModelAdapter) agentRunRequest(request quizmodel.Request) harness.AgentRunRequest {
	return harness.AgentRunRequest{
		ID:         request.ID,
		Purpose:    string(request.Action),
		ProjectDir: a.ProjectDir,
	}
}

func harnessQuizModelOutputSchema(request quizmodel.Request) ([]byte, error) {
	schema := quizmodel.OutputSchema(request).Schema

	content, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal harness output schema: %w", err)
	}

	return content, nil
}

func harnessQuizModelPrompt(request quizmodel.Request) (string, error) {
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

func workspaceDiscoveryInstructions(action quizmodel.Action) string {
	switch action {
	case quizmodel.ActionGenerateSpec:
		return decisionDocumentDiscoveryInstructions("the generated spec")
	case quizmodel.ActionGenerateAcceptanceFeature:
		return decisionDocumentDiscoveryInstructions("the generated acceptance feature")
	case quizmodel.ActionGeneratePlan,
		quizmodel.ActionGeneratePlanOutline,
		quizmodel.ActionGeneratePlanSliceDetail:
		return decisionDocumentDiscoveryInstructions("the generated plan")
	case quizmodel.ActionMaterializePlanTracking:
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
