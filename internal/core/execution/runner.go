package execution

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
	Run(ctx context.Context, command harnessruntime.Command, stdout, stderr io.Writer) error
}

// OSHarnessExecutor describes a package boundary value.
type OSHarnessExecutor struct{}

// Run executes the configured operation.
func (e OSHarnessExecutor) Run(ctx context.Context, command harnessruntime.Command, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, command.Program, command.Args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		return NewFailure(FailureHarness, fmt.Errorf("run harness command %s: %w", command.Program, err))
	}

	return nil
}

// Runner describes a package boundary value.
type Runner struct {
	Repository Repository
	Executor   HarnessExecutor
	EventSink  AgentEventSink
	Now        func() time.Time
}

// RunOptions describes a package boundary value.
type RunOptions struct {
	Record     Record
	Invocation harnessruntime.Invocation
	LoopID     string
	DryRun     bool
}

// NewRunner constructs the corresponding package service or value.
func NewRunner(repository Repository, executor HarnessExecutor) Runner {
	return Runner{
		Repository: repository,
		Executor:   executor,
		Now:        time.Now,
	}
}

// Run executes the configured operation.
func (r Runner) Run(ctx context.Context, opts RunOptions) (Record, error) {
	record := opts.Record
	if opts.DryRun {
		return record, nil
	}

	err := r.validate()
	if err != nil {
		return Record{}, err
	}

	err = createExecutionArtifactDir(record)
	if err != nil {
		return Record{}, err
	}

	output, err := openExecutionOutputArtifacts(record.Artifacts)
	if err != nil {
		return Record{}, err
	}

	defer func() {
		//nolint:errcheck // Main command result is already determined by execution.
		_ = output.log.Close()
	}()
	defer func() {
		//nolint:errcheck // Main command result is already determined by execution.
		_ = output.validation.Close()
	}()

	record, err = r.markExecutionRunning(ctx, record)
	if err != nil {
		return Record{}, err
	}

	err = r.runHarness(ctx, opts, record, output)
	if err != nil {
		return r.failExecution(ctx, record, err)
	}

	result, err := captureExecutionResult(record)
	if err != nil {
		r.emitExecutionAgentEvent(ctx, opts.LoopID, record, executionEventMessage{phase: AgentEventPhaseWorkFailed, level: AgentEventLevelError, message: fmt.Sprintf("%s result capture failed", record.TicketKey)})

		return r.failExecution(ctx, record, err)
	}

	r.emitChangedFilesDetected(ctx, opts.LoopID, record, result)

	return r.markExecutionSucceeded(ctx, record)
}

func (r Runner) validate() error {
	if r.Repository == nil {
		return fmt.Errorf("execution repository is required")
	}

	if r.Executor == nil {
		return fmt.Errorf("harness executor is required")
	}

	return nil
}

func createExecutionArtifactDir(record Record) error {
	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		return fmt.Errorf("create execution artifact directory: %w", err)
	}

	return nil
}

func (r Runner) markExecutionRunning(ctx context.Context, record Record) (Record, error) {
	startedAt := r.now()
	record.Status = StatusRunning
	record.StartedAt = &startedAt
	record.UpdatedAt = startedAt

	err := r.Repository.UpdateExecution(ctx, record)
	if err != nil {
		return Record{}, err
	}

	return record, nil
}

func (r Runner) markExecutionSucceeded(ctx context.Context, record Record) (Record, error) {
	finishedAt := r.now()
	record.Status = StatusSucceeded
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt

	err := r.Repository.UpdateExecution(ctx, record)
	if err != nil {
		return Record{}, err
	}

	return record, nil
}

type executionOutputArtifacts struct {
	log        *os.File
	validation *os.File
}

func openExecutionOutputArtifacts(artifacts Artifacts) (executionOutputArtifacts, error) {
	logFile, err := os.OpenFile(artifacts.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return executionOutputArtifacts{}, fmt.Errorf("open execution log artifact: %w", err)
	}

	validationFile, err := os.OpenFile(artifacts.ValidationPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		closeErr := logFile.Close()
		if closeErr != nil {
			return executionOutputArtifacts{}, fmt.Errorf("open validation artifact: %w; close execution log artifact: %w", err, closeErr)
		}

		return executionOutputArtifacts{}, fmt.Errorf("open validation artifact: %w", err)
	}

	return executionOutputArtifacts{log: logFile, validation: validationFile}, nil
}

func (r Runner) runHarness(ctx context.Context, opts RunOptions, record Record, output executionOutputArtifacts) error {
	err := r.Executor.Run(ctx, opts.Invocation.Build, output.log, output.log)
	if err != nil {
		failure := harnessFailure(err)

		r.emitExecutionAgentEvent(ctx, opts.LoopID, record, executionEventMessage{phase: AgentEventPhaseWorkFailed, level: AgentEventLevelError, message: fmt.Sprintf("%s runtime preparation failed", record.TicketKey)})

		return failure
	}

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, executionEventMessage{phase: AgentEventPhaseRuntimePrepared, level: AgentEventLevelInfo, message: fmt.Sprintf("%s runtime prepared", record.TicketKey)})

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, executionEventMessage{phase: AgentEventPhaseWorkerStarted, level: AgentEventLevelInfo, message: fmt.Sprintf("%s worker started", record.TicketKey)})

	err = r.Executor.Run(ctx, opts.Invocation.Run, output.log, output.log)
	if err != nil {
		failure := workerFailure(err)

		r.emitExecutionAgentEvent(ctx, opts.LoopID, record, executionEventMessage{phase: AgentEventPhaseWorkerFinished, level: AgentEventLevelError, message: fmt.Sprintf("%s worker failed", record.TicketKey)})

		return failure
	}

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, executionEventMessage{phase: AgentEventPhaseWorkerFinished, level: AgentEventLevelInfo, message: fmt.Sprintf("%s worker finished", record.TicketKey)})

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, executionEventMessage{phase: AgentEventPhaseValidationStarted, level: AgentEventLevelInfo, message: fmt.Sprintf("%s validation started: %s", record.TicketKey, record.Runtime.ValidationCommand)})

	err = r.Executor.Run(ctx, opts.Invocation.Validation, output.validation, output.validation)
	if err != nil {
		failure := validationFailure(err)

		r.emitExecutionAgentEvent(ctx, opts.LoopID, record, executionEventMessage{phase: AgentEventPhaseValidationFinished, level: AgentEventLevelError, message: fmt.Sprintf("%s validation failed: %s", record.TicketKey, record.Runtime.ValidationCommand)})

		return failure
	}

	r.emitExecutionAgentEvent(ctx, opts.LoopID, record, executionEventMessage{phase: AgentEventPhaseValidationFinished, level: AgentEventLevelInfo, message: fmt.Sprintf("%s validation passed: %s", record.TicketKey, record.Runtime.ValidationCommand)})

	return nil
}

func harnessFailure(err error) error {
	var failure Failure
	if errors.As(err, &failure) {
		return err
	}

	return NewFailure(FailureHarness, err)
}

func workerFailure(err error) error {
	var failure Failure
	if errors.As(err, &failure) {
		if failure.Kind == FailureWorker {
			return err
		}

		return NewFailure(FailureWorker, failure.Err)
	}

	return NewFailure(FailureWorker, err)
}

func validationFailure(err error) error {
	var failure Failure
	if errors.As(err, &failure) {
		if failure.Kind == FailureValidation {
			return err
		}

		return NewFailure(FailureValidation, failure.Err)
	}

	return NewFailure(FailureValidation, err)
}

func captureExecutionResult(record Record) (Result, error) {
	result, err := LoadResult(record.Artifacts.ResultPath)
	if err != nil {
		return Result{}, NewFailure(FailureResultCapture, err)
	}

	err = validateExecutionResultForRecord(result, record)
	if err != nil {
		return Result{}, err
	}

	return result, nil
}

func validateExecutionResultForRecord(result Result, record Record) error {
	if result.ExecutionID != record.ID {
		return NewFailure(
			FailureResultCapture,
			fmt.Errorf("execution result id mismatch: %s", result.ExecutionID),
		)
	}

	if result.TicketKey != record.TicketKey {
		return NewFailure(
			FailureResultCapture,
			fmt.Errorf("execution result ticket key mismatch: %s", result.TicketKey),
		)
	}

	if result.Status == ResultStatusFailed {
		return NewFailure(FailureValidation, fmt.Errorf("execution result failed: %s", result.Summary))
	}

	return nil
}

func (r Runner) failExecution(ctx context.Context, record Record, cause error) (Record, error) {
	finishedAt := r.now()
	record.Status = StatusFailed
	record.FinishedAt = &finishedAt
	record.UpdatedAt = finishedAt
	record.FailureReason = cause.Error()

	err := r.Repository.UpdateExecution(ctx, record)
	if err != nil {
		return Record{}, err
	}

	return record, cause
}

func (r Runner) emitChangedFilesDetected(ctx context.Context, loopID string, record Record, result Result) {
	if len(result.ChangedFiles) == 0 {
		return
	}

	r.emitExecutionAgentEvent(ctx, loopID, record, executionEventMessage{phase: AgentEventPhaseChangedFilesDetected, level: AgentEventLevelInfo, message: fmt.Sprintf("%s changed files detected: %d", record.TicketKey, len(result.ChangedFiles))})
}

type executionEventMessage struct {
	phase   AgentEventPhase
	level   AgentEventLevel
	message string
}

func (r Runner) emitExecutionAgentEvent(ctx context.Context, loopID string, record Record, message executionEventMessage) {
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
		Phase:         message.phase,
		Level:         message.level,
		Message:       message.message,
		Refs:          executionAgentEventRefs(record),
	})
	if err != nil {
		return
	}
}

func executionAgentEventRefs(record Record) AgentEventRefs {
	return compactAgentEventRefs(AgentEventRefs{
		"ticket_path":           record.TicketPath,
		"context_path":          filepath.Join(record.Artifacts.RunDir, "context.md"),
		"result_path":           record.Artifacts.ResultPath,
		"runtime_log_path":      record.Artifacts.LogPath,
		"validation_command":    record.Runtime.ValidationCommand,
		"validation_log_path":   record.Artifacts.ValidationPath,
		"execution_report_path": ReportPath(record),
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

func (r Runner) now() time.Time {
	if r.Now != nil {
		return r.Now().UTC()
	}

	return time.Now().UTC()
}
