package core

import (
	"context"
	"fmt"
	"time"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// TicketResultCommentWriter defines a package boundary interface.
type TicketResultCommentWriter interface {
	WriteResultComment(ctx context.Context, opts WriteResultCommentOptions) error
}

// WriteResultCommentOptions describes a package boundary value.
type WriteResultCommentOptions struct {
	Application ResultApplication
	CreatedAt   time.Time
}

type ticketResultCommentWriter struct {
	commenter workflowcore.TicketCommenter
}

// NewTicketResultCommentWriter constructs the corresponding package service or value.
func NewTicketResultCommentWriter() TicketResultCommentWriter {
	return &ticketResultCommentWriter{
		commenter: workflowcore.NewTicketCommenter(),
	}
}

// WriteResultComment writes the requested output.
func (w *ticketResultCommentWriter) WriteResultComment(ctx context.Context, opts WriteResultCommentOptions) error {
	if opts.Application.TicketPath == "" {
		return fmt.Errorf("result application ticket path is required")
	}

	if opts.Application.Comment.Body == "" {
		return fmt.Errorf("result application comment body is required")
	}

	return w.commenter.AddComment(ctx, workflowcore.AddTicketCommentOptions{
		Path:      opts.Application.TicketPath,
		Author:    opts.Application.Comment.Author,
		Body:      opts.Application.Comment.Body,
		CreatedAt: opts.CreatedAt,
	})
}
