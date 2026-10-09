package app

import (
	"context"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
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
	s.ctx = ctx

	if !validWorkflowControlIdentity(id) || revision == 0 {
		return result, core.ErrWorkflowConflict
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

	result.Attempted = true

	run, err := s.recoverRetainedWork(ctx, result.Status)
	if err != nil {
		return result, err
	}

	if run.StopIntent != "" || run.Status == "paused" || run.Status == "cancelled" {
		return result, nil
	}

	progress, err := s.ContinueWorkExtent(ctx, id, 1, run.Revision)
	result.Steps = progress.Steps

	return result, err
}

func retainedWorkRunnable(status WorkflowWorkStatus) bool {
	if status.Status == "running" {
		return true
	}

	if status.StopIntent != "" || status.ReviewState == "waiting_approval" || status.ReviewState == "human_required" {
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

func (s service) recoverRetainedWork(ctx context.Context, status WorkflowWorkStatus) (core.WorkflowInvocation, error) {
	run, err := s.ReadWorkflow(ctx, status.ID)
	if err != nil {
		return run, err
	}

	if run.Revision != status.Revision {
		return run, core.ErrWorkflowConflict
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
