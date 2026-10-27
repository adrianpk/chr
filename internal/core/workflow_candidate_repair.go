package core

import (
	"context"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowCandidateRepairBinding retains actionable combined-result evidence.
// It cannot authorize publication or rewind accepted source task execution.
type WorkflowCandidateRepairBinding struct {
	Unit            workflowstate.Ref   `json:"unit"`
	EntryID         string              `json:"entry_id"`
	ClaimGeneration uint64              `json:"claim_generation"`
	CandidateHash   string              `json:"candidate_hash"`
	FindingRef      WorkflowEvidenceRef `json:"finding_ref"`
	RecordRef       WorkflowEvidenceRef `json:"record_ref"`
}

// WorkflowCandidateRepair records the host result and its exact findings before
// a bounded merger may act. Failed gate results remain non-candidates.
type WorkflowCandidateRepair struct {
	Schema      string                       `json:"schema"`
	Reason      string                       `json:"reason"`
	Candidate   WorkflowIntegrationCandidate `json:"candidate"`
	FindingRefs []WorkflowEvidenceRef        `json:"finding_refs"`
}

// ReadWorkflowCandidateRepair returns retained findings for host request
// assembly. Reading never reserves a role or advances the queue.
func (s *WorkflowStore) ReadWorkflowCandidateRepair(ctx context.Context, id string) (repair WorkflowCandidateRepair, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		repair, err = owner.readWorkflowCandidateRepair(run)

		return err
	})

	return repair, err
}

func (s *WorkflowStore) repairWorkflowReviewedSlice(ctx context.Context, id, key string, provider WorkflowPullRequestProvider) error {
	native := false

	err := s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		input, _, err := owner.readWorkflowReview(run)
		if err != nil || input.Native == nil {
			return err
		}

		native = true

		return owner.repairWorkflowNativeReview(ctx, run, key, provider)
	})
	if err != nil || native {
		return err
	}

	_, err = s.RepairWorkflowSlice(ctx, id, key)

	return err
}

func (s *WorkflowStore) repairWorkflowNativeReview(ctx context.Context, run workflowInvocationFile, key string, provider WorkflowPullRequestProvider) error {
	if run.Review.State != "needs_repair" || run.Review.RepairTaskKey != key {
		return workflowstate.ErrConflict
	}

	if run.CandidateRepair != nil && workflowCandidateRepairFindingMatches(*run.CandidateRepair, run.Review) {
		_, err := s.readWorkflowCandidateRepair(run)

		return err
	}

	_, fresh, err := s.workflowReviewInputs(ctx, run, provider)
	if err != nil || checkWorkflowReviewInputHash(fresh, run.Review.InputHash) != nil {
		return errorsOrConflict(err)
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil {
		return err
	}

	ref := WorkflowEvidenceRef{Path: run.Review.ResultRef, Hash: run.Review.ResultHash}

	return s.retainWorkflowCandidateRepair(ctx, &run, candidate, "actionable_review_finding", []WorkflowEvidenceRef{ref})
}

func workflowCandidateRepairFindingMatches(b WorkflowCandidateRepairBinding, review *WorkflowReviewBinding) bool {
	return review != nil && b.Unit == review.Unit && b.FindingRef == (WorkflowEvidenceRef{Path: review.ResultRef, Hash: review.ResultHash})
}

func (s *WorkflowStore) readWorkflowCandidateRepair(run workflowInvocationFile) (WorkflowCandidateRepair, error) {
	var repair WorkflowCandidateRepair

	b := run.CandidateRepair
	if b == nil || run.Integration == nil || run.Lane == nil || b.EntryID != run.Integration.EntryID || b.Unit != run.Lane.Unit {
		return repair, workflowstate.ErrConflict
	}

	if err := s.readWorkflowNativeEvidence(run.ID, b.RecordRef, &repair); err != nil {
		return repair, err
	}

	if !validWorkflowCandidateRepairRecord(run, repair) {
		return repair, workflowstate.ErrConflict
	}

	for _, ref := range repair.FindingRefs {
		if _, err := s.readWorkflowReviewImage(ref.Path, ref.Hash); err != nil {
			return repair, err
		}
	}

	return repair, nil
}

func validWorkflowCandidateRepairRecord(run workflowInvocationFile, repair WorkflowCandidateRepair) bool {
	b := run.CandidateRepair
	if repair.Schema != "chr-candidate-repair/v1" || repair.Candidate.InvocationID != run.ID || repair.Candidate.Input.Native == nil || repair.Candidate.Input.Source.Unit != b.Unit || repair.Candidate.Input.Native.ClaimGeneration != b.ClaimGeneration || len(repair.FindingRefs) == 0 {
		return false
	}

	if !slices.Contains([]string{"combined_gate_regression", "actionable_review_finding"}, repair.Reason) {
		return false
	}

	if !slices.Contains(repair.FindingRefs, b.FindingRef) {
		return false
	}

	return true
}
