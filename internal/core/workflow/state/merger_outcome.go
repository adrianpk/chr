package state

// MergerOutcome describes an untrusted proposal, never a validation or publication grant.
type MergerOutcome string

// Merger proposal outcomes preserve their existing protocol strings.
const (
	MergerResolved      MergerOutcome = "resolved"
	MergerNeedsDecision MergerOutcome = "needs_decision"
	MergerFailed        MergerOutcome = "failed"
)

// MergerRepairOutcome selects a replay/fix plan or requests an explicit decision.
type MergerRepairOutcome string

// Repair plan outcomes preserve their existing retained strings.
const (
	MergerRepairCombined       MergerRepairOutcome = "combined"
	MergerRepairReplayStep     MergerRepairOutcome = "replay_step"
	MergerRepairIntegrationFix MergerRepairOutcome = "integration_fix"
	MergerRepairNeedsDecision  MergerRepairOutcome = "needs_decision"
)
