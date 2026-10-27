package ticket

import "strings"

const (
	// TicketStatusBacklog defines a package boundary constant.
	TicketStatusBacklog = "backlog"
	// TicketStatusReadyForDev defines a package boundary constant.
	TicketStatusReadyForDev = "ready_for_dev"
	// TicketStatusInProgress defines a package boundary constant.
	TicketStatusInProgress = "in_progress"
	// TicketStatusReadyForReview defines a package boundary constant.
	TicketStatusReadyForReview = "ready_for_review"
	// TicketStatusReadyForMerge defines a package boundary constant.
	TicketStatusReadyForMerge = "ready_for_merge"
	// TicketStatusDone defines a package boundary constant.
	TicketStatusDone = "done"

	// TicketStatusOpenLegacy defines a package boundary constant.
	TicketStatusOpenLegacy = "open"
	// TicketStatusReadyLegacy defines a package boundary constant.
	TicketStatusReadyLegacy = "ready"
	// TicketStatusReviewingLegacy defines a package boundary constant.
	TicketStatusReviewingLegacy = "reviewing"
	// TicketStatusBlockedLegacy defines a package boundary constant.
	TicketStatusBlockedLegacy = "blocked"
	// TicketStatusSolvedLegacy defines a package boundary constant.
	TicketStatusSolvedLegacy = "solved"
	// TicketStatusMergedLegacy defines a package boundary constant.
	TicketStatusMergedLegacy = "merged"
	// TicketStatusClosedLegacy defines a package boundary constant.
	TicketStatusClosedLegacy = "closed"
)

var normalBoardTicketStatuses = map[string]struct{}{
	TicketStatusBacklog:        {},
	TicketStatusReadyForDev:    {},
	TicketStatusInProgress:     {},
	TicketStatusReadyForReview: {},
	TicketStatusReadyForMerge:  {},
	TicketStatusDone:           {},
}

var legacyTicketStatusAliases = map[string]string{
	TicketStatusOpenLegacy:      TicketStatusBacklog,
	TicketStatusReadyLegacy:     TicketStatusReadyForDev,
	TicketStatusReviewingLegacy: TicketStatusReadyForReview,
	TicketStatusSolvedLegacy:    TicketStatusReadyForMerge,
	TicketStatusMergedLegacy:    TicketStatusDone,
	TicketStatusClosedLegacy:    TicketStatusDone,
}

// IsValidTicketStatus executes the corresponding package operation.
func IsValidTicketStatus(status string) bool {
	status = strings.TrimSpace(status)
	if _, ok := normalBoardTicketStatuses[status]; ok {
		return true
	}

	if _, ok := legacyTicketStatusAliases[status]; ok {
		return true
	}

	return status == TicketStatusBlockedLegacy
}

// IsNormalBoardTicketStatus executes the corresponding package operation.
func IsNormalBoardTicketStatus(status string) bool {
	_, ok := normalBoardTicketStatuses[strings.TrimSpace(status)]

	return ok
}

// BoardTicketStatusFor executes the corresponding package operation.
func BoardTicketStatusFor(status string) (string, bool) {
	status = strings.TrimSpace(status)
	if IsNormalBoardTicketStatus(status) {
		return status, true
	}

	alias, ok := legacyTicketStatusAliases[status]

	return alias, ok
}
