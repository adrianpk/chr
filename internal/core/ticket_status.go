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
	Status    ticketcore.Status
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

	status := ticketcore.Status(strings.TrimSpace(string(opts.Status)))
	if !ticketcore.ValidStatus(status) {
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

	updated, err := ticketmetadata.UpdateFields(string(content), map[string]string{
		"status":     opsmetadata.Quote(string(status)),
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

func validateTicketStatusTransitionMetadata(content string, status ticketcore.Status) error {
	if status != ticketcore.StatusReadyForReview && status != ticketcore.StatusReadyForMerge {
		return nil
	}

	frontmatter, err := ticketmetadata.Parse(content)
	if err != nil {
		return err
	}

	if status == ticketcore.StatusReadyForMerge {
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
	reviewRound, err := ticketmetadata.OptionalInteger(frontmatter.Metadata["review_round"])
	if err != nil {
		return err
	}

	ticket := ticketcore.Indexed{
		IndexedIdentity: ticketcore.IndexedIdentity{Key: opsmetadata.Unquote(frontmatter.Metadata["key"])}, IndexedClassification: ticketcore.IndexedClassification{Status: ticketcore.StatusReadyForMerge}, IndexedSource: ticketcore.IndexedSource{Branch: opsmetadata.Unquote(frontmatter.Metadata["branch"])}, IndexedReview: ticketcore.IndexedReview{
			ReviewRefs:   frontmatter.ReviewRefs,
			ReviewState:  ticketcore.ReviewState(opsmetadata.Unquote(frontmatter.Metadata["review_state"])),
			ReviewResult: ticketcore.ReviewResult(opsmetadata.Unquote(frontmatter.Metadata["review_result"])),
			ReviewRound:  reviewRound,
		},
	}

	_, err = ticketcore.ValidateMergeHandoff(ticket)

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
