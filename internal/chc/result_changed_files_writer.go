package chc

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type TicketResultChangedFilesWriter interface {
	WriteChangedFiles(ctx context.Context, opts WriteChangedFilesOptions) error
}

type WriteChangedFilesOptions struct {
	Application ResultApplication
	CreatedAt   time.Time
}

type ticketResultChangedFilesWriter struct {
	commenter TicketCommenter
}

func NewTicketResultChangedFilesWriter() TicketResultChangedFilesWriter {
	return &ticketResultChangedFilesWriter{
		commenter: NewTicketCommenter(),
	}
}

func (w *ticketResultChangedFilesWriter) WriteChangedFiles(
	ctx context.Context,
	opts WriteChangedFilesOptions,
) error {
	if len(opts.Application.ChangedFiles) == 0 {
		return nil
	}

	if opts.Application.TicketPath == "" {
		return fmt.Errorf("result application ticket path is required")
	}

	return w.commenter.AddComment(ctx, AddTicketCommentOptions{
		Path:      opts.Application.TicketPath,
		Author:    opts.Application.Comment.Author,
		Body:      RenderResultApplicationChangedFilesBody(opts.Application),
		CreatedAt: opts.CreatedAt,
	})
}

func RenderResultApplicationChangedFilesBody(application ResultApplication) string {
	var output strings.Builder

	output.WriteString("Execution ")
	output.WriteString(application.ExecutionID)
	output.WriteString(" changed files:")

	for _, changedFile := range application.ChangedFiles {
		path := strings.TrimSpace(changedFile.Path)
		if path == "" {
			continue
		}

		output.WriteString("\n\n- ")
		output.WriteString(path)

		summary := strings.TrimSpace(changedFile.Summary)
		if summary != "" {
			output.WriteString(": ")
			output.WriteString(summary)
		}
	}

	return output.String()
}
