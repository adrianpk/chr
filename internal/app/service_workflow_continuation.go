package app

import (
	"context"
	"fmt"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// ContinueWorkflow resumes an existing invocation through the common bounded
// operation. Adapters cannot replace its selector, target or frozen policy.
func (s service) ContinueWorkflow(ctx context.Context, id string, request workflowcore.ContinuationRequest, reviewer []string) (workflowcore.ContinuationProgress, error) {
	reviewDependencies, err := s.workflowReviewDependencies(id)
	store, provider := reviewDependencies.Store, reviewDependencies.Provider

	if err != nil {
		return workflowcore.ContinuationProgress{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowcore.ContinuationProgress{}, err
	}

	if run.HarnessBinding != nil {
		request.Worker, err = store.WorkflowHarnessWorker(ctx, id)
		if err != nil {
			return workflowcore.ContinuationProgress{Invocation: run}, err
		}
	}

	ports, err := s.workflowContinuationPorts(run, provider, reviewer)
	if err != nil {
		return workflowcore.ContinuationProgress{Invocation: run}, err
	}

	return store.ContinueWorkflow(ctx, id, request, ports)
}

func (s service) workflowContinuationPorts(run workflowcore.Invocation, provider workflowcore.PullRequestProvider, reviewer []string) (workflowcore.ContinuationPorts, error) {
	observer := s.workflowIntegrationObserver(run)

	target, ok := observer.(workflowcore.TargetGit)
	if !ok {
		return workflowcore.ContinuationPorts{}, fmt.Errorf("configured observer cannot hydrate continuation targets")
	}

	preparer := s.workflowPreparationGit
	if preparer == nil {
		preparer = workflowcore.NewWorkflowPreparationGit(run.Selection.Target.Repository, run.Selection.Target.Repository)
	}

	ports := workflowcore.ContinuationPorts{Publication: workflowcore.PublicationPorts{Provider: provider, Preparer: preparer, Observer: observer, Reviewer: reviewer}, Target: target}
	bindWorkflowNativePublicationPorts(&ports.Publication, run)

	if source, ok := observer.(workflowcore.SourceGit); ok {
		ports.Source = source
	}

	return ports, nil
}
