package core

import (
	"fmt"
	"strings"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

const (
	// TicketMergeStatePending defines a package boundary constant.
	TicketMergeStatePending = "pending"
	// TicketMergeStateMerged defines a package boundary constant.
	TicketMergeStateMerged = "merged"
	// TicketMergeStateBlocked defines a package boundary constant.
	TicketMergeStateBlocked = "blocked"

	// TicketMergeResultMerged defines a package boundary constant.
	TicketMergeResultMerged = "merged"
	// TicketMergeResultBlocked defines a package boundary constant.
	TicketMergeResultBlocked = "blocked"
	// TicketMergeResultNotMergeable defines a package boundary constant.
	TicketMergeResultNotMergeable = "not_mergeable"

	// TicketMergeActorMerger defines a package boundary constant.
	TicketMergeActorMerger = "merger"
)

// TicketMergeHandoff describes a package boundary value.
type TicketMergeHandoff struct {
	TicketKey string
	Branch    string
	ReviewRef ticketmetadata.ReviewRef
}

// ValidateTicketMergeHandoff validates the corresponding package boundary value.
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

func ticketMergeHandoffPullRequestRef(refs []ticketmetadata.ReviewRef) (ticketmetadata.ReviewRef, error) {
	for _, ref := range refs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return ref, nil
		}
	}

	return ticketmetadata.ReviewRef{}, fmt.Errorf("ticket merge handoff requires a pull_request review ref")
}
