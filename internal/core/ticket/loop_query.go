package ticket

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

var defaultLoopEligibleStatuses = []string{string(StatusReadyForDev)}

// LoopQueryOptions selects the project, eligible statuses, metadata filters and order.
type LoopQueryOptions struct {
	ProjectID string
	Statuses  []string
	Filters   workflowstate.TicketFilters
	Order     workflowstate.TicketOrder
}

// LoopQuery reads eligible tickets through the existing index boundary.
type LoopQuery struct {
	Index Index
}

// NewLoopQuery constructs a query over the supplied ticket index.
func NewLoopQuery(index Index) LoopQuery {
	return LoopQuery{Index: index}
}

// EligibleTickets executes the corresponding package operation.
func (q LoopQuery) EligibleTickets(ctx context.Context, opts LoopQueryOptions) ([]Indexed, error) {
	if opts.ProjectID == "" {
		return nil, fmt.Errorf("project id is required")
	}

	if q.Index == nil {
		return nil, fmt.Errorf("ticket index is required")
	}

	statuses, err := LoopEligibleStatuses(opts.Statuses)
	if err != nil {
		return nil, err
	}

	filters, err := NormalizeLoopFilters(opts.Filters)
	if err != nil {
		return nil, err
	}

	order, err := NormalizeLoopOrder(opts.Order)
	if err != nil {
		return nil, err
	}

	tickets, err := q.Index.ListProject(ctx, opts.ProjectID)
	if err != nil {
		return nil, err
	}

	eligible := make([]Indexed, 0, len(tickets))
	for _, ticket := range tickets {
		if slices.Contains(statuses, string(ticket.Status)) && MatchesLoopFilters(filters, ticket) {
			eligible = append(eligible, ticket)
		}
	}

	slices.SortFunc(eligible, func(left, right Indexed) int { return CompareLoopOrder(order, left, right) })

	return eligible, nil
}

// NormalizeLoopFilters validates discrete filters and preserves their original set order.
func NormalizeLoopFilters(filters workflowstate.TicketFilters) (workflowstate.TicketFilters, error) {
	var err error

	filters.Priorities, err = normalizeTicketLoopValues(filters.Priorities, "priority", ValidPriority)
	if err != nil {
		return workflowstate.TicketFilters{}, err
	}

	filters.Severities, err = normalizeTicketLoopValues(filters.Severities, "severity", ValidSeverity)
	if err != nil {
		return workflowstate.TicketFilters{}, err
	}

	filters.Types, err = normalizeTicketLoopValues(filters.Types, "type", ValidType)
	if err != nil {
		return workflowstate.TicketFilters{}, err
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

// MatchesLoopFilters applies the shared ticket selection predicate.
func MatchesLoopFilters(filters workflowstate.TicketFilters, ticket Indexed) bool {
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

// NormalizeLoopOrder selects the legacy default and rejects unsupported order modes.
func NormalizeLoopOrder(order workflowstate.TicketOrder) (workflowstate.TicketOrder, error) {
	if order == "" {
		return workflowstate.TicketOrderCreatedAt, nil
	}

	switch order {
	case workflowstate.TicketOrderKey,
		workflowstate.TicketOrderCreatedAt,
		workflowstate.TicketOrderPriority,
		workflowstate.TicketOrderSeverity,
		workflowstate.TicketOrderPrioritySeverity:
		return order, nil
	default:
		return "", fmt.Errorf("invalid ticket loop order: %s", order)
	}
}

// CompareLoopOrder applies the selected ticket order with the existing deterministic ties.
func CompareLoopOrder(order workflowstate.TicketOrder, left, right Indexed) int {
	switch order {
	case workflowstate.TicketOrderCreatedAt:
		return compareIndexedTicketsByCreatedAt(left, right)
	case workflowstate.TicketOrderPriority:
		return compareIndexedTicketsByPriority(left, right)
	case workflowstate.TicketOrderSeverity:
		return compareIndexedTicketsBySeverity(left, right)
	case workflowstate.TicketOrderPrioritySeverity:
		return firstNonZero(
			cmp.Compare(PriorityRank(left.Priority), PriorityRank(right.Priority)),
			cmp.Compare(severityRank(left.Severity), severityRank(right.Severity)),
			compareIndexedTicketsByCreatedAt(left, right),
		)
	default:
		return CompareKeys(left, right)
	}
}

func compareIndexedTicketsByPriority(left, right Indexed) int {
	return firstNonZero(
		cmp.Compare(PriorityRank(left.Priority), PriorityRank(right.Priority)),
		compareIndexedTicketsByCreatedAt(left, right),
	)
}

func compareIndexedTicketsBySeverity(left, right Indexed) int {
	return firstNonZero(
		cmp.Compare(severityRank(left.Severity), severityRank(right.Severity)),
		compareIndexedTicketsByCreatedAt(left, right),
	)
}

func compareIndexedTicketsByCreatedAt(left, right Indexed) int {
	if left.CreatedAt != nil && right.CreatedAt != nil {
		switch {
		case left.CreatedAt.Before(*right.CreatedAt):
			return -1
		case left.CreatedAt.After(*right.CreatedAt):
			return 1
		default:
			return CompareKeys(left, right)
		}
	}

	if left.CreatedAt != nil {
		return -1
	}

	if right.CreatedAt != nil {
		return 1
	}

	return CompareKeys(left, right)
}

// LoopEligibleStatuses validates the requested statuses and copies the default when absent.
func LoopEligibleStatuses(statuses []string) ([]string, error) {
	if len(statuses) == 0 {
		return slices.Clone(defaultLoopEligibleStatuses), nil
	}

	eligible := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if !ValidStatus(Status(status)) {
			return nil, fmt.Errorf("invalid ticket status: %s", status)
		}

		if !slices.Contains(eligible, status) {
			eligible = append(eligible, status)
		}
	}

	return eligible, nil
}

// PriorityRank orders ticket priorities, placing unclassified values last.
func PriorityRank(priority string) int {
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

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}

	return 0
}

// CompareKeys orders same-prefix numeric keys numerically and other keys lexically.
func CompareKeys(left, right Indexed) int {
	leftKey := splitTicketKey(left.Key)
	rightKey := splitTicketKey(right.Key)

	if leftKey.valid && rightKey.valid && leftKey.prefix == rightKey.prefix {
		return cmp.Compare(leftKey.number, rightKey.number)
	}

	if left.Key < right.Key {
		return -1
	}

	if left.Key > right.Key {
		return 1
	}

	return 0
}

type ticketKeyParts struct {
	prefix string
	number int
	valid  bool
}

func splitTicketKey(key string) ticketKeyParts {
	prefix, numberText, ok := strings.Cut(key, "-")
	if !ok || prefix == "" || numberText == "" {
		return ticketKeyParts{}
	}

	number, err := strconv.Atoi(numberText)
	if err != nil {
		return ticketKeyParts{}
	}

	return ticketKeyParts{prefix: prefix, number: number, valid: true}
}
