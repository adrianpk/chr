package workflow

const (
	// TicketMergePreflightBlockKindOperational defines a package boundary constant.
	TicketMergePreflightBlockKindOperational = "operational"
	// TicketMergePreflightBlockKindRepair defines a package boundary constant.
	TicketMergePreflightBlockKindRepair = "repair"

	// TicketMergePreflightBlockMissingHandoff defines a package boundary constant.
	TicketMergePreflightBlockMissingHandoff = "missing_handoff"
	// TicketMergePreflightBlockProviderUnreadable defines a package boundary constant.
	TicketMergePreflightBlockProviderUnreadable = "provider_unreadable"
	// TicketMergePreflightBlockPullRequestClosed defines a package boundary constant.
	TicketMergePreflightBlockPullRequestClosed = "pull_request_closed"
	// TicketMergePreflightBlockWrongBase defines a package boundary constant.
	TicketMergePreflightBlockWrongBase = "wrong_base"
	// TicketMergePreflightBlockHeadMismatch defines a package boundary constant.
	TicketMergePreflightBlockHeadMismatch = "head_mismatch"
	// TicketMergePreflightBlockNotMergeable defines a package boundary constant.
	TicketMergePreflightBlockNotMergeable = "not_mergeable"
	// TicketMergePreflightBlockRepairPending defines a package boundary constant.
	TicketMergePreflightBlockRepairPending = "serial_repair_pending"
	// TicketMergePreflightBlockInvalidMergeStrategy defines a package boundary constant.
	TicketMergePreflightBlockInvalidMergeStrategy = "invalid_merge_strategy"

	// TicketMergeStrategyMerge defines a package boundary constant.
	TicketMergeStrategyMerge = "merge"
	// TicketMergeStrategyRebase defines a package boundary constant.
	TicketMergeStrategyRebase = "rebase"
	// TicketMergeStrategySquash defines a package boundary constant.
	TicketMergeStrategySquash = "squash"
)
