package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
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
	if !ticketcore.IsValidTicketStatus(status) {
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
	if status != ticketcore.TicketStatusReadyForReview && status != ticketcore.TicketStatusReadyForMerge {
		return nil
	}

	frontmatter, err := ticketmetadata.Parse(content)
	if err != nil {
		return err
	}

	if status == ticketcore.TicketStatusReadyForMerge {
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

func validateReadyForMergeTransitionMetadata(frontmatter ticketmetadata.Frontmatter) error {
	reviewRound, err := ticketmetadata.OptionalTicketInt(frontmatter.Metadata["review_round"])
	if err != nil {
		return err
	}

	ticket := ticketcore.IndexedTicket{
		IndexedTicketIdentity: ticketcore.IndexedTicketIdentity{Key: opsmetadata.Unquote(frontmatter.Metadata["key"])}, IndexedTicketClassification: ticketcore.IndexedTicketClassification{Status: ticketcore.TicketStatusReadyForMerge}, IndexedTicketSource: ticketcore.IndexedTicketSource{Branch: opsmetadata.Unquote(frontmatter.Metadata["branch"])}, IndexedTicketReview: ticketcore.IndexedTicketReview{
			ReviewRefs:   frontmatter.ReviewRefs,
			ReviewState:  opsmetadata.Unquote(frontmatter.Metadata["review_state"]),
			ReviewResult: opsmetadata.Unquote(frontmatter.Metadata["review_result"]),
			ReviewRound:  reviewRound,
		},
	}

	_, err = ValidateTicketMergeHandoff(ticket)

	return err
}

func hasPullRequestReviewRef(refs []ticketmetadata.ReviewRef) bool {
	for _, ref := range refs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return true
		}
	}

	return false
}
