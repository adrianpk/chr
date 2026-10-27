package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
)

// HarnessExecutor defines a package boundary interface.
type HarnessExecutor interface {
	Run(ctx context.Context, command harnessruntime.HarnessCommand, stdout, stderr io.Writer) error
}

// OSHarnessExecutor describes a package boundary value.
type OSHarnessExecutor struct{}

// Run executes the configured operation.
func (e OSHarnessExecutor) Run(ctx context.Context, command harnessruntime.HarnessCommand, stdout, stderr io.Writer) error {
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
	Invocation harnessruntime.HarnessInvocation
	LoopID     string
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

	err = r.runHarness(ctx, opts, record, logFile, validationFile)
	if err != nil {
		return r.failExecution(ctx, record, err)
	}

	result, err := captureExecutionResult(record)
	if err != nil {
		r.emitExecutionAgentEvent(ctx, opts.LoopID, record, AgentEventPhaseWorkFailed, AgentEventLevelError,
			fmt.Sprintf("%s result capture failed", record.TicketKey))

		return r.failExecution(ctx, record, err)
	}

	r.emitChangedFilesDetected(ctx, opts.LoopID, record, result)

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

func (r ExecutionRunner) runHarness(ctx context.Context, opts RunExecutionOptions, record ExecutionRecord, log, validationLog io.Writer) error {
	err := r.Executor.Run(ctx, opts.Invocation.Build, log, log)
	if err != nil {
		failure := harnessFailure(err)

		r.emitExecutionAgentEvent(ctx, opts.LoopID, record, AgentEventPhaseWorkFailed, AgentEventLevelError,
			fmt.Sprintf("%s runtime preparation failed", record.TicketKey))

		return failure
	}

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, AgentEventPhaseRuntimePrepared, AgentEventLevelInfo,
		fmt.Sprintf("%s runtime prepared", record.TicketKey))

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, AgentEventPhaseWorkerStarted, AgentEventLevelInfo,
		fmt.Sprintf("%s worker started", record.TicketKey))

	err = r.Executor.Run(ctx, opts.Invocation.Run, log, log)
	if err != nil {
		failure := workerFailure(err)

		r.emitExecutionAgentEvent(ctx, opts.LoopID, record, AgentEventPhaseWorkerFinished, AgentEventLevelError,
			fmt.Sprintf("%s worker failed", record.TicketKey))

		return failure
	}

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, AgentEventPhaseWorkerFinished, AgentEventLevelInfo,
		fmt.Sprintf("%s worker finished", record.TicketKey))

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, AgentEventPhaseValidationStarted, AgentEventLevelInfo,
		fmt.Sprintf("%s validation started: %s", record.TicketKey, record.Runtime.ValidationCommand))

	err = r.Executor.Run(ctx, opts.Invocation.Validation, validationLog, validationLog)
	if err != nil {
		failure := validationFailure(err)

		r.emitExecutionAgentEvent(ctx, opts.LoopID, record, AgentEventPhaseValidationFinished, AgentEventLevelError,
			fmt.Sprintf("%s validation failed: %s", record.TicketKey, record.Runtime.ValidationCommand))

		return failure
	}

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, AgentEventPhaseValidationFinished, AgentEventLevelInfo,
		fmt.Sprintf("%s validation passed: %s", record.TicketKey, record.Runtime.ValidationCommand))

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

func captureExecutionResult(record ExecutionRecord) (ExecutionResult, error) {
	result, err := LoadExecutionResult(record.Artifacts.ResultPath)
	if err != nil {
		return ExecutionResult{}, NewExecutionFailure(ExecutionFailureResultCapture, err)
	}

	err = validateExecutionResultForRecord(result, record)
	if err != nil {
		return ExecutionResult{}, err
	}

	return result, nil
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

func (r ExecutionRunner) emitChangedFilesDetected(ctx context.Context, loopID string, record ExecutionRecord, result ExecutionResult) {
	if len(result.ChangedFiles) == 0 {
		return
	}

	r.emitExecutionAgentEvent(ctx, loopID, record, AgentEventPhaseChangedFilesDetected, AgentEventLevelInfo,
		fmt.Sprintf("%s changed files detected: %d", record.TicketKey, len(result.ChangedFiles)))
}

func (r ExecutionRunner) emitExecutionAgentEvent(ctx context.Context, loopID string, record ExecutionRecord, phase AgentEventPhase, level AgentEventLevel, message string) {
	if r.EventSink == nil {
		return
	}

	err := r.EventSink.EmitAgentEvent(ctx, AgentEvent{
		SchemaVersion: AgentEventSchemaVersion,
		Timestamp:     time.Now().UTC(),
		ProjectID:     record.ProjectID,
		LoopID:        loopID,
		ExecutionID:   record.ID,
		TicketKey:     record.TicketKey,
		Phase:         phase,
		Level:         level,
		Message:       message,
		Refs:          executionAgentEventRefs(record),
	})
	if err != nil {
		return
	}
}

func executionAgentEventRefs(record ExecutionRecord) AgentEventRefs {
	return compactAgentEventRefs(AgentEventRefs{
		"ticket_path":           record.TicketPath,
		"context_path":          filepath.Join(record.Artifacts.RunDir, "context.md"),
		"result_path":           record.Artifacts.ResultPath,
		"runtime_log_path":      record.Artifacts.LogPath,
		"validation_command":    record.Runtime.ValidationCommand,
		"validation_log_path":   record.Artifacts.ValidationPath,
		"execution_report_path": ExecutionReportPath(record),
	})
}

func compactAgentEventRefs(refs AgentEventRefs) AgentEventRefs {
	for key, value := range refs {
		if value == "" {
			delete(refs, key)
		}
	}

	if len(refs) == 0 {
		return nil
	}

	return refs
}

func (r ExecutionRunner) now() time.Time {
	if r.Now != nil {
		return r.Now().UTC()
	}

	return time.Now().UTC()
}
