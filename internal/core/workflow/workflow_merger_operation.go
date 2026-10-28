package workflow

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Private merger values bind only the observations/resources used together.
// Contexts, cancellation and Store mutation authority remain with the caller.
type workflowClaimedMerger struct {
	Claim    *Claim
	Prepared workflowPreparedTask
}

type workflowMergerAdmissionResult struct {
	Run    workflowInvocationFile
	Entry  IntegrationEntry
	Budget IntegrationBudget
	Seed   string
}

type workflowMergerWorkspace struct {
	Lane workflowstate.Lane
	Root string
}

type workflowMergerClaimInput struct {
	Observation workflowClaimObservation
	Budget      IntegrationBudget
	Input       workflowstate.MergerInput
	Harness     MergerHarness
	Workspace   workflowMergerWorkspace
}

type workflowMergerRepairCommit struct {
	State      MergerRepairState
	Repair     workflowstate.IntegrationTreeRepair
	Acceptance *MergerAcceptance
}

type workflowMergerRegenerationInput struct {
	State    MergerRepairState
	Conflict workflowstate.MergerConflict
	Tree     string
}

type workflowMergerRepairInput struct {
	State  MergerRepairState
	Repair workflowstate.IntegrationTreeRepair
}

type workflowMergerPathMode uint8

const (
	workflowMergerDeclaredPaths workflowMergerPathMode = iota
	workflowMergerExactPaths
)

type workflowMergerPathCheck struct {
	Before string
	After  string
	Mode   workflowMergerPathMode
}
