package app

import (
	"context"
	"fmt"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// PrepareWorkflowSliceIntegration runs the shared guarded combined-result gate.
func (s service) PrepareWorkflowSliceIntegration(ctx context.Context, id string) (workflowcore.IntegrationCandidate, error) {
	store, provider, err := s.workflowPreparationDependencies(id)
	if err != nil {
		return workflowcore.IntegrationCandidate{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowcore.IntegrationCandidate{}, err
	}

	git := s.workflowPreparationGit
	if git == nil {
		git = workflowcore.NewWorkflowPreparationGit(run.Selection.Target.Repository, run.Selection.Target.Repository)
	}

	return store.PrepareWorkflowSliceIntegration(ctx, id, provider, git)
}

// CheckWorkflowPublicationGuard returns exact current authority for submission.
func (s service) CheckWorkflowPublicationGuard(ctx context.Context, id string) (workflowcore.PublicationGuard, error) {
	store, provider, err := s.workflowPreparationDependencies(id)
	if err != nil {
		return workflowcore.PublicationGuard{}, err
	}

	return store.CheckWorkflowPublicationGuard(ctx, id, provider)
}

func (s service) workflowPreparationDependencies(id string) (*workflowcore.Store, workflowcore.PublicationGuardProvider, error) {
	store, provider, err := s.workflowReviewDependencies(id)
	if err != nil {
		return nil, nil, err
	}

	guarded, ok := provider.(workflowcore.PublicationGuardProvider)
	if !ok {
		return nil, nil, fmt.Errorf("%w: configured provider has no guarded integration contract", workflowcore.ErrWorkflowAtomicGuard)
	}

	return store, guarded, nil
}
