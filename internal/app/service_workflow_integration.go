package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// VerifyWorkflowSliceIntegration checks provider and remote Git delivery facts.
func (s service) VerifyWorkflowSliceIntegration(ctx context.Context, id, method string) (workflowstate.SliceDeliveryProof, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return workflowstate.SliceDeliveryProof{}, err
	}

	return store.VerifyWorkflowSliceIntegration(ctx, id, method, provider, s.workflowIntegrationObserver(run))
}

func (s service) workflowIntegrationProvider() (workflowcore.PullRequestProvider, error) {
	provider := s.workflowPullRequests
	if provider == nil {
		config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
		if err != nil {
			return nil, err
		}

		if err := core.ValidateReviewWorkflowConfig(config.Review); err != nil {
			return nil, err
		}

		provider = workflowcore.NewRemoteReviewAdapter(config.Review.Providers, nil)
	}

	return provider, nil
}

func (s service) workflowIntegrationObserver(run workflowstate.Invocation) workflowcore.IntegrationGit {
	git := s.workflowIntegrationGit
	if git == nil {
		git = workflowcore.NewWorkflowIntegrationGit(run.Selection.Target.Repository, run.Selection.Target.Repository)
	}

	return git
}
