package ticket

const (
	// DefaultReviewRepairLimit is the default maximum number of repair executions after review requests changes.
	DefaultReviewRepairLimit = 1
)

// NeedsReviewRepair reports whether a ticket is waiting for review-requested repair work.
func NeedsReviewRepair(ticket Indexed) bool {
	return ticket.Status == StatusReadyForDev &&
		(ticket.ReviewState == "changes_requested" || ticket.ReviewResult == "needs_changes")
}

// ReviewRepairCount returns how many repair executions have already been consumed.
func ReviewRepairCount(ticket Indexed) int {
	if !NeedsReviewRepair(ticket) || ticket.ReviewRound <= 1 {
		return 0
	}

	return ticket.ReviewRound - 1
}

// ReviewRepairAllowed reports whether a repair ticket can still run another repair execution.
func ReviewRepairAllowed(ticket Indexed, limit int) bool {
	return NeedsReviewRepair(ticket) && ReviewRepairCount(ticket) < normalizedReviewRepairLimit(limit)
}

// ReviewRepairLimitExceeded reports whether a repair ticket has exhausted its repair executions.
func ReviewRepairLimitExceeded(ticket Indexed, limit int) bool {
	return NeedsReviewRepair(ticket) && ReviewRepairCount(ticket) >= normalizedReviewRepairLimit(limit)
}

func normalizedReviewRepairLimit(limit int) int {
	if limit < 0 {
		return 0
	}

	return limit
}
