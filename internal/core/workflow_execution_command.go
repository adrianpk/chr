package core

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"slices"
	"syscall"
	"time"
)

const workflowCommandOutputLimit = 16 << 10

type workflowCommandOutput struct {
	buffer    bytes.Buffer
	truncated bool
}

func (output *workflowCommandOutput) Write(data []byte) (int, error) {
	length := len(data)

	remaining := workflowCommandOutputLimit - output.buffer.Len()
	if length > remaining {
		data = data[:remaining]
		output.truncated = true
	}

	_, err := output.buffer.Write(data)

	return length, err
}

func workflowRunCommand(ctx context.Context, claim *WorkflowClaim, prepared workflowPreparedTask, command []string) (WorkflowCommandResult, error) {
	stdout, stderr := &workflowCommandOutput{}, &workflowCommandOutput{}
	process := exec.CommandContext(ctx, command[0], command[1:]...)
	process.Dir = prepared.input.Lane.Worktree
	process.Stdin = bytes.NewReader(prepared.data)
	process.Stdout, process.Stderr = stdout, stderr

	process.Env = append(os.Environ(), "CHR_WORKFLOW_INPUT="+filepathForWorkflowInput(claim, prepared))
	process.ExtraFiles = []*os.File{claim.reservation}
	process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	process.WaitDelay = time.Second
	process.Cancel = func() error { return workflowKillCommandGroup(process) }
	err := process.Run()

	result := WorkflowCommandResult{Command: slices.Clone(command), ExitCode: -1, Stdout: stdout.buffer.String(), Stderr: stderr.buffer.String(), Truncated: stdout.truncated || stderr.truncated}
	if process.ProcessState != nil {
		result.ExitCode = process.ProcessState.ExitCode()
	}

	if process.Process != nil && !workflowCommandGroupStopped(process.Process.Pid) {
		claim.execution.Lock()
		claim.execution.groupPID = process.Process.Pid
		claim.execution.Unlock()

		err = errors.Join(err, ErrWorkflowOwned)
	}

	return result, err
}

func filepathForWorkflowInput(claim *WorkflowClaim, prepared workflowPreparedTask) string {
	return claim.store.layout.Root() + string(os.PathSeparator) + prepared.inputRef
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
	return errors.Is(syscall.Kill(-pid, 0), syscall.ESRCH)
}
