package chc

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type TicketStatusUpdater interface {
	UpdateStatus(ctx context.Context, opts UpdateTicketStatusOptions) error
}

type UpdateTicketStatusOptions struct {
	Path      string
	Status    string
	UpdatedAt time.Time
}

type ticketStatusUpdater struct{}

func NewTicketStatusUpdater() TicketStatusUpdater {
	return &ticketStatusUpdater{}
}

func (u *ticketStatusUpdater) UpdateStatus(ctx context.Context, opts UpdateTicketStatusOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status := strings.TrimSpace(opts.Status)
	if !IsValidTicketStatus(status) {
		return fmt.Errorf("invalid ticket status: %s", opts.Status)
	}

	content, err := os.ReadFile(opts.Path)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	updatedAt := opts.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated, err := updateTicketMetadata(string(content), map[string]string{
		"status":     yamlString(status),
		"updated_at": yamlTimestamp(updatedAt),
	})
	if err != nil {
		return err
	}

	if err := os.WriteFile(opts.Path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}

func IsValidTicketStatus(status string) bool {
	switch status {
	case "open", "ready", "in_progress", "reviewing", "blocked", "solved", "merged", "closed":
		return true
	default:
		return false
	}
}
