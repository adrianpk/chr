package core

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// TicketResultValidationWriter defines a package boundary interface.
type TicketResultValidationWriter interface {
	WriteResultValidation(ctx context.Context, opts WriteResultValidationOptions) error
}

// WriteResultValidationOptions describes a package boundary value.
type WriteResultValidationOptions struct {
	Application ResultApplication
	CreatedAt   time.Time
}

type ticketResultValidationWriter struct{}

// NewTicketResultValidationWriter constructs the corresponding package service or value.
func NewTicketResultValidationWriter() TicketResultValidationWriter {
	return &ticketResultValidationWriter{}
}

// WriteResultValidation writes the requested output.
func (w *ticketResultValidationWriter) WriteResultValidation(
	ctx context.Context,
	opts WriteResultValidationOptions,
) error {
	err := ctx.Err()
	if err != nil {
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

	updated, err := addResultValidationEvidence(string(content), opts.Application, createdAt)
	if err != nil {
		return err
	}

	err = os.WriteFile(opts.Application.TicketPath, []byte(updated), 0o644)
	if err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}

func addResultValidationEvidence(content string, application ResultApplication, createdAt time.Time) (string, error) {
	updatedContent, err := ticketmetadata.UpdateUpdatedAt(content, createdAt)
	if err != nil {
		return "", err
	}

	validationHeading := "# Validation\n"

	validationIndex := strings.Index(updatedContent, validationHeading)
	if validationIndex < 0 {
		return "", fmt.Errorf("ticket validation section is required")
	}

	insertAt := validationIndex + len(validationHeading)
	evidence := renderResultValidationEvidence(application, createdAt)

	return updatedContent[:insertAt] + "\n" + evidence + updatedContent[insertAt:], nil
}

func renderResultValidationEvidence(application ResultApplication, createdAt time.Time) string {
	var output strings.Builder

	output.WriteString("## ")
	output.WriteString(createdAt.UTC().Format(time.RFC3339))
	output.WriteString(" ")
	output.WriteString(application.ExecutionID)
	output.WriteString("\n\n")
	output.WriteString("- command: ")
	output.WriteString(application.Evidence.ValidationCommand)
	output.WriteString("\n")
	output.WriteString("- status: ")
	output.WriteString(string(application.Execution.Status))
	output.WriteString("\n\n")

	return output.String()
}
