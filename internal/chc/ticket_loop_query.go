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
	Order     TicketLoopOrder
}

type TicketLoopFilters struct {
	Priorities   []string
	Severities   []string
	ProjectAreas []string
	Tags         []string
	Types        []string
	UseCases     []string
}

type TicketLoopOrder string

const (
	TicketLoopOrderKey              TicketLoopOrder = "key"
	TicketLoopOrderCreatedAt        TicketLoopOrder = "created_at"
	TicketLoopOrderPriority         TicketLoopOrder = "priority"
	TicketLoopOrderSeverity         TicketLoopOrder = "severity"
	TicketLoopOrderPrioritySeverity TicketLoopOrder = "priority_severity"
)

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

	order, err := normalizeTicketLoopOrder(opts.Order)
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

	slices.SortFunc(eligible, order.compare)

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

func normalizeTicketLoopOrder(order TicketLoopOrder) (TicketLoopOrder, error) {
	if order == "" {
		return TicketLoopOrderKey, nil
	}

	switch order {
	case TicketLoopOrderKey,
		TicketLoopOrderCreatedAt,
		TicketLoopOrderPriority,
		TicketLoopOrderSeverity,
		TicketLoopOrderPrioritySeverity:
		return order, nil
	default:
		return "", fmt.Errorf("invalid ticket loop order: %s", order)
	}
}

func (order TicketLoopOrder) compare(left, right IndexedTicket) int {
	switch order {
	case TicketLoopOrderCreatedAt:
		return compareIndexedTicketsByCreatedAt(left, right)
	case TicketLoopOrderPriority:
		return compareIndexedTicketsByPriority(left, right)
	case TicketLoopOrderSeverity:
		return compareIndexedTicketsBySeverity(left, right)
	case TicketLoopOrderPrioritySeverity:
		return firstNonZero(
			compareIndexedTicketsByPriority(left, right),
			compareIndexedTicketsBySeverity(left, right),
			compareIndexedTicketsByKey(left, right),
		)
	default:
		return compareIndexedTicketsByKey(left, right)
	}
}

func compareIndexedTicketsByPriority(left, right IndexedTicket) int {
	return firstNonZero(
		compareRank(priorityRank(left.Priority), priorityRank(right.Priority)),
		compareIndexedTicketsByCreatedAt(left, right),
	)
}

func compareIndexedTicketsBySeverity(left, right IndexedTicket) int {
	return firstNonZero(
		compareRank(severityRank(left.Severity), severityRank(right.Severity)),
		compareIndexedTicketsByCreatedAt(left, right),
	)
}

func compareIndexedTicketsByCreatedAt(left, right IndexedTicket) int {
	if left.CreatedAt != nil && right.CreatedAt != nil {
		switch {
		case left.CreatedAt.Before(*right.CreatedAt):
			return -1
		case left.CreatedAt.After(*right.CreatedAt):
			return 1
		}
	}

	if left.CreatedAt != nil {
		return -1
	}

	if right.CreatedAt != nil {
		return 1
	}

	return compareIndexedTicketsByKey(left, right)
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

func priorityRank(priority string) int {
	switch priority {
	case "critical":
		return 0
	case "high":
		return 1
	case "normal":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

func severityRank(severity string) int {
	switch severity {
	case "critical":
		return 0
	case "high":
		return 1
	case "normal":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

func compareRank(left, right int) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}

	return 0
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
