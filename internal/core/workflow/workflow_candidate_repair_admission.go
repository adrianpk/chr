package workflow

import (
	"context"
	"slices"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowCandidateRepairBinding(run workflowInvocationFile) error {
	b := run.CandidateRepair
	if b == nil {
		return nil
	}

	if !slices.Contains([]string{"slice", "ticket"}, b.Unit.Kind) || b.Unit.ID == "" || !workflowstate.ValidFingerprint(b.EntryID) || b.ClaimGeneration == 0 || !workflowstate.ValidFingerprint(b.CandidateHash) {
		return workflowstate.ErrConflict
	}

	for _, ref := range []workflowstate.EvidenceRef{b.RecordRef, b.FindingRef} {
		if workflowstate.CheckEvidenceRef(run.ID, ref) != nil {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func (s *Store) checkWorkflowCandidateRepairAdmission(ctx context.Context, run workflowInvocationFile, entry IntegrationEntry, input workflowstate.MergerInput, provider PullRequestProvider) error {
	b := run.CandidateRepair
	if b == nil {
		return nil
	}

	if !workflowCandidateRepairFenceMatches(*b, run, entry, input) {
		return workflowstate.ErrConflict
	}

	repair, err := s.readWorkflowCandidateRepair(run)
	if err != nil {
		return err
	}

	if input.Reason != repair.Reason || !workflowmerge.RepairMatches(input, repair.Candidate.Input.Git) {
		return workflowstate.ErrConflict
	}

	for _, ref := range repair.FindingRefs {
		if !slices.Contains(input.FindingRefs, ref) {
			return workflowstate.ErrConflict
		}
	}

	guard, ok := provider.(PublicationGuardProvider)
	if !ok {
		return workflowstate.ErrAtomicGuard
	}

	return s.checkWorkflowCandidateFailureSource(ctx, run, repair.Candidate, guard)
}

func (s *Store) checkWorkflowCandidateFailureSource(ctx context.Context, run workflowInvocationFile, candidate workflowstate.IntegrationCandidate, provider PublicationGuardProvider) error {
	_, source, err := s.workflowReviewSourceInputs(ctx, run, provider)
	if err != nil || workflowreview.CheckWorkflowReviewInputHash(source, run.Preparation.SourceHash) != nil {
		return errorsOrConflict(err)
	}

	capability, err := observeWorkflowPreparedCapability(ctx, provider, candidate.Input)
	if err != nil || capability != candidate.Input.Capability {
		return errorsOrConflict(err)
	}

	_, source, err = s.workflowReviewSourceInputs(ctx, run, provider)
	if err != nil || workflowreview.CheckWorkflowReviewInputHash(source, run.Preparation.SourceHash) != nil {
		return errorsOrConflict(err)
	}

	return s.checkWorkflowPreparationGit(ctx, candidate)
}

func workflowCandidateRepairFenceMatches(b CandidateRepairBinding, run workflowInvocationFile, entry IntegrationEntry, input workflowstate.MergerInput) bool {
	return b.EntryID == entry.ID && b.ClaimGeneration == entry.ClaimGeneration && b.CandidateHash == workflowMergerCandidateHash(run, entry) && input.RepairCheckpointRef != nil
}
