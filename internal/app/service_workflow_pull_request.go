package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// EnsureWorkflowSlicePullRequest creates or reuses the slice's shared provider PR.
func (s service) EnsureWorkflowSlicePullRequest(ctx context.Context, id string, request core.WorkflowSlicePullRequestRequest) (core.WorkflowPullRequestSnapshot, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowPullRequestSnapshot{}, err
	}

	provider := s.workflowPullRequests
	if provider == nil {
		config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
		if err != nil {
			return core.WorkflowPullRequestSnapshot{}, err
		}

		if err := core.ValidateReviewWorkflowConfig(config.Review); err != nil {
			return core.WorkflowPullRequestSnapshot{}, err
		}

		provider = core.NewRemoteReviewAdapter(config.Review.Providers, nil)
	}

	return store.EnsureWorkflowSlicePullRequest(ctx, id, request, provider)
}
