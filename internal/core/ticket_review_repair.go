package core

const (
	// DefaultTicketReviewRepairLimit is the default maximum number of repair executions after review requests changes.
	DefaultTicketReviewRepairLimit = 1
)

// IsTicketReviewRepair reports whether a ticket is waiting for review-requested repair work.
func IsTicketReviewRepair(ticket IndexedTicket) bool {
	return ticket.Status == TicketStatusReadyForDev &&
		(ticket.ReviewState == "changes_requested" || ticket.ReviewResult == "needs_changes")
}

// TicketReviewRepairCount returns how many repair executions have already been consumed.
func TicketReviewRepairCount(ticket IndexedTicket) int {
	if !IsTicketReviewRepair(ticket) || ticket.ReviewRound <= 1 {
		return 0
	}

	return ticket.ReviewRound - 1
}

// TicketReviewRepairAllowed reports whether a repair ticket can still run another repair execution.
func TicketReviewRepairAllowed(ticket IndexedTicket, limit int) bool {
	return IsTicketReviewRepair(ticket) && TicketReviewRepairCount(ticket) < normalizedTicketReviewRepairLimit(limit)
}

// TicketReviewRepairLimitExceeded reports whether a repair ticket has exhausted its repair executions.
func TicketReviewRepairLimitExceeded(ticket IndexedTicket, limit int) bool {
	return IsTicketReviewRepair(ticket) && TicketReviewRepairCount(ticket) >= normalizedTicketReviewRepairLimit(limit)
}

func normalizedTicketReviewRepairLimit(limit int) int {
	if limit < 0 {
		return 0
	}

	return limit
}
