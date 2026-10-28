package workflow

import (
	"context"
	"slices"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CandidateRepairBinding retains actionable combined-result evidence.
// It cannot authorize publication or rewind accepted source task execution.
type CandidateRepairBinding struct {
	Unit            workflowstate.Ref         `json:"unit"`
	EntryID         string                    `json:"entry_id"`
	ClaimGeneration uint64                    `json:"claim_generation"`
	CandidateHash   string                    `json:"candidate_hash"`
	FindingRef      workflowstate.EvidenceRef `json:"finding_ref"`
	RecordRef       workflowstate.EvidenceRef `json:"record_ref"`
}

// CandidateRepair records the host result and its exact findings before
// a bounded merger may act. Failed gate results remain non-candidates.
type CandidateRepair struct {
	Schema      string                             `json:"schema"`
	Reason      string                             `json:"reason"`
	Candidate   workflowstate.IntegrationCandidate `json:"candidate"`
	FindingRefs []workflowstate.EvidenceRef        `json:"finding_refs"`
}

// ReadWorkflowCandidateRepair returns retained findings for host request
// assembly. Reading never reserves a role or advances the queue.
func (s *Store) ReadWorkflowCandidateRepair(ctx context.Context, id string) (repair CandidateRepair, err error) {
	err = s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		repair, err = invocationLocation.Owner.readWorkflowCandidateRepair(invocationLocation.Run)

		return err
	})

	return repair, err
}

func (s *Store) repairWorkflowReviewedSlice(ctx context.Context, id, key string, provider PullRequestProvider) error {
	native := false

	err := s.withLock(ctx, func() error {
		invocationLocation2, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return err
		}

		reviewRecord, err := invocationLocation2.Owner.readWorkflowReview(invocationLocation2.Run)

		input := reviewRecord.Input
		if err != nil || input.Native == nil {
			return err
		}

		native = true

		return invocationLocation2.Owner.repairWorkflowNativeReview(ctx, invocationLocation2.Run, key, provider)
	})
	if err != nil || native {
		return err
	}

	_, err = s.RepairWorkflowSlice(ctx, id, key)

	return err
}

func (s *Store) repairWorkflowNativeReview(ctx context.Context, run workflowInvocationFile, key string, provider PullRequestProvider) error {
	if run.Review.State != "needs_repair" || run.Review.RepairTaskKey != key {
		return workflowstate.ErrConflict
	}

	if run.CandidateRepair != nil && workflowCandidateRepairFindingMatches(*run.CandidateRepair, run.Review) {
		_, err := s.readWorkflowCandidateRepair(run)

		return err
	}

	reviewInputs, err := s.workflowReviewInputs(ctx, run, provider)

	fresh := reviewInputs.Input
	if err != nil || workflowreview.CheckInputHash(fresh, run.Review.InputHash) != nil {
		return errorsOrConflict(err)
	}

	candidate, err := s.readWorkflowPreparation(run)
	if err != nil {
		return err
	}

	ref := workflowstate.EvidenceRef{Path: run.Review.ResultRef, Hash: run.Review.ResultHash}

	return s.retainWorkflowCandidateRepair(ctx, &run, candidate, "actionable_review_finding", []workflowstate.EvidenceRef{ref})
}

func workflowCandidateRepairFindingMatches(b CandidateRepairBinding, review *workflowstate.ReviewBinding) bool {
	return review != nil && b.Unit == review.Unit && b.FindingRef == (workflowstate.EvidenceRef{Path: review.ResultRef, Hash: review.ResultHash})
}

func (s *Store) readWorkflowCandidateRepair(run workflowInvocationFile) (CandidateRepair, error) {
	var repair CandidateRepair

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

func validWorkflowCandidateRepairRecord(run workflowInvocationFile, repair CandidateRepair) bool {
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
