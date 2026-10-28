package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) registerWorkflowProcess(ctx context.Context, claim *Claim, prepared workflowPreparedTask, command []string, pid int) (registered Process, index int, err error) {
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
		registered = Process{CapacitySlot: workflowWorkerSlotNumber(claim.reservation.Name()), Identity: identity, GroupID: pid, Command: command, State: "running", StdoutRef: base + ".stdout", StderrRef: base + ".stderr"}
		attempt.Processes = append(attempt.Processes, registered)

		if err := claim.store.commitWorkflowProcessStart(ctx, &run, prepared); err != nil {
			return err
		}

		return registerWorkflowProcessCapacity(ctx, claim, attempt.Processes[index])
	})

	return registered, index, err
}

func (s *Store) finishWorkflowProcess(ctx context.Context, claim *Claim, index int, expected Process, result CommandResult) error {
	return s.withLock(ctx, func() error {
		run, _, err := s.checkWorkflowCredential(ctx, claim, false)
		if err != nil {
			return err
		}

		attempt := &run.Attempts[len(run.Attempts)-1]
		if index >= len(attempt.Processes) {
			return workflowstate.ErrConflict
		}

		process := &attempt.Processes[index]
		if !workflowProcessBindingMatches(*process, expected) {
			return errors.Join(workflowstate.ErrConflict, ErrWorkflowOwned)
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

func (s *Store) hashWorkflowProcessOutput(process *Process) error {
	stdout, err := opsfile.Read(filepath.Join(s.layout.Root(), process.StdoutRef))
	if err != nil {
		return err
	}

	stderr, err := opsfile.Read(filepath.Join(s.layout.Root(), process.StderrRef))
	if err != nil {
		return err
	}

	process.StdoutHash, process.StderrHash = contenthash.Sum(stdout), contenthash.Sum(stderr)

	return nil
}

func (s *Store) workflowProcessOutput(claim *Claim, index int) (stdout, stderr *opsfile.CommandOutput, err error) {
	base := workflowAttemptFile(claim.InvocationID, claim.AttemptID, fmt.Sprintf("command-%d", index+1))

	stdout, err = openWorkflowCommandOutput(filepath.Join(s.layout.Root(), base+".stdout"))
	if err != nil {
		return nil, nil, err
	}

	stderr, err = openWorkflowCommandOutput(filepath.Join(s.layout.Root(), base+".stderr"))
	if err != nil {
		return nil, nil, errors.Join(err, stdout.File.Close())
	}

	return stdout, stderr, nil
}

func openWorkflowCommandOutput(path string) (*opsfile.CommandOutput, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)

	return &opsfile.CommandOutput{File: file}, err
}

func workflowProcessBindingMatches(observed, expected Process) bool {
	return observed.CapacitySlot == expected.CapacitySlot && observed.Identity == expected.Identity && observed.GroupID == expected.GroupID && observed.StdoutRef == expected.StdoutRef && observed.StderrRef == expected.StderrRef && observed.State == "running" && slices.Equal(observed.Command, expected.Command)
}
