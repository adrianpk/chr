package app

import (
	"context"
	"fmt"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s service) workflowStore() (*workflowcore.Store, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir, Namespace: s.workflowNamespace})
	if err != nil {
		return nil, err
	}

	return workflowcore.NewWorkflowStore(layout, s.projectContext.ProjectID), nil
}

// StartWorkflow admits or attaches to the same durable ticket-native invocation.
func (s service) StartWorkflow(ctx context.Context, request workflowcore.StartRequest) (workflowcore.StartResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.StartResult{}, err
	}

	request.ResolveHarness = s.prepareWorkflowHarness

	return store.StartWorkflow(ctx, request)
}

// ReadWorkflow returns a retained checkpoint without owner credentials.
func (s service) ReadWorkflow(ctx context.Context, id string) (workflowcore.Invocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.Invocation{}, err
	}

	return store.ReadWorkflow(ctx, id)
}

// ClaimWorkflow uses the shared process-safe claim and serial capacity boundary.
func (s service) ClaimWorkflow(ctx context.Context, id string) (*workflowcore.Claim, error) {
	store, err := s.workflowStore()
	if err != nil {
		return nil, err
	}

	return store.ClaimWorkflow(ctx, id)
}

// CheckWorkflowClaim checks a private credential against its durable fence.
func (s service) CheckWorkflowClaim(ctx context.Context, claim *workflowcore.Claim) error {
	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	return store.CheckWorkflowClaim(ctx, claim)
}

// CheckpointWorkflow yields at a guarded boundary without recording task success.
func (s service) CheckpointWorkflow(ctx context.Context, claim *workflowcore.Claim) error {
	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	return store.CheckpointWorkflow(ctx, claim)
}

// ExecuteWorkflowTask runs an owned isolated task and the frozen host gate.
func (s service) ExecuteWorkflowTask(ctx context.Context, claim *workflowcore.Claim, worker []string) (workflowstate.TaskResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.TaskResult{}, err
	}

	return store.ExecuteWorkflowTask(ctx, claim, worker)
}

// StopWorkflow persists pause or cancel intent through the shared domain owner.
func (s service) StopWorkflow(ctx context.Context, id, intent string, expected ...uint64) (workflowcore.Invocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.Invocation{}, err
	}

	return store.StopWorkflow(ctx, id, intent, expected...)
}

// ResumeWorkflow continues a completed pause with its consumed budget retained.
func (s service) ResumeWorkflow(ctx context.Context, id string, expected ...uint64) (workflowcore.Invocation, error) {
	return s.resumeWorkflowChecked(ctx, id, nil, expected...)
}

func (s service) resumeWorkflowChecked(ctx context.Context, id string, check workflowcore.AdmissionCheck, expected ...uint64) (workflowcore.Invocation, error) {
	revision, err := workflowControlRevision(expected)
	if err != nil {
		return workflowcore.Invocation{}, err
	}

	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.Invocation{}, err
	}

	run, err := store.ReadWorkflow(ctx, id)
	if err != nil {
		return workflowcore.Invocation{}, err
	}

	ports := workflowcore.ResumePorts{ExpectedRevision: revision}
	if run.Blocker != nil && run.Blocker.NextActionOwner == "prerequisite_owner" {
		ports, err = s.workflowResumePorts(run)
		if err != nil {
			return workflowcore.Invocation{}, err
		}

		ports.ExpectedRevision = revision
	}

	ports.AuthorizationCheck = check

	return store.ResumeWorkflow(ctx, id, ports)
}

func workflowControlRevision(expected []uint64) (uint64, error) {
	if len(expected) == 0 {
		return 0, nil
	}

	if len(expected) != 1 || expected[0] == 0 {
		return 0, workflowstate.ErrConflict
	}

	return expected[0], nil
}

func (s service) workflowResumePorts(run workflowcore.Invocation) (workflowcore.ResumePorts, error) {
	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
	if err != nil {
		return workflowcore.ResumePorts{}, err
	}

	provider, err := s.workflowIntegrationProvider()

	target, ok := s.workflowIntegrationObserver(run).(workflowcore.TargetGit)
	if !ok {
		return workflowcore.ResumePorts{}, fmt.Errorf("configured observer cannot hydrate dependency targets")
	}

	return workflowcore.ResumePorts{Provider: provider, Target: target, TargetBranch: config.Work.BranchPolicy.IntegrationBranch}, err
}

// ReconcileWorkflow recovers dead or expired ownership after process exclusion.
func (s service) ReconcileWorkflow(ctx context.Context, id string, expected ...uint64) (workflowcore.Invocation, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.Invocation{}, err
	}

	return store.ReconcileWorkflow(ctx, id, expected...)
}

// PrepareWorkflowManualTask freezes an owned handoff for manual execution.
func (s service) PrepareWorkflowManualTask(ctx context.Context, claim *workflowcore.Claim, worker []string) (workflowcore.ManualTask, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.ManualTask{}, err
	}

	return store.PrepareWorkflowManualTask(ctx, claim, worker)
}

// ExecuteWorkflowManualTask supervises the frozen harness without completing work.
func (s service) ExecuteWorkflowManualTask(ctx context.Context, claim *workflowcore.Claim) (workflowstate.CommandResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.CommandResult{}, err
	}

	return store.ExecuteWorkflowManualTask(ctx, claim)
}

// ReconcileWorkflowTask host-validates an explicit manual candidate before status.
func (s service) ReconcileWorkflowTask(ctx context.Context, claim *workflowcore.Claim, candidate workflowcore.ManualCandidate) (workflowstate.TaskResult, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowstate.TaskResult{}, err
	}

	return store.ReconcileWorkflowTask(ctx, claim, candidate)
}
