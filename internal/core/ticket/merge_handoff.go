package ticket

import (
	"fmt"
	"strings"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// MergeHandoff describes a package boundary value.
type MergeHandoff struct {
	TicketKey string
	Branch    string
	ReviewRef ticketmetadata.ReviewRef
}

// ValidateMergeHandoff validates the corresponding package boundary value.
func ValidateMergeHandoff(ticket Indexed) (MergeHandoff, error) {
	if strings.TrimSpace(ticket.Key) == "" {
		return MergeHandoff{}, fmt.Errorf("ticket key is required")
	}

	if ticket.Status != StatusReadyForMerge {
		return MergeHandoff{}, fmt.Errorf("ticket merge handoff requires status %s: %s", StatusReadyForMerge, ticket.Status)
	}

	if ticket.ReviewResult != ReviewResultPass || ticket.ReviewState != ReviewStatePass {
		return MergeHandoff{}, fmt.Errorf("ticket merge handoff requires passed review metadata")
	}

	if strings.TrimSpace(ticket.Branch) == "" {
		return MergeHandoff{}, fmt.Errorf("ticket merge handoff requires branch metadata")
	}

	ref, err := mergeHandoffPullRequestRef(ticket.ReviewRefs)
	if err != nil {
		return MergeHandoff{}, err
	}

	return MergeHandoff{
		TicketKey: ticket.Key,
		Branch:    ticket.Branch,
		ReviewRef: ref,
	}, nil
}

func mergeHandoffPullRequestRef(refs []ticketmetadata.ReviewRef) (ticketmetadata.ReviewRef, error) {
	for _, ref := range refs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return ref, nil
		}
	}

	return ticketmetadata.ReviewRef{}, fmt.Errorf("ticket merge handoff requires a pull_request review ref")
}
