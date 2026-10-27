package core

import (
	"fmt"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// IntegrationCycleMode defines the policy used by one ticket integration cycle.
type IntegrationCycleMode string

// IntegrationCyclePhase defines the current or final phase of one integration cycle.
type IntegrationCyclePhase string

// IntegrationCycleStatus defines the lifecycle status of one integration cycle.
type IntegrationCycleStatus string

// IntegrationCycleBlocker defines the compact reason a cycle cannot continue.
type IntegrationCycleBlocker string

// IntegrationReviewDecision defines the outcome of an integration review phase.
type IntegrationReviewDecision string

// IntegrationCyclePolicy describes how one selected ticket should move through integration.
type IntegrationCyclePolicy struct {
	Mode                IntegrationCycleMode
	RequiresReview      bool
	RequiresHumanReview bool
	RequiresMerge       bool
}

// IntegrationCyclePlan describes the next ticket integration cycle without side effects.
type IntegrationCyclePlan struct {
	ProjectID string
	Ticket    ticketcore.Indexed
	Policy    IntegrationCyclePolicy
	Phase     IntegrationCyclePhase
	Status    IntegrationCycleStatus
	Blocker   IntegrationCycleBlocker
}

// IntegrationCyclePlanOptions describes input for planning one selected ticket cycle.
type IntegrationCyclePlanOptions struct {
	ProjectID    string
	Ticket       ticketcore.Indexed
	ReviewConfig ReviewWorkflowConfig
}

const (
	// IntegrationCycleModeHumanReview stops after development handoff for human review.
	IntegrationCycleModeHumanReview IntegrationCycleMode = "human_review"
	// IntegrationCycleModeAgenticReview runs high-level agentic review before merge.
	IntegrationCycleModeAgenticReview IntegrationCycleMode = "agentic_review"
	// IntegrationCycleModeTrustedAutoMerge skips separate review but still merges before continuing.
	IntegrationCycleModeTrustedAutoMerge IntegrationCycleMode = "trusted_auto_merge"
)

const (
	// IntegrationCyclePhaseSelected identifies a selected ticket before execution.
	IntegrationCyclePhaseSelected IntegrationCyclePhase = "selected"
	// IntegrationCyclePhaseDeveloping identifies active ticket development.
	IntegrationCyclePhaseDeveloping IntegrationCyclePhase = "developing"
	// IntegrationCyclePhaseDevelopmentApplied identifies applied development results.
	IntegrationCyclePhaseDevelopmentApplied IntegrationCyclePhase = "development_applied"
	// IntegrationCyclePhaseReviewHandoff identifies human-review handoff.
	IntegrationCyclePhaseReviewHandoff IntegrationCyclePhase = "review_handoff"
	// IntegrationCyclePhaseReviewing identifies active agentic review.
	IntegrationCyclePhaseReviewing IntegrationCyclePhase = "reviewing"
	// IntegrationCyclePhaseRepairing identifies repair work after review findings.
	IntegrationCyclePhaseRepairing IntegrationCyclePhase = "repairing"
	// IntegrationCyclePhaseReadyForMerge identifies a reviewed ticket ready to merge.
	IntegrationCyclePhaseReadyForMerge IntegrationCyclePhase = "ready_for_merge"
	// IntegrationCyclePhaseMerging identifies active merge integration.
	IntegrationCyclePhaseMerging IntegrationCyclePhase = "merging"
	// IntegrationCyclePhaseDone identifies a completed cycle.
	IntegrationCyclePhaseDone IntegrationCyclePhase = "done"
	// IntegrationCyclePhaseBlocked identifies a safely blocked cycle.
	IntegrationCyclePhaseBlocked IntegrationCyclePhase = "blocked"
)

const (
	// IntegrationCycleStatusPending identifies a planned cycle.
	IntegrationCycleStatusPending IntegrationCycleStatus = "pending"
	// IntegrationCycleStatusRunning identifies an active cycle.
	IntegrationCycleStatusRunning IntegrationCycleStatus = "running"
	// IntegrationCycleStatusDone identifies a cycle whose ticket reached done.
	IntegrationCycleStatusDone IntegrationCycleStatus = "done"
	// IntegrationCycleStatusBlocked identifies a cycle blocked by review or merge policy.
	IntegrationCycleStatusBlocked IntegrationCycleStatus = "blocked"
	// IntegrationCycleStatusFailed identifies infrastructure failure before normal blocking.
	IntegrationCycleStatusFailed IntegrationCycleStatus = "failed"
)

const (
	// IntegrationCycleBlockerNone means no blocker is recorded.
	IntegrationCycleBlockerNone IntegrationCycleBlocker = ""
	// IntegrationCycleBlockerHumanReviewPending means human review must complete before later work.
	IntegrationCycleBlockerHumanReviewPending IntegrationCycleBlocker = "human_review_pending"
	// IntegrationCycleBlockerNeedsRepair means review found blocking repair work.
	IntegrationCycleBlockerNeedsRepair IntegrationCycleBlocker = "needs_repair"
	// IntegrationCycleBlockerInsufficientEvidence means review could not find enough evidence.
	IntegrationCycleBlockerInsufficientEvidence IntegrationCycleBlocker = "insufficient_evidence"
	// IntegrationCycleBlockerHumanRequired means the reviewer cannot decide safely.
	IntegrationCycleBlockerHumanRequired IntegrationCycleBlocker = "human_required"
	// IntegrationCycleBlockerRepairLimitExceeded means finite repair budget was exhausted.
	IntegrationCycleBlockerRepairLimitExceeded IntegrationCycleBlocker = "repair_limit_exceeded"
	// IntegrationCycleBlockerMergeBlocked means merge preflight or provider execution blocked.
	IntegrationCycleBlockerMergeBlocked IntegrationCycleBlocker = "merge_blocked"
	// IntegrationCycleBlockerPolicyInvalid means cycle policy could not be resolved.
	IntegrationCycleBlockerPolicyInvalid IntegrationCycleBlocker = "policy_invalid"
)

const (
	// IntegrationReviewDecisionPass allows the ticket to proceed toward merge.
	IntegrationReviewDecisionPass IntegrationReviewDecision = "pass"
	// IntegrationReviewDecisionNeedsChanges sends the same ticket back to repair.
	IntegrationReviewDecisionNeedsChanges IntegrationReviewDecision = "needs_changes"
	// IntegrationReviewDecisionInsufficientEvidence blocks until evidence is improved.
	IntegrationReviewDecisionInsufficientEvidence IntegrationReviewDecision = "insufficient_evidence"
	// IntegrationReviewDecisionHumanRequired blocks for human judgement.
	IntegrationReviewDecisionHumanRequired IntegrationReviewDecision = "human_required"
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

// Valid reports whether the cycle phase is supported.
func (p IntegrationCyclePhase) Valid() bool {
	switch p {
	case IntegrationCyclePhaseSelected,
		IntegrationCyclePhaseDeveloping,
		IntegrationCyclePhaseDevelopmentApplied,
		IntegrationCyclePhaseReviewHandoff,
		IntegrationCyclePhaseReviewing,
		IntegrationCyclePhaseRepairing,
		IntegrationCyclePhaseReadyForMerge,
		IntegrationCyclePhaseMerging,
		IntegrationCyclePhaseDone,
		IntegrationCyclePhaseBlocked:
		return true
	default:
		return false
	}
}

// Valid reports whether the cycle status is supported.
func (s IntegrationCycleStatus) Valid() bool {
	switch s {
	case IntegrationCycleStatusPending,
		IntegrationCycleStatusRunning,
		IntegrationCycleStatusDone,
		IntegrationCycleStatusBlocked,
		IntegrationCycleStatusFailed:
		return true
	default:
		return false
	}
}

// Valid reports whether the blocker value is supported.
func (b IntegrationCycleBlocker) Valid() bool {
	switch b {
	case IntegrationCycleBlockerNone,
		IntegrationCycleBlockerHumanReviewPending,
		IntegrationCycleBlockerNeedsRepair,
		IntegrationCycleBlockerInsufficientEvidence,
		IntegrationCycleBlockerHumanRequired,
		IntegrationCycleBlockerRepairLimitExceeded,
		IntegrationCycleBlockerMergeBlocked,
		IntegrationCycleBlockerPolicyInvalid:
		return true
	default:
		return false
	}
}

// Valid reports whether the review decision is supported.
func (d IntegrationReviewDecision) Valid() bool {
	switch d {
	case IntegrationReviewDecisionPass,
		IntegrationReviewDecisionNeedsChanges,
		IntegrationReviewDecisionInsufficientEvidence,
		IntegrationReviewDecisionHumanRequired:
		return true
	default:
		return false
	}
}

// IntegrationCyclePolicyFromReviewConfig maps review workflow config to cycle behavior.
func IntegrationCyclePolicyFromReviewConfig(config ReviewWorkflowConfig) (IntegrationCyclePolicy, error) {
	if err := ValidateReviewWorkflowConfig(config); err != nil {
		return IntegrationCyclePolicy{}, err
	}

	switch config.Mode {
	case workflowstate.ReviewHuman:
		return IntegrationCyclePolicy{
			Mode:                IntegrationCycleModeHumanReview,
			RequiresReview:      true,
			RequiresHumanReview: true,
			RequiresMerge:       false,
		}, nil
	case workflowstate.ReviewAgentic:
		return IntegrationCyclePolicy{
			Mode:           IntegrationCycleModeAgenticReview,
			RequiresReview: true,
			RequiresMerge:  true,
		}, nil
	case workflowstate.ReviewTrustedAutoMerge:
		return IntegrationCyclePolicy{
			Mode:          IntegrationCycleModeTrustedAutoMerge,
			RequiresMerge: true,
		}, nil
	default:
		return IntegrationCyclePolicy{}, fmt.Errorf("unsupported review workflow mode: %s", config.Mode)
	}
}

// PlanIntegrationCycle describes one selected ticket cycle without mutating state.
func PlanIntegrationCycle(opts IntegrationCyclePlanOptions) (IntegrationCyclePlan, error) {
	if opts.ProjectID == "" {
		return IntegrationCyclePlan{}, fmt.Errorf("integration cycle project id is required")
	}

	if opts.Ticket.Key == "" {
		return IntegrationCyclePlan{}, fmt.Errorf("integration cycle ticket key is required")
	}

	policy, err := IntegrationCyclePolicyFromReviewConfig(opts.ReviewConfig)
	if err != nil {
		return IntegrationCyclePlan{}, err
	}

	return IntegrationCyclePlan{
		ProjectID: opts.ProjectID,
		Ticket:    opts.Ticket,
		Policy:    policy,
		Phase:     IntegrationCyclePhaseSelected,
		Status:    IntegrationCycleStatusPending,
		Blocker:   IntegrationCycleBlockerNone,
	}, nil
}
