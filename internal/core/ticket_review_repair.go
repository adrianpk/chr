package core

import ticketcore "forge.adrianpk.com/chr/internal/core/ticket"

const (
	// DefaultTicketReviewRepairLimit is the default maximum number of repair executions after review requests changes.
	DefaultTicketReviewRepairLimit = 1
)

// IsTicketReviewRepair reports whether a ticket is waiting for review-requested repair work.
func IsTicketReviewRepair(ticket ticketcore.IndexedTicket) bool {
	return ticket.Status == ticketcore.TicketStatusReadyForDev &&
		(ticket.ReviewState == "changes_requested" || ticket.ReviewResult == "needs_changes")
}

// TicketReviewRepairCount returns how many repair executions have already been consumed.
func TicketReviewRepairCount(ticket ticketcore.IndexedTicket) int {
	if !IsTicketReviewRepair(ticket) || ticket.ReviewRound <= 1 {
		return 0
	}

	return ticket.ReviewRound - 1
}

// TicketReviewRepairAllowed reports whether a repair ticket can still run another repair execution.
func TicketReviewRepairAllowed(ticket ticketcore.IndexedTicket, limit int) bool {
	return IsTicketReviewRepair(ticket) && TicketReviewRepairCount(ticket) < normalizedTicketReviewRepairLimit(limit)
}

// TicketReviewRepairLimitExceeded reports whether a repair ticket has exhausted its repair executions.
func TicketReviewRepairLimitExceeded(ticket ticketcore.IndexedTicket, limit int) bool {
	return IsTicketReviewRepair(ticket) && TicketReviewRepairCount(ticket) >= normalizedTicketReviewRepairLimit(limit)
}

func normalizedTicketReviewRepairLimit(limit int) int {
	if limit < 0 {
		return 0
	}

	return limit
}
