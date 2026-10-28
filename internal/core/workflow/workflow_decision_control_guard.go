package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowDecisionControlGuard(ctx context.Context, run workflowInvocationFile, request workflowstate.ControlMutationRequest) (IntegrationBudgetGuard, error) {
	if err := checkWorkflowRevision(run.Invocation, []uint64{request.ExpectedRevision}); err != nil {
		return IntegrationBudgetGuard{}, err
	}

	if err := checkWorkflowControlReview(run, request.ReviewHash); err != nil {
		return IntegrationBudgetGuard{}, err
	}

	if request.Integration == nil {
		if run.Integration != nil {
			return IntegrationBudgetGuard{}, workflowstate.ErrConflict
		}

		return IntegrationBudgetGuard{}, nil
	}

	observed, err := s.inspectWorkflowIntegration(ctx, run)
	if err != nil {
		return IntegrationBudgetGuard{}, err
	}

	return workflowDisplayedIntegrationGuard(run, observed, *request.Integration)
}

func workflowDisplayedIntegrationGuard(run workflowInvocationFile, observed *IntegrationInspection, shown workflowstate.ControlIntegrationGuard) (IntegrationBudgetGuard, error) {
	if observed == nil || observed.Entry == nil {
		return IntegrationBudgetGuard{}, workflowstate.ErrConflict
	}

	entry := observed.Entry
	if shown.EntryID != entry.ID || shown.EntryRevision != entry.Revision || shown.BudgetRevision != observed.Budget.Revision || shown.CandidateHash != IntegrationDisplayedCandidate(observed) {
		return IntegrationBudgetGuard{}, workflowstate.ErrConflict
	}

	return IntegrationBudgetGuard{InvocationID: run.ID, InvocationRevision: run.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, BudgetRevision: observed.Budget.Revision, ClaimGeneration: entry.ClaimGeneration}, nil
}

// IntegrationDisplayedCandidate is the immutable candidate named by
// operator controls, including a retained question's pre-publication candidate.
func IntegrationDisplayedCandidate(observed *IntegrationInspection) string {
	if observed == nil {
		return ""
	}

	if observed.Candidate != nil && observed.Candidate.Native != nil {
		return observed.Candidate.Native.CandidateHash
	}

	if observed.Candidate == nil && observed.Decisions != nil {
		return observed.Decisions.Input.CandidateHash
	}

	if observed.Entry != nil && observed.Entry.CandidateRef != nil {
		return observed.Entry.CandidateRef.Hash
	}

	return ""
}

func checkWorkflowControlReview(run workflowInvocationFile, expected string) error {
	reviewHash := ""
	if run.Review != nil {
		reviewHash = run.Review.InputHash
	}

	if expected != reviewHash {
		return workflowstate.ErrConflict
	}

	return nil
}
