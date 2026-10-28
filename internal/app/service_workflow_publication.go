package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// AdvanceWorkflowSliceIntegration reconciles delivery before guarded publication.
// Adapters share this operation; it never starts another task or slice.
func (s service) AdvanceWorkflowSliceIntegration(ctx context.Context, id string, reviewer []string) (workflowcore.IntegrationProgress, error) {
	reviewDependencies, err := s.workflowReviewDependencies(id)
	store, provider := reviewDependencies.Store, reviewDependencies.Provider

	if err != nil {
		return workflowcore.IntegrationProgress{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowcore.IntegrationProgress{}, err
	}

	preparer := s.workflowPreparationGit
	if preparer == nil {
		preparer = workflowcore.NewWorkflowPreparationGit(run.Selection.Target.Repository, run.Selection.Target.Repository)
	}

	ports := workflowcore.PublicationPorts{Provider: provider, Preparer: preparer, Observer: s.workflowIntegrationObserver(run), Reviewer: reviewer}
	bindWorkflowNativePublicationPorts(&ports, run)

	return store.AdvanceWorkflowSliceIntegration(ctx, id, ports)
}

func bindWorkflowNativePublicationPorts(ports *workflowcore.PublicationPorts, run workflowstate.Invocation) {
	ports.NativePublisher = workflowcore.NewWorkflowGitPublisher(run.Selection.Target.Repository, run.Selection.Target.Repository)
	if publisher, ok := ports.Provider.(workflowcore.GitPublisher); ok {
		ports.NativePublisher = publisher
	}

	if observer, ok := ports.Provider.(workflowcore.NativePublicationObserver); ok {
		ports.NativeObserver = observer
	}
}
