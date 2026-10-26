package core

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
)

// ProvingObjective describes a package boundary value.
type ProvingObjective struct {
	ID             string
	TicketKey      string
	Kind           string
	ArtifactRoot   string
	ContextPath    string
	ResultPath     string
	LogPath        string
	ValidationPath string
	MaxAttempts    int
	Body           string
}

// ProvingSandboxRunOptions describes a package boundary value.
type ProvingSandboxRunOptions struct {
	Worker        harness.Worker
	ProjectDir    string
	ObjectivePath string
	DryRun        bool
	Executor      HarnessExecutor
	Stdout        io.Writer
	Stderr        io.Writer
}

// ProvingSandboxRunResult describes a package boundary value.
type ProvingSandboxRunResult struct {
	Objective  ProvingObjective
	Plan       RuntimePlan
	Build      HarnessCommand
	Run        HarnessCommand
	TicketPath string
	DryRun     bool
}

// ProvingValidationCommandResult describes a package boundary value.
type ProvingValidationCommandResult struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
}

// ProvingValidationResult describes a package boundary value.
type ProvingValidationResult struct {
	Objective ProvingObjective
	Commands  []ProvingValidationCommandResult
	Passed    bool
	LogPath   string
}

// ProvingValidationRunner defines a package boundary type.
type ProvingValidationRunner func(context.Context, string, ProvingObjective) (ProvingValidationResult, error)

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

// LoadProvingObjective loads the corresponding package data.
func LoadProvingObjective(path string) (ProvingObjective, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ProvingObjective{}, fmt.Errorf("read proving objective: %w", err)
	}

	objective, err := ParseProvingObjective(string(content))
	if err != nil {
		return ProvingObjective{}, fmt.Errorf("parse proving objective %s: %w", path, err)
	}

	return objective, nil
}

// ParseProvingObjective parses the corresponding package data.
func ParseProvingObjective(content string) (ProvingObjective, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return ProvingObjective{}, fmt.Errorf("objective frontmatter is required")
	}

	rest := strings.TrimPrefix(normalized, "---\n")

	parts := strings.SplitN(rest, "\n---\n", 2)
	if len(parts) != 2 {
		return ProvingObjective{}, fmt.Errorf("objective frontmatter terminator is required")
	}

	fields := parseSimpleFrontmatter(parts[0])

	maxAttempts, err := strconv.Atoi(fields["max_attempts"])
	if err != nil || maxAttempts <= 0 {
		return ProvingObjective{}, fmt.Errorf("max_attempts must be a positive integer")
	}

	objective := ProvingObjective{
		ID:             fields["objective_id"],
		TicketKey:      fields["ticket_key"],
		Kind:           fields["kind"],
		ArtifactRoot:   fields["artifact_root"],
		ContextPath:    fields["context_path"],
		ResultPath:     fields["result_path"],
		LogPath:        fields["log_path"],
		ValidationPath: fields["validation_path"],
		MaxAttempts:    maxAttempts,
		Body:           strings.TrimSpace(parts[1]),
	}

	err = validateProvingObjective(objective)
	if err != nil {
		return ProvingObjective{}, err
	}

	return objective, nil
}

// RunProvingSandbox executes the corresponding package operation.
func RunProvingSandbox(ctx context.Context, opts ProvingSandboxRunOptions) (ProvingSandboxRunResult, error) {
	source, err := loadProvingObjectiveForProject(opts.ProjectDir, opts.ObjectivePath)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return RunProvingSandboxForObjective(ctx, opts, source.ProjectDir, source.Objective)
}

// RunProvingSandboxForObjective executes the corresponding package operation.
func RunProvingSandboxForObjective(
	ctx context.Context,
	opts ProvingSandboxRunOptions,
	projectDir string,
	objective ProvingObjective,
) (ProvingSandboxRunResult, error) {
	result, err := PlanProvingSandboxRunForObjective(projectDir, objective, opts.Worker)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	result.DryRun = opts.DryRun
	if opts.DryRun {
		return result, nil
	}

	logFile, err := prepareProvingArtifacts(result)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	defer func() {
		//nolint:errcheck // Main command result is already determined by execution.
		_ = logFile.Close()
	}()

	err = runProvingSandboxCommands(ctx, result, opts, logFile)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return result, nil
}

// RunProvingValidation executes the corresponding package operation.
func RunProvingValidation(ctx context.Context, projectDir, objectivePath string) (ProvingValidationResult, error) {
	source, err := loadProvingObjectiveForProject(projectDir, objectivePath)
	if err != nil {
		return ProvingValidationResult{}, err
	}

	return RunProvingValidationForObjective(ctx, source.ProjectDir, source.Objective)
}

// RunProvingValidationForObjective executes the corresponding package operation.
func RunProvingValidationForObjective(
	ctx context.Context,
	projectDir string,
	objective ProvingObjective,
) (ProvingValidationResult, error) {
	planned, err := planProvingRuntime(projectDir, objective, nil)
	if err != nil {
		return ProvingValidationResult{}, err
	}

	artifact, err := ensureProvingOpsArtifact(projectDir, planned.Objective)
	if err != nil {
		return ProvingValidationResult{}, err
	}

	err = os.MkdirAll(filepath.Dir(planned.Plan.ValidationPath), 0o755)
	if err != nil {
		return ProvingValidationResult{}, fmt.Errorf("create proving validation directory: %w", err)
	}

	commands := ProvingValidationCommands()
	result := ProvingValidationResult{
		Objective: planned.Objective,
		Passed:    true,
		LogPath:   planned.Plan.ValidationPath,
	}

	for _, command := range commands {
		commandResult := runProvingValidationCommand(ctx, planned.Plan.ProjectDir, command)
		result.Commands = append(result.Commands, commandResult)

		if commandResult.ExitCode != 0 {
			result.Passed = false
		}
	}

	err = os.WriteFile(result.LogPath, []byte(RenderProvingValidationResult(result)), 0o644)
	if err != nil {
		return ProvingValidationResult{}, fmt.Errorf("write proving validation log: %w", err)
	}

	if err := RefreshOpsArtifactMetadata(artifact.Dir, provingValidationArtifactStatus(result), time.Now()); err != nil {
		return ProvingValidationResult{}, err
	}

	return result, nil
}

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

// ProvingValidationCommands executes the corresponding package operation.
func ProvingValidationCommands() []string {
	return []string{
		"make compile",
		"make lint",
		"make test",
		"make verify",
	}
}

func prepareProvingArtifacts(result ProvingSandboxRunResult) (*os.File, error) {
	artifact, err := ensureProvingOpsArtifact(result.Plan.ProjectDir, result.Objective)
	if err != nil {
		return nil, err
	}

	for _, path := range []string{result.Plan.WorkerContextPath, result.Plan.ResultPath, result.Plan.LogPath, result.Plan.ValidationPath} {
		err = os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			return nil, fmt.Errorf("create proving artifact directory: %w", err)
		}
	}

	err = os.WriteFile(result.Plan.WorkerContextPath, []byte(RenderProvingWorkerContext(result)), 0o644)
	if err != nil {
		return nil, fmt.Errorf("write proving worker context: %w", err)
	}

	logFile, err := os.OpenFile(result.Plan.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open proving runtime log: %w", err)
	}

	if err := RefreshOpsArtifactMetadata(artifact.Dir, "running", time.Now()); err != nil {
		if closeErr := logFile.Close(); closeErr != nil {
			return nil, fmt.Errorf("%w; close proving runtime log: %w", err, closeErr)
		}

		return nil, err
	}

	return logFile, nil
}

func runProvingSandboxCommands(
	ctx context.Context,
	result ProvingSandboxRunResult,
	opts ProvingSandboxRunOptions,
	logFile io.Writer,
) error {
	executor := opts.Executor
	if executor == nil {
		executor = OSHarnessExecutor{}
	}

	stdout := provingOutput(opts.Stdout, logFile)
	stderr := provingOutput(opts.Stderr, logFile)

	err := executor.Run(ctx, result.Build, stdout, stderr)
	if err != nil {
		return err
	}

	return executor.Run(ctx, result.Run, stdout, stderr)
}

func provingOutput(configured, fallback io.Writer) io.Writer {
	if configured != nil {
		return configured
	}

	return fallback
}

func runProvingValidationCommand(ctx context.Context, projectDir, command string) ProvingValidationCommandResult {
	var stdout bytes.Buffer

	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "/bin/sh", "-lc", command)
	cmd.Dir = projectDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return ProvingValidationCommandResult{
		Command:  command,
		ExitCode: commandExitCode(err),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return -1
}

// RenderProvingValidationResult renders the corresponding package data.
func RenderProvingValidationResult(result ProvingValidationResult) string {
	var output strings.Builder

	output.WriteString("# Proving Validation Result\n\n")
	output.WriteString("objective_id: ")
	output.WriteString(result.Objective.ID)
	output.WriteString("\n")
	output.WriteString("ticket_key: ")
	output.WriteString(result.Objective.TicketKey)
	output.WriteString("\n")
	output.WriteString("passed: ")
	output.WriteString(strconv.FormatBool(result.Passed))
	output.WriteString("\n\n")

	for _, command := range result.Commands {
		output.WriteString("## Command: ")
		output.WriteString(command.Command)
		output.WriteString("\n\n")
		output.WriteString("exit_code: ")
		output.WriteString(strconv.Itoa(command.ExitCode))
		output.WriteString("\n\n")
		output.WriteString("### Stdout\n\n")
		output.WriteString(command.Stdout)

		if !strings.HasSuffix(command.Stdout, "\n") {
			output.WriteString("\n")
		}

		output.WriteString("\n### Stderr\n\n")
		output.WriteString(command.Stderr)

		if !strings.HasSuffix(command.Stderr, "\n") {
			output.WriteString("\n")
		}

		output.WriteString("\n")
	}

	return output.String()
}

// RenderProvingValidationFeedback renders the corresponding package data.
func RenderProvingValidationFeedback(result ProvingValidationResult) string {
	var output strings.Builder

	output.WriteString("The previous attempt did not pass validation.\n\n")
	output.WriteString("Validation artifact: ")
	output.WriteString(result.LogPath)
	output.WriteString("\n\n")

	for _, command := range result.Commands {
		if command.ExitCode == 0 {
			continue
		}

		output.WriteString("## Failing Command: ")
		output.WriteString(command.Command)
		output.WriteString("\n\n")
		output.WriteString("exit_code: ")
		output.WriteString(strconv.Itoa(command.ExitCode))
		output.WriteString("\n\n")

		if strings.TrimSpace(command.Stdout) != "" {
			output.WriteString("### Stdout\n\n")
			output.WriteString(strings.TrimSpace(command.Stdout))
			output.WriteString("\n\n")
		}

		if strings.TrimSpace(command.Stderr) != "" {
			output.WriteString("### Stderr\n\n")
			output.WriteString(strings.TrimSpace(command.Stderr))
			output.WriteString("\n\n")
		}
	}

	output.WriteString("Fix the implementation and any missing tests. Do not weaken validation.\n")

	return output.String()
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

// ProvingValidationFailureSignature executes the corresponding package operation.
func ProvingValidationFailureSignature(result ProvingValidationResult) string {
	hash := sha256.New()

	for _, command := range result.Commands {
		if command.ExitCode == 0 {
			continue
		}

		hash.Write([]byte(command.Command))
		hash.Write([]byte{0})
		hash.Write([]byte(strconv.Itoa(command.ExitCode)))
		hash.Write([]byte{0})
		hash.Write([]byte(command.Stdout))
		hash.Write([]byte{0})
		hash.Write([]byte(command.Stderr))
		hash.Write([]byte{0})
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// PlanProvingSandboxRun executes the corresponding package operation.
func PlanProvingSandboxRun(projectDir, objectivePath string, worker harness.Worker) (ProvingSandboxRunResult, error) {
	source, err := loadProvingObjectiveForProject(projectDir, objectivePath)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return PlanProvingSandboxRunForObjective(source.ProjectDir, source.Objective, worker)
}

// PlanProvingSandboxRunForObjective executes the corresponding package operation.
func PlanProvingSandboxRunForObjective(projectDir string, objective ProvingObjective, worker harness.Worker) (ProvingSandboxRunResult, error) {
	if worker == nil {
		return ProvingSandboxRunResult{}, fmt.Errorf("worker adapter is required")
	}

	result, err := planProvingRuntime(projectDir, objective, worker.RuntimeMounts())
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	invocation, err := NewHarnessInvocation(HarnessInvocationOptions{
		ExecutionID: objective.ID, TicketKey: objective.TicketKey,
		TicketPath: result.TicketPath, Plan: result.Plan, Worker: worker,
	})
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	result.Build, result.Run = invocation.Build, invocation.Run

	return result, nil
}

func planProvingRuntime(projectDir string, objective ProvingObjective, mounts []harness.RuntimeMount) (ProvingSandboxRunResult, error) {
	config, err := LoadWorkspaceConfig(projectDir)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	plan, err := PlanRuntime(RuntimePlanOptions{
		ExtraMounts: mounts,
		ProjectDir:  projectDir,
		Config:      config.Runtime,
		TicketKey:   objective.TicketKey,
		RunID:       objective.ID,
	})
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	plan.RunArtifactDir = resolveProjectPath(projectDir, objective.ArtifactRoot)
	plan.WorkerContextPath = resolveProjectPath(projectDir, objective.ContextPath)
	plan.ResultPath = resolveProjectPath(projectDir, objective.ResultPath)
	plan.LogPath = resolveProjectPath(projectDir, objective.LogPath)
	plan.ValidationPath = resolveProjectPath(projectDir, objective.ValidationPath)

	ticketPath, err := findProvingTicketPath(projectDir, objective.TicketKey)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return ProvingSandboxRunResult{Objective: objective, Plan: plan, TicketPath: ticketPath}, nil
}

type provingObjectiveSource struct {
	ProjectDir string
	Objective  ProvingObjective
}

func loadProvingObjectiveForProject(projectDir, objectivePath string) (provingObjectiveSource, error) {
	projectDir, err := resolveProjectDir(projectDir)
	if err != nil {
		return provingObjectiveSource{}, err
	}

	objective, err := LoadProvingObjective(resolveProjectPath(projectDir, objectivePath))
	if err != nil {
		return provingObjectiveSource{}, err
	}

	return provingObjectiveSource{ProjectDir: projectDir, Objective: objective}, nil
}

func ensureProvingOpsArtifact(projectDir string, objective ProvingObjective) (OpsArtifact, error) {
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return OpsArtifact{}, err
	}

	id := filepath.Base(filepath.Clean(filepath.FromSlash(objective.ArtifactRoot)))

	return EnsureOpsArtifactEnvelope(CreateOpsArtifactOptions{
		Layout:      layout,
		Type:        "run",
		Title:       "Proving " + objective.ID,
		Producer:    "chr",
		ExternalIDs: []string{objective.ID},
	}, id)
}

func provingValidationArtifactStatus(result ProvingValidationResult) string {
	if result.Passed {
		return "passed"
	}

	return "failed"
}

func provingAttemptObjective(
	objective ProvingObjective,
	attemptNumber int,
	previousValidation *ProvingValidationResult,
) ProvingObjective {
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

	if err := RefreshOpsArtifactMetadata(artifact.Dir, result.Status, time.Now()); err != nil {
		return err
	}

	return nil
}

// RenderProvingWorkerContext renders the corresponding package data.
func RenderProvingWorkerContext(result ProvingSandboxRunResult) string {
	var output strings.Builder

	output.WriteString("# Proving Worker Context\n\n")
	output.WriteString("## Objective Identity\n\n")
	output.WriteString("- objective_id: ")
	output.WriteString(result.Objective.ID)
	output.WriteString("\n")
	output.WriteString("- ticket_key: ")
	output.WriteString(result.Objective.TicketKey)
	output.WriteString("\n")
	output.WriteString("- kind: ")
	output.WriteString(result.Objective.Kind)
	output.WriteString("\n\n")
	output.WriteString("## Artifact Paths\n\n")
	output.WriteString("- context_path: ")
	output.WriteString(containerProjectPath(result.Plan, result.Plan.WorkerContextPath))
	output.WriteString("\n")
	output.WriteString("- result_path: ")
	output.WriteString(containerProjectPath(result.Plan, result.Plan.ResultPath))
	output.WriteString("\n")
	output.WriteString("- log_path: ")
	output.WriteString(containerProjectPath(result.Plan, result.Plan.LogPath))
	output.WriteString("\n")
	output.WriteString("- validation_path: ")
	output.WriteString(containerProjectPath(result.Plan, result.Plan.ValidationPath))
	output.WriteString("\n\n")
	output.WriteString("## Objective\n\n")
	output.WriteString(result.Objective.Body)
	output.WriteString("\n\n")
	output.WriteString("## Output Contract\n\n")
	output.WriteString("Write the result JSON to the result_path above.\n")
	output.WriteString("Do not remove or weaken tests to make validation pass.\n")

	return output.String()
}

func parseSimpleFrontmatter(content string) map[string]string {
	fields := map[string]string{}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return fields
}

func validateProvingObjective(objective ProvingObjective) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "objective_id", value: objective.ID},
		{name: "ticket_key", value: objective.TicketKey},
		{name: "kind", value: objective.Kind},
		{name: "artifact_root", value: objective.ArtifactRoot},
		{name: "context_path", value: objective.ContextPath},
		{name: "result_path", value: objective.ResultPath},
		{name: "log_path", value: objective.LogPath},
		{name: "validation_path", value: objective.ValidationPath},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}

	if strings.TrimSpace(objective.Body) == "" {
		return fmt.Errorf("objective body is required")
	}

	return nil
}

func findProvingTicketPath(projectDir, ticketKey string) (string, error) {
	ticketsDir := UserPath(projectDir, "tickets")

	entries, err := os.ReadDir(ticketsDir)
	if err != nil {
		return "", fmt.Errorf("read proving tickets directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasPrefix(entry.Name(), ticketKey+"-") && strings.HasSuffix(entry.Name(), ".md") {
			return filepath.Join(ticketsDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("ticket %s not found", ticketKey)
}
