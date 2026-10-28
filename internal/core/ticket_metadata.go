package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// TicketMetadataUpdater updates editable ticket metadata fields.
type TicketMetadataUpdater interface {
	UpdateMetadata(ctx context.Context, opts UpdateTicketMetadataOptions) error
}

// UpdateTicketMetadataOptions describes editable ticket metadata.
type UpdateTicketMetadataOptions struct {
	Path      string
	Title     string
	Synopsis  string
	UpdatedAt time.Time
}

type ticketMetadataUpdater struct{}

// NewTicketMetadataUpdater constructs a ticket metadata updater.
func NewTicketMetadataUpdater() TicketMetadataUpdater {
	return &ticketMetadataUpdater{}
}

// UpdateMetadata updates editable ticket metadata in place.
func (u *ticketMetadataUpdater) UpdateMetadata(ctx context.Context, opts UpdateTicketMetadataOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if strings.TrimSpace(opts.Title) == "" {
		return fmt.Errorf("ticket title is required")
	}

	if strings.TrimSpace(opts.Synopsis) == "" {
		return fmt.Errorf("ticket synopsis is required")
	}

	content, err := workflowcore.ReadLegacyTicketFile(opts.Path)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	updatedAt := opts.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated, err := ticketmetadata.UpdateFields(string(content), map[string]string{
		"title":      opsmetadata.Quote(strings.TrimSpace(opts.Title)),
		"synopsis":   opsmetadata.Quote(strings.TrimSpace(opts.Synopsis)),
		"updated_at": opsmetadata.Timestamp(updatedAt),
	})
	if err != nil {
		return err
	}

	if err := workflowcore.ReplaceLegacyTicketFile(ctx, opts.Path, content, []byte(updated)); err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}
