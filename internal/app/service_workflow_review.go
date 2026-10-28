package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// ReviewWorkflowSlice uses the shared policy, evidence and supervised role path.
func (s service) ReviewWorkflowSlice(ctx context.Context, id string, reviewer []string) (workflowstate.ReviewBinding, error) {
	dependencies, err := s.workflowReviewDependencies(id)
	if err != nil {
		return workflowstate.ReviewBinding{}, err
	}

	return dependencies.Store.ReviewWorkflowSlice(ctx, id, dependencies.Provider, reviewer)
}

// ApproveWorkflowSlice binds explicit approval to the current displayed hash.
func (s service) ApproveWorkflowSlice(ctx context.Context, id, bindingHash string, expected ...uint64) (workflowstate.ReviewBinding, error) {
	dependencies, err := s.workflowReviewDependencies(id)
	if err != nil {
		return workflowstate.ReviewBinding{}, err
	}

	return dependencies.Store.ApproveWorkflowSlice(ctx, id, bindingHash, dependencies.Provider, expected...)
}

// CheckWorkflowReviewGrant rechecks authorization for a guarded publication.
func (s service) CheckWorkflowReviewGrant(ctx context.Context, id string) (workflowstate.ReviewInput, error) {
	dependencies, err := s.workflowReviewDependencies(id)
	if err != nil {
		return workflowstate.ReviewInput{}, err
	}

	return dependencies.Store.CheckWorkflowReviewGrant(ctx, id, dependencies.Provider)
}

type workflowReviewDependencies struct {
	Store    *workflowcore.Store
	Provider workflowcore.PullRequestProvider
}

func (s service) workflowReviewDependencies(id string) (result workflowReviewDependencies, err error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowReviewDependencies{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return workflowReviewDependencies{}, err
	}

	if remote, ok := provider.(workflowcore.RemoteReviewAdapter); ok {
		config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
		if err != nil {
			return workflowReviewDependencies{}, err
		}

		if config.Work.Integration.PublicationProfile == workspaceconfig.WorkflowPublicationGitAtomicDualRef {
			provider = workflowcore.NewWorkflowNativeReviewAdapter(remote, store, id, s.env.ConfigDir())
		}
	}

	return workflowReviewDependencies{Store: store, Provider: provider}, nil
}
