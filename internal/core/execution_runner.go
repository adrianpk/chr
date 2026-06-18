package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// HarnessExecutor defines a package boundary interface.
type HarnessExecutor interface {
	Run(ctx context.Context, command HarnessCommand, stdout, stderr io.Writer) error
}

// OSHarnessExecutor describes a package boundary value.
type OSHarnessExecutor struct{}

// Run executes the configured operation.
func (e OSHarnessExecutor) Run(ctx context.Context, command HarnessCommand, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, command.Program, command.Args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		return NewExecutionFailure(ExecutionFailureHarness, fmt.Errorf("run harness command %s: %w", command.Program, err))
	}

	return nil
}

// ExecutionRunner describes a package boundary value.
type ExecutionRunner struct {
	Repository ExecutionRepository
	Executor   HarnessExecutor
	EventSink  AgentEventSink
	Now        func() time.Time
}

// RunExecutionOptions describes a package boundary value.
type RunExecutionOptions struct {
	Record     ExecutionRecord
	Invocation HarnessInvocation
	DryRun     bool
}

// NewExecutionRunner constructs the corresponding package service or value.
func NewExecutionRunner(repository ExecutionRepository, executor HarnessExecutor) ExecutionRunner {
	return ExecutionRunner{
		Repository: repository,
		Executor:   executor,
		Now:        time.Now,
	}
}

// Run executes the configured operation.
func (r ExecutionRunner) Run(ctx context.Context, opts RunExecutionOptions) (ExecutionRecord, error) {
	record := opts.Record
	if opts.DryRun {
		return record, nil
	}

	err := r.validate()
	if err != nil {
		return ExecutionRecord{}, err
	}

	err = createExecutionArtifactDir(record)
	if err != nil {
		return ExecutionRecord{}, err
	}

	logFile, validationFile, err := openExecutionOutputArtifacts(record.Artifacts)
	if err != nil {
		return ExecutionRecord{}, err
	}

	defer func() {
		//nolint:errcheck // Main command result is already determined by execution.
		_ = logFile.Close()
	}()
	defer func() {
		//nolint:errcheck // Main command result is already determined by execution.
		_ = validationFile.Close()
	}()

	record, err = r.markExecutionRunning(ctx, record)
	if err != nil {
		return ExecutionRecord{}, err
	}

	err = r.runHarness(ctx, opts.Invocation, logFile, validationFile)
	if err != nil {
		return r.failExecution(ctx, record, err)
	}

	err = captureExecutionResult(record)
	if err != nil {
		return r.failExecution(ctx, record, err)
	}

	return r.markExecutionSucceeded(ctx, record)
}

func (r ExecutionRunner) validate() error {
	if r.Repository == nil {
		return fmt.Errorf("execution repository is required")
	}

	if r.Executor == nil {
		return fmt.Errorf("harness executor is required")
	}

	return nil
}

func createExecutionArtifactDir(record ExecutionRecord) error {
	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		return fmt.Errorf("create execution artifact directory: %w", err)
	}

	return nil
}

func (r ExecutionRunner) markExecutionRunning(
	ctx context.Context,
	record ExecutionRecord,
) (ExecutionRecord, error) {
	startedAt := r.now()
	record.Status = ExecutionStatusRunning
	record.StartedAt = &startedAt
	record.UpdatedAt = startedAt

	err := r.Repository.UpdateExecution(ctx, record)
	if err != nil {
		return ExecutionRecord{}, err
	}

	return record, nil
}

func (r ExecutionRunner) markExecutionSucceeded(
	ctx context.Context,
	record ExecutionRecord,
) (ExecutionRecord, error) {
	finishedAt := r.now()
	record.Status = ExecutionStatusSucceeded
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt

	err := r.Repository.UpdateExecution(ctx, record)
	if err != nil {
		return ExecutionRecord{}, err
	}

	return record, nil
}

func openExecutionOutputArtifacts(artifacts ExecutionArtifacts) (*os.File, *os.File, error) {
	logFile, err := os.OpenFile(artifacts.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open execution log artifact: %w", err)
	}

	validationFile, err := os.OpenFile(artifacts.ValidationPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		closeErr := logFile.Close()
		if closeErr != nil {
			return nil, nil, fmt.Errorf("open validation artifact: %w; close execution log artifact: %w", err, closeErr)
		}

		return nil, nil, fmt.Errorf("open validation artifact: %w", err)
	}

	return logFile, validationFile, nil
}

func (r ExecutionRunner) runHarness(ctx context.Context, invocation HarnessInvocation, log, validationLog io.Writer) error {
	err := r.Executor.Run(ctx, invocation.Build, log, log)
	if err != nil {
		return harnessFailure(err)
	}

	err = r.Executor.Run(ctx, invocation.Run, log, log)
	if err != nil {
		return workerFailure(err)
	}

	err = r.Executor.Run(ctx, invocation.Validation, validationLog, validationLog)
	if err != nil {
		return validationFailure(err)
	}

	return nil
}

func harnessFailure(err error) error {
	var failure ExecutionFailure
	if errors.As(err, &failure) {
		return err
	}

	return NewExecutionFailure(ExecutionFailureHarness, err)
}

func workerFailure(err error) error {
	var failure ExecutionFailure
	if errors.As(err, &failure) {
		if failure.Kind == ExecutionFailureWorker {
			return err
		}

		return NewExecutionFailure(ExecutionFailureWorker, failure.Err)
	}

	return NewExecutionFailure(ExecutionFailureWorker, err)
}

func validationFailure(err error) error {
	var failure ExecutionFailure
	if errors.As(err, &failure) {
		if failure.Kind == ExecutionFailureValidation {
			return err
		}

		return NewExecutionFailure(ExecutionFailureValidation, failure.Err)
	}

	return NewExecutionFailure(ExecutionFailureValidation, err)
}

func captureExecutionResult(record ExecutionRecord) error {
	result, err := LoadExecutionResult(record.Artifacts.ResultPath)
	if err != nil {
		return NewExecutionFailure(ExecutionFailureResultCapture, err)
	}

	return validateExecutionResultForRecord(result, record)
}

func validateExecutionResultForRecord(result ExecutionResult, record ExecutionRecord) error {
	if result.ExecutionID != record.ID {
		return NewExecutionFailure(
			ExecutionFailureResultCapture,
			fmt.Errorf("execution result id mismatch: %s", result.ExecutionID),
		)
	}

	if result.TicketKey != record.TicketKey {
		return NewExecutionFailure(
			ExecutionFailureResultCapture,
			fmt.Errorf("execution result ticket key mismatch: %s", result.TicketKey),
		)
	}

	if result.Status == ExecutionResultStatusFailed {
		return NewExecutionFailure(ExecutionFailureValidation, fmt.Errorf("execution result failed: %s", result.Summary))
	}

	return nil
}

func (r ExecutionRunner) failExecution(ctx context.Context, record ExecutionRecord, cause error) (ExecutionRecord, error) {
	finishedAt := r.now()
	record.Status = ExecutionStatusFailed
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt
	record.FailureReason = cause.Error()

	err := r.Repository.UpdateExecution(ctx, record)
	if err != nil {
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
