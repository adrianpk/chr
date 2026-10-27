package ticket

import "strings"

// Status is a ticket lifecycle value; legacy spellings remain valid input.
type Status string

const (
	// StatusBacklog defines a package boundary constant.
	StatusBacklog Status = "backlog"
	// StatusReadyForDev defines a package boundary constant.
	StatusReadyForDev Status = "ready_for_dev"
	// StatusInProgress defines a package boundary constant.
	StatusInProgress Status = "in_progress"
	// StatusReadyForReview defines a package boundary constant.
	StatusReadyForReview Status = "ready_for_review"
	// StatusReadyForMerge defines a package boundary constant.
	StatusReadyForMerge Status = "ready_for_merge"
	// StatusDone defines a package boundary constant.
	StatusDone Status = "done"

	// StatusOpenLegacy defines a package boundary constant.
	StatusOpenLegacy Status = "open"
	// StatusReadyLegacy defines a package boundary constant.
	StatusReadyLegacy Status = "ready"
	// StatusReviewingLegacy defines a package boundary constant.
	StatusReviewingLegacy Status = "reviewing"
	// StatusBlockedLegacy defines a package boundary constant.
	StatusBlockedLegacy Status = "blocked"
	// StatusSolvedLegacy defines a package boundary constant.
	StatusSolvedLegacy Status = "solved"
	// StatusMergedLegacy defines a package boundary constant.
	StatusMergedLegacy Status = "merged"
	// StatusClosedLegacy defines a package boundary constant.
	StatusClosedLegacy Status = "closed"
)

var normalBoardTicketStatuses = map[Status]struct{}{
	StatusBacklog:        {},
	StatusReadyForDev:    {},
	StatusInProgress:     {},
	StatusReadyForReview: {},
	StatusReadyForMerge:  {},
	StatusDone:           {},
}

var legacyTicketStatusAliases = map[Status]Status{
	StatusOpenLegacy:      StatusBacklog,
	StatusReadyLegacy:     StatusReadyForDev,
	StatusReviewingLegacy: StatusReadyForReview,
	StatusSolvedLegacy:    StatusReadyForMerge,
	StatusMergedLegacy:    StatusDone,
	StatusClosedLegacy:    StatusDone,
}

// ValidStatus executes the corresponding package operation.
func ValidStatus(status Status) bool {
	status = Status(strings.TrimSpace(string(status)))
	if _, ok := normalBoardTicketStatuses[status]; ok {
		return true
	}

	if _, ok := legacyTicketStatusAliases[status]; ok {
		return true
	}

	return status == StatusBlockedLegacy
}

// NormalBoardStatus executes the corresponding package operation.
func NormalBoardStatus(status Status) bool {
	_, ok := normalBoardTicketStatuses[Status(strings.TrimSpace(string(status)))]

	return ok
}

// BoardStatus executes the corresponding package operation.
func BoardStatus(status Status) (Status, bool) {
	status = Status(strings.TrimSpace(string(status)))
	if NormalBoardStatus(status) {
		return status, true
	}

	alias, ok := legacyTicketStatusAliases[status]

	return alias, ok
}
