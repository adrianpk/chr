package core

import (
	"context"
	"time"
)

// WorkflowInspection observes retained authorization and current task records.
// Adapters project it without exposing process identities or owner credentials.
type WorkflowInspection struct {
	Invocation WorkflowInvocation
	Policy     WorkflowReviewPolicy
	Tasks      []WorkflowRecord
}

// InspectWorkflow does not replay transitions, reconcile ownership or admit work.
func (s *WorkflowStore) InspectWorkflow(ctx context.Context, id string) (result WorkflowInspection, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		result.Invocation = run.WorkflowInvocation

		result.Policy, readErr = owner.workflowReviewPolicy(run)
		if readErr != nil {
			return readErr
		}

		snapshot, readErr := owner.readSnapshot(ctx)
		result.Tasks = snapshot.Records

		return readErr
	})

	return result, err
}

// InspectWorkflowPolicy reports current publication policy without changing state.
func (s *WorkflowStore) InspectWorkflowPolicy(ctx context.Context) (policy WorkflowReviewPolicy, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		var readErr error

		policy, readErr = s.workflowReviewPolicy(workflowInvocationFile{})

		return readErr
	})

	return policy, err
}
