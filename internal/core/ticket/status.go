package ticket

import "strings"

const (
	// StatusBacklog defines a package boundary constant.
	StatusBacklog = "backlog"
	// StatusReadyForDev defines a package boundary constant.
	StatusReadyForDev = "ready_for_dev"
	// StatusInProgress defines a package boundary constant.
	StatusInProgress = "in_progress"
	// StatusReadyForReview defines a package boundary constant.
	StatusReadyForReview = "ready_for_review"
	// StatusReadyForMerge defines a package boundary constant.
	StatusReadyForMerge = "ready_for_merge"
	// StatusDone defines a package boundary constant.
	StatusDone = "done"

	// StatusOpenLegacy defines a package boundary constant.
	StatusOpenLegacy = "open"
	// StatusReadyLegacy defines a package boundary constant.
	StatusReadyLegacy = "ready"
	// StatusReviewingLegacy defines a package boundary constant.
	StatusReviewingLegacy = "reviewing"
	// StatusBlockedLegacy defines a package boundary constant.
	StatusBlockedLegacy = "blocked"
	// StatusSolvedLegacy defines a package boundary constant.
	StatusSolvedLegacy = "solved"
	// StatusMergedLegacy defines a package boundary constant.
	StatusMergedLegacy = "merged"
	// StatusClosedLegacy defines a package boundary constant.
	StatusClosedLegacy = "closed"
)

var normalBoardTicketStatuses = map[string]struct{}{
	StatusBacklog:        {},
	StatusReadyForDev:    {},
	StatusInProgress:     {},
	StatusReadyForReview: {},
	StatusReadyForMerge:  {},
	StatusDone:           {},
}

var legacyTicketStatusAliases = map[string]string{
	StatusOpenLegacy:      StatusBacklog,
	StatusReadyLegacy:     StatusReadyForDev,
	StatusReviewingLegacy: StatusReadyForReview,
	StatusSolvedLegacy:    StatusReadyForMerge,
	StatusMergedLegacy:    StatusDone,
	StatusClosedLegacy:    StatusDone,
}

// ValidStatus executes the corresponding package operation.
func ValidStatus(status string) bool {
	status = strings.TrimSpace(status)
	if _, ok := normalBoardTicketStatuses[status]; ok {
		return true
	}

	if _, ok := legacyTicketStatusAliases[status]; ok {
		return true
	}

	return status == StatusBlockedLegacy
}

// NormalBoardStatus executes the corresponding package operation.
func NormalBoardStatus(status string) bool {
	_, ok := normalBoardTicketStatuses[strings.TrimSpace(status)]

	return ok
}

// BoardStatus executes the corresponding package operation.
func BoardStatus(status string) (string, bool) {
	status = strings.TrimSpace(status)
	if NormalBoardStatus(status) {
		return status, true
	}

	alias, ok := legacyTicketStatusAliases[status]

	return alias, ok
}
