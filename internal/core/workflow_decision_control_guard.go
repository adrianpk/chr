package core

import "context"

func (s *WorkflowStore) workflowDecisionControlGuard(ctx context.Context, run workflowInvocationFile, request WorkflowControlMutationRequest) (WorkflowIntegrationBudgetGuard, error) {
	if err := checkWorkflowRevision(run.WorkflowInvocation, []uint64{request.ExpectedRevision}); err != nil {
		return WorkflowIntegrationBudgetGuard{}, err
	}

	if err := checkWorkflowControlReview(run, request.ReviewHash); err != nil {
		return WorkflowIntegrationBudgetGuard{}, err
	}

	if request.Integration == nil {
		if run.Integration != nil {
			return WorkflowIntegrationBudgetGuard{}, ErrWorkflowConflict
		}

		return WorkflowIntegrationBudgetGuard{}, nil
	}

	observed, err := s.inspectWorkflowIntegration(ctx, run)
	if err != nil {
		return WorkflowIntegrationBudgetGuard{}, err
	}

	return workflowDisplayedIntegrationGuard(run, observed, *request.Integration)
}

func workflowDisplayedIntegrationGuard(run workflowInvocationFile, observed *WorkflowIntegrationInspection, shown WorkflowControlIntegrationGuard) (WorkflowIntegrationBudgetGuard, error) {
	if observed == nil || observed.Entry == nil {
		return WorkflowIntegrationBudgetGuard{}, ErrWorkflowConflict
	}

	entry := observed.Entry
	if shown.EntryID != entry.ID || shown.EntryRevision != entry.Revision || shown.BudgetRevision != observed.Budget.Revision || shown.CandidateHash != WorkflowIntegrationDisplayedCandidate(observed) {
		return WorkflowIntegrationBudgetGuard{}, ErrWorkflowConflict
	}

	return WorkflowIntegrationBudgetGuard{InvocationID: run.ID, InvocationRevision: run.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, BudgetRevision: observed.Budget.Revision, ClaimGeneration: entry.ClaimGeneration}, nil
}

// WorkflowIntegrationDisplayedCandidate is the immutable candidate named by
// operator controls, including a retained question's pre-publication candidate.
func WorkflowIntegrationDisplayedCandidate(observed *WorkflowIntegrationInspection) string {
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
		return ErrWorkflowConflict
	}

	return nil
}
