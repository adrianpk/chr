package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func (s *WorkflowStore) registerWorkflowProcess(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, command []string, pid int) (registered WorkflowProcess, index int, err error) {
	identity, err := workflowProcessIdentity(pid)
	if err != nil {
		return registered, 0, err
	}

	err = s.withLock(ctx, func() error {
		run, record, readErr := s.checkWorkflowClaim(ctx, claim)
		if readErr != nil {
			return readErr
		}

		if run.StopIntent != "" {
			return ErrWorkflowStop
		}

		if readErr = checkWorkflowExecutionBinding(run, record, prepared); readErr != nil {
			return readErr
		}

		attempt := &run.Attempts[len(run.Attempts)-1]

		index = len(attempt.Processes)
		if index >= 17 {
			return fmt.Errorf("workflow command process limit exceeded")
		}

		base := workflowAttemptFile(run.ID, attempt.ID, fmt.Sprintf("command-%d", index+1))
		registered = WorkflowProcess{Identity: identity, GroupID: pid, Command: command, State: "running", StdoutRef: base + ".stdout", StderrRef: base + ".stderr"}
		attempt.Processes = append(attempt.Processes, registered)

		if err := claim.store.commitWorkflowState(&run); err != nil {
			return err
		}

		return registerWorkflowProcessCapacity(ctx, claim, attempt.Processes[index])
	})

	return registered, index, err
}

func (s *WorkflowStore) finishWorkflowProcess(ctx context.Context, claim *WorkflowClaim, index int, expected WorkflowProcess, result WorkflowCommandResult) error {
	return s.withLock(ctx, func() error {
		run, _, err := s.checkWorkflowCredential(ctx, claim, false)
		if err != nil {
			return err
		}

		attempt := &run.Attempts[len(run.Attempts)-1]
		if index >= len(attempt.Processes) {
			return ErrWorkflowConflict
		}

		process := &attempt.Processes[index]
		if !workflowProcessBindingMatches(*process, expected) {
			return errors.Join(ErrWorkflowConflict, ErrWorkflowOwned)
		}

		alive, err := workflowProcessGroupAlive(*process)
		if err != nil {
			return err
		}

		if alive {
			return ErrWorkflowOwned
		}

		process.State, process.Result = "stopped", &result
		if err := claim.store.hashWorkflowProcessOutput(process); err != nil {
			return err
		}

		return claim.store.commitWorkflowState(&run)
	})
}

func (s *WorkflowStore) hashWorkflowProcessOutput(process *WorkflowProcess) error {
	stdout, err := readWorkflowFile(filepath.Join(s.layout.Root(), process.StdoutRef))
	if err != nil {
		return err
	}

	stderr, err := readWorkflowFile(filepath.Join(s.layout.Root(), process.StderrRef))
	if err != nil {
		return err
	}

	process.StdoutHash, process.StderrHash = workflowHash(stdout), workflowHash(stderr)

	return nil
}

func (s *WorkflowStore) workflowProcessOutput(claim *WorkflowClaim, index int) (stdout, stderr *workflowCommandOutput, err error) {
	base := workflowAttemptFile(claim.InvocationID, claim.AttemptID, fmt.Sprintf("command-%d", index+1))

	stdout, err = openWorkflowCommandOutput(filepath.Join(s.layout.Root(), base+".stdout"))
	if err != nil {
		return nil, nil, err
	}

	stderr, err = openWorkflowCommandOutput(filepath.Join(s.layout.Root(), base+".stderr"))
	if err != nil {
		return nil, nil, errors.Join(err, stdout.file.Close())
	}

	return stdout, stderr, nil
}

func openWorkflowCommandOutput(path string) (*workflowCommandOutput, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)

	return &workflowCommandOutput{file: file}, err
}

func workflowProcessBindingMatches(observed, expected WorkflowProcess) bool {
	return observed.Identity == expected.Identity && observed.GroupID == expected.GroupID && observed.StdoutRef == expected.StdoutRef && observed.StderrRef == expected.StderrRef && observed.State == "running" && slices.Equal(observed.Command, expected.Command)
}
