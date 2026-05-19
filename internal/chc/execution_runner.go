package chc

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type HarnessExecutor interface {
	Run(ctx context.Context, command HarnessCommand, stdout, stderr io.Writer) error
}

type OSHarnessExecutor struct{}

func (e OSHarnessExecutor) Run(ctx context.Context, command HarnessCommand, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, command.Program, command.Args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run harness command %s: %w", command.Program, err)
	}

	return nil
}

type ExecutionRunner struct {
	Repository ExecutionRepository
	Executor   HarnessExecutor
	Now        func() time.Time
}

type RunExecutionOptions struct {
	Record     ExecutionRecord
	Invocation HarnessInvocation
	DryRun     bool
}

func NewExecutionRunner(repository ExecutionRepository, executor HarnessExecutor) ExecutionRunner {
	return ExecutionRunner{
		Repository: repository,
		Executor:   executor,
		Now:        time.Now,
	}
}

func (r ExecutionRunner) Run(ctx context.Context, opts RunExecutionOptions) (ExecutionRecord, error) {
	record := opts.Record
	if opts.DryRun {
		return record, nil
	}

	if r.Repository == nil {
		return ExecutionRecord{}, fmt.Errorf("execution repository is required")
	}

	if r.Executor == nil {
		return ExecutionRecord{}, fmt.Errorf("harness executor is required")
	}

	if err := os.MkdirAll(record.Artifacts.RunDir, 0o755); err != nil {
		return ExecutionRecord{}, fmt.Errorf("create execution artifact directory: %w", err)
	}

	logFile, err := os.OpenFile(record.Artifacts.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return ExecutionRecord{}, fmt.Errorf("open execution log artifact: %w", err)
	}

	defer func() {
		//nolint:errcheck // Main command result is already determined by execution.
		_ = logFile.Close()
	}()

	startedAt := r.now()
	record.Status = ExecutionStatusRunning
	record.StartedAt = &startedAt
	record.UpdatedAt = startedAt

	if err := r.Repository.UpdateExecution(ctx, record); err != nil {
		return ExecutionRecord{}, err
	}

	if err := r.runHarness(ctx, opts.Invocation, logFile); err != nil {
		return r.failExecution(ctx, record, err)
	}

	finishedAt := r.now()
	record.Status = ExecutionStatusSucceeded
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt

	if err := r.Repository.UpdateExecution(ctx, record); err != nil {
		return ExecutionRecord{}, err
	}

	return record, nil
}

func (r ExecutionRunner) runHarness(ctx context.Context, invocation HarnessInvocation, log io.Writer) error {
	if err := r.Executor.Run(ctx, invocation.Build, log, log); err != nil {
		return err
	}

	if err := r.Executor.Run(ctx, invocation.Run, log, log); err != nil {
		return err
	}

	return nil
}

func (r ExecutionRunner) failExecution(ctx context.Context, record ExecutionRecord, cause error) (ExecutionRecord, error) {
	finishedAt := r.now()
	record.Status = ExecutionStatusFailed
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt
	record.FailureReason = cause.Error()

	if err := r.Repository.UpdateExecution(ctx, record); err != nil {
		return ExecutionRecord{}, err
	}

	return record, cause
}

func (r ExecutionRunner) now() time.Time {
	if r.Now != nil {
		return r.Now().UTC()
	}

	return time.Now().UTC()
}
