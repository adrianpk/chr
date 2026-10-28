package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// EnsureWorkflowSlicePullRequest creates or reuses the slice's shared provider PR.
func (s service) EnsureWorkflowSlicePullRequest(ctx context.Context, id string, request workflowcore.SlicePullRequestRequest) (workflowstate.PullRequestSnapshot, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	provider := s.workflowPullRequests
	if provider == nil {
		config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
		if err != nil {
			return workflowstate.PullRequestSnapshot{}, err
		}

		if err := core.ValidateReviewWorkflowConfig(config.Review); err != nil {
			return workflowstate.PullRequestSnapshot{}, err
		}

		provider = workflowcore.NewRemoteReviewAdapter(config.Review.Providers, nil)
	}

	return store.EnsureWorkflowSlicePullRequest(ctx, id, request, provider)
}
