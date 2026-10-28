package app

import (
	"context"
	"fmt"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// CloseWorkflowSet shares set-owned closing validation across operation adapters.
func (s service) CloseWorkflowSet(ctx context.Context, request workflowcloseout.Request) (workflowcloseout.Result, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcloseout.Result{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return workflowcloseout.Result{}, err
	}

	observer := s.workflowIntegrationObserver(workflowstate.Invocation{InvocationInput: workflowstate.InvocationInput{Selection: workflowstate.PreviewRequest{Target: workflowstate.PreviewTarget{Repository: request.Repository}}}})

	target, ok := observer.(workflowcore.TargetGit)
	if !ok {
		return workflowcloseout.Result{}, fmt.Errorf("configured observer cannot hydrate closing targets")
	}

	return store.CloseWorkflowSet(ctx, request, workflowcore.FrontierPorts{Provider: provider, Observer: observer, Target: target})
}

// ReconcileWorkflowClosing recovers a dead set-closing owner without rerunning work.
func (s service) ReconcileWorkflowClosing(ctx context.Context, path string) (workflowstate.Record, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.Record{}, err
	}

	return store.ReconcileWorkflowClosing(ctx, path)
}
