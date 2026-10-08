package core

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *WorkflowStore) loadWorkflowManualTask(ctx context.Context, claim *WorkflowClaim, execute bool) (prepared workflowPreparedTask, err error) {
	err = s.withLock(ctx, func() error {
		run, record, readErr := s.checkWorkflowClaim(ctx, claim)
		if readErr != nil {
			return readErr
		}

		if run.StopIntent != "" {
			return ErrWorkflowStop
		}

		attempt := run.Attempts[len(run.Attempts)-1]
		if readErr = workflowAttemptQuiescent(attempt); readErr != nil {
			return readErr
		}

		prepared, readErr = claim.store.loadWorkflowManualInput(run, record, execute)

		return readErr
	})

	return prepared, err
}

func (s *WorkflowStore) loadWorkflowManualInput(run workflowInvocationFile, record WorkflowRecord, execute bool) (workflowPreparedTask, error) {
	attempt := run.Attempts[len(run.Attempts)-1]

	prepared, err := s.readWorkflowManualInput(attempt)
	if err != nil {
		return prepared, err
	}

	if err := checkWorkflowExecutionBinding(run, record, prepared); err != nil {
		return prepared, err
	}

	if err := s.checkWorkflowPreparedInput(prepared); err != nil {
		return prepared, err
	}

	if execute && (len(attempt.Processes) != 0 || len(prepared.input.Worker) == 0) {
		return prepared, ErrWorkflowOwned
	}

	return prepared, nil
}

func (s *WorkflowStore) readWorkflowManualInput(attempt WorkflowAttempt) (workflowPreparedTask, error) {
	path, err := s.safePath(attempt.InputRef)
	if err != nil || attempt.InputRef == "" {
		return workflowPreparedTask{}, ErrWorkflowConflict
	}

	data, err := readWorkflowFile(path)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	if workflowHash(data) != attempt.InputHash {
		return workflowPreparedTask{}, ErrWorkflowConflict
	}

	var input WorkflowTaskInput
	if err := json.Unmarshal(data, &input); err != nil {
		return workflowPreparedTask{}, err
	}

	if input.Schema != "chr-task-input/v1" || input.ExecutionMode != "manual" {
		return workflowPreparedTask{}, ErrWorkflowOwned
	}

	if err := validateWorkflowInputWorker(input.ExecutionMode, input.Worker); err != nil {
		return workflowPreparedTask{}, err
	}

	return workflowPreparedTask{input: input, inputRef: attempt.InputRef, inputHash: attempt.InputHash, data: data, leaseUntil: attempt.LeaseUntil}, nil
}

func validateWorkflowManualCandidate(candidate WorkflowManualCandidate) error {
	if len(candidate.InputHash) != 64 || !workflowCommitID(candidate.Commit) || !workflowCommitID(candidate.Tree) {
		return fmt.Errorf("manual result requires frozen input hash and full commit/tree identities")
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowManualCandidate(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, candidate WorkflowManualCandidate) error {
	if candidate.InputHash != prepared.inputHash || candidate.Branch != prepared.input.Lane.Branch {
		return ErrWorkflowConflict
	}

	commit, tree, err := claim.store.workflowLaneCandidate(ctx, prepared.input.Lane)
	if err != nil {
		return err
	}

	if commit != candidate.Commit || tree != candidate.Tree {
		return fmt.Errorf("%w: reported manual candidate differs from actual Git image", ErrWorkflowConflict)
	}

	return s.checkWorkflowValidationImage(ctx, claim, prepared, WorkflowTaskResult{WorkflowTaskValidation: WorkflowTaskValidation{Commit: commit, Tree: tree}})
}

func (s *WorkflowStore) workflowManualWorkerResult(ctx context.Context, claim *WorkflowClaim) (WorkflowCommandResult, error) {
	run, err := s.ReadWorkflow(ctx, claim.InvocationID)
	if err != nil {
		return WorkflowCommandResult{}, err
	}

	processes := run.Attempts[len(run.Attempts)-1].Processes
	if len(processes) == 0 {
		return WorkflowCommandResult{}, nil
	}

	if len(processes) != 1 || processes[0].Result == nil || processes[0].Result.ExitCode != 0 {
		return WorkflowCommandResult{}, ErrWorkflowOwned
	}

	return *processes[0].Result, nil
}
