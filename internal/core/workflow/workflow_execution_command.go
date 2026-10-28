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
	"strings"
	"syscall"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowRunCommand(ctx context.Context, claim *Claim, prepared workflowPreparedTask, command []string) (result workflowstate.CommandResult, err error) {
	run, err := claim.store.ReadWorkflow(ctx, claim.InvocationID)
	if err != nil {
		return workflowstate.CommandResult{}, err
	}

	index := len(run.Attempts[len(run.Attempts)-1].Processes)

	commandOutput, err := claim.store.workflowProcessOutput(claim, index)
	if err != nil {
		return workflowstate.CommandResult{}, err
	}

	defer func() {
		err = errors.Join(err, commandOutput.Stdout.File.Close(), commandOutput.Stderr.File.Close())
	}()

	gate, permit, err := os.Pipe()
	if err != nil {
		return workflowstate.CommandResult{}, err
	}

	defer func() { _ = gate.Close(); _ = permit.Close() }() //nolint:errcheck

	supervisedProcess := workflowSupervisedCommand(ctx, claim, prepared, command, gate)

	process, guard := supervisedProcess.Process, supervisedProcess.Guard
	process.Stdin = bytes.NewReader(prepared.data)

	process.Stdout, process.Stderr = commandOutput.Stdout, commandOutput.Stderr
	if err := process.Start(); err != nil {
		return workflowstate.CommandResult{}, err
	}

	if err := guard.remember(process.Process.Pid); err != nil {
		return workflowstate.CommandResult{}, errors.Join(err, permit.Close(), process.Wait())
	}

	return runWorkflowSupervisedProcess(ctx, claim, prepared, workflowProcessExecution{Command: command, Process: process, Permit: permit, Index: index, Output: workflowCommandOutput{Stdout: commandOutput.Stdout, Stderr: commandOutput.Stderr}})
}

func filepathForWorkflowInput(claim *Claim, prepared workflowPreparedTask) string {
	return filepath.Join(claim.store.layout.Root(), prepared.inputRef)
}

func workflowKillCommandGroup(command *exec.Cmd) error {
	if command.Process == nil {
		return os.ErrProcessDone
	}

	err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}

	return err
}

func workflowCommandGroupStopped(pid int) bool {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return false
	}

	alive, err := workflowProcessGroupAlive(Process{Identity: workflowstate.ProcessIdentity{BootID: strings.TrimSpace(string(boot))}, GroupID: pid})

	return err == nil && !alive
}

type workflowProcessExecution struct {
	Command []string
	Process *exec.Cmd
	Permit  *os.File
	Index   int
	Output  workflowCommandOutput
}

func runWorkflowSupervisedProcess(ctx context.Context, claim *Claim, prepared workflowPreparedTask, execution workflowProcessExecution) (workflowstate.CommandResult, error) {
	registeredProcess, registerErr := claim.store.registerWorkflowProcess(ctx, claim, prepared, execution.Command, execution.Process.Process.Pid)

	if registerErr == nil && registeredProcess.Index != execution.Index {
		registerErr = workflowstate.ErrConflict
	}

	if registerErr == nil {
		_, registerErr = fmt.Fprintln(execution.Permit, "run")
	}

	closeErr := execution.Permit.Close()
	err := errors.Join(registerErr, closeErr, execution.Process.Wait())

	result := workflowstate.CommandResult{Command: slices.Clone(execution.Command), ExitCode: -1, Stdout: execution.Output.Stdout.Buffer.String(), Stderr: execution.Output.Stderr.Buffer.String(), Truncated: execution.Output.Stdout.Truncated || execution.Output.Stderr.Truncated}
	if execution.Process.ProcessState != nil {
		result.ExitCode = execution.Process.ProcessState.ExitCode()
	}

	if registerErr == nil {
		var quiescenceErr error
		if prepared.input.Merger != nil || prepared.input.AgentCall != nil {
			quiescenceErr = waitWorkflowMergerProcessGroup(ctx, registeredProcess.Process, prepared.leaseUntil)
		}

		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		finishErr := claim.store.finishWorkflowProcess(cleanup, claim, execution.Index, registeredProcess.Process, result)
		if errors.Is(finishErr, ErrWorkflowOwned) {
			claim.execution.Lock()
			claim.execution.groupPID = execution.Process.Process.Pid
			claim.execution.Unlock()
		}

		err = errors.Join(err, quiescenceErr, finishErr)
	}

	return result, err
}
