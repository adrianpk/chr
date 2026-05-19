package chc

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type TicketCommenter interface {
	AddComment(ctx context.Context, opts AddTicketCommentOptions) error
}

type AddTicketCommentOptions struct {
	Path      string
	Author    string
	Body      string
	CreatedAt time.Time
}

type ticketCommenter struct{}

func NewTicketCommenter() TicketCommenter {
	return &ticketCommenter{}
}

func (c *ticketCommenter) AddComment(ctx context.Context, opts AddTicketCommentOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if strings.TrimSpace(opts.Body) == "" {
		return fmt.Errorf("ticket comment body is required")
	}

	content, err := os.ReadFile(opts.Path)
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

	if err := os.WriteFile(opts.Path, []byte(updated), 0o644); err != nil {
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
	updatedContent, err := updateTicketUpdatedAt(content, comment.CreatedAt)
	if err != nil {
		return "", err
	}

	renderedComment := renderTicketComment(comment)
	commentsHeading := "# Comments\n"

	commentsIndex := strings.Index(updatedContent, commentsHeading)
	if commentsIndex < 0 {
		return "", fmt.Errorf("ticket comments section is required")
	}

	insertAt := commentsIndex + len(commentsHeading)

	nextHeadingIndex := strings.Index(updatedContent[insertAt:], "\n# ")
	if nextHeadingIndex < 0 {
		return updatedContent[:insertAt] + "\n" + renderedComment, nil
	}

	insertAt += nextHeadingIndex + 1

	return updatedContent[:insertAt] + renderedComment + "\n" + updatedContent[insertAt:], nil
}

func updateTicketUpdatedAt(content string, updatedAt time.Time) (string, error) {
	lineStart := strings.Index(content, "updated_at:")
	if lineStart < 0 {
		return "", fmt.Errorf("ticket updated_at metadata is required")
	}

	lineEnd := strings.IndexByte(content[lineStart:], '\n')
	if lineEnd < 0 {
		return "", fmt.Errorf("ticket updated_at metadata line is incomplete")
	}

	lineEnd += lineStart
	line := "updated_at: " + yamlTimestamp(updatedAt)

	return content[:lineStart] + line + content[lineEnd:], nil
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
	output.WriteString(strconv.Quote(body))
	output.WriteString("\n")

	return output.String()
}
