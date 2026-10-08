package app

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
)

// PrepareWorkflowSliceIntegration runs the shared guarded combined-result gate.
func (s service) PrepareWorkflowSliceIntegration(ctx context.Context, id string) (core.WorkflowIntegrationCandidate, error) {
	store, provider, err := s.workflowPreparationDependencies()
	if err != nil {
		return core.WorkflowIntegrationCandidate{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return core.WorkflowIntegrationCandidate{}, err
	}

	git := s.workflowPreparationGit
	if git == nil {
		git = core.NewWorkflowPreparationGit(run.Selection.Target.Repository, run.Selection.Target.Repository)
	}

	return store.PrepareWorkflowSliceIntegration(ctx, id, provider, git)
}

// CheckWorkflowPublicationGuard returns exact current authority for submission.
func (s service) CheckWorkflowPublicationGuard(ctx context.Context, id string) (core.WorkflowPublicationGuard, error) {
	store, provider, err := s.workflowPreparationDependencies()
	if err != nil {
		return core.WorkflowPublicationGuard{}, err
	}

	return store.CheckWorkflowPublicationGuard(ctx, id, provider)
}

func (s service) workflowPreparationDependencies() (*core.WorkflowStore, core.WorkflowPublicationGuardProvider, error) {
	store, provider, err := s.workflowReviewDependencies()
	if err != nil {
		return nil, nil, err
	}

	guarded, ok := provider.(core.WorkflowPublicationGuardProvider)
	if !ok {
		return nil, nil, fmt.Errorf("%w: configured provider has no guarded integration contract", core.ErrWorkflowAtomicGuard)
	}

	return store, guarded, nil
}
