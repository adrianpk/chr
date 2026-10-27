package core

import (
	"context"
	"errors"
	"time"
)

// WorkflowManualTask exposes immutable manual inputs; the claim remains private.
type WorkflowManualTask struct {
	Input      WorkflowTaskInput `json:"input"`
	InputRef   string            `json:"input_ref"`
	InputHash  string            `json:"input_hash"`
	LeaseUntil time.Time         `json:"lease_until"`
}

// WorkflowManualCandidate reports an exact Git image for host verification.
type WorkflowManualCandidate struct {
	InputHash string `json:"input_hash"`
	Branch    string `json:"branch"`
	Commit    string `json:"commit"`
	Tree      string `json:"tree"`
}

// PrepareWorkflowManualTask freezes a manual handoff and retains serial ownership.
// Supply a harness argv to execute it through ExecuteWorkflowManualTask; an empty
// argv permits reconciliation of a quiescent candidate without launching a worker.
func (s *WorkflowStore) PrepareWorkflowManualTask(ctx context.Context, claim *WorkflowClaim, worker []string) (WorkflowManualTask, error) {
	if err := beginWorkflowClaimExecution(claim); err != nil {
		return WorkflowManualTask{}, err
	}
	defer endWorkflowClaimExecution(claim)

	prepared, err := s.prepareWorkflowTaskMode(ctx, claim, worker, workflowPreparationManual)

	return WorkflowManualTask{Input: prepared.input, InputRef: prepared.inputRef, InputHash: prepared.inputHash, LeaseUntil: prepared.leaseUntil}, err
}

// ExecuteWorkflowManualTask supervises the frozen harness once. Success retains
// ownership for explicit candidate reconciliation; it does not validate the task.
func (s *WorkflowStore) ExecuteWorkflowManualTask(ctx context.Context, claim *WorkflowClaim) (WorkflowCommandResult, error) {
	if err := beginWorkflowClaimExecution(claim); err != nil {
		return WorkflowCommandResult{}, err
	}
	defer endWorkflowClaimExecution(claim)

	prepared, err := s.loadWorkflowManualTask(ctx, claim, true)
	if err != nil {
		return WorkflowCommandResult{}, err
	}

	watched, stop := s.watchWorkflowStop(ctx, claim)
	defer stop()

	workerContext, cancel := context.WithDeadline(watched, prepared.leaseUntil)
	defer cancel()

	observed, executionErr := workflowRunAgentCommand(workerContext, claim, prepared, prepared.input.Worker)
	if errors.Is(executionErr, ErrWorkflowOwned) {
		return observed, executionErr
	}

	if executionErr == nil && workerContext.Err() == nil {
		return observed, nil
	}

	result := newWorkflowTaskResult(prepared)

	result.Worker, result.Outcome, result.FinishedAt = observed, "task_validation_failed", time.Now().UTC()
	if executionErr != nil {
		result.Failure = workflowFailureText(executionErr)
	}

	cleanup, done := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer done()

	err = s.finishWorkflowTask(cleanup, claim, prepared, &result, workerContext.Err() != nil)

	return observed, errors.Join(executionErr, err)
}

// ReconcileWorkflowTask verifies a reported manual Git image with the frozen host
// gate and the same guarded result writer used by automatic execution.
func (s *WorkflowStore) ReconcileWorkflowTask(ctx context.Context, claim *WorkflowClaim, candidate WorkflowManualCandidate) (WorkflowTaskResult, error) {
	if err := beginWorkflowClaimExecution(claim); err != nil {
		return WorkflowTaskResult{}, err
	}
	defer endWorkflowClaimExecution(claim)

	if err := validateWorkflowManualCandidate(candidate); err != nil {
		return WorkflowTaskResult{}, err
	}

	if result, found, err := s.reuseWorkflowManualResult(ctx, claim, candidate); found || err != nil {
		return result, err
	}

	prepared, err := s.loadWorkflowManualTask(ctx, claim, false)
	if err != nil {
		return WorkflowTaskResult{}, err
	}

	if err := s.checkWorkflowManualCandidate(ctx, claim, prepared, candidate); err != nil {
		return WorkflowTaskResult{}, err
	}

	watched, stop := s.watchWorkflowStop(ctx, claim)
	defer stop()

	validationContext, cancel := context.WithDeadline(watched, prepared.leaseUntil)
	defer cancel()

	result := newWorkflowTaskResult(prepared)

	result.Worker, err = s.workflowManualWorkerResult(ctx, claim)
	if err != nil {
		return result, err
	}

	validationErr := s.validateWorkflowTask(validationContext, claim, prepared, &result)
	result.Successful, result.Outcome, result.IntegrationPending = validationErr == nil, "task_validated", true

	result.FinishedAt = time.Now().UTC()
	if validationErr != nil {
		result.Outcome, result.Failure = "task_validation_failed", workflowFailureText(validationErr)
	}

	if errors.Is(validationErr, ErrWorkflowOwned) {
		return result, validationErr
	}

	cleanup, done := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer done()

	err = s.finishWorkflowTask(cleanup, claim, prepared, &result, validationContext.Err() != nil)

	return result, errors.Join(validationErr, err)
}
