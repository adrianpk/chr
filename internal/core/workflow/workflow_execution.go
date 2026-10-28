package workflow

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// workflowTaskHost owns task and gate process lifetime, outside persistence.
type workflowTaskHost struct {
	store *Store
}

// ExecuteTask runs one owned task, then applies host validation evidence.
// Workers receive frozen JSON on stdin and an isolated cwd; they do not own status.
func ExecuteTask(ctx context.Context, store *Store, claim *Claim, worker []string) (workflowstate.TaskResult, error) {
	return (workflowTaskHost{store: store}).execute(ctx, claim, worker, nil)
}

// ExecuteTaskWithHarness enforces and settles a certified whole-call grant.
func ExecuteTaskWithHarness(ctx context.Context, store *Store, claim *Claim, harness AgentHarness) (workflowstate.TaskResult, error) {
	return (workflowTaskHost{store: store}).execute(ctx, claim, harness.Command, &harness)
}

func (host workflowTaskHost) execute(ctx context.Context, claim *Claim, worker []string, harness *AgentHarness) (workflowstate.TaskResult, error) {
	if err := beginWorkflowClaimExecution(claim); err != nil {
		return workflowstate.TaskResult{}, err
	}
	defer endWorkflowClaimExecution(claim)

	prepared, err := host.store.prepareWorkflowTaskHarness(ctx, claim, worker, workflowPreparationAutomatic, harness)
	if err != nil {
		return workflowstate.TaskResult{}, err
	}

	watched, stop := context.WithCancel(ctx)

	watchWorkflowStop(watched, host.store, workflowStopWatch{Claim: claim, Cancel: stop})
	defer stop()

	result, executionErr := host.run(watched, claim, prepared)
	if errors.Is(executionErr, workflowstate.ErrOwned) {
		return result, executionErr
	}

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	err = host.store.finishWorkflowTask(cleanup, claim, prepared, &result, watched.Err() != nil)

	return result, errors.Join(executionErr, err)
}

func workflowClaimQuiescent(claim *Claim) bool {
	return !claim.execution.running && (claim.execution.groupPID == 0 || workflowCommandGroupStopped(claim.execution.groupPID))
}

func (host workflowTaskHost) run(ctx context.Context, claim *Claim, prepared workflowPreparedTask) (workflowstate.TaskResult, error) {
	workerContext, cancel := context.WithDeadline(ctx, prepared.leaseUntil)
	defer cancel()

	input := prepared.input
	result := newWorkflowTaskResult(prepared)
	worker, err := workflowRunAgentCommand(workerContext, claim, prepared, input.Worker)

	result.Worker = worker
	if err == nil {
		err = host.validate(workerContext, claim, prepared, &result)
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
		return workflowstate.ErrOwned
	}

	claim.execution.Lock()
	defer claim.execution.Unlock()

	if !workflowClaimQuiescent(claim) {
		return workflowstate.ErrOwned
	}

	claim.execution.running = true

	return nil
}

func endWorkflowClaimExecution(claim *Claim) {
	claim.execution.Lock()
	claim.execution.running = false
	claim.execution.Unlock()
}

func newWorkflowTaskResult(prepared workflowPreparedTask) workflowstate.TaskResult {
	input := prepared.input

	return workflowstate.TaskResult{TaskValidation: workflowstate.TaskValidation{Schema: "chr-task-validation/v1", Key: input.TaskKey, ContractHash: input.ContractHash, Branch: input.Lane.Branch}, ExecutionMode: input.ExecutionMode, AttemptID: input.AttemptID, Generation: input.Generation, InputRef: prepared.inputRef, InputHash: prepared.inputHash}
}
