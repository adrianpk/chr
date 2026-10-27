package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// ApplyTrustedAutoMergeReviewOptions describes trusted auto-merge handoff input.
type ApplyTrustedAutoMergeReviewOptions struct {
	Ticket     ticketcore.IndexedTicket
	ReviewedAt time.Time
}

// ApplyTrustedAutoMergeReview writes pass metadata for trusted auto-merge.
func ApplyTrustedAutoMergeReview(ctx context.Context, opts ApplyTrustedAutoMergeReviewOptions) (TicketReviewCommandResult, error) {
	if strings.TrimSpace(opts.Ticket.Path) == "" {
		return TicketReviewCommandResult{}, fmt.Errorf("trusted auto-merge ticket path is required")
	}

	ref, ok := firstTicketPullRequestReviewRef(opts.Ticket.ReviewRefs)
	if !ok {
		return TicketReviewCommandResult{}, fmt.Errorf("trusted auto-merge requires a pull_request review ref")
	}

	result := TicketReviewCommandResult{
		TicketKey:    opts.Ticket.Key,
		ReviewRef:    ref,
		ReviewState:  TicketReviewCommandResultPass,
		ReviewResult: TicketReviewCommandResultPass,
		ReviewRound:  opts.Ticket.ReviewRound + 1,
		Status:       ticketcore.TicketStatusReadyForMerge,
	}

	ticket := opts.Ticket
	ticket.Status = result.Status
	ticket.ReviewState = result.ReviewState
	ticket.ReviewResult = result.ReviewResult

	if _, err := ValidateTicketMergeHandoff(ticket); err != nil {
		return TicketReviewCommandResult{}, err
	}

	if err := applyTicketReviewCommandResult(ctx, opts.Ticket.Path, result, opts.ReviewedAt); err != nil {
		return TicketReviewCommandResult{}, err
	}

	return result, nil
}
