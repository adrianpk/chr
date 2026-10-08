package core

import (
	"context"
	"errors"
	"time"
)

// ExecuteWorkflowTask runs one owned task, then applies host validation evidence.
// Workers receive frozen JSON on stdin and an isolated cwd; they do not own status.
func (s *WorkflowStore) ExecuteWorkflowTask(ctx context.Context, claim *WorkflowClaim, worker []string) (WorkflowTaskResult, error) {
	if claim == nil || claim.store == nil || claim.execution == nil {
		return WorkflowTaskResult{}, ErrWorkflowOwned
	}

	claim.execution.Lock()
	if !workflowClaimQuiescent(claim) {
		claim.execution.Unlock()

		return WorkflowTaskResult{}, ErrWorkflowOwned
	}

	claim.execution.running = true

	claim.execution.Unlock()
	defer func() {
		claim.execution.Lock()
		claim.execution.running = false
		claim.execution.Unlock()
	}()

	prepared, err := s.prepareWorkflowTask(ctx, claim, worker)
	if err != nil {
		return WorkflowTaskResult{}, err
	}

	result, executionErr := s.runWorkflowTask(ctx, claim, prepared)
	if errors.Is(executionErr, ErrWorkflowOwned) {
		return result, executionErr
	}

	err = s.finishWorkflowTask(ctx, claim, prepared, &result)

	return result, errors.Join(executionErr, err)
}

func workflowClaimQuiescent(claim *WorkflowClaim) bool {
	return !claim.execution.running && (claim.execution.groupPID == 0 || workflowCommandGroupStopped(claim.execution.groupPID))
}

func (s *WorkflowStore) runWorkflowTask(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask) (WorkflowTaskResult, error) {
	workerContext, cancel := context.WithDeadline(ctx, prepared.leaseUntil)
	defer cancel()

	input := prepared.input
	result := WorkflowTaskResult{WorkflowTaskValidation: WorkflowTaskValidation{Schema: "chr-task-validation/v1", Key: input.TaskKey, ContractHash: input.ContractHash, Branch: input.Lane.Branch}, AttemptID: input.AttemptID, Generation: input.Generation, InputRef: prepared.inputRef, InputHash: prepared.inputHash}
	worker, err := workflowRunCommand(workerContext, claim, prepared, input.Worker)

	result.Worker = worker
	if err == nil {
		err = s.validateWorkflowTask(workerContext, claim, prepared, &result)
	}

	result.Successful = err == nil
	result.Outcome = "task_validated"
	result.IntegrationPending = true

	result.FinishedAt = time.Now().UTC()
	if err != nil {
		result.Outcome = "task_validation_failed"
		result.Failure = workflowFailureText(err)
	}

	return result, err
}

func workflowFailureText(err error) string {
	text := err.Error()
	if len(text) > 4096 {
		return text[:4096]
	}

	return text
}
