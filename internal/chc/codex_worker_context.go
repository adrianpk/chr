package chc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CodexWorkerContextMaterializer interface {
	MaterializeContext(ctx context.Context, opts MaterializeCodexWorkerContextOptions) error
}

type MaterializeCodexWorkerContextOptions struct {
	Path           string
	ExecutionID    string
	TicketKey      string
	TicketPath     string
	ResultPath     string
	LogPath        string
	ValidationPath string
}

type codexWorkerContextMaterializer struct{}

func NewCodexWorkerContextMaterializer() CodexWorkerContextMaterializer {
	return &codexWorkerContextMaterializer{}
}

func MaterializeCodexWorkerContextForExecution(
	ctx context.Context,
	record ExecutionRecord,
	plan RuntimePlan,
) error {
	return NewCodexWorkerContextMaterializer().MaterializeContext(ctx, MaterializeCodexWorkerContextOptions{
		Path:           plan.WorkerContextPath,
		ExecutionID:    record.ID,
		TicketKey:      record.TicketKey,
		TicketPath:     containerProjectPath(plan, record.TicketPath),
		ResultPath:     containerProjectPath(plan, plan.ResultPath),
		LogPath:        containerProjectPath(plan, plan.LogPath),
		ValidationPath: containerProjectPath(plan, plan.ValidationPath),
	})
}

func (m *codexWorkerContextMaterializer) MaterializeContext(
	ctx context.Context,
	opts MaterializeCodexWorkerContextOptions,
) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	err = validateCodexWorkerContextOptions(opts)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(opts.Path), 0o755)
	if err != nil {
		return fmt.Errorf("create codex worker context directory: %w", err)
	}

	content := RenderCodexWorkerContext(opts)

	err = os.WriteFile(opts.Path, []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("write codex worker context: %w", err)
	}

	return nil
}

func RenderCodexWorkerContext(opts MaterializeCodexWorkerContextOptions) string {
	var output strings.Builder

	output.WriteString("# Codex Worker Context\n\n")
	output.WriteString("## Ticket\n\n")
	output.WriteString("- execution_id: ")
	output.WriteString(opts.ExecutionID)
	output.WriteString("\n")
	output.WriteString("- ticket_key: ")
	output.WriteString(opts.TicketKey)
	output.WriteString("\n")
	output.WriteString("- ticket_path: ")
	output.WriteString(opts.TicketPath)
	output.WriteString("\n\n")
	output.WriteString("## Artifacts\n\n")
	output.WriteString("- result_path: ")
	output.WriteString(opts.ResultPath)
	output.WriteString("\n")
	output.WriteString("- log_path: ")
	output.WriteString(opts.LogPath)
	output.WriteString("\n")
	output.WriteString("- validation_path: ")
	output.WriteString(opts.ValidationPath)
	output.WriteString("\n\n")
	output.WriteString("## Result Contract\n\n")
	output.WriteString("Write a JSON object to `result_path` with these fields:\n\n")
	output.WriteString("- `execution_id`: must match the execution ID above.\n")
	output.WriteString("- `ticket_key`: must match the ticket key above.\n")
	output.WriteString("- `status`: `succeeded` or `failed`.\n")
	output.WriteString("- `summary`: concise human-readable result summary.\n")
	output.WriteString("- `changed_files`: optional list of workspace-relative changed file summaries.\n")

	return output.String()
}

func validateCodexWorkerContextOptions(opts MaterializeCodexWorkerContextOptions) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "context path", value: opts.Path},
		{name: "execution id", value: opts.ExecutionID},
		{name: "ticket key", value: opts.TicketKey},
		{name: "ticket path", value: opts.TicketPath},
		{name: "result path", value: opts.ResultPath},
		{name: "log path", value: opts.LogPath},
		{name: "validation path", value: opts.ValidationPath},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("codex worker %s is required", field.name)
		}
	}

	return nil
}
