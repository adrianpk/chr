package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/workspace"
)

type codexDumpCurationProvider struct {
	ProjectDir string
	Command    string
	Model      string
	Timeout    time.Duration
	Runner     workspace.AgentRunner
}

func (p codexDumpCurationProvider) CurateDump(
	ctx context.Context,
	request DumpCurationRequest,
) (DumpCurationResult, error) {
	runRequest, err := p.agentRunRequest(request)
	if err != nil {
		return DumpCurationResult{}, err
	}

	run, err := workspace.RunCodexAgent(ctx, p.agentRunner(), runRequest)
	if err != nil {
		return DumpCurationResult{}, err
	}

	curation, err := decodeDumpCurationPayload(run.Output)
	if err != nil {
		workspace.RecordAgentRunResponseEvent(ctx, runRequest, run, "parse_error", err)

		return DumpCurationResult{}, err
	}

	workspace.RecordAgentRunResponseEvent(ctx, runRequest, run, "ready", nil)

	return curation, nil
}

func (p codexDumpCurationProvider) agentRunRequest(request DumpCurationRequest) (workspace.AgentRunRequest, error) {
	prompt, err := codexDumpCurationPrompt(request)
	if err != nil {
		return workspace.AgentRunRequest{}, err
	}

	schema, err := json.MarshalIndent(dumpCurationJSONSchema(), "", "  ")
	if err != nil {
		return workspace.AgentRunRequest{}, fmt.Errorf("marshal Codex dump curation schema: %w", err)
	}

	projectDir := strings.TrimSpace(p.ProjectDir)
	if projectDir == "" {
		projectDir = request.ProjectDir
	}

	return workspace.AgentRunRequest{
		ID:                 dumpCurationRequestID(request),
		Purpose:            "dump_curation",
		ProjectDir:         projectDir,
		Prompt:             prompt,
		OutputSchema:       schema,
		OutputName:         "output.json",
		Program:            p.Command,
		Model:              p.Model,
		Timeout:            p.Timeout,
		DeferResponseTrace: true,
	}, nil
}

func (p codexDumpCurationProvider) agentRunner() workspace.AgentRunner {
	if p.Runner != nil {
		return p.Runner
	}

	return workspace.CodexAgentRunner{}
}

func codexDumpCurationPrompt(request DumpCurationRequest) (string, error) {
	input, err := json.MarshalIndent(openAIDumpCurationProvider{}.openAIInput(request), "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal Codex dump curation input: %w", err)
	}

	var prompt strings.Builder
	prompt.WriteString("You are the ChrysoCoord dump curation backend.\n")
	prompt.WriteString("Curate the requested dump and return only the final JSON object matching the provided output schema.\n")
	prompt.WriteString("Do not edit files, run commands, create commits, open pull requests, or ask follow-up questions.\n\n")
	prompt.WriteString("Instructions:\n")
	prompt.WriteString(dumpCurationInstructions())
	prompt.WriteString("\n\nRequest payload:\n```json\n")
	prompt.Write(input)
	prompt.WriteString("\n```\n")

	return prompt.String(), nil
}

func dumpCurationRequestID(request DumpCurationRequest) string {
	return "dump-curation-" + dumpSlug(firstNonEmptyDumpCurationString(request.Title, "dump"))
}
