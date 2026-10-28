package workflow

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowClaimedPreparation struct {
	Claim    *Claim
	Prepared workflowPreparedTask
	Input    workflowstate.PreparationInput
}

type workflowValidationPreparation struct {
	Claimed workflowClaimedPreparation
	Round   int
	Cached  *workflowstate.IntegrationCandidate
}

type workflowPreparationObservation struct {
	Observation workflowSliceObservation
	Source      workflowstate.ReviewInput
	Capability  workflowstate.PublicationCapability
}

type workflowPreparationServices struct {
	Provider PublicationGuardProvider
	Git      PreparationGit
}

type workflowPreparedIntegration struct {
	Input workflowstate.PreparationInput
	Lane  workflowstate.Lane
}

type workflowPreparationGitRequest struct {
	Preflight workflowPreparationObservation
	Lane      workflowstate.Lane
}

type workflowPreparationInputChange struct {
	Record      workflowstate.Record
	Observation workflowSliceObservation
	Preparation workflowPreparedIntegration
}

type workflowPreparationFailureMode uint8

const (
	workflowPreparationSourceCheck workflowPreparationFailureMode = iota
	workflowPreparationTargetDrift
	workflowPreparationCandidateRepair
)
