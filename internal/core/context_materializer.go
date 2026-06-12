package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ContextPackageMaterializer defines a package boundary interface.
type ContextPackageMaterializer interface {
	MaterializeContextPackage(ctx context.Context, opts MaterializeContextPackageOptions) error
}

// MaterializeContextPackageOptions describes a package boundary value.
type MaterializeContextPackageOptions struct {
	Composition       ContextCompositionPlan
	Plan              RuntimePlan
	Agent             string
	Project           string
	Protocols         string
	Documents         string
	CodeContext       string
	PreviousAttempt   string
	AdditionalSources []ContextSourceInventoryEntry
}

type contextPackageMaterializer struct{}

// NewContextPackageMaterializer constructs the corresponding package service or value.
func NewContextPackageMaterializer() ContextPackageMaterializer {
	return &contextPackageMaterializer{}
}

// MaterializeContextPackage executes the corresponding package operation.
func (m *contextPackageMaterializer) MaterializeContextPackage(
	ctx context.Context,
	opts MaterializeContextPackageOptions,
) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	err = validateMaterializeContextPackageOptions(opts)
	if err != nil {
		return err
	}

	err = os.MkdirAll(opts.Composition.Paths.ContextDir, 0o755)
	if err != nil {
		return fmt.Errorf("create context artifact directory: %w", err)
	}

	files := contextPackageFiles(opts)
	for path, content := range files {
		err = os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			return fmt.Errorf("create context package parent directory: %w", err)
		}

		err = os.WriteFile(path, []byte(content), 0o644)
		if err != nil {
			return fmt.Errorf("write context package artifact: %w", err)
		}
	}

	return nil
}

// RenderMainContextPackage renders the corresponding package data.
func RenderMainContextPackage(opts MaterializeContextPackageOptions) string {
	var output strings.Builder

	output.WriteString("# Codex Worker Context\n\n")
	writeMainContextExecution(&output, opts)
	writeMainContextPackageLinks(&output, opts)
	writeMainContextReadOrder(&output)
	writeMainContextArtifacts(&output, opts)
	writeMainContextResultContract(&output)
	writeMainContextIncludedSources(&output, opts)

	return output.String()
}

// RenderContextPackageDryRun renders the corresponding package data.
func RenderContextPackageDryRun(opts MaterializeContextPackageOptions) string {
	var output strings.Builder

	sources := contextPackageSources(opts)

	if len(sources) == 0 {
		output.WriteString("No context sources were recorded.\n")

		return output.String()
	}

	for _, source := range sources {
		output.WriteString("- ")
		output.WriteString(source.ID)
		output.WriteString(" [")
		output.WriteString(string(source.Kind))
		output.WriteString("/")
		output.WriteString(string(source.Class))
		output.WriteString("]")

		if source.Path != "" {
			output.WriteString(" ")
			output.WriteString(source.Path)
		}

		if source.Included {
			output.WriteString(" included")
		} else {
			output.WriteString(" skipped")
		}

		if source.Reason != "" {
			output.WriteString(" (")
			output.WriteString(source.Reason)
			output.WriteString(")")
		}

		output.WriteString("\n")
	}

	return output.String()
}

func writeMainContextExecution(output *strings.Builder, opts MaterializeContextPackageOptions) {
	output.WriteString("## Execution\n\n")
	output.WriteString("- execution_id: ")
	output.WriteString(opts.Composition.ExecutionID)
	output.WriteString("\n")
	output.WriteString("- ticket_key: ")
	output.WriteString(opts.Composition.TicketKey)
	output.WriteString("\n")
	output.WriteString("- ticket_path: ")
	output.WriteString(containerProjectPath(opts.Plan, opts.Composition.TicketPath))
	output.WriteString("\n\n")
}

func writeMainContextPackageLinks(output *strings.Builder, opts MaterializeContextPackageOptions) {
	paths := opts.Composition.Paths

	output.WriteString("## Context Package\n\n")
	writeContextArtifactLink(output, opts.Plan, "agent", paths.AgentPath)
	writeContextArtifactLink(output, opts.Plan, "project", paths.ProjectPath)
	writeContextArtifactLink(output, opts.Plan, "protocols", paths.ProtocolsPath)
	writeContextArtifactLink(output, opts.Plan, "documents", paths.DocumentsPath)
	writeContextArtifactLink(output, opts.Plan, "code_context", paths.CodeContextPath)
	writeContextArtifactLink(output, opts.Plan, "previous_attempt", paths.PreviousAttemptPath)
	writeContextArtifactLink(output, opts.Plan, "source_inventory", paths.SourceInventoryPath)
	output.WriteString("\n")
}

func writeMainContextReadOrder(output *strings.Builder) {
	output.WriteString("## Read Order\n\n")
	output.WriteString("1. Read `agent` and `protocols` first.\n")
	output.WriteString("2. Read `project` and `documents` for ticket and use-case context.\n")
	output.WriteString("3. Read `code_context` for repository orientation.\n")
	output.WriteString("4. Read `previous_attempt` before retrying failed work.\n")
	output.WriteString("5. Use `source_inventory` to inspect included and skipped context sources.\n\n")
}

func writeMainContextArtifacts(output *strings.Builder, opts MaterializeContextPackageOptions) {
	output.WriteString("## Artifacts\n\n")
	writeContextArtifactLink(output, opts.Plan, "result_path", opts.Plan.ResultPath)
	writeContextArtifactLink(output, opts.Plan, "log_path", opts.Plan.LogPath)
	writeContextArtifactLink(output, opts.Plan, "validation_path", opts.Plan.ValidationPath)
	output.WriteString("\n")
}

func writeMainContextResultContract(output *strings.Builder) {
	output.WriteString("## Result Contract\n\n")
	output.WriteString("Write a JSON object to `result_path` with these fields:\n\n")
	output.WriteString("- `execution_id`: must match the execution ID above.\n")
	output.WriteString("- `ticket_key`: must match the ticket key above.\n")
	output.WriteString("- `status`: `succeeded` or `failed`.\n")
	output.WriteString("- `summary`: concise human-readable result summary.\n")
	output.WriteString("- `changed_files`: optional list of workspace-relative changed file summaries.\n\n")
}

func writeMainContextIncludedSources(output *strings.Builder, opts MaterializeContextPackageOptions) {
	output.WriteString("## Included Sources\n\n")
	output.WriteString(RenderContextPackageDryRun(opts))
}

func contextPackageFiles(opts MaterializeContextPackageOptions) map[string]string {
	paths := opts.Composition.Paths

	return map[string]string{
		paths.MainContextPath:     RenderMainContextPackage(opts),
		paths.AgentPath:           opts.Agent,
		paths.ProjectPath:         opts.Project,
		paths.ProtocolsPath:       opts.Protocols,
		paths.DocumentsPath:       opts.Documents,
		paths.CodeContextPath:     opts.CodeContext,
		paths.PreviousAttemptPath: opts.PreviousAttempt,
		paths.SourceInventoryPath: RenderSourceInventory(contextPackageSources(opts)),
	}
}

func contextPackageSources(opts MaterializeContextPackageOptions) []ContextSourceInventoryEntry {
	sources := append([]ContextSourceInventoryEntry(nil), opts.Composition.Sources...)
	sources = append(sources, opts.AdditionalSources...)

	return sources
}

func validateMaterializeContextPackageOptions(opts MaterializeContextPackageOptions) error {
	if opts.Plan.ProjectDir == "" {
		return fmt.Errorf("context package runtime project dir is required")
	}

	required := []struct {
		name  string
		value string
	}{
		{name: "main context path", value: opts.Composition.Paths.MainContextPath},
		{name: "context dir", value: opts.Composition.Paths.ContextDir},
		{name: "agent path", value: opts.Composition.Paths.AgentPath},
		{name: "project path", value: opts.Composition.Paths.ProjectPath},
		{name: "protocols path", value: opts.Composition.Paths.ProtocolsPath},
		{name: "documents path", value: opts.Composition.Paths.DocumentsPath},
		{name: "code context path", value: opts.Composition.Paths.CodeContextPath},
		{name: "previous attempt path", value: opts.Composition.Paths.PreviousAttemptPath},
		{name: "source inventory path", value: opts.Composition.Paths.SourceInventoryPath},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("context package %s is required", field.name)
		}
	}

	return nil
}

func writeContextArtifactLink(output *strings.Builder, plan RuntimePlan, name, hostPath string) {
	output.WriteString("- ")
	output.WriteString(name)
	output.WriteString(": `")
	output.WriteString(containerProjectPath(plan, hostPath))
	output.WriteString("`")
	output.WriteString(" (host: `")
	output.WriteString(hostPath)
	output.WriteString("`)\n")
}
