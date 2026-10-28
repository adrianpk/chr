package workflow

import (
	"context"
	"slices"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// LegacyWorkQueue separates shared eligible order from display metadata.
type LegacyWorkQueue struct {
	Tickets, Eligible []ticketcore.Indexed
	Excluded          int
}

// ReadLegacyWorkQueue reads authoritative files without refreshing SQL.
func ReadLegacyWorkQueue(ctx context.Context, layout opscore.Layout, projectID string) (result LegacyWorkQueue, err error) {
	result.Tickets, err = NewWorkflowStore(layout, projectID).readBoardTickets(ctx, false)
	if err != nil {
		return result, err
	}

	tickets, excluded, err := FilterLegacyWorkTickets(ctx, result.Tickets)
	if err != nil {
		return result, err
	}

	result.Excluded = excluded

	for _, ticket := range tickets {
		if ticket.Status == ticketcore.StatusReadyForDev {
			result.Eligible = append(result.Eligible, ticket)
		}
	}

	order, err := ticketcore.NormalizeLoopOrder(workflowstate.TicketOrderCreatedAt)
	if err != nil {
		return result, err
	}

	slices.SortFunc(result.Eligible, func(left, right ticketcore.Indexed) int { return ticketcore.CompareLoopOrder(order, left, right) })

	return result, nil
}
