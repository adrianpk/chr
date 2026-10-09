package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// AdvanceWorkflowSliceIntegration reconciles delivery before guarded publication.
// Adapters share this operation; it never starts another task or slice.
func (s service) AdvanceWorkflowSliceIntegration(ctx context.Context, id string, reviewer []string) (core.WorkflowIntegrationProgress, error) {
	store, provider, err := s.workflowReviewDependencies()
	if err != nil {
		return core.WorkflowIntegrationProgress{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return core.WorkflowIntegrationProgress{}, err
	}

	preparer := s.workflowPreparationGit
	if preparer == nil {
		preparer = core.NewWorkflowPreparationGit(run.Selection.Target.Repository, run.Selection.Target.Repository)
	}

	ports := core.WorkflowPublicationPorts{Provider: provider, Preparer: preparer, Observer: s.workflowIntegrationObserver(run), Reviewer: reviewer}

	return store.AdvanceWorkflowSliceIntegration(ctx, id, ports)
}
