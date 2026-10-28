package workflow

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ManualTask exposes immutable manual inputs; the claim remains private.
type ManualTask struct {
	Input      TaskInput `json:"input"`
	InputRef   string    `json:"input_ref"`
	InputHash  string    `json:"input_hash"`
	LeaseUntil time.Time `json:"lease_until"`
}

// ManualCandidate reports an exact Git image for host verification.
type ManualCandidate struct {
	InputHash string `json:"input_hash"`
	Branch    string `json:"branch"`
	Commit    string `json:"commit"`
	Tree      string `json:"tree"`
}

// PrepareManualTask freezes a manual handoff and retains serial ownership.
// Supply a harness argv to execute it through ExecuteManualTask; an empty
// argv permits reconciliation of a quiescent candidate without launching a worker.
func PrepareManualTask(ctx context.Context, store *Store, claim *Claim, worker []string) (ManualTask, error) {
	if err := beginWorkflowClaimExecution(claim); err != nil {
		return ManualTask{}, err
	}
	defer endWorkflowClaimExecution(claim)

	prepared, err := store.prepareWorkflowTaskMode(ctx, claim, worker, workflowPreparationManual)

	return ManualTask{Input: prepared.input, InputRef: prepared.inputRef, InputHash: prepared.inputHash, LeaseUntil: prepared.leaseUntil}, err
}

// ExecuteManualTask supervises the frozen harness once. Success retains
// ownership for explicit candidate reconciliation; it does not validate the task.
func ExecuteManualTask(ctx context.Context, store *Store, claim *Claim) (workflowstate.CommandResult, error) {
	if err := beginWorkflowClaimExecution(claim); err != nil {
		return workflowstate.CommandResult{}, err
	}
	defer endWorkflowClaimExecution(claim)

	prepared, err := store.loadWorkflowManualTask(ctx, claim, true)
	if err != nil {
		return workflowstate.CommandResult{}, err
	}

	watched, stop := context.WithCancel(ctx)

	watchWorkflowStop(watched, store, workflowStopWatch{Claim: claim, Cancel: stop})
	defer stop()

	workerContext, cancel := context.WithDeadline(watched, prepared.leaseUntil)
	defer cancel()

	observed, executionErr := workflowRunAgentCommand(workerContext, claim, prepared, prepared.input.Worker)
	if errors.Is(executionErr, workflowstate.ErrOwned) {
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

	err = store.finishWorkflowTask(cleanup, claim, prepared, &result, workerContext.Err() != nil)

	return observed, errors.Join(executionErr, err)
}

// ReconcileTask verifies a reported manual Git image with the frozen host
// gate and the same guarded result writer used by automatic execution.
func ReconcileTask(ctx context.Context, store *Store, claim *Claim, candidate ManualCandidate) (workflowstate.TaskResult, error) {
	if err := beginWorkflowClaimExecution(claim); err != nil {
		return workflowstate.TaskResult{}, err
	}
	defer endWorkflowClaimExecution(claim)

	if err := validateWorkflowManualCandidate(candidate); err != nil {
		return workflowstate.TaskResult{}, err
	}

	{
		manualReplay, err := store.reuseWorkflowManualResult(ctx, claim, candidate)

		if manualReplay.Found || err != nil {
			return manualReplay.Result, err
		}
	}

	prepared, err := store.loadWorkflowManualTask(ctx, claim, false)
	if err != nil {
		return workflowstate.TaskResult{}, err
	}

	if err := store.checkWorkflowManualCandidate(ctx, claim, prepared, candidate); err != nil {
		return workflowstate.TaskResult{}, err
	}

	watched, stop := context.WithCancel(ctx)

	watchWorkflowStop(watched, store, workflowStopWatch{Claim: claim, Cancel: stop})
	defer stop()

	validationContext, cancel := context.WithDeadline(watched, prepared.leaseUntil)
	defer cancel()

	result := newWorkflowTaskResult(prepared)

	result.Worker, err = store.workflowManualWorkerResult(ctx, claim)
	if err != nil {
		return result, err
	}

	validationErr := (workflowTaskHost{store: store}).validate(validationContext, claim, prepared, &result)
	result.Successful, result.Outcome, result.IntegrationPending = validationErr == nil, "task_validated", true

	result.FinishedAt = time.Now().UTC()
	if validationErr != nil {
		result.Outcome, result.Failure = "task_validation_failed", workflowFailureText(validationErr)
	}

	if errors.Is(validationErr, workflowstate.ErrOwned) {
		return result, validationErr
	}

	cleanup, done := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer done()

	err = store.finishWorkflowTask(cleanup, claim, prepared, &result, validationContext.Err() != nil)

	return result, errors.Join(validationErr, err)
}
