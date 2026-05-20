package chc

import (
	"context"
	"fmt"
	"slices"
)

var defaultLoopEligibleStatuses = []string{"ready"}

type TicketLoopQueryOptions struct {
	ProjectID string
	Statuses  []string
}

type TicketLoopQuery struct {
	Index TicketIndex
}

func NewTicketLoopQuery(index TicketIndex) TicketLoopQuery {
	return TicketLoopQuery{Index: index}
}

func (q TicketLoopQuery) EligibleTickets(ctx context.Context, opts TicketLoopQueryOptions) ([]IndexedTicket, error) {
	if opts.ProjectID == "" {
		return nil, fmt.Errorf("project id is required")
	}

	if q.Index == nil {
		return nil, fmt.Errorf("ticket index is required")
	}

	statuses, err := loopEligibleStatuses(opts.Statuses)
	if err != nil {
		return nil, err
	}

	tickets, err := q.Index.ListProjectTickets(ctx, opts.ProjectID)
	if err != nil {
		return nil, err
	}

	eligible := make([]IndexedTicket, 0, len(tickets))
	for _, ticket := range tickets {
		if slices.Contains(statuses, ticket.Status) {
			eligible = append(eligible, ticket)
		}
	}

	slices.SortFunc(eligible, compareIndexedTicketsByKey)

	return eligible, nil
}

func loopEligibleStatuses(statuses []string) ([]string, error) {
	if len(statuses) == 0 {
		return slices.Clone(defaultLoopEligibleStatuses), nil
	}

	eligible := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if !IsValidTicketStatus(status) {
			return nil, fmt.Errorf("invalid ticket status: %s", status)
		}

		if !slices.Contains(eligible, status) {
			eligible = append(eligible, status)
		}
	}

	return eligible, nil
}

func compareIndexedTicketsByKey(left, right IndexedTicket) int {
	if left.Key < right.Key {
		return -1
	}

	if left.Key > right.Key {
		return 1
	}

	return 0
}
