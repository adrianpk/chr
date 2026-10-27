package model

import (
	"fmt"
	"strings"
)

// ValidatePlanTrackingPayload validates the corresponding package boundary value.
func ValidatePlanTrackingPayload(payload PlanTrackingPayload) (PlanTrackingPayload, error) {
	normalized := PlanTrackingPayload{
		Title:  strings.TrimSpace(payload.Title),
		Body:   strings.TrimSpace(payload.Body),
		Status: strings.TrimSpace(payload.Status),
	}

	if normalized.Title == "" {
		return PlanTrackingPayload{}, fmt.Errorf("plan tracking title is required")
	}

	if normalized.Body == "" {
		return PlanTrackingPayload{}, fmt.Errorf("plan tracking body is required")
	}

	if normalized.Status != "tracking" {
		return PlanTrackingPayload{}, fmt.Errorf("plan tracking status must be tracking")
	}

	if len(payload.Tickets) == 0 {
		return PlanTrackingPayload{}, fmt.Errorf("plan tracking tickets are required")
	}

	normalized.Tickets = make([]PlanTrackingTicketPayload, 0, len(payload.Tickets))
	for index, ticket := range payload.Tickets {
		normalizedTicket, err := ValidatePlanTrackingTicketPayload(ticket)
		if err != nil {
			return PlanTrackingPayload{}, fmt.Errorf("plan tracking ticket %d: %w", index+1, err)
		}

		normalized.Tickets = append(normalized.Tickets, normalizedTicket)
	}

	return normalized, nil
}

// ValidatePlanTrackingTicketPayload validates the corresponding package boundary value.
func ValidatePlanTrackingTicketPayload(payload PlanTrackingTicketPayload) (PlanTrackingTicketPayload, error) {
	normalized := PlanTrackingTicketPayload{
		SliceID:      strings.TrimSpace(payload.SliceID),
		Title:        strings.TrimSpace(payload.Title),
		Synopsis:     strings.TrimSpace(payload.Synopsis),
		Summary:      strings.TrimSpace(payload.Summary),
		Body:         strings.TrimSpace(payload.Body),
		Priority:     normalizePlanTrackingPriority(payload.Priority),
		Severity:     normalizePlanTrackingSeverity(payload.Severity),
		Type:         normalizePlanTrackingType(payload.Type),
		Tags:         normalizeTags(payload.Tags),
		Dependencies: normalizePlanTrackingDependencies(payload.Dependencies),
	}

	if normalized.SliceID == "" {
		return PlanTrackingTicketPayload{}, fmt.Errorf("slice id is required")
	}

	if normalized.Title == "" {
		return PlanTrackingTicketPayload{}, fmt.Errorf("ticket title is required")
	}

	if normalized.Summary == "" {
		return PlanTrackingTicketPayload{}, fmt.Errorf("ticket summary is required")
	}

	if normalized.Synopsis == "" {
		return PlanTrackingTicketPayload{}, fmt.Errorf("ticket synopsis is required")
	}

	if normalized.Body == "" {
		return PlanTrackingTicketPayload{}, fmt.Errorf("ticket body is required")
	}

	return normalized, nil
}

func normalizePlanTrackingPriority(priority string) string {
	switch strings.TrimSpace(priority) {
	case "low", "normal", "high", "critical":
		return strings.TrimSpace(priority)
	default:
		return "normal"
	}
}

func normalizePlanTrackingSeverity(severity string) string {
	switch strings.TrimSpace(severity) {
	case "low", "normal", "high", "critical":
		return strings.TrimSpace(severity)
	default:
		return "normal"
	}
}

func normalizePlanTrackingType(ticketType string) string {
	switch strings.TrimSpace(ticketType) {
	case "task", "bug", "feature", "cleanup", "research", "docs":
		return strings.TrimSpace(ticketType)
	default:
		return "task"
	}
}

//nolint:wsl_v5 // Small normalization loop is clearer inline.
func normalizePlanTrackingDependencies(dependencies []string) []string {
	normalized := []string{}
	for _, dependency := range dependencies {
		dependency = strings.TrimSpace(dependency)
		if dependency != "" {
			normalized = append(normalized, dependency)
		}
	}

	return normalized
}
