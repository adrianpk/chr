package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/harness"
)

// HarnessAuthoringQuizQuestionProvider describes a package boundary value.
type HarnessAuthoringQuizQuestionProvider struct {
	Program string
	Runner  harness.AgentRunner
}

// GenerateAuthoringQuizQuestions generates the requested artifact.
func (p HarnessAuthoringQuizQuestionProvider) GenerateAuthoringQuizQuestions(ctx context.Context, request authoringcore.QuizGenerationRequest) (authoringcore.QuizGenerationResult, error) {
	err := authoringcore.ValidateQuizGenerationRequest(request)
	if err != nil {
		return authoringcore.QuizGenerationResult{}, err
	}

	program := strings.TrimSpace(p.Program)
	if _, _, bound := HarnessGenerationSelection(ctx); bound {
		program = ""
	}

	payload, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return authoringcore.QuizGenerationResult{}, fmt.Errorf("marshal authoring quiz request: %w", err)
	}

	prompt := authoringQuizHarnessPrompt(string(payload))

	run, err := RunAgent(ctx, p.Runner, harness.AgentRunRequest{
		ID:                 request.ID,
		Purpose:            request.Action,
		ProjectDir:         request.ProjectDir,
		Prompt:             prompt,
		OutputName:         "output.json",
		Program:            program,
		ReadOnlyFiles:      true,
		DeferResponseTrace: true,
	})
	if err != nil {
		return authoringcore.QuizGenerationResult{}, fmt.Errorf("invoke authoring quiz provider: %w", err)
	}

	result, err := authoringcore.ParseQuizGenerationResult(run.Output)
	result.Provider = run.Provider
	result.Strategy = authoringcore.QuizGenerationStrategyDeep

	result.RawText = string(run.Output)
	if err == nil {
		err = authoringcore.ValidateQuizGenerationResult(result)
	}

	err = ValidateHarnessGenerationOutput(ctx, harness.AgentRunRequest{ID: request.ID, Purpose: request.Action, ProjectDir: request.ProjectDir}, err)
	if err != nil {
		recordAgentRunResponseEvent(ctx, harness.AgentRunRequest{
			ID:         request.ID,
			Purpose:    request.Action,
			ProjectDir: request.ProjectDir,
		}, run, "parse_error", err)

		return authoringcore.QuizGenerationResult{}, err
	}

	recordAgentRunResponseEvent(ctx, harness.AgentRunRequest{
		ID:         request.ID,
		Purpose:    request.Action,
		ProjectDir: request.ProjectDir,
	}, run, "ready", nil)

	return result, nil
}

func authoringQuizHarnessPrompt(payload string) string {
	return strings.TrimSpace(`
You generate context-aware authoring quiz questions for chr.

Use the provided structured request and context sections.
Ask concrete, project-specific questions that help produce the requested target artifact.
Do not ask generic questionnaire questions.
For spec targets, cover problem, actors, workflows, acceptance criteria,
constraints, non-goals, validation evidence, impacted code/docs, and open risks
when the project context supports those areas.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return strict JSON only. Do not include Markdown or prose outside JSON.

The JSON response must have exactly one top-level key: "questions".
Each question object must contain:
- id
- question
- tags
- follow_up_depth
- completion_recommended

Request:
`) + "\n\n" + payload + "\n"
}
