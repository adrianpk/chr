package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"time"
)

func (s *WorkflowStore) prepareWorkflowTask(ctx context.Context, claim *WorkflowClaim, worker []string) (prepared workflowPreparedTask, err error) {
	return s.prepareWorkflowTaskMode(ctx, claim, worker, "automatic")
}

func (s *WorkflowStore) prepareWorkflowTaskMode(ctx context.Context, claim *WorkflowClaim, worker []string, mode string) (prepared workflowPreparedTask, err error) {
	return s.prepareWorkflowTaskHarness(ctx, claim, worker, mode, nil)
}

func (s *WorkflowStore) prepareWorkflowTaskHarness(ctx context.Context, claim *WorkflowClaim, worker []string, mode string, harness *WorkflowAgentHarness) (prepared workflowPreparedTask, err error) {
	if claim == nil || claim.purpose != "" {
		return prepared, ErrWorkflowOwned
	}

	err = s.withLock(ctx, func() error {
		run, record, readErr := s.checkWorkflowClaim(ctx, claim)
		if readErr != nil {
			return readErr
		}

		worker, readErr = workflowBoundTaskWorker(run, worker, mode)
		if readErr != nil {
			return readErr
		}

		prepared, readErr = claim.store.prepareWorkflowAttempt(ctx, run, record, worker, mode, harness)

		return readErr
	})

	return prepared, err
}

func (s *WorkflowStore) prepareWorkflowAttempt(ctx context.Context, run workflowInvocationFile, record WorkflowRecord, worker []string, mode string, harness *WorkflowAgentHarness) (workflowPreparedTask, error) {
	attempt := &run.Attempts[len(run.Attempts)-1]
	if attempt.InputRef != "" {
		return workflowPreparedTask{}, fmt.Errorf("%w: attempt input is already prepared", ErrWorkflowOwned)
	}

	lane, err := s.retainedWorkflowLane(ctx, run, record)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	if lane == nil {
		value := s.newWorkflowLane(run, record)
		lane = &value
	}

	run.Lane = lane
	if err := s.commitWorkflowState(&run); err != nil {
		return workflowPreparedTask{}, err
	}

	if err := s.ensureWorkflowLane(ctx, *lane); err != nil {
		return workflowPreparedTask{}, err
	}

	input, err := s.workflowTaskInput(ctx, run, record, worker)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	input.ExecutionMode = mode

	return s.saveWorkflowTaskInput(ctx, run, input, harness)
}

func (s *WorkflowStore) workflowTaskInput(ctx context.Context, run workflowInvocationFile, record WorkflowRecord, worker []string) (WorkflowTaskInput, error) {
	attempt := run.Attempts[len(run.Attempts)-1]

	input := WorkflowTaskInput{Schema: "chr-task-input/v1", InvocationID: run.ID, AttemptID: attempt.ID, Generation: attempt.Generation, TaskKey: record.Task.Key, InputRevision: attempt.InputRevision, SourceHash: record.SourceHash, ContractHash: record.Task.ContractHash, Body: string(record.Body), Target: run.ExecutionSelection().Target, Lane: *run.Lane, Worker: slices.Clone(worker)}
	if run.HarnessBinding != nil && attempt.Purpose == "" {
		input.HarnessBinding = cloneHarnessBinding(run.HarnessBinding)
		input.Schema = "chr-task-input/v2"
	}

	if err := json.Unmarshal(record.Domain["execution_contract"], &input.Contract); err != nil {
		return input, err
	}

	var err error

	input.OrderMap, err = s.requiredWorkflowOrderMap(ctx, record)
	if err != nil {
		return input, err
	}

	input.Prerequisites, input.EvidenceHashes, err = s.workflowPrerequisiteImages(ctx, run, record)

	return input, err
}

func (s *WorkflowStore) saveWorkflowTaskInput(ctx context.Context, run workflowInvocationFile, input WorkflowTaskInput, harness *WorkflowAgentHarness) (workflowPreparedTask, error) {
	prepared, file, budgetFile, err := s.encodeWorkflowAgentInput(ctx, &run, input, harness)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(run)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	files := []workflowFileChange{file, state}
	if budgetFile.Path != "" {
		files = append(files, budgetFile)
	}

	if err := s.commitFiles(files); err != nil {
		return workflowPreparedTask{}, err
	}

	return prepared, nil
}

func encodeWorkflowTaskInput(run *workflowInvocationFile, input WorkflowTaskInput) (workflowPreparedTask, workflowFileChange, error) {
	var err error

	input.ValidationTimeoutSeconds = run.Policy.ValidationTimeoutSeconds

	input.ScopeHash, err = workflowExecutionScopeHash(*run)
	if err != nil {
		return workflowPreparedTask{}, workflowFileChange{}, err
	}

	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil || len(data) >= WorkflowFileLimit {
		return workflowPreparedTask{}, workflowFileChange{}, fmt.Errorf("workflow input exceeds file bounds: %w", errorsOrConflict(err))
	}

	data = append(data, '\n')
	ref := workflowAttemptFile(run.ID, input.AttemptID, "input.json")
	attempt := &run.Attempts[len(run.Attempts)-1]
	lane := input.Lane
	attempt.Lane, attempt.InputRef, attempt.InputHash = &lane, ref, workflowHash(data)

	return workflowPreparedTask{input: input, inputRef: ref, inputHash: attempt.InputHash, data: data, leaseUntil: attempt.LeaseUntil}, workflowFileChange{Path: ref, After: data}, nil
}

func workflowAttemptFile(invocationID, attemptID, name string) string {
	return filepath.ToSlash(filepath.Join("run", invocationID, "extension", "chr", "attempts", attemptID, name))
}

func (s *WorkflowStore) commitWorkflowState(run *workflowInvocationFile, extra ...workflowFileChange) error {
	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	if err := s.commitFiles(append(extra, state)); err != nil {
		return err
	}

	run.sourceHash = workflowHash(state.After)

	return nil
}

func validateWorkflowInputWorker(mode string, worker []string) error {
	if mode == "manual" && len(worker) == 0 {
		return nil
	}

	if mode != "manual" && mode != "automatic" {
		return ErrWorkflowConflict
	}

	return validateWorkflowCommand(worker)
}
