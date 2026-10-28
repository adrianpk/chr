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

type workflowRegisteredProcess struct {
	Process workflowstate.Process
	Index   int
}

type workflowCommandOutput struct {
	Stdout *opsfile.CommandOutput
	Stderr *opsfile.CommandOutput
}

func (s *Store) registerWorkflowProcess(ctx context.Context, claim *Claim, prepared workflowPreparedTask, command []string, pid int) (observed workflowRegisteredProcess, err error) {
	var (
		registered workflowstate.Process
		index      int
	)

	identity, err := workflowProcessIdentity(pid)
	if err != nil {
		return workflowRegisteredProcess{Process: registered, Index: 0}, err
	}

	err = s.withLock(ctx, func() error {
		claimObservation, readErr := s.checkWorkflowClaim(ctx, claim)
		if readErr != nil {
			return readErr
		}

		if claimObservation.Run.StopIntent != "" {
			return workflowstate.ErrStop
		}

		if readErr = checkWorkflowExecutionBinding(claimObservation.Run, claimObservation.Record, prepared); readErr != nil {
			return readErr
		}

		attempt := &claimObservation.Run.Attempts[len(claimObservation.Run.Attempts)-1]

		index = len(attempt.Processes)
		if index >= 17 {
			return fmt.Errorf("workflow command process limit exceeded")
		}

		base := workflowAttemptFile(claimObservation.Run.ID, attempt.ID, fmt.Sprintf("command-%d", index+1))
		registered = workflowstate.Process{CapacitySlot: workflowWorkerSlotNumber(claim.reservation.Name()), Identity: identity, GroupID: pid, Command: command, State: workflowstate.ProcessRunning, StdoutRef: base + ".stdout", StderrRef: base + ".stderr"}
		attempt.Processes = append(attempt.Processes, registered)

		if err := claim.store.commitWorkflowProcessStart(ctx, &claimObservation.Run, prepared); err != nil {
			return err
		}

		return registerWorkflowProcessCapacity(ctx, claim, attempt.Processes[index])
	})

	return workflowRegisteredProcess{Process: registered, Index: index}, err
}

func (s *Store) finishWorkflowProcess(ctx context.Context, claim *Claim, index int, expected workflowstate.Process, result workflowstate.CommandResult) error {
	return s.withLock(ctx, func() error {
		claimObservation2, err := s.checkWorkflowCredential(ctx, claim, workflowCredentialIdentity)
		if err != nil {
			return err
		}

		attempt := &claimObservation2.Run.Attempts[len(claimObservation2.Run.Attempts)-1]
		if index >= len(attempt.Processes) {
			return workflowstate.ErrConflict
		}

		process := &attempt.Processes[index]
		if !workflowProcessBindingMatches(*process, expected) {
			return errors.Join(workflowstate.ErrConflict, workflowstate.ErrOwned)
		}

		alive, err := workflowProcessGroupAlive(*process)
		if err != nil {
			return err
		}

		if alive {
			return workflowstate.ErrOwned
		}

		process.State, process.Result = workflowstate.ProcessStopped, &result
		if err := claim.store.hashWorkflowProcessOutput(process); err != nil {
			return err
		}

		return claim.store.commitWorkflowState(&claimObservation2.Run)
	})
}

func (s *Store) hashWorkflowProcessOutput(process *workflowstate.Process) error {
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

func (s *Store) workflowProcessOutput(claim *Claim, index int) (observed workflowCommandOutput, err error) {
	var (
		stdout *opsfile.CommandOutput
		stderr *opsfile.CommandOutput
	)

	base := workflowAttemptFile(claim.InvocationID, claim.AttemptID, fmt.Sprintf("command-%d", index+1))

	stdout, err = openWorkflowCommandOutput(filepath.Join(s.layout.Root(), base+".stdout"))
	if err != nil {
		return workflowCommandOutput{Stdout: nil, Stderr: nil}, err
	}

	stderr, err = openWorkflowCommandOutput(filepath.Join(s.layout.Root(), base+".stderr"))
	if err != nil {
		return workflowCommandOutput{Stdout: nil, Stderr: nil}, errors.Join(err, stdout.File.Close())
	}

	return workflowCommandOutput{Stdout: stdout, Stderr: stderr}, nil
}

func openWorkflowCommandOutput(path string) (*opsfile.CommandOutput, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)

	return &opsfile.CommandOutput{File: file}, err
}

func workflowProcessBindingMatches(observed, expected workflowstate.Process) bool {
	return observed.CapacitySlot == expected.CapacitySlot && observed.Identity == expected.Identity && observed.GroupID == expected.GroupID && observed.StdoutRef == expected.StdoutRef && observed.StderrRef == expected.StderrRef && observed.State == workflowstate.ProcessRunning && slices.Equal(observed.Command, expected.Command)
}
