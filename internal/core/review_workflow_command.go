package core

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type workflowProviderCommandRunner struct{}

type workflowProviderOutput struct {
	buffer   bytes.Buffer
	overflow bool
}

func (output *workflowProviderOutput) Write(data []byte) (int, error) {
	length := len(data)

	remaining := WorkflowFileLimit - output.buffer.Len()
	if length > remaining {
		data, output.overflow = data[:remaining], true
	}

	_, err := output.buffer.Write(data)

	return length, err
}

func (runner workflowProviderCommandRunner) Output(ctx context.Context, command string, args ...string) ([]byte, error) {
	return runner.Input(ctx, command, nil, args...)
}

func (workflowProviderCommandRunner) Input(ctx context.Context, command string, input []byte, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	process := exec.CommandContext(ctx, command, args...)
	stdout, stderr := &workflowProviderOutput{}, &workflowCommandOutput{}
	process.Stdin, process.Stdout, process.Stderr = bytes.NewReader(input), stdout, stderr

	process.WaitDelay = time.Second
	if err := process.Run(); err != nil {
		return nil, fmt.Errorf("run workflow provider: %w", err)
	}

	if stdout.overflow {
		return nil, fmt.Errorf("workflow provider response exceeds file bounds")
	}

	return stdout.buffer.Bytes(), nil
}

func (adapter RemoteReviewAdapter) workflowProviderRunner() ReviewRemoteCommandRunner {
	switch adapter.runner.(type) {
	case OSReviewRemoteCommandRunner, *OSReviewRemoteCommandRunner:
		return workflowProviderCommandRunner{}
	default:
		return adapter.runner
	}
}
