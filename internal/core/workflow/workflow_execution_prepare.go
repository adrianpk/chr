package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowEncodedTaskInput struct {
	Prepared workflowPreparedTask
	Change   workflowstate.FileChange
}

type workflowPreparationMode string

const (
	workflowPreparationAutomatic workflowPreparationMode = "automatic"
	workflowPreparationManual    workflowPreparationMode = "manual"
)

const workflowTaskInstructions = `Complete the execution_contract in lane.worktree on the assigned lane.branch.
Stay within the task scope and preserve unrelated work. Run the required validation commands.
Commit the task changes with a Conventional Commit message and leave the worktree clean before returning success.
Do not push, merge, switch branches, or edit chr workflow state. The host owns validation and integration.
If the task, validation, or commit cannot be completed, report the failure instead of claiming success.`

func (s *Store) prepareWorkflowTaskMode(ctx context.Context, claim *Claim, worker []string, mode workflowPreparationMode) (prepared workflowPreparedTask, err error) {
	return s.prepareWorkflowTaskHarness(ctx, claim, worker, mode, nil)
}

func (s *Store) prepareWorkflowTaskHarness(ctx context.Context, claim *Claim, worker []string, mode workflowPreparationMode, harness *AgentHarness) (prepared workflowPreparedTask, err error) {
	if claim == nil || claim.purpose != "" {
		return prepared, ErrWorkflowOwned
	}

	err = s.withLock(ctx, func() error {
		claimObservation, readErr := s.checkWorkflowClaim(ctx, claim)
		if readErr != nil {
			return readErr
		}

		worker, readErr = workflowBoundTaskWorker(claimObservation.Run, worker, mode)
		if readErr != nil {
			return readErr
		}

		prepared, readErr = claim.store.prepareWorkflowAttempt(ctx, workflowAttemptPreparation{Run: claimObservation.Run, Record: claimObservation.Record, Worker: worker, Mode: mode, Harness: harness})

		return readErr
	})

	return prepared, err
}

type workflowAttemptPreparation struct {
	Run     workflowInvocationFile
	Record  Record
	Worker  []string
	Mode    workflowPreparationMode
	Harness *AgentHarness
}

func (s *Store) prepareWorkflowAttempt(ctx context.Context, preparation workflowAttemptPreparation) (workflowPreparedTask, error) {
	attempt := &preparation.Run.Attempts[len(preparation.Run.Attempts)-1]
	if attempt.InputRef != "" {
		return workflowPreparedTask{}, fmt.Errorf("%w: attempt input is already prepared", ErrWorkflowOwned)
	}

	lane, err := s.retainedWorkflowLane(ctx, preparation.Run, preparation.Record)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	if lane == nil {
		value := s.newWorkflowLane(preparation.Run, preparation.Record)
		lane = &value
	}

	preparation.Run.Lane = lane
	if err := s.commitWorkflowState(&preparation.Run); err != nil {
		return workflowPreparedTask{}, err
	}

	if err := s.ensureWorkflowLane(ctx, *lane); err != nil {
		return workflowPreparedTask{}, err
	}

	input, err := s.workflowTaskInput(ctx, preparation.Run, preparation.Record, preparation.Worker)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	input.ExecutionMode = string(preparation.Mode)
	input.Instructions = workflowTaskInstructions

	return s.saveWorkflowTaskInput(ctx, preparation.Run, input, preparation.Harness)
}

func (s *Store) workflowTaskInput(ctx context.Context, run workflowInvocationFile, record Record, worker []string) (TaskInput, error) {
	attempt := run.Attempts[len(run.Attempts)-1]

	input := TaskInput{Schema: "chr-task-input/v1", InvocationID: run.ID, AttemptID: attempt.ID, Generation: attempt.Generation, TaskKey: record.Task.Key, InputRevision: attempt.InputRevision, SourceHash: record.SourceHash, ContractHash: record.Task.ContractHash, Body: string(record.Body), Target: run.ExecutionSelection().Target, Lane: *run.Lane, Worker: slices.Clone(worker)}
	if run.HarnessBinding != nil && attempt.Purpose == "" {
		input.HarnessBinding = harnesscore.CloneBinding(run.HarnessBinding)
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

	prerequisiteImages, prerequisiteImagesErr := s.workflowPrerequisiteImages(ctx, run, record)

	input.Prerequisites, input.EvidenceHashes, err = prerequisiteImages.Images, prerequisiteImages.EvidenceHashes, prerequisiteImagesErr

	return input, err
}

func (s *Store) saveWorkflowTaskInput(ctx context.Context, run workflowInvocationFile, input TaskInput, harness *AgentHarness) (workflowPreparedTask, error) {
	encodedAgentInput, err := s.encodeWorkflowAgentInput(ctx, &run, input, harness)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(run)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	files := []workflowstate.FileChange{encodedAgentInput.Input, state}
	if encodedAgentInput.Budget.Path != "" {
		files = append(files, encodedAgentInput.Budget)
	}

	if err := s.commitFiles(files); err != nil {
		return workflowPreparedTask{}, err
	}

	return encodedAgentInput.Prepared, nil
}

func encodeWorkflowTaskInput(run *workflowInvocationFile, input TaskInput) (workflowEncodedTaskInput, error) {
	var err error

	input.ValidationTimeoutSeconds = run.Policy.ValidationTimeoutSeconds

	input.ScopeHash, err = workflowExecutionScopeHash(*run)
	if err != nil {
		return workflowEncodedTaskInput{Prepared: workflowPreparedTask{}, Change: workflowstate.FileChange{}}, err
	}

	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return workflowEncodedTaskInput{Prepared: workflowPreparedTask{}, Change: workflowstate.FileChange{}}, fmt.Errorf("workflow input exceeds file bounds: %w", errorsOrConflict(err))
	}

	data = append(data, '\n')
	ref := workflowAttemptFile(run.ID, input.AttemptID, "input.json")
	attempt := &run.Attempts[len(run.Attempts)-1]
	lane := input.Lane
	attempt.Lane, attempt.InputRef, attempt.InputHash = &lane, ref, contenthash.Sum(data)

	return workflowEncodedTaskInput{Prepared: workflowPreparedTask{input: input, inputRef: ref, inputHash: attempt.InputHash, data: data, leaseUntil: attempt.LeaseUntil}, Change: workflowstate.FileChange{Path: ref, After: data}}, nil
}

func workflowAttemptFile(invocationID, attemptID, name string) string {
	return filepath.ToSlash(filepath.Join("run", invocationID, "extension", "chr", "attempts", attemptID, name))
}

func (s *Store) commitWorkflowState(run *workflowInvocationFile, extra ...workflowstate.FileChange) error {
	run.Revision++
	run.UpdatedAt = time.Now().UTC()

	state, err := workflowInvocationChange(*run)
	if err != nil {
		return err
	}

	if err := s.commitFiles(append(extra, state)); err != nil {
		return err
	}

	run.sourceHash = contenthash.Sum(state.After)

	return nil
}

func validateWorkflowInputWorker(mode workflowPreparationMode, worker []string) error {
	if mode == workflowPreparationManual && len(worker) == 0 {
		return nil
	}

	if mode != workflowPreparationManual && mode != workflowPreparationAutomatic {
		return workflowstate.ErrConflict
	}

	return workflowstate.ValidateCommand(worker)
}
