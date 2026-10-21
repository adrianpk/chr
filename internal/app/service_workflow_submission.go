package app

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
)

// SubmitWorkflowSliceIntegration invokes the shared durable publication protocol.
func (s service) SubmitWorkflowSliceIntegration(ctx context.Context, id string) (core.WorkflowSubmission, error) {
	store, provider, run, err := s.workflowSubmissionDependencies(ctx, id)
	if err != nil {
		return core.WorkflowSubmission{}, err
	}

	return store.SubmitWorkflowSliceIntegration(ctx, id, provider, s.workflowIntegrationObserver(run))
}

// ReconcileWorkflowSubmission observes a retained submission without retrying it.
func (s service) ReconcileWorkflowSubmission(ctx context.Context, id string) (core.WorkflowSubmission, error) {
	store, provider, run, err := s.workflowSubmissionDependencies(ctx, id)
	if err != nil {
		return core.WorkflowSubmission{}, err
	}

	return store.ReconcileWorkflowSubmission(ctx, id, provider, s.workflowIntegrationObserver(run))
}

// ReadWorkflowSubmission reports the common-directory submission checkpoint.
func (s service) ReadWorkflowSubmission(ctx context.Context, id string) (core.WorkflowSubmission, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowSubmission{}, err
	}

	return store.ReadWorkflowSubmission(ctx, id)
}

func (s service) workflowSubmissionDependencies(ctx context.Context, id string) (*core.WorkflowStore, core.WorkflowPublicationProvider, core.WorkflowInvocation, error) {
	store, guarded, err := s.workflowPreparationDependencies(id)
	if err != nil {
		return nil, nil, core.WorkflowInvocation{}, err
	}

	provider, ok := guarded.(core.WorkflowPublicationProvider)
	if !ok {
		return nil, nil, core.WorkflowInvocation{}, fmt.Errorf("%w: configured provider has no durable publication contract", core.ErrWorkflowAtomicGuard)
	}

	run, err := store.ReadWorkflow(ctx, id)

	return store, provider, run, err
}
