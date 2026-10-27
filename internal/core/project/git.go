package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
)

// GitReader executes the existing bounded Git transport with caller-owned directory and environment.
type GitReader struct {
	Directory   string
	Environment []string
}

// Output retains trimmed original Git output.
func (reader GitReader) Output(ctx context.Context, args ...string) (string, error) {
	data, err := reader.RawOutput(ctx, args...)

	return strings.TrimSpace(data), err
}

// RawOutput retains original Git output bytes without trimming.
func (reader GitReader) RawOutput(ctx context.Context, args ...string) (string, error) {
	return reader.RawInput(ctx, "", args...)
}

// RawInput retains original bounded Git execution with explicit input.
func (reader GitReader) RawInput(ctx context.Context, input string, args ...string) (string, error) {
	command := reader.Process(ctx, args...)
	command.Stdin = strings.NewReader(input)
	stdout, stderr := &opsfile.ProviderOutput{}, &opsfile.CommandOutput{}
	command.Stdout, command.Stderr = stdout, stderr

	if err := command.Run(); err != nil {
		return "", fmt.Errorf("integration Git observation failed: %w", err)
	}

	if stdout.Overflow {
		return "", fmt.Errorf("integration Git output exceeds file bounds")
	}

	return stdout.Buffer.String(), nil
}

// Process retains the original Git process flags, environment and cancellation.
func (reader GitReader) Process(ctx context.Context, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, "git", append([]string{"--no-replace-objects", "--literal-pathspecs", "-c", "core.hooksPath=/dev/null", "-C", reader.Directory}, args...)...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}

		return err
	}

	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_NO_REPLACE_OBJECTS=1")
	if reader.Environment != nil {
		command.Env = reader.Environment
	}

	command.WaitDelay = time.Second

	return command
}
