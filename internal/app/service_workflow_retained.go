package app

import (
	"context"
	"errors"
	"time"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// RetainedWorkProgress reports one guarded activation of existing authorization.
type RetainedWorkProgress struct {
	Status    WorkflowWorkStatus
	Attempted bool
	Steps     int
}

// ResumeRetainedWork never admits tickets or changes the frozen extent. Operator
// stops and review waits remain observational; dead owners use common recovery.
func (s service) ResumeRetainedWork(ctx context.Context, id string, revision uint64) (result RetainedWorkProgress, err error) {
	return s.ResumeRetainedWorkGuarded(ctx, id, revision, nil)
}

// ResumeRetainedWorkGuarded rechecks volatile caller authority before recovery
// and under the common continuation acceptance lock. It grants no new extent.
func (s service) ResumeRetainedWorkGuarded(ctx context.Context, id string, revision uint64, check workflowcore.AdmissionCheck) (result RetainedWorkProgress, err error) {
	s.ctx = ctx

	if !validWorkflowControlIdentity(id) || revision == 0 {
		return result, workflowstate.ErrConflict
	}

	defer func() {
		readCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()

		current, readErr := s.WorkflowStatus(readCtx, id, WorkflowStatusRequest{Limit: 8})
		if readErr == nil {
			result.Status = current
		}

		err = errors.Join(err, readErr)
	}()

	result.Status, err = s.WorkflowStatus(ctx, id, WorkflowStatusRequest{Limit: 8, ExpectedRevision: revision})
	if err != nil || !retainedWorkRunnable(result.Status) {
		return result, err
	}

	ready, err := s.retainedWorkCapacityAvailable(ctx, result.Status)
	if err != nil || !ready {
		return result, err
	}

	if err := retainedWorkAuthorization(ctx, check); err != nil {
		return result, err
	}

	result.Attempted = true

	return s.advanceRetainedWork(ctx, result, check)
}

func (s service) advanceRetainedWork(ctx context.Context, result RetainedWorkProgress, check workflowcore.AdmissionCheck) (RetainedWorkProgress, error) {
	run, err := s.recoverRetainedWork(ctx, result.Status)
	if err != nil {
		return result, err
	}

	if run.StopIntent != "" || run.Status == "paused" || run.Status == "cancelled" {
		return result, nil
	}

	progress, err := s.continueWorkExtent(ctx, run.ID, 1, check, run.Revision)
	result.Steps = progress.Steps

	return result, err
}

func retainedWorkRunnable(status WorkflowWorkStatus) bool {
	if status.Status == "running" {
		return true
	}

	if retainedWorkAwaitingAuthority(status) {
		return false
	}

	switch status.Status {
	case "pending", "waiting_capacity", "waiting_review", "waiting_integration":
		return true
	case "paused":
		return status.Outcome == "recovered"
	default:
		return false
	}
}

func retainedWorkAwaitingAuthority(status WorkflowWorkStatus) bool {
	if status.StopIntent != "" || status.ReviewState == "waiting_approval" || status.ReviewState == "human_required" {
		return true
	}

	if status.Blocker != nil && status.Blocker.NextActionOwner != "host" {
		return true
	}

	if status.Integration == nil {
		return false
	}

	return status.Integration.Decision != nil && status.Integration.Decision.State == "waiting"
}

func (s service) retainedWorkCapacityAvailable(ctx context.Context, status WorkflowWorkStatus) (bool, error) {
	// A running invocation must reach common recovery, which proves whether its
	// current owner is alive. Capacity observations cannot take over that owner.
	if status.Status == "running" {
		return true, nil
	}

	if status.Integration != nil {
		return status.Integration.AvailableRoles > 0, nil
	}

	store, err := s.workflowStore()
	if err != nil {
		return false, err
	}

	capacity, err := store.QueryWorkflowIntegrationCapacity(ctx)

	return capacity.Available > 0, err
}

func retainedWorkAuthorization(ctx context.Context, check workflowcore.AdmissionCheck) error {
	if check == nil {
		return nil
	}

	return check(ctx)
}

func (s service) recoverRetainedWork(ctx context.Context, status WorkflowWorkStatus) (workflowcore.Invocation, error) {
	run, err := s.ReadWorkflow(ctx, status.ID)
	if err != nil {
		return run, err
	}

	if run.Revision != status.Revision {
		return run, workflowstate.ErrConflict
	}

	if run.Status == "running" {
		run, err = s.ReconcileWorkflow(ctx, run.ID, run.Revision)
		if err != nil {
			return run, err
		}
	}

	if run.Status == "paused" && run.StopIntent == "" && run.Outcome == "recovered" {
		return s.ResumeWorkflow(ctx, run.ID, run.Revision)
	}

	return run, nil
}
