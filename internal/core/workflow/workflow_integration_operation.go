package workflow

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Private operation results preserve related observations and partial outcomes.
// These values carry no context, cancellation handle or new mutation authority.
type (
	workflowAcceptedRepair struct {
		State *MergerRepairState
		Refs  []workflowstate.EvidenceRef
	}
	workflowContinuationStep struct {
		Steps  int
		Reason string
	}
	workflowDeliveryPreparation struct {
		Observation workflowSliceObservation
		Candidate   SliceCandidate
	}
	workflowBudgetObservation struct {
		Entry  IntegrationEntry
		Budget IntegrationBudget
	}
	workflowTokenTotals   struct{ Settled, Reserved int64 }
	workflowRoleUsage     struct{ Consumed, Limit int64 }
	workflowDecisionRetry struct {
		Record IntegrationDecisions
		Found  bool
	}
	workflowRebaseResults struct {
		Commits []string
		Base    string
	}
	workflowPreparedOccupancy struct {
		Count    int
		Retained bool
	}
	workflowPreparationTarget struct{ Repository, Ref string }
	workflowTaskSelection     struct {
		Item   IntegrationSelection
		Reason string
	}
	workflowTaskSelections struct {
		Items    []IntegrationSelection
		Excluded map[string]int
	}
	workflowPreparationEvidence struct {
		Change workflowstate.FileChange
		Ref    workflowstate.EvidenceRef
	}
)

type workflowFreshDecision struct {
	Request IntegrationAnswerRequest
	Answer  workflowstate.IntegrationAnswer
}
type workflowIntegrationAdmissionSnapshot struct {
	Runs        []workflowInvocationFile
	Entries     []IntegrationEntry
	Coordinator workflowCoordinator
}

type workflowIntegrationRunUsage struct {
	Budget IntegrationBudget
	Usage  IntegrationUsage
}
type workflowPreparationAuthorship struct {
	Author            workflowstate.IntegrationIdentity
	Encoding, Message string
}
type workflowPreparationCommit struct {
	Tree       string
	Parents    []string
	Authorship workflowPreparationAuthorship
}
