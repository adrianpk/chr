package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// ReviewWorkflowSlice uses the shared policy, evidence and supervised role path.
func (s service) ReviewWorkflowSlice(ctx context.Context, id string, reviewer []string) (core.WorkflowReviewBinding, error) {
	store, provider, err := s.workflowReviewDependencies(id)
	if err != nil {
		return core.WorkflowReviewBinding{}, err
	}

	return store.ReviewWorkflowSlice(ctx, id, provider, reviewer)
}

// ApproveWorkflowSlice binds explicit approval to the current displayed hash.
func (s service) ApproveWorkflowSlice(ctx context.Context, id, bindingHash string, expected ...uint64) (core.WorkflowReviewBinding, error) {
	store, provider, err := s.workflowReviewDependencies(id)
	if err != nil {
		return core.WorkflowReviewBinding{}, err
	}

	return store.ApproveWorkflowSlice(ctx, id, bindingHash, provider, expected...)
}

// CheckWorkflowReviewGrant rechecks authorization for a guarded publication.
func (s service) CheckWorkflowReviewGrant(ctx context.Context, id string) (core.WorkflowReviewInput, error) {
	store, provider, err := s.workflowReviewDependencies(id)
	if err != nil {
		return core.WorkflowReviewInput{}, err
	}

	return store.CheckWorkflowReviewGrant(ctx, id, provider)
}

func (s service) workflowReviewDependencies(id string) (*core.WorkflowStore, core.WorkflowPullRequestProvider, error) {
	store, err := s.workflowStore()
	if err != nil {
		return nil, nil, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return nil, nil, err
	}

	if remote, ok := provider.(core.RemoteReviewAdapter); ok {
		config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
		if err != nil {
			return nil, nil, err
		}

		if config.Work.Integration.PublicationProfile == core.WorkflowPublicationGitAtomicDualRef {
			provider = core.NewWorkflowNativeReviewAdapter(remote, store, id, s.env.ConfigDir())
		}
	}

	return store, provider, nil
}
