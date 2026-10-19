package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness"
)

type harnessDumpCurationProvider struct {
	ProjectDir string
	Command    string
	Model      string
	Timeout    time.Duration
	Runner     harness.AgentRunner
}

func (p harnessDumpCurationProvider) CurateDump(ctx context.Context, request DumpCurationRequest) (DumpCurationResult, error) {
	runRequest, err := p.agentRunRequest(request)
	if err != nil {
		return DumpCurationResult{}, err
	}

	run, err := core.RunAgent(ctx, p.Runner, runRequest)
	if err != nil {
		return DumpCurationResult{}, err
	}

	curation, err := decodeDumpCurationPayload(run.Output)
	if err != nil {
		core.RecordAgentRunResponseEvent(ctx, runRequest, run, "parse_error", err)

		return DumpCurationResult{}, err
	}

	core.RecordAgentRunResponseEvent(ctx, runRequest, run, "ready", nil)

	return curation, nil
}

func (p harnessDumpCurationProvider) agentRunRequest(request DumpCurationRequest) (harness.AgentRunRequest, error) {
	prompt, err := harnessDumpCurationPrompt(request)
	if err != nil {
		return harness.AgentRunRequest{}, err
	}

	schema, err := json.MarshalIndent(dumpCurationJSONSchema(), "", "  ")
	if err != nil {
		return harness.AgentRunRequest{}, fmt.Errorf("marshal harness dump curation schema: %w", err)
	}

	projectDir := strings.TrimSpace(p.ProjectDir)
	if projectDir == "" {
		projectDir = request.ProjectDir
	}

	return harness.AgentRunRequest{
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

func harnessDumpCurationPrompt(request DumpCurationRequest) (string, error) {
	input, err := json.MarshalIndent(openAIDumpCurationProvider{}.openAIInput(request), "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal harness dump curation input: %w", err)
	}

	var prompt strings.Builder
	prompt.WriteString("You are the Chryso dump curation backend.\n")
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
