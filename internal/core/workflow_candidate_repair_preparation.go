package core

import (
	"context"
	"encoding/json"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowRepairedPreparationGit attaches a retained host-accepted result to a
// fresh isolated gate checkout. It cannot replace the result or method.
type WorkflowRepairedPreparationGit interface {
	WorkflowPreparationGit
	PrepareWorkflowRepairedIntegration(context.Context, WorkflowPreparationQuery, WorkflowMergerRepairState) (WorkflowIntegrationProof, error)
}

func (s *WorkflowStore) acceptedWorkflowCandidateRepair(run workflowInvocationFile) (*WorkflowMergerRepairState, []workflowstate.WorkflowEvidenceRef, error) {
	if run.Integration == nil {
		return nil, nil, nil
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return nil, nil, err
	}

	for i := len(run.Attempts) - 1; i >= 0; i-- {
		attempt := run.Attempts[i]
		if attempt.Purpose != workflowMergerPurpose || attempt.Lane == nil || attempt.Lane.Unit != entry.Unit {
			continue
		}

		state, refs, err := s.readWorkflowAcceptedRepair(run, entry, attempt)
		if state != nil || err != nil {
			return state, refs, err
		}
	}

	return nil, nil, nil
}

func (s *WorkflowStore) readWorkflowAcceptedRepair(run workflowInvocationFile, entry WorkflowIntegrationEntry, attempt WorkflowAttempt) (*WorkflowMergerRepairState, []workflowstate.WorkflowEvidenceRef, error) {
	ref := workflowstate.WorkflowEvidenceRef{Path: attempt.ResultRef, Hash: attempt.ResultHash}

	var result WorkflowMergerAttempt
	if err := s.readWorkflowNativeEvidence(run.ID, ref, &result); err != nil {
		return nil, nil, err
	}

	// Superseded generations remain retained history, never a new-base result.
	if result.Input.ClaimGeneration < entry.ClaimGeneration {
		return nil, nil, nil
	}

	if !validWorkflowAcceptedRepair(run, entry, attempt, result) {
		return nil, nil, workflowstate.ErrConflict
	}

	a := result.Acceptance

	refs, err := s.workflowCandidateRepairAttemptRefs(run, *a.Next.Result)

	return a.Next, refs, err
}

func (s *WorkflowStore) workflowCandidateRepairAttemptRefs(run workflowInvocationFile, proof WorkflowIntegrationProof) ([]workflowstate.WorkflowEvidenceRef, error) {
	refs := []workflowstate.WorkflowEvidenceRef{}

	if proof.Provenance == nil {
		return nil, workflowstate.ErrConflict
	}

	for _, repair := range proof.Provenance.Repairs {
		ref, err := s.workflowCandidateRepairAttemptRef(run, repair.Input.InputHash)
		if err != nil {
			return nil, err
		}

		refs = append(refs, ref)
	}

	return refs, nil
}

func (s *WorkflowStore) workflowCandidateRepairAttemptRef(run workflowInvocationFile, hash string) (workflowstate.WorkflowEvidenceRef, error) {
	for _, attempt := range run.Attempts {
		if attempt.Purpose != workflowMergerPurpose || attempt.ResultRef == "" {
			continue
		}

		data, err := s.readWorkflowReviewImage(attempt.ResultRef, attempt.ResultHash)
		if err != nil {
			return workflowstate.WorkflowEvidenceRef{}, err
		}

		var result WorkflowMergerAttempt
		if err := json.Unmarshal(data, &result); err != nil {
			return workflowstate.WorkflowEvidenceRef{}, err
		}

		if result.Input.InputHash == hash && result.Acceptance != nil && result.Acceptance.Outcome == "accepted_pending_validation" {
			return workflowstate.WorkflowEvidenceRef{Path: attempt.ResultRef, Hash: attempt.ResultHash}, nil
		}
	}

	return workflowstate.WorkflowEvidenceRef{}, workflowstate.ErrConflict
}

func validWorkflowAcceptedRepair(run workflowInvocationFile, entry WorkflowIntegrationEntry, attempt WorkflowAttempt, result WorkflowMergerAttempt) bool {
	a := result.Acceptance
	if attempt.Outcome != "checkpointed" || a == nil || a.Outcome != "accepted_pending_validation" || a.AttemptID != attempt.ID || a.InputHash != result.Input.InputHash || a.Next == nil || a.Next.Result == nil {
		return false
	}

	return workflowAcceptedRepairFenceMatches(run, entry, result)
}

func workflowAcceptedRepairFenceMatches(run workflowInvocationFile, entry WorkflowIntegrationEntry, result WorkflowMergerAttempt) bool {
	input := result.Input

	return input.EntryID == entry.ID && input.Unit == entry.Unit && input.ClaimGeneration == entry.ClaimGeneration && input.CandidateHash == workflowMergerCandidateHash(run, entry) && workflowMergerRepairIdentity(input, *result.Acceptance.Next.Result)
}
