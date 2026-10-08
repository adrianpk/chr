package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

func (s service) workflowStore() (*core.WorkflowStore, error) {
	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir})
	if err != nil {
		return nil, err
	}

	return core.NewWorkflowStore(layout, s.projectContext.ProjectID), nil
}

// StartWorkflow admits or attaches to the same durable ticket-native invocation.
func (s service) StartWorkflow(ctx context.Context, request core.WorkflowStartRequest) (core.WorkflowStartResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowStartResult{}, err
	}

	return store.StartWorkflow(ctx, request)
}

// ReadWorkflow returns a retained checkpoint without owner credentials.
func (s service) ReadWorkflow(ctx context.Context, id string) (core.WorkflowInvocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowInvocation{}, err
	}

	return store.ReadWorkflow(ctx, id)
}

// ClaimWorkflow uses the shared process-safe claim and serial capacity boundary.
func (s service) ClaimWorkflow(ctx context.Context, id string) (*core.WorkflowClaim, error) {
	store, err := s.workflowStore()
	if err != nil {
		return nil, err
	}

	return store.ClaimWorkflow(ctx, id)
}

// CheckWorkflowClaim checks a private credential against its durable fence.
func (s service) CheckWorkflowClaim(ctx context.Context, claim *core.WorkflowClaim) error {
	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	return store.CheckWorkflowClaim(ctx, claim)
}

// CheckpointWorkflow yields at a guarded boundary without recording task success.
func (s service) CheckpointWorkflow(ctx context.Context, claim *core.WorkflowClaim) error {
	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	return store.CheckpointWorkflow(ctx, claim)
}

// ExecuteWorkflowTask runs an owned isolated task and the frozen host gate.
func (s service) ExecuteWorkflowTask(ctx context.Context, claim *core.WorkflowClaim, worker []string) (core.WorkflowTaskResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowTaskResult{}, err
	}

	return store.ExecuteWorkflowTask(ctx, claim, worker)
}

// StopWorkflow persists pause or cancel intent through the shared domain owner.
func (s service) StopWorkflow(ctx context.Context, id, intent string) (core.WorkflowInvocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowInvocation{}, err
	}

	return store.StopWorkflow(ctx, id, intent)
}

// ResumeWorkflow continues a completed pause with its consumed budget retained.
func (s service) ResumeWorkflow(ctx context.Context, id string) (core.WorkflowInvocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowInvocation{}, err
	}

	return store.ResumeWorkflow(ctx, id)
}

// ReconcileWorkflow recovers dead or expired ownership after process exclusion.
func (s service) ReconcileWorkflow(ctx context.Context, id string) (core.WorkflowInvocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowInvocation{}, err
	}

	return store.ReconcileWorkflow(ctx, id)
}
