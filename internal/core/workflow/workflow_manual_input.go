package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) loadWorkflowManualTask(ctx context.Context, claim *Claim, execute bool) (prepared workflowPreparedTask, err error) {
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

func (s *Store) loadWorkflowManualInput(run workflowInvocationFile, record Record, execute bool) (workflowPreparedTask, error) {
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

func (s *Store) readWorkflowManualInput(attempt Attempt) (workflowPreparedTask, error) {
	path, err := s.safePath(attempt.InputRef)
	if err != nil || attempt.InputRef == "" {
		return workflowPreparedTask{}, workflowstate.ErrConflict
	}

	data, err := opsfile.Read(path)
	if err != nil {
		return workflowPreparedTask{}, err
	}

	if contenthash.Sum(data) != attempt.InputHash {
		return workflowPreparedTask{}, workflowstate.ErrConflict
	}

	var input TaskInput
	if err := json.Unmarshal(data, &input); err != nil {
		return workflowPreparedTask{}, err
	}

	if validateWorkflowTaskHarnessSchema(input) != nil || input.ExecutionMode != "manual" {
		return workflowPreparedTask{}, ErrWorkflowOwned
	}

	if err := validateWorkflowInputWorker(workflowPreparationMode(input.ExecutionMode), input.Worker); err != nil {
		return workflowPreparedTask{}, err
	}

	return workflowPreparedTask{input: input, inputRef: attempt.InputRef, inputHash: attempt.InputHash, data: data, leaseUntil: attempt.LeaseUntil}, nil
}

func validateWorkflowManualCandidate(candidate ManualCandidate) error {
	if len(candidate.InputHash) != 64 || !workflowCommitID(candidate.Commit) || !workflowCommitID(candidate.Tree) {
		return fmt.Errorf("manual result requires frozen input hash and full commit/tree identities")
	}

	return nil
}

func (s *Store) checkWorkflowManualCandidate(ctx context.Context, claim *Claim, prepared workflowPreparedTask, candidate ManualCandidate) error {
	if candidate.InputHash != prepared.inputHash || candidate.Branch != prepared.input.Lane.Branch {
		return workflowstate.ErrConflict
	}

	commit, tree, err := claim.store.workflowLaneCandidate(ctx, prepared.input.Lane)
	if err != nil {
		return err
	}

	if commit != candidate.Commit || tree != candidate.Tree {
		return fmt.Errorf("%w: reported manual candidate differs from actual Git image", workflowstate.ErrConflict)
	}

	return s.checkWorkflowValidationImage(ctx, claim, prepared, TaskResult{TaskValidation: TaskValidation{Commit: commit, Tree: tree}})
}

func (s *Store) workflowManualWorkerResult(ctx context.Context, claim *Claim) (CommandResult, error) {
	run, err := s.ReadWorkflow(ctx, claim.InvocationID)
	if err != nil {
		return CommandResult{}, err
	}

	processes := run.Attempts[len(run.Attempts)-1].Processes
	if len(processes) == 0 {
		return CommandResult{}, nil
	}

	if len(processes) != 1 || processes[0].Result == nil || processes[0].Result.ExitCode != 0 {
		return CommandResult{}, ErrWorkflowOwned
	}

	return *processes[0].Result, nil
}
