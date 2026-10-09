package app

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core"
)

// AdvanceWorkflowFrontier refreshes the exact retained extent after finalization.
// It does not execute work, create a PR, submit integration or close a set.
func (s service) AdvanceWorkflowFrontier(ctx context.Context, id string) (core.WorkflowInvocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowInvocation{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return core.WorkflowInvocation{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return core.WorkflowInvocation{}, err
	}

	observer := s.workflowIntegrationObserver(run)

	target, ok := observer.(core.WorkflowTargetGit)
	if !ok {
		return run, fmt.Errorf("configured integration observer cannot hydrate continuation targets")
	}

	return store.AdvanceWorkflowFrontier(ctx, id, core.WorkflowFrontierPorts{Provider: provider, Observer: observer, Target: target})
}
