package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

// TicketStatusUpdater defines a package boundary interface.
type TicketStatusUpdater interface {
	UpdateStatus(ctx context.Context, opts UpdateTicketStatusOptions) error
}

// UpdateTicketStatusOptions describes a package boundary value.
type UpdateTicketStatusOptions struct {
	Path      string
	Status    string
	UpdatedAt time.Time
}

const (
	// TicketStatusBacklog defines a package boundary constant.
	TicketStatusBacklog = "backlog"
	// TicketStatusReadyForDev defines a package boundary constant.
	TicketStatusReadyForDev = "ready_for_dev"
	// TicketStatusInProgress defines a package boundary constant.
	TicketStatusInProgress = "in_progress"
	// TicketStatusReadyForReview defines a package boundary constant.
	TicketStatusReadyForReview = "ready_for_review"
	// TicketStatusReadyForMerge defines a package boundary constant.
	TicketStatusReadyForMerge = "ready_for_merge"
	// TicketStatusDone defines a package boundary constant.
	TicketStatusDone = "done"

	// TicketStatusOpenLegacy defines a package boundary constant.
	TicketStatusOpenLegacy = "open"
	// TicketStatusReadyLegacy defines a package boundary constant.
	TicketStatusReadyLegacy = "ready"
	// TicketStatusReviewingLegacy defines a package boundary constant.
	TicketStatusReviewingLegacy = "reviewing"
	// TicketStatusBlockedLegacy defines a package boundary constant.
	TicketStatusBlockedLegacy = "blocked"
	// TicketStatusSolvedLegacy defines a package boundary constant.
	TicketStatusSolvedLegacy = "solved"
	// TicketStatusMergedLegacy defines a package boundary constant.
	TicketStatusMergedLegacy = "merged"
	// TicketStatusClosedLegacy defines a package boundary constant.
	TicketStatusClosedLegacy = "closed"
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

// NewTicketStatusUpdater constructs the corresponding package service or value.
func NewTicketStatusUpdater() TicketStatusUpdater {
	return &ticketStatusUpdater{}
}

// UpdateStatus executes the corresponding package operation.
func (u *ticketStatusUpdater) UpdateStatus(ctx context.Context, opts UpdateTicketStatusOptions) error {
	err := ctx.Err()
	if err != nil {
		return err
	}

	status := strings.TrimSpace(opts.Status)
	if !IsValidTicketStatus(status) {
		return fmt.Errorf("invalid ticket status: %s", opts.Status)
	}

	content, err := readLegacyTicketFile(opts.Path)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	err = validateTicketStatusTransitionMetadata(string(content), status)
	if err != nil {
		return err
	}

	updatedAt := opts.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated, err := updateTicketMetadata(string(content), map[string]string{
		"status":     opsmetadata.Quote(status),
		"updated_at": opsmetadata.Timestamp(updatedAt),
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

func validateTicketStatusTransitionMetadata(content, status string) error {
	if status != TicketStatusReadyForReview && status != TicketStatusReadyForMerge {
		return nil
	}

	frontmatter, err := parseTicketFrontmatter(content)
	if err != nil {
		return err
	}

	if status == TicketStatusReadyForMerge {
		return validateReadyForMergeTransitionMetadata(frontmatter)
	}

	branch := opsmetadata.Unquote(frontmatter.Metadata["branch"])
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("ticket ready_for_review requires branch metadata")
	}

	if !hasPullRequestReviewRef(frontmatter.ReviewRefs) {
		return fmt.Errorf("ticket ready_for_review requires a pull_request review ref")
	}

	return nil
}

func validateReadyForMergeTransitionMetadata(frontmatter ticketFrontmatter) error {
	reviewRound, err := optionalTicketInt(frontmatter.Metadata["review_round"])
	if err != nil {
		return err
	}

	ticket := IndexedTicket{
		IndexedTicketIdentity: IndexedTicketIdentity{Key: opsmetadata.Unquote(frontmatter.Metadata["key"])}, IndexedTicketClassification: IndexedTicketClassification{Status: TicketStatusReadyForMerge}, IndexedTicketSource: IndexedTicketSource{Branch: opsmetadata.Unquote(frontmatter.Metadata["branch"])}, IndexedTicketReview: IndexedTicketReview{
			ReviewRefs:   frontmatter.ReviewRefs,
			ReviewState:  opsmetadata.Unquote(frontmatter.Metadata["review_state"]),
			ReviewResult: opsmetadata.Unquote(frontmatter.Metadata["review_result"]),
			ReviewRound:  reviewRound,
		},
	}

	_, err = ValidateTicketMergeHandoff(ticket)

	return err
}

func hasPullRequestReviewRef(refs []TicketReviewRef) bool {
	for _, ref := range refs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return true
		}
	}

	return false
}

// IsValidTicketStatus executes the corresponding package operation.
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

// IsNormalBoardTicketStatus executes the corresponding package operation.
func IsNormalBoardTicketStatus(status string) bool {
	_, ok := normalBoardTicketStatuses[strings.TrimSpace(status)]

	return ok
}

// BoardTicketStatusFor executes the corresponding package operation.
func BoardTicketStatusFor(status string) (string, bool) {
	status = strings.TrimSpace(status)
	if IsNormalBoardTicketStatus(status) {
		return status, true
	}

	alias, ok := legacyTicketStatusAliases[status]

	return alias, ok
}
