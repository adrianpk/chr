package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// ReviewWorkflowSlice uses the shared policy, evidence and supervised role path.
func (s service) ReviewWorkflowSlice(ctx context.Context, id string, reviewer []string) (workflowstate.ReviewBinding, error) {
	store, provider, err := s.workflowReviewDependencies(id)
	if err != nil {
		return workflowstate.ReviewBinding{}, err
	}

	return store.ReviewWorkflowSlice(ctx, id, provider, reviewer)
}

// ApproveWorkflowSlice binds explicit approval to the current displayed hash.
func (s service) ApproveWorkflowSlice(ctx context.Context, id, bindingHash string, expected ...uint64) (workflowstate.ReviewBinding, error) {
	store, provider, err := s.workflowReviewDependencies(id)
	if err != nil {
		return workflowstate.ReviewBinding{}, err
	}

	return store.ApproveWorkflowSlice(ctx, id, bindingHash, provider, expected...)
}

// CheckWorkflowReviewGrant rechecks authorization for a guarded publication.
func (s service) CheckWorkflowReviewGrant(ctx context.Context, id string) (workflowstate.ReviewInput, error) {
	store, provider, err := s.workflowReviewDependencies(id)
	if err != nil {
		return workflowstate.ReviewInput{}, err
	}

	return store.CheckWorkflowReviewGrant(ctx, id, provider)
}

func (s service) workflowReviewDependencies(id string) (*workflowcore.Store, workflowcore.PullRequestProvider, error) {
	store, err := s.workflowStore()
	if err != nil {
		return nil, nil, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return nil, nil, err
	}

	if remote, ok := provider.(workflowcore.RemoteReviewAdapter); ok {
		config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
		if err != nil {
			return nil, nil, err
		}

		if config.Work.Integration.PublicationProfile == workspaceconfig.WorkflowPublicationGitAtomicDualRef {
			provider = workflowcore.NewWorkflowNativeReviewAdapter(remote, store, id, s.env.ConfigDir())
		}
	}

	return store, provider, nil
}
