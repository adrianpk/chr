package app

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
)

// ContinueWorkflow resumes an existing invocation through the common bounded
// operation. Adapters cannot replace its selector, target or frozen policy.
func (s service) ContinueWorkflow(ctx context.Context, id string, request core.WorkflowContinuationRequest, reviewer []string) (core.WorkflowContinuationProgress, error) {
	store, provider, err := s.workflowReviewDependencies()
	if err != nil {
		return core.WorkflowContinuationProgress{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return core.WorkflowContinuationProgress{}, err
	}

	observer := s.workflowIntegrationObserver(run)

	target, ok := observer.(core.WorkflowTargetGit)
	if !ok {
		return core.WorkflowContinuationProgress{Invocation: run}, fmt.Errorf("configured observer cannot hydrate continuation targets")
	}

	preparer := s.workflowPreparationGit
	if preparer == nil {
		preparer = core.NewWorkflowPreparationGit(run.Selection.Target.Repository, run.Selection.Target.Repository)
	}

	ports := core.WorkflowContinuationPorts{Publication: core.WorkflowPublicationPorts{Provider: provider, Preparer: preparer, Observer: observer, Reviewer: reviewer}, Target: target}

	return store.ContinueWorkflow(ctx, id, request, ports)
}
