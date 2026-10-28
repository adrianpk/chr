package app

import (
	"context"
	"fmt"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CloseWorkflowSet shares set-owned closing validation across operation adapters.
func (s service) CloseWorkflowSet(ctx context.Context, request workflowcore.SetClosureRequest) (workflowcore.SetClosureResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.SetClosureResult{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return workflowcore.SetClosureResult{}, err
	}

	observer := s.workflowIntegrationObserver(workflowcore.Invocation{InvocationInput: workflowcore.InvocationInput{Selection: workflowstate.PreviewRequest{Target: workflowstate.PreviewTarget{Repository: request.Repository}}}})

	target, ok := observer.(workflowcore.TargetGit)
	if !ok {
		return workflowcore.SetClosureResult{}, fmt.Errorf("configured observer cannot hydrate closing targets")
	}

	return store.CloseWorkflowSet(ctx, request, workflowcore.FrontierPorts{Provider: provider, Observer: observer, Target: target})
}

// ReconcileWorkflowClosing recovers a dead set-closing owner without rerunning work.
func (s service) ReconcileWorkflowClosing(ctx context.Context, path string) (workflowcore.Record, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.Record{}, err
	}

	return store.ReconcileWorkflowClosing(ctx, path)
}
