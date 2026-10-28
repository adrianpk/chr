package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// TicketCommenter defines a package boundary interface.
type TicketCommenter interface {
	AddComment(ctx context.Context, opts AddTicketCommentOptions) error
}

// AddTicketCommentOptions describes a package boundary value.
type AddTicketCommentOptions struct {
	Path      string
	Author    string
	Body      string
	CreatedAt time.Time
}

type ticketCommenter struct{}

// NewTicketCommenter constructs the corresponding package service or value.
func NewTicketCommenter() TicketCommenter {
	return &ticketCommenter{}
}

// AddComment executes the corresponding package operation.
func (c *ticketCommenter) AddComment(ctx context.Context, opts AddTicketCommentOptions) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	if strings.TrimSpace(opts.Body) == "" {
		return fmt.Errorf("ticket comment body is required")
	}

	content, err := readLegacyTicketFile(opts.Path)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	createdAt := opts.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	updated, err := addTicketComment(string(content), ticketComment{
		Author:    opts.Author,
		Body:      opts.Body,
		CreatedAt: createdAt,
	})
	if err != nil {
		return err
	}

	err = writeLegacyTicketFile(ctx, opts.Path, content, []byte(updated))
	if err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}

type ticketComment struct {
	Author    string
	Body      string
	CreatedAt time.Time
}

func addTicketComment(content string, comment ticketComment) (string, error) {
	updatedContent, err := ticketmetadata.UpdateUpdatedAt(content, comment.CreatedAt)
	if err != nil {
		return "", err
	}

	renderedComment := renderTicketComment(comment)
	commentsHeading := "# Comments\n"

	commentsIndex := strings.Index(updatedContent, commentsHeading)
	if commentsIndex < 0 {
		updatedContent = addMissingTicketCommentsSection(updatedContent)
		commentsIndex = strings.Index(updatedContent, commentsHeading)
	}

	insertAt := commentsIndex + len(commentsHeading)

	nextHeadingIndex := strings.Index(updatedContent[insertAt:], "\n# ")
	if nextHeadingIndex < 0 {
		return strings.TrimRight(updatedContent, "\n") + "\n\n" + renderedComment, nil
	}

	insertAt += nextHeadingIndex + 1

	return updatedContent[:insertAt] + renderedComment + "\n" + updatedContent[insertAt:], nil
}

func addMissingTicketCommentsSection(content string) string {
	blockerHeading := "# Blocker\n"
	blockerIndex := strings.Index(content, blockerHeading)

	if blockerIndex >= 0 {
		return content[:blockerIndex] + "# Comments\n\n" + content[blockerIndex:]
	}

	validationHeading := "# Validation\n"
	validationIndex := strings.Index(content, validationHeading)

	if validationIndex >= 0 {
		return content[:validationIndex] + "# Comments\n\n" + content[validationIndex:]
	}

	return strings.TrimRight(content, "\n") + "\n\n# Comments\n"
}

func renderTicketComment(comment ticketComment) string {
	author := strings.TrimSpace(comment.Author)
	if author == "" {
		author = "user"
	}

	body := strings.TrimSpace(comment.Body)

	var output strings.Builder
	output.WriteString("## ")
	output.WriteString(comment.CreatedAt.UTC().Format(time.RFC3339))
	output.WriteString(" ")
	output.WriteString(author)
	output.WriteString("\n\n")
	output.WriteString(body)
	output.WriteString("\n")

	return output.String()
}
