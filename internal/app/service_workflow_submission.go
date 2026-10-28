package app

import (
	"context"
	"fmt"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SubmitWorkflowSliceIntegration invokes the shared durable publication protocol.
func (s service) SubmitWorkflowSliceIntegration(ctx context.Context, id string) (workflowstate.Submission, error) {
	store, provider, run, err := s.workflowSubmissionDependencies(ctx, id)
	if err != nil {
		return workflowstate.Submission{}, err
	}

	return store.SubmitWorkflowSliceIntegration(ctx, id, provider, s.workflowIntegrationObserver(run))
}

// ReconcileWorkflowSubmission observes a retained submission without retrying it.
func (s service) ReconcileWorkflowSubmission(ctx context.Context, id string) (workflowstate.Submission, error) {
	store, provider, run, err := s.workflowSubmissionDependencies(ctx, id)
	if err != nil {
		return workflowstate.Submission{}, err
	}

	if run.Integration != nil {
		ports, err := s.workflowContinuationPorts(run, provider, nil)
		if err != nil {
			return workflowstate.Submission{}, err
		}

		return store.ReconcileWorkflowNativeIntegration(ctx, id, ports.Publication)
	}

	return store.ReconcileWorkflowSubmission(ctx, id, provider, s.workflowIntegrationObserver(run))
}

// ReadWorkflowSubmission reports the common-directory submission checkpoint.
func (s service) ReadWorkflowSubmission(ctx context.Context, id string) (workflowstate.Submission, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.Submission{}, err
	}

	return store.ReadWorkflowSubmission(ctx, id)
}

func (s service) workflowSubmissionDependencies(ctx context.Context, id string) (*workflowcore.Store, workflowcore.PublicationProvider, workflowcore.Invocation, error) {
	store, guarded, err := s.workflowPreparationDependencies(id)
	if err != nil {
		return nil, nil, workflowcore.Invocation{}, err
	}

	provider, ok := guarded.(workflowcore.PublicationProvider)
	if !ok {
		return nil, nil, workflowcore.Invocation{}, fmt.Errorf("%w: configured provider has no durable publication contract", workflowcore.ErrWorkflowAtomicGuard)
	}

	run, err := store.ReadWorkflow(ctx, id)

	return store, provider, run, err
}
