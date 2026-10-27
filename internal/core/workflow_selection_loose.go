package core

import (
	"fmt"
	"slices"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func normalizeWorkflowPreviewRequest(request workflowstate.WorkflowPreviewRequest) (workflowstate.WorkflowPreviewRequest, error) {
	if err := validateWorkflowLooseInputBounds(request); err != nil {
		return workflowstate.WorkflowPreviewRequest{}, err
	}

	if request.Selector.Extent != "loose" {
		if workflowHasLooseOptions(request) {
			return workflowstate.WorkflowPreviewRequest{}, fmt.Errorf("loose filters and rank order require the loose extent")
		}

		return request, nil
	}

	var err error

	request.Filters, err = normalizeTicketLoopFilters(request.Filters)
	if err != nil {
		return workflowstate.WorkflowPreviewRequest{}, err
	}

	request.Statuses, err = loopEligibleStatuses(request.Statuses)
	if err != nil {
		return workflowstate.WorkflowPreviewRequest{}, err
	}

	request.Order, err = normalizeTicketLoopOrder(request.Order)
	if err != nil {
		return workflowstate.WorkflowPreviewRequest{}, err
	}

	request.MinPriority, request.MinSeverity = strings.TrimSpace(request.MinPriority), strings.TrimSpace(request.MinSeverity)
	if err := validateWorkflowMinimumRanks(request); err != nil {
		return workflowstate.WorkflowPreviewRequest{}, err
	}

	for _, values := range [][]string{request.Filters.Priorities, request.Filters.Severities, request.Filters.ProjectAreas, request.Filters.Tags, request.Filters.Types, request.Filters.UseCases, request.Statuses} {
		slices.Sort(values)
	}

	return request, nil
}

func workflowHasLooseOptions(request workflowstate.WorkflowPreviewRequest) bool {
	filters := request.Filters
	count := len(filters.Priorities) + len(filters.Severities) + len(filters.ProjectAreas) + len(filters.Tags) + len(filters.Types) + len(filters.UseCases) + len(request.Statuses)

	return count > 0 || request.Order != "" || request.MinPriority != "" || request.MinSeverity != ""
}

func validateWorkflowLooseInputBounds(request workflowstate.WorkflowPreviewRequest) error {
	filters := request.Filters
	for _, values := range [][]string{filters.Priorities, filters.Severities, filters.ProjectAreas, filters.Tags, filters.Types, filters.UseCases, request.Statuses} {
		if len(values) > workflowstate.PageLimit {
			return fmt.Errorf("workflow filter value bound exceeded")
		}

		for _, value := range values {
			if len(value) > 256 || strings.ContainsAny(value, "\x00\n\r") {
				return fmt.Errorf("invalid workflow filter value")
			}
		}
	}

	if len(request.MinPriority) > 32 || len(request.MinSeverity) > 32 {
		return fmt.Errorf("minimum rank bound exceeded")
	}

	return nil
}

func workflowLooseExclusion(request workflowstate.WorkflowPreviewRequest, ticket IndexedTicket) string {
	if !slices.Contains(request.Statuses, ticket.Status) {
		return "status_mismatch"
	}

	ticket.Priority, ticket.Severity = workflowRank(ticket.Priority), workflowRank(ticket.Severity)

	ticket.ProjectArea = strings.TrimSpace(ticket.ProjectArea)
	if !matchesTicketLoopFilters(request.Filters, ticket) {
		return "filter_mismatch"
	}

	if !workflowMeetsMinimum(ticket.Priority, request.MinPriority) || !workflowMeetsMinimum(ticket.Severity, request.MinSeverity) {
		return "below_minimum"
	}

	return ""
}

func workflowRank(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unclassified"
	}

	return value
}

func workflowMeetsMinimum(value, minimum string) bool {
	if minimum == "" {
		return true
	}

	return value != "unclassified" && priorityRank(value) <= priorityRank(minimum)
}

func (selection workflowSelection) orderLooseMembers(members []WorkflowPreviewMember, order workflowstate.TicketLoopOrder) {
	tickets := map[string]IndexedTicket{}
	for _, ticket := range selection.tickets {
		ticket.Priority, ticket.Severity = workflowRank(ticket.Priority), workflowRank(ticket.Severity)
		tickets[ticket.Key] = ticket
	}

	slices.SortFunc(members, func(left, right WorkflowPreviewMember) int {
		return compareTicketLoopOrder(order, tickets[left.Key], tickets[right.Key])
	})
}

func validateWorkflowMinimumRanks(request workflowstate.WorkflowPreviewRequest) error {
	for _, rank := range []string{request.MinPriority, request.MinSeverity} {
		if rank != "" && (!isValidTicketPriority(rank) || rank == "unclassified") {
			return fmt.Errorf("minimum rank must be critical, high, normal or low")
		}
	}

	return nil
}
