package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	"forge.adrianpk.com/chr/internal/harness"
)

// ProvingRalphLoopOptions describes a package boundary value.
type ProvingRalphLoopOptions struct {
	Worker           harness.Worker
	ProjectDir       string
	ObjectivePath    string
	MaxAttempts      int
	Executor         HarnessExecutor
	ValidationRunner ProvingValidationRunner
	Stdout           io.Writer
	Stderr           io.Writer
}

// ProvingRalphLoopResult describes a package boundary value.
type ProvingRalphLoopResult struct {
	Objective   ProvingObjective
	Attempts    []ProvingRalphLoopAttempt
	Status      string
	StopReason  string
	SummaryPath string
}

// ProvingRalphLoopAttempt describes a package boundary value.
type ProvingRalphLoopAttempt struct {
	Number           int
	ObjectiveID      string
	ArtifactRoot     string
	ContextPath      string
	ResultPath       string
	LogPath          string
	ValidationPath   string
	ValidationPassed bool
	FailureSignature string
	Error            string
}

const (
	// ProvingRalphLoopStatusPassed defines a package boundary constant.
	ProvingRalphLoopStatusPassed = "passed"
	// ProvingRalphLoopStatusFailed defines a package boundary constant.
	ProvingRalphLoopStatusFailed = "failed"
	// ProvingRalphLoopStatusBlocked defines a package boundary constant.
	ProvingRalphLoopStatusBlocked = "blocked"

	// ProvingRalphLoopStopValidationPassed defines a package boundary constant.
	ProvingRalphLoopStopValidationPassed = "validation_passed"
	// ProvingRalphLoopStopRepeatedFailure defines a package boundary constant.
	ProvingRalphLoopStopRepeatedFailure = "repeated_failure"
	// ProvingRalphLoopStopMaxAttempts defines a package boundary constant.
	ProvingRalphLoopStopMaxAttempts = "max_attempts"
	// ProvingRalphLoopStopAgentFailed defines a package boundary constant.
	ProvingRalphLoopStopAgentFailed = "agent_failed"
	// ProvingRalphLoopStopValidationFailed defines a package boundary constant.
	ProvingRalphLoopStopValidationFailed = "validation_failed"
)

// RunProvingRalphLoop executes the corresponding package operation.
func RunProvingRalphLoop(ctx context.Context, opts ProvingRalphLoopOptions) (ProvingRalphLoopResult, error) {
	source, err := loadProvingObjectiveForProject(opts.ProjectDir, opts.ObjectivePath)
	if err != nil {
		return ProvingRalphLoopResult{}, err
	}

	maxAttempts, err := provingRalphLoopMaxAttempts(opts, source.Objective)
	if err != nil {
		return ProvingRalphLoopResult{}, err
	}

	validationRunner := provingRalphLoopValidationRunner(opts)
	result := newProvingRalphLoopResult(source.ProjectDir, source.Objective)
	seenFailures := map[string]bool{}

	var previousValidation *ProvingValidationResult

	for attemptNumber := 1; attemptNumber <= maxAttempts; attemptNumber++ {
		attemptObjective := provingAttemptObjective(source.Objective, attemptNumber, previousValidation)
		outcome := runProvingRalphLoopAttempt(ctx, opts, provingAttemptInput{ProjectDir: source.ProjectDir, Number: attemptNumber, Objective: attemptObjective}, validationRunner)
		result.Attempts = append(result.Attempts, outcome.Attempt)

		if outcome.StopReason != "" {
			result.Status = outcome.Status
			result.StopReason = outcome.StopReason

			break
		}

		if seenFailures[outcome.Attempt.FailureSignature] {
			result.Status = ProvingRalphLoopStatusFailed
			result.StopReason = ProvingRalphLoopStopRepeatedFailure

			break
		}

		seenFailures[outcome.Attempt.FailureSignature] = true
		previousValidation = outcome.Validation
	}

	err = writeProvingRalphLoopSummary(source.ProjectDir, result)
	if err != nil {
		return ProvingRalphLoopResult{}, err
	}

	return result, nil
}

func provingRalphLoopMaxAttempts(opts ProvingRalphLoopOptions, objective ProvingObjective) (int, error) {
	maxAttempts := opts.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = objective.MaxAttempts
	}

	if maxAttempts <= 0 {
		return 0, fmt.Errorf("max attempts must be greater than zero")
	}

	return maxAttempts, nil
}

func provingRalphLoopValidationRunner(opts ProvingRalphLoopOptions) ProvingValidationRunner {
	if opts.ValidationRunner != nil {
		return opts.ValidationRunner
	}

	return RunProvingValidationForObjective
}

func newProvingRalphLoopResult(projectDir string, objective ProvingObjective) ProvingRalphLoopResult {
	summaryPath := filepath.ToSlash(filepath.Join(objective.ArtifactRoot, "evidence", "ralph-loop.md"))

	return ProvingRalphLoopResult{
		Objective:   objective,
		Status:      ProvingRalphLoopStatusFailed,
		StopReason:  ProvingRalphLoopStopMaxAttempts,
		SummaryPath: resolveProjectPath(projectDir, summaryPath),
	}
}

type provingAttemptInput struct {
	ProjectDir string
	Number     int
	Objective  ProvingObjective
}

type provingAttemptOutcome struct {
	Attempt    ProvingRalphLoopAttempt
	Validation *ProvingValidationResult
	Status     string
	StopReason string
}

func runProvingRalphLoopAttempt(ctx context.Context, opts ProvingRalphLoopOptions, input provingAttemptInput, validationRunner ProvingValidationRunner) provingAttemptOutcome {
	attempt := provingRalphLoopAttempt(input.Number, input.Objective)

	sandbox, sandboxErr := RunProvingSandboxForObjective(ctx, ProvingSandboxRunOptions{
		Worker:     opts.Worker,
		ProjectDir: input.ProjectDir,
		DryRun:     false,
		Executor:   opts.Executor,
		Stdout:     opts.Stdout,
		Stderr:     opts.Stderr,
	}, input.ProjectDir, input.Objective)
	if sandboxErr != nil {
		attempt.Error = sandboxErr.Error()

		return provingAttemptOutcome{Attempt: attempt, Status: ProvingRalphLoopStatusBlocked, StopReason: ProvingRalphLoopStopAgentFailed}
	}

	validation, validationErr := validationRunner(ctx, input.ProjectDir, input.Objective)
	if validationErr != nil {
		attempt.Error = validationErr.Error()

		return provingAttemptOutcome{Attempt: attempt, Status: ProvingRalphLoopStatusBlocked, StopReason: ProvingRalphLoopStopValidationFailed}
	}

	attempt.ContextPath = sandbox.Plan.WorkerContextPath
	attempt.ResultPath = sandbox.Plan.ResultPath
	attempt.LogPath = sandbox.Plan.LogPath
	attempt.ValidationPath = validation.LogPath
	attempt.ValidationPassed = validation.Passed
	attempt.FailureSignature = ProvingValidationFailureSignature(validation)

	if validation.Passed {
		return provingAttemptOutcome{Attempt: attempt, Validation: &validation, Status: ProvingRalphLoopStatusPassed, StopReason: ProvingRalphLoopStopValidationPassed}
	}

	return provingAttemptOutcome{Attempt: attempt, Validation: &validation}
}

// RenderProvingRalphLoopResult renders the corresponding package data.
func RenderProvingRalphLoopResult(result ProvingRalphLoopResult) string {
	var output strings.Builder

	writeProvingRalphLoopHeader(&output, result)

	for _, attempt := range result.Attempts {
		writeProvingRalphLoopAttempt(&output, attempt)
	}

	return output.String()
}

func writeProvingRalphLoopHeader(output *strings.Builder, result ProvingRalphLoopResult) {
	output.WriteString("# Proving Ralph Loop Result\n\n")
	output.WriteString("objective_id: ")
	output.WriteString(result.Objective.ID)
	output.WriteString("\n")
	output.WriteString("ticket_key: ")
	output.WriteString(result.Objective.TicketKey)
	output.WriteString("\n")
	output.WriteString("status: ")
	output.WriteString(result.Status)
	output.WriteString("\n")
	output.WriteString("stop_reason: ")
	output.WriteString(result.StopReason)
	output.WriteString("\n")
	output.WriteString("attempts: ")
	output.WriteString(strconv.Itoa(len(result.Attempts)))
	output.WriteString("\n\n")
}

func writeProvingRalphLoopAttempt(output *strings.Builder, attempt ProvingRalphLoopAttempt) {
	output.WriteString("## Attempt ")
	output.WriteString(strconv.Itoa(attempt.Number))
	output.WriteString("\n\n")
	output.WriteString("- objective_id: ")
	output.WriteString(attempt.ObjectiveID)
	output.WriteString("\n")
	output.WriteString("- artifact_root: ")
	output.WriteString(attempt.ArtifactRoot)
	output.WriteString("\n")
	output.WriteString("- context_path: ")
	output.WriteString(attempt.ContextPath)
	output.WriteString("\n")
	output.WriteString("- result_path: ")
	output.WriteString(attempt.ResultPath)
	output.WriteString("\n")
	output.WriteString("- log_path: ")
	output.WriteString(attempt.LogPath)
	output.WriteString("\n")
	output.WriteString("- validation_path: ")
	output.WriteString(attempt.ValidationPath)
	output.WriteString("\n")
	output.WriteString("- validation_passed: ")
	output.WriteString(strconv.FormatBool(attempt.ValidationPassed))
	output.WriteString("\n")
	output.WriteString("- failure_signature: ")
	output.WriteString(attempt.FailureSignature)
	output.WriteString("\n")

	if attempt.Error != "" {
		output.WriteString("- error: ")
		output.WriteString(attempt.Error)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func provingAttemptObjective(objective ProvingObjective, attemptNumber int, previousValidation *ProvingValidationResult) ProvingObjective {
	attempt := objective
	attempt.ID = fmt.Sprintf("%s-attempt-%d", objective.ID, attemptNumber)
	attempt.ContextPath = filepath.ToSlash(filepath.Join(objective.ArtifactRoot, "files", fmt.Sprintf("attempt-%d-context.md", attemptNumber)))
	attempt.ResultPath = filepath.ToSlash(filepath.Join(objective.ArtifactRoot, "evidence", fmt.Sprintf("attempt-%d-result.json", attemptNumber)))
	attempt.LogPath = filepath.ToSlash(filepath.Join(objective.ArtifactRoot, "evidence", fmt.Sprintf("attempt-%d-agent.log", attemptNumber)))
	attempt.ValidationPath = filepath.ToSlash(filepath.Join(objective.ArtifactRoot, "evidence", fmt.Sprintf("attempt-%d-validation.log", attemptNumber)))

	if previousValidation != nil {
		attempt.Body = strings.TrimSpace(objective.Body) + "\n\n## Previous Validation Failure\n\n" +
			RenderProvingValidationFeedback(*previousValidation)
	}

	return attempt
}

func provingRalphLoopAttempt(attemptNumber int, objective ProvingObjective) ProvingRalphLoopAttempt {
	return ProvingRalphLoopAttempt{
		Number:         attemptNumber,
		ObjectiveID:    objective.ID,
		ArtifactRoot:   objective.ArtifactRoot,
		ContextPath:    objective.ContextPath,
		ResultPath:     objective.ResultPath,
		LogPath:        objective.LogPath,
		ValidationPath: objective.ValidationPath,
	}
}

func writeProvingRalphLoopSummary(projectDir string, result ProvingRalphLoopResult) error {
	artifact, err := ensureProvingOpsArtifact(projectDir, result.Objective)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(result.SummaryPath), 0o755)
	if err != nil {
		return fmt.Errorf("create proving ralph loop summary directory: %w", err)
	}

	err = os.WriteFile(result.SummaryPath, []byte(RenderProvingRalphLoopResult(result)), 0o644)
	if err != nil {
		return fmt.Errorf("write proving ralph loop summary: %w", err)
	}

	if err := opscore.RefreshArtifactMetadata(artifact.Dir, result.Status, time.Now()); err != nil {
		return err
	}

	return nil
}
