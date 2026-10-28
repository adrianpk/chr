package workflow

import (
	"context"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// A live claim already identifies its canonical owner. Poll that owner's
// invocation under the same repository exclusion and read preparation as a
// public read, without enumerating unrelated invocations in every source lane.
// Lost credentials stop execution; this read cannot authorize a new effect.
func (s *Store) checkWorkflowClaimStop(ctx context.Context, claim *Claim) error {
	// Publication and closing observers can carry only an invocation identity;
	// without a claimed owner they retain the public scope-resolution path.
	if claim.store == nil {
		run, err := s.ReadWorkflow(ctx, claim.InvocationID)
		if err != nil {
			return err
		}

		if run.StopIntent != "" {
			return workflowstate.ErrStop
		}

		return nil
	}

	return checkWorkflowOwnedStop(ctx, claim)
}

func checkWorkflowOwnedStop(ctx context.Context, claim *Claim) error {
	owner := claim.store

	return owner.withRepositoryLock(ctx, func() error { return owner.prepareWorkflowRead(ctx) }, func() error {
		run, err := owner.readWorkflowInvocation(claim.InvocationID)
		if err != nil {
			return err
		}

		if run.OwnerRoot != owner.layout.Root() || run.ProjectID != owner.projectID || len(run.Attempts) == 0 {
			return workflowstate.ErrOwned
		}

		attempt := run.Attempts[len(run.Attempts)-1]
		if !workflowAttemptCredentialMatches(run, claim, attempt) {
			return workflowstate.ErrOwned
		}

		if run.StopIntent != "" {
			return workflowstate.ErrStop
		}

		return nil
	})
}
