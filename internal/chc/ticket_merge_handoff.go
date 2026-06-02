package chc

import (
	"fmt"
	"strings"
)

const (
	TicketMergeStatePending = "pending"
	TicketMergeStateMerged  = "merged"
	TicketMergeStateBlocked = "blocked"

	TicketMergeResultMerged       = "merged"
	TicketMergeResultBlocked      = "blocked"
	TicketMergeResultNotMergeable = "not_mergeable"

	TicketMergeActorMerger = "merger"
)

type TicketMergeHandoff struct {
	TicketKey string
	Branch    string
	ReviewRef TicketReviewRef
}

func ValidateTicketMergeHandoff(ticket IndexedTicket) (TicketMergeHandoff, error) {
	if strings.TrimSpace(ticket.Key) == "" {
		return TicketMergeHandoff{}, fmt.Errorf("ticket key is required")
	}

	if ticket.Status != TicketStatusReadyForMerge {
		return TicketMergeHandoff{}, fmt.Errorf("ticket merge handoff requires status %s: %s", TicketStatusReadyForMerge, ticket.Status)
	}

	if ticket.ReviewResult != TicketReviewCommandResultPass || ticket.ReviewState != TicketReviewCommandResultPass {
		return TicketMergeHandoff{}, fmt.Errorf("ticket merge handoff requires passed review metadata")
	}

	if strings.TrimSpace(ticket.Branch) == "" {
		return TicketMergeHandoff{}, fmt.Errorf("ticket merge handoff requires branch metadata")
	}

	ref, err := ticketMergeHandoffPullRequestRef(ticket.ReviewRefs)
	if err != nil {
		return TicketMergeHandoff{}, err
	}

	return TicketMergeHandoff{
		TicketKey: ticket.Key,
		Branch:    ticket.Branch,
		ReviewRef: ref,
	}, nil
}

func ticketMergeHandoffPullRequestRef(refs []TicketReviewRef) (TicketReviewRef, error) {
	for _, ref := range refs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return ref, nil
		}
	}

	return TicketReviewRef{}, fmt.Errorf("ticket merge handoff requires a pull_request review ref")
}
