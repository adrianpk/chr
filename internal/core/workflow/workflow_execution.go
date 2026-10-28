package workflow

import (
	"context"
	"errors"
	"time"
)

// ExecuteWorkflowTask runs one owned task, then applies host validation evidence.
// Workers receive frozen JSON on stdin and an isolated cwd; they do not own status.
func (s *Store) ExecuteWorkflowTask(ctx context.Context, claim *Claim, worker []string) (TaskResult, error) {
	return s.executeWorkflowTask(ctx, claim, worker, nil)
}

// ExecuteWorkflowTaskWithHarness enforces and settles a certified whole-call grant.
func (s *Store) ExecuteWorkflowTaskWithHarness(ctx context.Context, claim *Claim, harness AgentHarness) (TaskResult, error) {
	return s.executeWorkflowTask(ctx, claim, harness.Command, &harness)
}

func (s *Store) executeWorkflowTask(ctx context.Context, claim *Claim, worker []string, harness *AgentHarness) (TaskResult, error) {
	if err := beginWorkflowClaimExecution(claim); err != nil {
		return TaskResult{}, err
	}
	defer endWorkflowClaimExecution(claim)

	prepared, err := s.prepareWorkflowTaskHarness(ctx, claim, worker, workflowPreparationAutomatic, harness)
	if err != nil {
		return TaskResult{}, err
	}

	watched, stop := s.watchWorkflowStop(ctx, claim)
	defer stop()

	result, executionErr := s.runWorkflowTask(watched, claim, prepared)
	if errors.Is(executionErr, ErrWorkflowOwned) {
		return result, executionErr
	}

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	err = s.finishWorkflowTask(cleanup, claim, prepared, &result, watched.Err() != nil)

	return result, errors.Join(executionErr, err)
}

func workflowClaimQuiescent(claim *Claim) bool {
	return !claim.execution.running && (claim.execution.groupPID == 0 || workflowCommandGroupStopped(claim.execution.groupPID))
}

func (s *Store) runWorkflowTask(ctx context.Context, claim *Claim, prepared workflowPreparedTask) (TaskResult, error) {
	workerContext, cancel := context.WithDeadline(ctx, prepared.leaseUntil)
	defer cancel()

	input := prepared.input
	result := newWorkflowTaskResult(prepared)
	worker, err := workflowRunAgentCommand(workerContext, claim, prepared, input.Worker)

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

func beginWorkflowClaimExecution(claim *Claim) error {
	if claim == nil || claim.store == nil || claim.execution == nil {
		return ErrWorkflowOwned
	}

	claim.execution.Lock()
	defer claim.execution.Unlock()

	if !workflowClaimQuiescent(claim) {
		return ErrWorkflowOwned
	}

	claim.execution.running = true

	return nil
}

func endWorkflowClaimExecution(claim *Claim) {
	claim.execution.Lock()
	claim.execution.running = false
	claim.execution.Unlock()
}

func newWorkflowTaskResult(prepared workflowPreparedTask) TaskResult {
	input := prepared.input

	return TaskResult{TaskValidation: TaskValidation{Schema: "chr-task-validation/v1", Key: input.TaskKey, ContractHash: input.ContractHash, Branch: input.Lane.Branch}, ExecutionMode: input.ExecutionMode, AttemptID: input.AttemptID, Generation: input.Generation, InputRef: prepared.inputRef, InputHash: prepared.inputHash}
}
