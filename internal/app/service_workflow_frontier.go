package app

import (
	"context"
	"fmt"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// AdvanceWorkflowFrontier refreshes the exact retained extent after finalization.
// It does not execute work, create a PR, submit integration or close a set.
func (s service) AdvanceWorkflowFrontier(ctx context.Context, id string) (workflowcore.Invocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.Invocation{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowcore.Invocation{}, err
	}

	provider, err := s.workflowIntegrationProvider()
	if err != nil {
		return workflowcore.Invocation{}, err
	}

	observer := s.workflowIntegrationObserver(run)

	target, ok := observer.(workflowcore.TargetGit)
	if !ok {
		return run, fmt.Errorf("configured integration observer cannot hydrate continuation targets")
	}

	return store.AdvanceWorkflowFrontier(ctx, id, workflowcore.FrontierPorts{Provider: provider, Observer: observer, Target: target})
}
