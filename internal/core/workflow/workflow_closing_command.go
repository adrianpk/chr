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

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (host workflowClosingHost) runCommand(ctx context.Context, claim *workflowClosingClaim, directory string, command []string) (result workflowstate.CommandResult, err error) {
	result = workflowstate.CommandResult{Command: slices.Clone(command), ExitCode: -1}
	index := len(claim.state.Processes)
	base := workflowcloseout.File(claim.set.ID, claim.state.Generation, fmt.Sprintf("command-%d", index+1))

	stdout, err := openWorkflowCommandOutput(filepath.Join(host.store.layout.Root(), base+".stdout"))
	if err != nil {
		return result, err
	}

	defer func() { err = errors.Join(err, stdout.File.Close()) }()

	stderr, err := openWorkflowCommandOutput(filepath.Join(host.store.layout.Root(), base+".stderr"))
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, stderr.File.Close()) }()

	gate, permit, err := os.Pipe()
	if err != nil {
		return result, err
	}
	defer func() { _ = gate.Close(); _ = permit.Close() }() //nolint:errcheck // Fallback pipe cleanup; launch and wait paths report permission/process failures.

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

	return host.waitCommand(ctx, claim, workflowClosingCommandExecution{Command: command, Process: process, Permit: permit, Output: workflowCommandOutput{Stdout: stdout, Stderr: stderr}, Base: base})
}

type workflowClosingCommandExecution struct {
	Command []string
	Process *exec.Cmd
	Permit  *os.File
	Output  workflowCommandOutput
	Base    string
}

func (host workflowClosingHost) waitCommand(ctx context.Context, claim *workflowClosingClaim, execution workflowClosingCommandExecution) (workflowstate.CommandResult, error) {
	registered, err := host.store.registerWorkflowClosingProcess(ctx, claim, execution.Command, execution.Process.Process.Pid, execution.Base)
	if err == nil {
		_, err = fmt.Fprintln(execution.Permit, "run")
	}

	err = errors.Join(err, execution.Permit.Close(), execution.Process.Wait())

	result := workflowstate.CommandResult{Command: slices.Clone(execution.Command), ExitCode: -1, Stdout: execution.Output.Stdout.Buffer.String(), Stderr: execution.Output.Stderr.Buffer.String(), Truncated: execution.Output.Stdout.Truncated || execution.Output.Stderr.Truncated}
	if execution.Process.ProcessState != nil {
		result.ExitCode = execution.Process.ProcessState.ExitCode()
	}

	if registered.State != "" {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		err = errors.Join(err, host.store.finishWorkflowClosingProcess(cleanup, claim, registered))
	}

	return result, err
}

func (s *Store) registerWorkflowClosingProcess(ctx context.Context, claim *workflowClosingClaim, command []string, pid int, base string) (registered workflowstate.Process, err error) {
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

		registered = workflowstate.Process{CapacitySlot: workflowWorkerSlotNumber(claim.worker.reservation.Name()), Identity: identity, GroupID: pid, Command: slices.Clone(command), State: workflowstate.ProcessRunning, StdoutRef: base + ".stdout", StderrRef: base + ".stderr"}

		claim.state.Processes = append(claim.state.Processes, registered)
		if err := s.commitWorkflowClosingState(claim); err != nil {
			return err
		}

		return registerWorkflowProcessCapacity(ctx, claim.worker, registered)
	})

	return registered, err
}

func (s *Store) finishWorkflowClosingProcess(ctx context.Context, claim *workflowClosingClaim, expected workflowstate.Process) error {
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

		process.State = workflowstate.ProcessStopped
		if err := s.hashWorkflowProcessOutput(process); err != nil {
			return err
		}

		return s.commitWorkflowClosingState(claim)
	})
}
