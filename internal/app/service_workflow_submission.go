package app

import (
	"context"
	"fmt"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// SubmitWorkflowSliceIntegration invokes the shared durable publication protocol.
func (s service) SubmitWorkflowSliceIntegration(ctx context.Context, id string) (workflowstate.Submission, error) {
	submissionDependencies, err := s.workflowSubmissionDependencies(ctx, id)
	store, provider, run := submissionDependencies.Store, submissionDependencies.Provider, submissionDependencies.Invocation

	if err != nil {
		return workflowstate.Submission{}, err
	}

	return store.SubmitWorkflowSliceIntegration(ctx, id, provider, s.workflowIntegrationObserver(run))
}

// ReconcileWorkflowSubmission observes a retained submission without retrying it.
func (s service) ReconcileWorkflowSubmission(ctx context.Context, id string) (workflowstate.Submission, error) {
	submissionDependencies2, err := s.workflowSubmissionDependencies(ctx, id)
	store, provider, run := submissionDependencies2.Store, submissionDependencies2.Provider, submissionDependencies2.Invocation

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

type workflowSubmissionDependencies struct {
	Store      *workflowcore.Store
	Provider   workflowcore.PublicationProvider
	Invocation workflowstate.Invocation
}

func (s service) workflowSubmissionDependencies(ctx context.Context, id string) (workflowSubmissionDependencies, error) {
	preparationDependencies, err := s.workflowPreparationDependencies(id)

	store, guarded := preparationDependencies.Store, preparationDependencies.Provider

	if err != nil {
		return workflowSubmissionDependencies{Store: nil, Provider: nil, Invocation: workflowstate.Invocation{}}, err
	}

	provider, ok := guarded.(workflowcore.PublicationProvider)
	if !ok {
		return workflowSubmissionDependencies{Store: nil, Provider: nil, Invocation: workflowstate.Invocation{}}, fmt.Errorf("%w: configured provider has no durable publication contract", workflowstate.ErrAtomicGuard)
	}

	run, err := store.ReadWorkflow(ctx, id)

	return workflowSubmissionDependencies{Store: store, Provider: provider, Invocation: run}, err
}
