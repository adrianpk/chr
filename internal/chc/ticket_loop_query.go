package chc

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

var defaultLoopEligibleStatuses = []string{"ready"}

type TicketLoopQueryOptions struct {
	ProjectID string
	Statuses  []string
	Filters   TicketLoopFilters
}

type TicketLoopFilters struct {
	Priorities   []string
	Severities   []string
	ProjectAreas []string
	Tags         []string
	Types        []string
	UseCases     []string
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

	filters, err := normalizeTicketLoopFilters(opts.Filters)
	if err != nil {
		return nil, err
	}

	tickets, err := q.Index.ListProjectTickets(ctx, opts.ProjectID)
	if err != nil {
		return nil, err
	}

	eligible := make([]IndexedTicket, 0, len(tickets))
	for _, ticket := range tickets {
		if slices.Contains(statuses, ticket.Status) && filters.match(ticket) {
			eligible = append(eligible, ticket)
		}
	}

	slices.SortFunc(eligible, compareIndexedTicketsByKey)

	return eligible, nil
}

func normalizeTicketLoopFilters(filters TicketLoopFilters) (TicketLoopFilters, error) {
	var err error

	if filters.Priorities, err = normalizeTicketLoopValues(filters.Priorities, "priority", isValidTicketPriority); err != nil {
		return TicketLoopFilters{}, err
	}

	if filters.Severities, err = normalizeTicketLoopValues(filters.Severities, "severity", isValidTicketSeverity); err != nil {
		return TicketLoopFilters{}, err
	}

	if filters.Types, err = normalizeTicketLoopValues(filters.Types, "type", isValidTicketType); err != nil {
		return TicketLoopFilters{}, err
	}

	filters.ProjectAreas = normalizeTicketLoopFreeformValues(filters.ProjectAreas)
	filters.Tags = normalizeTicketLoopFreeformValues(filters.Tags)
	filters.UseCases = normalizeTicketLoopFreeformValues(filters.UseCases)

	return filters, nil
}

func normalizeTicketLoopValues(values []string, name string, valid func(string) bool) ([]string, error) {
	normalized := normalizeTicketLoopFreeformValues(values)
	for _, value := range normalized {
		if !valid(value) {
			return nil, fmt.Errorf("invalid ticket %s: %s", name, value)
		}
	}

	return normalized, nil
}

func normalizeTicketLoopFreeformValues(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || slices.Contains(normalized, value) {
			continue
		}

		normalized = append(normalized, value)
	}

	return normalized
}

func (filters TicketLoopFilters) match(ticket IndexedTicket) bool {
	return matchesTicketLoopValues(filters.Priorities, ticket.Priority) &&
		matchesTicketLoopValues(filters.Severities, ticket.Severity) &&
		matchesTicketLoopValues(filters.ProjectAreas, ticket.ProjectArea) &&
		matchesTicketLoopValues(filters.Types, ticket.Type) &&
		matchesTicketLoopValues(filters.UseCases, ticket.UseCase) &&
		matchesTicketLoopTags(filters.Tags, ticket.Tags)
}

func matchesTicketLoopValues(values []string, ticketValue string) bool {
	return len(values) == 0 || slices.Contains(values, ticketValue)
}

func matchesTicketLoopTags(filters, ticketTags []string) bool {
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		if !slices.Contains(ticketTags, filter) {
			return false
		}
	}

	return true
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
