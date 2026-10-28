package workflow

import (
	"context"
	"encoding/json"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// RepairedPreparationGit attaches a retained host-accepted result to a
// fresh isolated gate checkout. It cannot replace the result or method.
type RepairedPreparationGit interface {
	PreparationGit
	PrepareWorkflowRepairedIntegration(context.Context, PreparationQuery, MergerRepairState) (workflowstate.IntegrationProof, error)
}

func (s *Store) acceptedWorkflowCandidateRepair(run workflowInvocationFile) (workflowAcceptedRepair, error) {
	if run.Integration == nil {
		return workflowAcceptedRepair{State: nil, Refs: nil}, nil
	}

	entry, err := s.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		return workflowAcceptedRepair{State: nil, Refs: nil}, err
	}

	for i := len(run.Attempts) - 1; i >= 0; i-- {
		attempt := run.Attempts[i]
		if attempt.Purpose != workflowMergerPurpose || attempt.Lane == nil || attempt.Lane.Unit != entry.Unit {
			continue
		}

		acceptedRepair, err := s.readWorkflowAcceptedRepair(run, entry, attempt)

		state, refs := acceptedRepair.State, acceptedRepair.Refs

		if state != nil || err != nil {
			return workflowAcceptedRepair{State: state, Refs: refs}, err
		}
	}

	return workflowAcceptedRepair{State: nil, Refs: nil}, nil
}

func (s *Store) readWorkflowAcceptedRepair(run workflowInvocationFile, entry IntegrationEntry, attempt Attempt) (workflowAcceptedRepair, error) {
	ref := workflowstate.EvidenceRef{Path: attempt.ResultRef, Hash: attempt.ResultHash}

	var result MergerAttempt
	if err := s.readWorkflowNativeEvidence(run.ID, ref, &result); err != nil {
		return workflowAcceptedRepair{State: nil, Refs: nil}, err
	}

	// Superseded generations remain retained history, never a new-base result.
	if result.Input.ClaimGeneration < entry.ClaimGeneration {
		return workflowAcceptedRepair{State: nil, Refs: nil}, nil
	}

	if !validWorkflowAcceptedRepair(run, entry, attempt, result) {
		return workflowAcceptedRepair{State: nil, Refs: nil}, workflowstate.ErrConflict
	}

	a := result.Acceptance

	refs, err := s.workflowCandidateRepairAttemptRefs(run, *a.Next.Result)

	return workflowAcceptedRepair{State: a.Next, Refs: refs}, err
}

func (s *Store) workflowCandidateRepairAttemptRefs(run workflowInvocationFile, proof workflowstate.IntegrationProof) ([]workflowstate.EvidenceRef, error) {
	refs := []workflowstate.EvidenceRef{}

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

func (s *Store) workflowCandidateRepairAttemptRef(run workflowInvocationFile, hash string) (workflowstate.EvidenceRef, error) {
	for _, attempt := range run.Attempts {
		if attempt.Purpose != workflowMergerPurpose || attempt.ResultRef == "" {
			continue
		}

		data, err := s.readWorkflowReviewImage(attempt.ResultRef, attempt.ResultHash)
		if err != nil {
			return workflowstate.EvidenceRef{}, err
		}

		var result MergerAttempt
		if err := json.Unmarshal(data, &result); err != nil {
			return workflowstate.EvidenceRef{}, err
		}

		if result.Input.InputHash == hash && result.Acceptance != nil && result.Acceptance.Outcome == "accepted_pending_validation" {
			return workflowstate.EvidenceRef{Path: attempt.ResultRef, Hash: attempt.ResultHash}, nil
		}
	}

	return workflowstate.EvidenceRef{}, workflowstate.ErrConflict
}

func validWorkflowAcceptedRepair(run workflowInvocationFile, entry IntegrationEntry, attempt Attempt, result MergerAttempt) bool {
	a := result.Acceptance
	if attempt.Outcome != "checkpointed" || a == nil || a.Outcome != "accepted_pending_validation" || a.AttemptID != attempt.ID || a.InputHash != result.Input.InputHash || a.Next == nil || a.Next.Result == nil {
		return false
	}

	return workflowAcceptedRepairFenceMatches(run, entry, result)
}

func workflowAcceptedRepairFenceMatches(run workflowInvocationFile, entry IntegrationEntry, result MergerAttempt) bool {
	input := result.Input

	return input.EntryID == entry.ID && input.Unit == entry.Unit && input.ClaimGeneration == entry.ClaimGeneration && input.CandidateHash == workflowMergerCandidateHash(run, entry) && workflowmerge.RepairMatches(input, *result.Acceptance.Next.Result)
}
