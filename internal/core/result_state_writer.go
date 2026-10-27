package core

import (
	"context"
	"fmt"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// TicketResultStateWriter defines a package boundary interface.
type TicketResultStateWriter interface {
	WriteResultState(ctx context.Context, opts WriteResultStateOptions) error
}

// WriteResultStateOptions describes a package boundary value.
type WriteResultStateOptions struct {
	Application ResultApplication
	UpdatedAt   time.Time
}

type ticketResultStateWriter struct {
	statusUpdater TicketStatusUpdater
}

// NewTicketResultStateWriter constructs the corresponding package service or value.
func NewTicketResultStateWriter() TicketResultStateWriter {
	return &ticketResultStateWriter{
		statusUpdater: NewTicketStatusUpdater(),
	}
}

// WriteResultState writes the requested output.
func (w *ticketResultStateWriter) WriteResultState(ctx context.Context, opts WriteResultStateOptions) error {
	if opts.Application.TicketPath == "" {
		return fmt.Errorf("result application ticket path is required")
	}

	if opts.Application.TicketStatus == "" {
		return fmt.Errorf("result application ticket status is required")
	}

	return w.statusUpdater.UpdateStatus(ctx, UpdateTicketStatusOptions{
		Path:      opts.Application.TicketPath,
		Status:    ticketcore.Status(opts.Application.TicketStatus),
		UpdatedAt: opts.UpdatedAt,
	})
}
