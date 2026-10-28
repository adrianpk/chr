package workflow

import (
	"context"
	"fmt"
	"slices"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowAmendmentObservation struct {
	Set         workflowstate.Record
	Invocations []workflowInvocationFile
}

func (s *Store) observeWorkflowAmendment(ctx context.Context, request SetAmendmentRequest) (workflowAmendmentObservation, error) {
	set, err := s.readRecord(request.SetPath)
	if err != nil {
		return workflowAmendmentObservation{Set: set}, err
	}

	if !s.workflowAmendmentImageMatches(set, request) {
		return workflowAmendmentObservation{Set: set}, workflowstate.ErrConflict
	}

	if err := s.checkWorkflowAmendmentPlan(set, request); err != nil {
		return workflowAmendmentObservation{Set: set}, err
	}

	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return workflowAmendmentObservation{Set: set}, err
	}

	for _, run := range runs {
		if run.Namespace != set.Namespace || !workflowClaimResources(run)["set:"+set.ID] {
			continue
		}

		if err := s.checkWorkflowAmendmentOwner(ctx, run); err != nil {
			return workflowAmendmentObservation{Set: set}, err
		}
	}

	return workflowAmendmentObservation{Set: set, Invocations: runs}, nil
}

func (s *Store) workflowAmendmentImageMatches(set workflowstate.Record, request SetAmendmentRequest) bool {
	return set.Set != nil && set.Revision == request.ExpectedRevision && set.SourceHash == request.ExpectedSourceHash && set.ProjectID == s.projectID && set.Namespace == s.layout.Namespace()
}

func (s *Store) checkWorkflowAmendmentPlan(set workflowstate.Record, request SetAmendmentRequest) error {
	if set.Set.Status == "closing" {
		return ErrWorkflowOwned
	}

	plan := request.Plan
	if set.Set.PlanHash == plan.PlanHash || set.Set.UseCase != plan.UseCase || set.Set.UseCaseOrder != plan.UseCaseOrder {
		return fmt.Errorf("%w: changed plan must preserve set association and order", workflowstate.ErrConflict)
	}

	if string(set.Metadata["materialization_id"]) != fmt.Sprintf("%q", plan.MaterializationID) {
		return fmt.Errorf("%w: amendment changed materialization identity", workflowstate.ErrConflict)
	}

	if request.Reopen != (set.Set.Status == "completed") {
		return fmt.Errorf("completed sets require explicit reopening; active sets require amendment")
	}

	if len(set.Set.Amendments) >= workflowstate.RecordLimit || len(set.Set.ClosureHistory) >= workflowstate.RecordLimit {
		return fmt.Errorf("set amendment or closure history bound exhausted")
	}

	return s.checkWorkflowSetAmendmentEvidence(set)
}

func (s *Store) checkWorkflowAmendmentOwner(ctx context.Context, run workflowInvocationFile) error {
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

func (s *Store) checkWorkflowSetAmendmentEvidence(set workflowstate.Record) error {
	if err := workflowcloseout.CheckHistory(set.ID, set.Set.Closing, workflowClosingEvidence{store: s}); err != nil {
		return err
	}

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

		if _, err := opsfile.Read(path); err != nil {
			return err
		}
	}

	return nil
}
