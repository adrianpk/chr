package chc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CodexWorkerContextMaterializer defines a package boundary interface.
type CodexWorkerContextMaterializer interface {
	MaterializeContext(ctx context.Context, opts MaterializeCodexWorkerContextOptions) error
}

// MaterializeCodexWorkerContextOptions describes a package boundary value.
type MaterializeCodexWorkerContextOptions struct {
	Path           string
	ExecutionID    string
	TicketKey      string
	TicketPath     string
	ResultPath     string
	LogPath        string
	ValidationPath string
	RepairHandoff  *CodexWorkerRepairHandoff
}

// CodexWorkerRepairHandoff describes a package boundary value.
type CodexWorkerRepairHandoff struct {
	Branch            string
	PullRequestURL    string
	ReviewCommentsURL string
	ReviewState       string
	ReviewResult      string
	ReviewRound       int
}

type codexWorkerContextMaterializer struct{}

// NewCodexWorkerContextMaterializer constructs the corresponding package service or value.
func NewCodexWorkerContextMaterializer() CodexWorkerContextMaterializer {
	return &codexWorkerContextMaterializer{}
}

// MaterializeCodexWorkerContextForExecution executes the corresponding package operation.
func MaterializeCodexWorkerContextForExecution(
	ctx context.Context,
	record ExecutionRecord,
	plan RuntimePlan,
) error {
	repairHandoff, err := codexWorkerRepairHandoff(ctx, plan.ProjectDir, record.TicketPath)
	if err != nil {
		return err
	}

	return NewCodexWorkerContextMaterializer().MaterializeContext(ctx, MaterializeCodexWorkerContextOptions{
		Path:           plan.WorkerContextPath,
		ExecutionID:    record.ID,
		TicketKey:      record.TicketKey,
		TicketPath:     containerProjectPath(plan, record.TicketPath),
		ResultPath:     containerProjectPath(plan, plan.ResultPath),
		LogPath:        containerProjectPath(plan, plan.LogPath),
		ValidationPath: containerProjectPath(plan, plan.ValidationPath),
		RepairHandoff:  repairHandoff,
	})
}

// MaterializeContext executes the corresponding package operation.
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

// RenderCodexWorkerContext renders the corresponding package data.
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
	writeCodexWorkerRepairHandoff(&output, opts)
	output.WriteString("## Result Contract\n\n")
	output.WriteString("Write a JSON object to `result_path` with these fields:\n\n")
	output.WriteString("- `execution_id`: must match the execution ID above.\n")
	output.WriteString("- `ticket_key`: must match the ticket key above.\n")
	output.WriteString("- `status`: `succeeded` or `failed`.\n")
	output.WriteString("- `summary`: concise human-readable result summary.\n")
	output.WriteString("- `changed_files`: optional list of workspace-relative changed file summaries.\n")

	return output.String()
}

func writeCodexWorkerRepairHandoff(output *strings.Builder, opts MaterializeCodexWorkerContextOptions) {
	if opts.RepairHandoff == nil {
		return
	}

	handoff := opts.RepairHandoff

	output.WriteString("## Review Repair Handoff\n\n")
	output.WriteString("- branch: ")
	output.WriteString(handoff.Branch)
	output.WriteString("\n")
	output.WriteString("- pull_request: ")
	output.WriteString(handoff.PullRequestURL)
	output.WriteString("\n")
	output.WriteString("- review_comments_url: ")
	output.WriteString(handoff.ReviewCommentsURL)
	output.WriteString("\n")
	output.WriteString("- review_state: ")
	output.WriteString(handoff.ReviewState)
	output.WriteString("\n")
	output.WriteString("- review_result: ")
	output.WriteString(handoff.ReviewResult)
	output.WriteString("\n")
	output.WriteString("- review_round: ")
	output.WriteString(strconv.Itoa(handoff.ReviewRound))
	output.WriteString("\n\n")
	output.WriteString("Use the remote review comments as the source of detailed findings. Do not duplicate remote comments into the ticket.\n\n")
}

func codexWorkerRepairHandoff(ctx context.Context, projectDir, ticketPath string) (*CodexWorkerRepairHandoff, error) {
	ticket, err := NewTicketParser().ParseTicket(ctx, ParseTicketOptions{
		ProjectDir: projectDir,
		Path:       ticketPath,
		IndexedAt:  time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("read repair handoff ticket metadata: %w", err)
	}

	if !isReviewRepairTicket(ticket) {
		return nil, nil
	}

	ref, err := ticketReviewCommandPullRequestRef(ticket)
	if err != nil {
		return nil, fmt.Errorf("repair handoff requires pull request ref: %w", err)
	}

	if strings.TrimSpace(ticket.Branch) == "" {
		return nil, fmt.Errorf("repair handoff requires branch metadata")
	}

	if strings.TrimSpace(ticket.ReviewCommentsURL) == "" {
		return nil, fmt.Errorf("repair handoff requires review_comments_url")
	}

	return &CodexWorkerRepairHandoff{
		Branch:            ticket.Branch,
		PullRequestURL:    ref.URL,
		ReviewCommentsURL: ticket.ReviewCommentsURL,
		ReviewState:       ticket.ReviewState,
		ReviewResult:      ticket.ReviewResult,
		ReviewRound:       ticket.ReviewRound,
	}, nil
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
