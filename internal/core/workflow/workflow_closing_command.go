package workflow

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) runWorkflowClosingCommand(ctx context.Context, claim *workflowClosingClaim, directory string, command []string) (result workflowstate.CommandResult, err error) {
	result = workflowstate.CommandResult{Command: slices.Clone(command), ExitCode: -1}
	index := len(claim.state.Processes)
	base := workflowClosingFile(claim.set.ID, claim.state.Generation, fmt.Sprintf("command-%d", index+1))

	stdout, err := openWorkflowCommandOutput(filepath.Join(s.layout.Root(), base+".stdout"))
	if err != nil {
		return result, err
	}

	defer func() { err = errors.Join(err, stdout.File.Close()) }()

	stderr, err := openWorkflowCommandOutput(filepath.Join(s.layout.Root(), base+".stderr"))
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, stderr.File.Close()) }()

	gate, permit, err := os.Pipe()
	if err != nil {
		return result, err
	}
	defer func() { _ = gate.Close(); _ = permit.Close() }() //nolint:errcheck

	prepared := workflowClosingPrepared(claim, directory)
	supervisedProcess := workflowSupervisedCommand(ctx, claim.worker, prepared, command, gate)

	process, guard := supervisedProcess.Process, supervisedProcess.Guard

	process.Stdin, process.Stdout, process.Stderr = bytes.NewReader(claim.data), stdout, stderr
	if err := process.Start(); err != nil {
		return result, err
	}

	if err := guard.remember(process.Process.Pid); err != nil {
		return result, errors.Join(err, permit.Close(), process.Wait())
	}

	return s.waitWorkflowClosingCommand(ctx, claim, command, process, permit, stdout, stderr, base)
}

func (s *Store) waitWorkflowClosingCommand(ctx context.Context, claim *workflowClosingClaim, command []string, process *exec.Cmd, permit *os.File, stdout, stderr *opsfile.CommandOutput, base string) (workflowstate.CommandResult, error) {
	registered, err := s.registerWorkflowClosingProcess(ctx, claim, command, process.Process.Pid, base)
	if err == nil {
		_, err = fmt.Fprintln(permit, "run")
	}

	err = errors.Join(err, permit.Close(), process.Wait())

	result := workflowstate.CommandResult{Command: slices.Clone(command), ExitCode: -1, Stdout: stdout.Buffer.String(), Stderr: stderr.Buffer.String(), Truncated: stdout.Truncated || stderr.Truncated}
	if process.ProcessState != nil {
		result.ExitCode = process.ProcessState.ExitCode()
	}

	if registered.State != "" {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		err = errors.Join(err, s.finishWorkflowClosingProcess(cleanup, claim, registered))
	}

	return result, err
}

func (s *Store) registerWorkflowClosingProcess(ctx context.Context, claim *workflowClosingClaim, command []string, pid int, base string) (registered Process, err error) {
	identity, err := workflowProcessIdentity(pid)
	if err != nil {
		return registered, err
	}

	err = s.withLock(ctx, func() error {
		if err := s.checkWorkflowClosingClaim(ctx, claim, true); err != nil {
			return err
		}

		if len(claim.state.Processes) >= 16 {
			return workflowstate.ErrConflict
		}

		registered = Process{CapacitySlot: workflowWorkerSlotNumber(claim.worker.reservation.Name()), Identity: identity, GroupID: pid, Command: slices.Clone(command), State: ProcessRunning, StdoutRef: base + ".stdout", StderrRef: base + ".stderr"}

		claim.state.Processes = append(claim.state.Processes, registered)
		if err := s.commitWorkflowClosingState(ctx, claim); err != nil {
			return err
		}

		return registerWorkflowProcessCapacity(ctx, claim.worker, registered)
	})

	return registered, err
}

func (s *Store) finishWorkflowClosingProcess(ctx context.Context, claim *workflowClosingClaim, expected Process) error {
	return s.withLock(ctx, func() error {
		if err := s.checkWorkflowClosingCredential(claim); err != nil {
			return err
		}

		process := &claim.state.Processes[len(claim.state.Processes)-1]
		if !workflowProcessBindingMatches(*process, expected) {
			return workflowstate.ErrConflict
		}

		alive, err := workflowProcessGroupAlive(*process)
		if err != nil || alive {
			return errors.Join(err, ErrWorkflowCapacity)
		}

		process.State = ProcessStopped
		if err := s.hashWorkflowProcessOutput(process); err != nil {
			return err
		}

		return s.commitWorkflowClosingState(ctx, claim)
	})
}
