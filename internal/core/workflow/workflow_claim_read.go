package workflow

import "context"

// Credential checks hold repository exclusion and have already prepared reads.
// An exact owner needs no repository-wide search. Other roots and project
// aliases retain normal scope resolution; no credential or input check is skipped.
func (s *Store) locateWorkflowClaim(ctx context.Context, claim *Claim) (workflowInvocationFile, *Store, error) {
	if claim.store == nil || claim.store.layout.Root() != s.layout.Root() || claim.store.projectID != s.projectID {
		return s.locateWorkflowInvocation(ctx, claim.InvocationID)
	}

	run, err := s.readWorkflowInvocation(claim.InvocationID)
	if err != nil {
		return run, s, err
	}

	if run.OwnerRoot != s.layout.Root() || run.ProjectID != s.projectID {
		return run, s, ErrWorkflowOwned
	}

	return run, s, nil
}
