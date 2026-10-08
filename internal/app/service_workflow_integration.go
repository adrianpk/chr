package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// VerifyWorkflowSliceIntegration checks provider and remote Git delivery facts.
func (s service) VerifyWorkflowSliceIntegration(ctx context.Context, id, method string) (core.WorkflowSliceDeliveryProof, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowSliceDeliveryProof{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return core.WorkflowSliceDeliveryProof{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return core.WorkflowSliceDeliveryProof{}, err
	}

	return store.VerifyWorkflowSliceIntegration(ctx, id, method, provider, s.workflowIntegrationObserver(run))
}

func (s service) workflowIntegrationProvider() (core.WorkflowPullRequestProvider, error) {
	provider := s.workflowPullRequests
	if provider == nil {
		config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
		if err != nil {
			return nil, err
		}

		if err := core.ValidateReviewWorkflowConfig(config.Review); err != nil {
			return nil, err
		}

		provider = core.NewRemoteReviewAdapter(config.Review.Providers, nil)
	}

	return provider, nil
}

func (s service) workflowIntegrationObserver(run core.WorkflowInvocation) core.WorkflowIntegrationGit {
	git := s.workflowIntegrationGit
	if git == nil {
		git = core.NewWorkflowIntegrationGit(run.Selection.Target.Repository, run.Selection.Target.Repository)
	}

	return git
}
