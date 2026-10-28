package core

import (
	"context"
	"fmt"
	"time"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// TicketEvidenceStrategyChecklistWriter writes evidence strategy blocks to tickets.
type TicketEvidenceStrategyChecklistWriter interface {
	UpsertEvidenceStrategyChecklist(ctx context.Context, opts UpsertTicketEvidenceStrategyChecklistOptions) error
}

// UpsertTicketEvidenceStrategyChecklistOptions describes a ticket checklist write.
type UpsertTicketEvidenceStrategyChecklistOptions struct {
	Path      string
	Checklist EvidenceStrategyChecklist
	UpdatedAt time.Time
}

type ticketEvidenceStrategyChecklistWriter struct{}

// NewTicketEvidenceStrategyChecklistWriter constructs the ticket evidence checklist writer.
func NewTicketEvidenceStrategyChecklistWriter() TicketEvidenceStrategyChecklistWriter {
	return &ticketEvidenceStrategyChecklistWriter{}
}

// UpsertEvidenceStrategyChecklist writes or updates the ticket evidence strategy block.
func (w *ticketEvidenceStrategyChecklistWriter) UpsertEvidenceStrategyChecklist(ctx context.Context, opts UpsertTicketEvidenceStrategyChecklistOptions) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	if opts.Path == "" {
		return fmt.Errorf("ticket path is required")
	}

	content, err := readLegacyTicketFile(opts.Path)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	updatedAt := opts.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updatedContent, err := ticketmetadata.UpdateUpdatedAt(string(content), updatedAt)
	if err != nil {
		return err
	}

	updatedContent, err = UpsertEvidenceStrategyChecklist(updatedContent, opts.Checklist)
	if err != nil {
		return err
	}

	err = writeLegacyTicketFile(ctx, opts.Path, content, []byte(updatedContent))
	if err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}
