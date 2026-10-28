package workflow

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
)

type workflowProviderCommandRunner struct{}

func (runner workflowProviderCommandRunner) Output(ctx context.Context, command string, args ...string) ([]byte, error) {
	return runner.Input(ctx, command, nil, args...)
}

func (workflowProviderCommandRunner) Input(ctx context.Context, command string, input []byte, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	process := exec.CommandContext(ctx, command, args...)
	stdout, stderr := &opsfile.ProviderOutput{}, &opsfile.CommandOutput{}
	process.Stdin, process.Stdout, process.Stderr = bytes.NewReader(input), stdout, stderr

	process.WaitDelay = time.Second
	if err := process.Run(); err != nil {
		return nil, fmt.Errorf("run workflow provider: %w", err)
	}

	if stdout.Overflow {
		return nil, fmt.Errorf("workflow provider response exceeds file bounds")
	}

	return stdout.Buffer.Bytes(), nil
}

func (adapter RemoteReviewAdapter) workflowProviderRunner() ReviewRemoteCommandRunner {
	switch adapter.runner.(type) {
	case OSReviewRemoteCommandRunner, *OSReviewRemoteCommandRunner:
		return workflowProviderCommandRunner{}
	default:
		return adapter.runner
	}
}
