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

const (
	TicketStatusBacklog        = "backlog"
	TicketStatusReadyForDev    = "ready_for_dev"
	TicketStatusInProgress     = "in_progress"
	TicketStatusReadyForReview = "ready_for_review"
	TicketStatusReadyForMerge  = "ready_for_merge"
	TicketStatusDone           = "done"

	TicketStatusOpenLegacy      = "open"
	TicketStatusReadyLegacy     = "ready"
	TicketStatusReviewingLegacy = "reviewing"
	TicketStatusBlockedLegacy   = "blocked"
	TicketStatusSolvedLegacy    = "solved"
	TicketStatusMergedLegacy    = "merged"
	TicketStatusClosedLegacy    = "closed"
)

var normalBoardTicketStatuses = map[string]struct{}{
	TicketStatusBacklog:        {},
	TicketStatusReadyForDev:    {},
	TicketStatusInProgress:     {},
	TicketStatusReadyForReview: {},
	TicketStatusReadyForMerge:  {},
	TicketStatusDone:           {},
}

var legacyTicketStatusAliases = map[string]string{
	TicketStatusOpenLegacy:      TicketStatusBacklog,
	TicketStatusReadyLegacy:     TicketStatusReadyForDev,
	TicketStatusReviewingLegacy: TicketStatusReadyForReview,
	TicketStatusSolvedLegacy:    TicketStatusReadyForMerge,
	TicketStatusMergedLegacy:    TicketStatusDone,
	TicketStatusClosedLegacy:    TicketStatusDone,
}

type ticketStatusUpdater struct{}

func NewTicketStatusUpdater() TicketStatusUpdater {
	return &ticketStatusUpdater{}
}

func (u *ticketStatusUpdater) UpdateStatus(ctx context.Context, opts UpdateTicketStatusOptions) error {
	err := ctx.Err()
	if err != nil {
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

	err = os.WriteFile(opts.Path, []byte(updated), 0o644)
	if err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}

func IsValidTicketStatus(status string) bool {
	status = strings.TrimSpace(status)
	if _, ok := normalBoardTicketStatuses[status]; ok {
		return true
	}

	if _, ok := legacyTicketStatusAliases[status]; ok {
		return true
	}

	return status == TicketStatusBlockedLegacy
}

func IsNormalBoardTicketStatus(status string) bool {
	_, ok := normalBoardTicketStatuses[strings.TrimSpace(status)]

	return ok
}

func BoardTicketStatusFor(status string) (string, bool) {
	status = strings.TrimSpace(status)
	if IsNormalBoardTicketStatus(status) {
		return status, true
	}

	alias, ok := legacyTicketStatusAliases[status]

	return alias, ok
}
