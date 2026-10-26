package app

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
)

// CloseWorkflowSet shares set-owned closing validation across operation adapters.
func (s service) CloseWorkflowSet(ctx context.Context, request core.WorkflowSetClosureRequest) (core.WorkflowSetClosureResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowSetClosureResult{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return core.WorkflowSetClosureResult{}, err
	}

	observer := s.workflowIntegrationObserver(core.WorkflowInvocation{WorkflowInvocationInput: core.WorkflowInvocationInput{Selection: core.WorkflowPreviewRequest{Target: core.WorkflowPreviewTarget{Repository: request.Repository}}}})

	target, ok := observer.(core.WorkflowTargetGit)
	if !ok {
		return core.WorkflowSetClosureResult{}, fmt.Errorf("configured observer cannot hydrate closing targets")
	}

	return store.CloseWorkflowSet(ctx, request, core.WorkflowFrontierPorts{Provider: provider, Observer: observer, Target: target})
}

// ReconcileWorkflowClosing recovers a dead set-closing owner without rerunning work.
func (s service) ReconcileWorkflowClosing(ctx context.Context, path string) (core.WorkflowRecord, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowRecord{}, err
	}

	return store.ReconcileWorkflowClosing(ctx, path)
}
