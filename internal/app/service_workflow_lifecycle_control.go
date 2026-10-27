package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s service) executeWorkflowLifecycleResult(ctx context.Context, store *core.WorkflowStore, request WorkflowControlRequest) (*workflowstate.WorkflowControlResult, error) {
	ports, err := s.workflowLifecycleControlPorts(ctx, store, request)
	if err != nil {
		return nil, err
	}

	native := workflowNativeControlRequest(request)

	result, err := store.ControlWorkflowLifecycle(ctx, native, ports, request.AuthorizationCheck)
	if result.RequestID == "" {
		return nil, err
	}

	return &result, err
}

func (s service) workflowLifecycleControlPorts(ctx context.Context, store *core.WorkflowStore, request WorkflowControlRequest) (core.WorkflowLifecycleControlPorts, error) {
	var ports core.WorkflowLifecycleControlPorts
	if request.Action != "continue" && request.Action != "resume" {
		return ports, nil
	}

	run, err := store.ReadWorkflow(ctx, request.ID)
	if err != nil {
		return ports, err
	}

	if request.Action == "resume" {
		if run.Blocker != nil && run.Blocker.NextActionOwner == "prerequisite_owner" {
			ports.Resume, err = s.workflowResumePorts(run)
		}

		return ports, err
	}

	ports.Continuation, err = s.workflowContinuationInput(request.StepLimit)
	if err != nil {
		return ports, err
	}

	_, provider, err := s.workflowReviewDependencies(request.ID)
	if err != nil {
		return ports, err
	}

	ports.Execution, err = s.workflowContinuationPorts(run, provider, ports.Continuation.Worker)

	return ports, err
}
