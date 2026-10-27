package state

// IntegrationCycleMode defines the policy used by one ticket integration cycle.
type IntegrationCycleMode string

const (
	// IntegrationCycleModeHumanReview stops after development handoff for human review.
	IntegrationCycleModeHumanReview IntegrationCycleMode = "human_review"
	// IntegrationCycleModeAgenticReview runs high-level agentic review before merge.
	IntegrationCycleModeAgenticReview IntegrationCycleMode = "agentic_review"
	// IntegrationCycleModeTrustedAutoMerge skips separate review but still merges before continuing.
	IntegrationCycleModeTrustedAutoMerge IntegrationCycleMode = "trusted_auto_merge"
)

// Valid reports whether the cycle mode is supported.
func (m IntegrationCycleMode) Valid() bool {
	switch m {
	case IntegrationCycleModeHumanReview,
		IntegrationCycleModeAgenticReview,
		IntegrationCycleModeTrustedAutoMerge:
		return true
	default:
		return false
	}
}
