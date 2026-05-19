package chc

import (
	"fmt"
	"strings"
)

func ValidateIndexedTicket(ticket IndexedTicket) error {
	if err := validateRequiredTicketMetadata(ticket); err != nil {
		return err
	}

	if err := validateTicketClassification(ticket); err != nil {
		return err
	}

	return validateTicketTimes(ticket)
}

func validateRequiredTicketMetadata(ticket IndexedTicket) error {
	if strings.TrimSpace(ticket.Key) == "" {
		return fmt.Errorf("ticket key is required")
	}

	if strings.TrimSpace(ticket.Title) == "" {
		return fmt.Errorf("ticket title is required")
	}

	if !IsValidTicketStatus(ticket.Status) {
		return fmt.Errorf("invalid ticket status: %s", ticket.Status)
	}

	return nil
}

func validateTicketClassification(ticket IndexedTicket) error {
	if !isValidTicketPriority(ticket.Priority) {
		return fmt.Errorf("invalid ticket priority: %s", ticket.Priority)
	}

	if !isValidTicketSeverity(ticket.Severity) {
		return fmt.Errorf("invalid ticket severity: %s", ticket.Severity)
	}

	if !isValidTicketType(ticket.Type) {
		return fmt.Errorf("invalid ticket type: %s", ticket.Type)
	}

	for _, tag := range ticket.Tags {
		if strings.TrimSpace(tag) == "" {
			return fmt.Errorf("ticket tags cannot contain empty values")
		}
	}

	return nil
}

func validateTicketTimes(ticket IndexedTicket) error {
	if ticket.CreatedAt == nil {
		return fmt.Errorf("ticket created_at is required")
	}

	if ticket.UpdatedAt == nil {
		return fmt.Errorf("ticket updated_at is required")
	}

	if ticket.UpdatedAt.Before(*ticket.CreatedAt) {
		return fmt.Errorf("ticket updated_at cannot be before created_at")
	}

	return nil
}

func isValidTicketPriority(priority string) bool {
	switch strings.TrimSpace(priority) {
	case "", "low", "normal", "high", "critical":
		return true
	default:
		return false
	}
}

func isValidTicketSeverity(severity string) bool {
	switch strings.TrimSpace(severity) {
	case "", "low", "normal", "high", "critical":
		return true
	default:
		return false
	}
}

func isValidTicketType(ticketType string) bool {
	switch strings.TrimSpace(ticketType) {
	case "", "task", "bug", "feature", "cleanup", "research", "docs":
		return true
	default:
		return false
	}
}
