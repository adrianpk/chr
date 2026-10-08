package core

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
)

const workflowCommandOutputLimit = 16 << 10

type workflowCommandOutput struct {
	buffer    bytes.Buffer
	truncated bool
	file      *os.File
}

func (output *workflowCommandOutput) Write(data []byte) (int, error) {
	length := len(data)

	remaining := workflowCommandOutputLimit - output.buffer.Len()
	if length > remaining {
		data = data[:remaining]
		output.truncated = true
	}

	_, err := output.buffer.Write(data)
	if output.file != nil {
		_, writeErr := output.file.Write(data)
		err = errors.Join(err, writeErr, output.file.Sync())
	}

	return length, err
}

func workflowRunCommand(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, command []string) (result WorkflowCommandResult, err error) {
	run, err := claim.store.ReadWorkflow(ctx, claim.InvocationID)
	if err != nil {
		return WorkflowCommandResult{}, err
	}

	index := len(run.Attempts[len(run.Attempts)-1].Processes)

	stdout, stderr, err := claim.store.workflowProcessOutput(claim, index)
	if err != nil {
		return WorkflowCommandResult{}, err
	}

	defer func() {
		err = errors.Join(err, stdout.file.Close(), stderr.file.Close())
	}()

	gate, permit, err := os.Pipe()
	if err != nil {
		return WorkflowCommandResult{}, err
	}

	defer func() { _ = gate.Close(); _ = permit.Close() }() //nolint:errcheck

	process, guard := workflowSupervisedCommand(ctx, claim, prepared, command, gate)
	process.Stdin = bytes.NewReader(prepared.data)

	process.Stdout, process.Stderr = stdout, stderr
	if err := process.Start(); err != nil {
		return WorkflowCommandResult{}, err
	}

	if err := guard.remember(process.Process.Pid); err != nil {
		return WorkflowCommandResult{}, errors.Join(err, permit.Close(), process.Wait())
	}

	return runWorkflowSupervisedProcess(ctx, claim, prepared, command, process, permit, index, stdout, stderr)
}

func filepathForWorkflowInput(claim *WorkflowClaim, prepared workflowPreparedTask) string {
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

	alive, err := workflowProcessGroupAlive(WorkflowProcess{Identity: WorkflowProcessIdentity{BootID: strings.TrimSpace(string(boot))}, GroupID: pid})

	return err == nil && !alive
}

func runWorkflowSupervisedProcess(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, command []string, process *exec.Cmd, permit *os.File, index int, stdout, stderr *workflowCommandOutput) (WorkflowCommandResult, error) {
	ownership, registered, registerErr := claim.store.registerWorkflowProcess(ctx, claim, prepared, command, process.Process.Pid)
	if registerErr == nil && registered != index {
		registerErr = ErrWorkflowConflict
	}

	if registerErr == nil {
		_, registerErr = fmt.Fprintln(permit, "run")
	}

	closeErr := permit.Close()
	err := errors.Join(registerErr, closeErr, process.Wait())

	result := WorkflowCommandResult{Command: slices.Clone(command), ExitCode: -1, Stdout: stdout.buffer.String(), Stderr: stderr.buffer.String(), Truncated: stdout.truncated || stderr.truncated}
	if process.ProcessState != nil {
		result.ExitCode = process.ProcessState.ExitCode()
	}

	if registerErr == nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		finishErr := claim.store.finishWorkflowProcess(cleanup, claim, index, ownership, result)
		if errors.Is(finishErr, ErrWorkflowOwned) {
			claim.execution.Lock()
			claim.execution.groupPID = process.Process.Pid
			claim.execution.Unlock()
		}

		err = errors.Join(err, finishErr)
	}

	return result, err
}
