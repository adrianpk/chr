package chc

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type TicketResultBlockedReasonWriter interface {
	WriteBlockedReason(ctx context.Context, opts WriteBlockedReasonOptions) error
}

type WriteBlockedReasonOptions struct {
	Application ResultApplication
	CreatedAt   time.Time
}

type ticketResultBlockedReasonWriter struct{}

func NewTicketResultBlockedReasonWriter() TicketResultBlockedReasonWriter {
	return &ticketResultBlockedReasonWriter{}
}

func (w *ticketResultBlockedReasonWriter) WriteBlockedReason(
	ctx context.Context,
	opts WriteBlockedReasonOptions,
) error {
	if opts.Application.TicketStatus != TicketResultStatusBlocked {
		return nil
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if opts.Application.TicketPath == "" {
		return fmt.Errorf("result application ticket path is required")
	}

	content, err := os.ReadFile(opts.Application.TicketPath)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	createdAt := opts.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	updated, err := addResultBlockedReason(string(content), opts.Application, createdAt)
	if err != nil {
		return err
	}

	if err := os.WriteFile(opts.Application.TicketPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}

func addResultBlockedReason(content string, application ResultApplication, createdAt time.Time) (string, error) {
	updatedContent, err := updateTicketUpdatedAt(content, createdAt)
	if err != nil {
		return "", err
	}

	blockerHeading := "# Blocker\n"

	blockerIndex := strings.Index(updatedContent, blockerHeading)
	if blockerIndex < 0 {
		return "", fmt.Errorf("ticket blocker section is required")
	}

	insertAt := blockerIndex + len(blockerHeading)
	reason := renderResultBlockedReason(application, createdAt)

	return updatedContent[:insertAt] + "\n" + reason + updatedContent[insertAt:], nil
}

func renderResultBlockedReason(application ResultApplication, createdAt time.Time) string {
	blockedReason := application.BlockedReason

	var output strings.Builder
	output.WriteString("## ")
	output.WriteString(createdAt.UTC().Format(time.RFC3339))
	output.WriteString(" ")
	output.WriteString(application.ExecutionID)
	output.WriteString("\n\n")
	output.WriteString("- kind: ")
	output.WriteString(string(blockedReason.Kind))
	output.WriteString("\n")
	output.WriteString("- reason: ")
	output.WriteString(blockedReason.Reason)
	output.WriteString("\n")
	output.WriteString("- result: ")
	output.WriteString(blockedReason.ResultPath)
	output.WriteString("\n")
	output.WriteString("- log: ")
	output.WriteString(blockedReason.LogPath)
	output.WriteString("\n")
	output.WriteString("- validation: ")
	output.WriteString(blockedReason.ValidationPath)
	output.WriteString("\n\n")

	return output.String()
}
