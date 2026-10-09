package core

import (
	"context"
	"fmt"
	"slices"
)

func (s *WorkflowStore) observeWorkflowAmendment(ctx context.Context, request WorkflowSetAmendmentRequest) (WorkflowRecord, []workflowInvocationFile, error) {
	set, err := s.readRecord(request.SetPath)
	if err != nil {
		return set, nil, err
	}

	if !s.workflowAmendmentImageMatches(set, request) {
		return set, nil, ErrWorkflowConflict
	}

	if err := s.checkWorkflowAmendmentPlan(set, request); err != nil {
		return set, nil, err
	}

	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return set, nil, err
	}

	for _, run := range runs {
		if run.Namespace != set.Namespace || !workflowClaimResources(run)["set:"+set.ID] {
			continue
		}

		if err := s.checkWorkflowAmendmentOwner(ctx, run); err != nil {
			return set, nil, err
		}
	}

	return set, runs, nil
}

func (s *WorkflowStore) workflowAmendmentImageMatches(set WorkflowRecord, request WorkflowSetAmendmentRequest) bool {
	return set.Set != nil && set.Revision == request.ExpectedRevision && set.SourceHash == request.ExpectedSourceHash && set.ProjectID == s.projectID && set.Namespace == s.layout.Namespace()
}

func (s *WorkflowStore) checkWorkflowAmendmentPlan(set WorkflowRecord, request WorkflowSetAmendmentRequest) error {
	plan := request.Plan
	if set.Set.PlanHash == plan.PlanHash || set.Set.UseCase != plan.UseCase || set.Set.UseCaseOrder != plan.UseCaseOrder {
		return fmt.Errorf("%w: changed plan must preserve set association and order", ErrWorkflowConflict)
	}

	if string(set.Metadata["materialization_id"]) != fmt.Sprintf("%q", plan.MaterializationID) {
		return fmt.Errorf("%w: amendment changed materialization identity", ErrWorkflowConflict)
	}

	if request.Reopen != (set.Set.Status == "completed") {
		return fmt.Errorf("completed sets require explicit reopening; active sets require amendment")
	}

	if len(set.Set.Amendments) >= WorkflowRecordLimit || len(set.Set.ClosureHistory) >= WorkflowRecordLimit {
		return fmt.Errorf("set amendment or closure history bound exhausted")
	}

	return s.checkWorkflowSetAmendmentEvidence(set)
}

func (s *WorkflowStore) checkWorkflowAmendmentOwner(ctx context.Context, run workflowInvocationFile) error {
	for _, attempt := range run.Attempts {
		if slices.Contains([]string{"running", "checkpointed"}, attempt.Outcome) {
			return ErrWorkflowOwned
		}

		if err := workflowAttemptQuiescent(attempt); err != nil {
			return err
		}
	}

	if workflowInvocationTerminal(run.Status) {
		return nil
	}

	if run.OwnerRoot != s.layout.Root() {
		return fmt.Errorf("%w: amend in the invocation's owning checkout", ErrWorkflowOwned)
	}

	if run.Lane != nil || run.Submission != nil {
		return fmt.Errorf("%w: advance or reconcile the retained unit before amendment", ErrWorkflowOwned)
	}

	owner, err := s.workflowOwnerStore(run)
	if err != nil {
		return err
	}

	return owner.checkWorkflowInputs(ctx, run)
}

func (s *WorkflowStore) checkWorkflowSetAmendmentEvidence(set WorkflowRecord) error {
	for _, binding := range set.Set.Amendments {
		if err := s.checkWorkflowAmendmentEvidence(set, binding); err != nil {
			return err
		}
	}

	if set.Set.ClosureRef != "" {
		path, err := s.safePath(set.Set.ClosureRef)
		if err != nil {
			return err
		}

		if _, err := readWorkflowFile(path); err != nil {
			return err
		}
	}

	return nil
}
