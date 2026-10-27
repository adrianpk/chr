package ticket

import (
	"fmt"
	"strings"

	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

// ValidateIndexed validates the corresponding package boundary value.
func ValidateIndexed(ticket Indexed) error {
	err := validateRequiredTicketMetadata(ticket)
	if err != nil {
		return err
	}

	err = validateTicketClassification(ticket)
	if err != nil {
		return err
	}

	err = validateTicketReviewMetadata(ticket)
	if err != nil {
		return err
	}

	err = validateTicketMergeMetadata(ticket)
	if err != nil {
		return err
	}

	return validateTicketTimes(ticket)
}

func validateRequiredTicketMetadata(ticket Indexed) error {
	if strings.TrimSpace(ticket.Key) == "" {
		return fmt.Errorf("ticket key is required")
	}

	if strings.TrimSpace(ticket.Title) == "" {
		return fmt.Errorf("ticket title is required")
	}

	if !ValidStatus(ticket.Status) {
		return fmt.Errorf("invalid ticket status: %s", ticket.Status)
	}

	return nil
}

func validateTicketClassification(ticket Indexed) error {
	if !ValidPriority(ticket.Priority) {
		return fmt.Errorf("invalid ticket priority: %s", ticket.Priority)
	}

	if !ValidSeverity(ticket.Severity) {
		return fmt.Errorf("invalid ticket severity: %s", ticket.Severity)
	}

	if !ValidType(ticket.Type) {
		return fmt.Errorf("invalid ticket type: %s", ticket.Type)
	}

	for _, tag := range ticket.Tags {
		if strings.TrimSpace(tag) == "" {
			return fmt.Errorf("ticket tags cannot contain empty values")
		}
	}

	return nil
}

func validateTicketTimes(ticket Indexed) error {
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

// ValidPriority preserves the shared ticket metadata contract.
func ValidPriority(priority string) bool {
	switch strings.TrimSpace(priority) {
	case "", "unclassified", "low", "normal", "high", "critical":
		return true
	default:
		return false
	}
}

// ValidSeverity preserves the shared ticket metadata contract.
func ValidSeverity(severity string) bool {
	switch strings.TrimSpace(severity) {
	case "", "unclassified", "low", "normal", "high", "critical":
		return true
	default:
		return false
	}
}

// ValidType preserves the shared ticket metadata contract.
func ValidType(ticketType string) bool {
	switch strings.TrimSpace(ticketType) {
	case "", "task", "bug", "feature", "cleanup", "research", "docs":
		return true
	default:
		return false
	}
}

func validateTicketReviewMetadata(ticket Indexed) error {
	if !isValidTicketReviewState(ticket.ReviewState) {
		return fmt.Errorf("invalid ticket review_state: %s", ticket.ReviewState)
	}

	if !isValidTicketReviewResult(ticket.ReviewResult) {
		return fmt.Errorf("invalid ticket review_result: %s", ticket.ReviewResult)
	}

	if ticket.ReviewRound < 0 {
		return fmt.Errorf("ticket review_round cannot be negative")
	}

	for _, ref := range ticket.ReviewRefs {
		if err := ValidateReviewRef(ref); err != nil {
			return err
		}
	}

	for _, finding := range ticket.BlockingFindings {
		if strings.TrimSpace(finding) == "" {
			return fmt.Errorf("ticket blocking_findings cannot contain empty values")
		}
	}

	return nil
}

// ValidateReviewRef preserves the shared ticket metadata contract.
func ValidateReviewRef(ref ticketmetadata.ReviewRef) error {
	if strings.TrimSpace(ref.Type) == "" {
		return fmt.Errorf("ticket review ref type is required")
	}

	if strings.TrimSpace(ref.Provider) == "" {
		return fmt.Errorf("ticket review ref provider is required")
	}

	if strings.TrimSpace(ref.URL) == "" {
		return fmt.Errorf("ticket review ref url is required")
	}

	return nil
}

func isValidTicketReviewState(state ReviewState) bool {
	switch ReviewState(strings.TrimSpace(string(state))) {
	case "", "pass", "changes_requested", "insufficient_evidence", "needs_human_decision":
		return true
	default:
		return false
	}
}

func isValidTicketReviewResult(result ReviewResult) bool {
	switch ReviewResult(strings.TrimSpace(string(result))) {
	case "", "pass", "needs_changes", "insufficient_evidence", "needs_human_decision":
		return true
	default:
		return false
	}
}

func validateTicketMergeMetadata(ticket Indexed) error {
	if !isValidTicketMergeState(ticket.MergeState) {
		return fmt.Errorf("invalid ticket merge_state: %s", ticket.MergeState)
	}

	if !isValidTicketMergeResult(ticket.MergeResult) {
		return fmt.Errorf("invalid ticket merge_result: %s", ticket.MergeResult)
	}

	return nil
}

func isValidTicketMergeState(state MergeState) bool {
	switch MergeState(strings.TrimSpace(string(state))) {
	case "", "pending", "merged", "blocked":
		return true
	default:
		return false
	}
}

func isValidTicketMergeResult(result MergeResult) bool {
	switch MergeResult(strings.TrimSpace(string(result))) {
	case "", "merged", "blocked", "not_mergeable":
		return true
	default:
		return false
	}
}
