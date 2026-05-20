package chc

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
)

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

type ProvingSandboxRunOptions struct {
	ProjectDir    string
	ObjectivePath string
	DryRun        bool
	Executor      HarnessExecutor
	Stdout        io.Writer
	Stderr        io.Writer
}

type ProvingSandboxRunResult struct {
	Objective  ProvingObjective
	Plan       RuntimePlan
	Build      HarnessCommand
	Run        HarnessCommand
	TicketPath string
	DryRun     bool
}

type ProvingValidationCommandResult struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
}

type ProvingValidationResult struct {
	Objective ProvingObjective
	Commands  []ProvingValidationCommandResult
	Passed    bool
	LogPath   string
}

type ProvingValidationRunner func(context.Context, string, ProvingObjective) (ProvingValidationResult, error)

type ProvingRalphLoopOptions struct {
	ProjectDir       string
	ObjectivePath    string
	MaxAttempts      int
	Executor         HarnessExecutor
	ValidationRunner ProvingValidationRunner
	Stdout           io.Writer
	Stderr           io.Writer
}

type ProvingRalphLoopResult struct {
	Objective   ProvingObjective
	Attempts    []ProvingRalphLoopAttempt
	Status      string
	StopReason  string
	SummaryPath string
}

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
	ProvingRalphLoopStatusPassed  = "passed"
	ProvingRalphLoopStatusFailed  = "failed"
	ProvingRalphLoopStatusBlocked = "blocked"

	ProvingRalphLoopStopValidationPassed = "validation_passed"
	ProvingRalphLoopStopRepeatedFailure  = "repeated_failure"
	ProvingRalphLoopStopMaxAttempts      = "max_attempts"
	ProvingRalphLoopStopAgentFailed      = "agent_failed"
	ProvingRalphLoopStopValidationFailed = "validation_failed"
)

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

	if err := validateProvingObjective(objective); err != nil {
		return ProvingObjective{}, err
	}

	return objective, nil
}

func RunProvingSandbox(ctx context.Context, opts ProvingSandboxRunOptions) (ProvingSandboxRunResult, error) {
	projectDir, objective, err := loadProvingObjectiveForProject(opts.ProjectDir, opts.ObjectivePath)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return RunProvingSandboxForObjective(ctx, opts, projectDir, objective)
}

func RunProvingSandboxForObjective(
	ctx context.Context,
	opts ProvingSandboxRunOptions,
	projectDir string,
	objective ProvingObjective,
) (ProvingSandboxRunResult, error) {
	result, err := PlanProvingSandboxRunForObjective(projectDir, objective)
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

	if err := runProvingSandboxCommands(ctx, result, opts, logFile); err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return result, nil
}

func RunProvingValidation(ctx context.Context, projectDir, objectivePath string) (ProvingValidationResult, error) {
	projectDir, objective, err := loadProvingObjectiveForProject(projectDir, objectivePath)
	if err != nil {
		return ProvingValidationResult{}, err
	}

	return RunProvingValidationForObjective(ctx, projectDir, objective)
}

func RunProvingValidationForObjective(
	ctx context.Context,
	projectDir string,
	objective ProvingObjective,
) (ProvingValidationResult, error) {
	planned, err := PlanProvingSandboxRunForObjective(projectDir, objective)
	if err != nil {
		return ProvingValidationResult{}, err
	}

	if err := os.MkdirAll(filepath.Dir(planned.Plan.ValidationPath), 0o755); err != nil {
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

	if err := os.WriteFile(result.LogPath, []byte(RenderProvingValidationResult(result)), 0o644); err != nil {
		return ProvingValidationResult{}, fmt.Errorf("write proving validation log: %w", err)
	}

	return result, nil
}

func RunProvingRalphLoop(ctx context.Context, opts ProvingRalphLoopOptions) (ProvingRalphLoopResult, error) {
	projectDir, objective, err := loadProvingObjectiveForProject(opts.ProjectDir, opts.ObjectivePath)
	if err != nil {
		return ProvingRalphLoopResult{}, err
	}

	maxAttempts, err := provingRalphLoopMaxAttempts(opts, objective)
	if err != nil {
		return ProvingRalphLoopResult{}, err
	}

	validationRunner := provingRalphLoopValidationRunner(opts)
	result := newProvingRalphLoopResult(projectDir, objective)
	seenFailures := map[string]bool{}

	var previousValidation *ProvingValidationResult

	for attemptNumber := 1; attemptNumber <= maxAttempts; attemptNumber++ {
		attemptObjective := provingAttemptObjective(objective, attemptNumber, previousValidation)
		attempt, validation, status, stopReason := runProvingRalphLoopAttempt(
			ctx,
			opts,
			projectDir,
			attemptNumber,
			attemptObjective,
			validationRunner,
		)
		result.Attempts = append(result.Attempts, attempt)

		if stopReason != "" {
			result.Status = status
			result.StopReason = stopReason

			break
		}

		if seenFailures[attempt.FailureSignature] {
			result.Status = ProvingRalphLoopStatusFailed
			result.StopReason = ProvingRalphLoopStopRepeatedFailure

			break
		}

		seenFailures[attempt.FailureSignature] = true
		previousValidation = validation
	}

	if err := writeProvingRalphLoopSummary(result); err != nil {
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
	summaryPath := filepath.ToSlash(filepath.Join(objective.ArtifactRoot, "ralph-loop.md"))

	return ProvingRalphLoopResult{
		Objective:   objective,
		Status:      ProvingRalphLoopStatusFailed,
		StopReason:  ProvingRalphLoopStopMaxAttempts,
		SummaryPath: resolveProjectPath(projectDir, summaryPath),
	}
}

func runProvingRalphLoopAttempt(
	ctx context.Context,
	opts ProvingRalphLoopOptions,
	projectDir string,
	attemptNumber int,
	attemptObjective ProvingObjective,
	validationRunner ProvingValidationRunner,
) (ProvingRalphLoopAttempt, *ProvingValidationResult, string, string) {
	attempt := provingRalphLoopAttempt(attemptNumber, attemptObjective)

	sandbox, sandboxErr := RunProvingSandboxForObjective(ctx, ProvingSandboxRunOptions{
		ProjectDir: projectDir,
		DryRun:     false,
		Executor:   opts.Executor,
		Stdout:     opts.Stdout,
		Stderr:     opts.Stderr,
	}, projectDir, attemptObjective)
	if sandboxErr != nil {
		attempt.Error = sandboxErr.Error()

		return attempt, nil, ProvingRalphLoopStatusBlocked, ProvingRalphLoopStopAgentFailed
	}

	validation, validationErr := validationRunner(ctx, projectDir, attemptObjective)
	if validationErr != nil {
		attempt.Error = validationErr.Error()

		return attempt, nil, ProvingRalphLoopStatusBlocked, ProvingRalphLoopStopValidationFailed
	}

	attempt.ContextPath = sandbox.Plan.WorkerContextPath
	attempt.ResultPath = sandbox.Plan.ResultPath
	attempt.LogPath = sandbox.Plan.LogPath
	attempt.ValidationPath = validation.LogPath
	attempt.ValidationPassed = validation.Passed
	attempt.FailureSignature = ProvingValidationFailureSignature(validation)

	if validation.Passed {
		return attempt, &validation, ProvingRalphLoopStatusPassed, ProvingRalphLoopStopValidationPassed
	}

	return attempt, &validation, "", ""
}

func ProvingValidationCommands() []string {
	return []string{
		"make compile",
		"make lint",
		"make test",
		"make verify",
	}
}

func prepareProvingArtifacts(result ProvingSandboxRunResult) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(result.Plan.WorkerContextPath), 0o755); err != nil {
		return nil, fmt.Errorf("create proving artifact directory: %w", err)
	}

	if err := os.WriteFile(result.Plan.WorkerContextPath, []byte(RenderProvingWorkerContext(result)), 0o644); err != nil {
		return nil, fmt.Errorf("write proving worker context: %w", err)
	}

	logFile, err := os.OpenFile(result.Plan.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open proving runtime log: %w", err)
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

	if err := executor.Run(ctx, result.Build, stdout, stderr); err != nil {
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

func PlanProvingSandboxRun(projectDir, objectivePath string) (ProvingSandboxRunResult, error) {
	projectDir, objective, err := loadProvingObjectiveForProject(projectDir, objectivePath)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return PlanProvingSandboxRunForObjective(projectDir, objective)
}

func PlanProvingSandboxRunForObjective(projectDir string, objective ProvingObjective) (ProvingSandboxRunResult, error) {
	config, err := LoadWorkspaceConfig(projectDir)
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	plan, err := PlanRuntime(RuntimePlanOptions{
		ProjectDir: projectDir,
		Config:     config.Runtime,
		TicketKey:  objective.TicketKey,
		RunID:      objective.ID,
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

	invocation, err := NewHarnessInvocation(HarnessInvocationOptions{
		ExecutionID: objective.ID,
		TicketKey:   objective.TicketKey,
		TicketPath:  ticketPath,
		Plan:        plan,
	})
	if err != nil {
		return ProvingSandboxRunResult{}, err
	}

	return ProvingSandboxRunResult{
		Objective:  objective,
		Plan:       plan,
		Build:      invocation.Build,
		Run:        invocation.Run,
		TicketPath: ticketPath,
	}, nil
}

func loadProvingObjectiveForProject(projectDir, objectivePath string) (string, ProvingObjective, error) {
	projectDir, err := resolveProjectDir(projectDir)
	if err != nil {
		return "", ProvingObjective{}, err
	}

	objective, err := LoadProvingObjective(resolveProjectPath(projectDir, objectivePath))
	if err != nil {
		return "", ProvingObjective{}, err
	}

	return projectDir, objective, nil
}

func provingAttemptObjective(
	objective ProvingObjective,
	attemptNumber int,
	previousValidation *ProvingValidationResult,
) ProvingObjective {
	attempt := objective
	attempt.ID = fmt.Sprintf("%s-attempt-%d", objective.ID, attemptNumber)
	attempt.ArtifactRoot = filepath.ToSlash(filepath.Join(objective.ArtifactRoot, fmt.Sprintf("attempt-%d", attemptNumber)))
	attempt.ContextPath = filepath.ToSlash(filepath.Join(attempt.ArtifactRoot, "context.md"))
	attempt.ResultPath = filepath.ToSlash(filepath.Join(attempt.ArtifactRoot, "result.json"))
	attempt.LogPath = filepath.ToSlash(filepath.Join(attempt.ArtifactRoot, "agent.log"))
	attempt.ValidationPath = filepath.ToSlash(filepath.Join(attempt.ArtifactRoot, "validation.log"))

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

func writeProvingRalphLoopSummary(result ProvingRalphLoopResult) error {
	if err := os.MkdirAll(filepath.Dir(result.SummaryPath), 0o755); err != nil {
		return fmt.Errorf("create proving ralph loop summary directory: %w", err)
	}

	if err := os.WriteFile(result.SummaryPath, []byte(RenderProvingRalphLoopResult(result)), 0o644); err != nil {
		return fmt.Errorf("write proving ralph loop summary: %w", err)
	}

	return nil
}

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
	ticketsDir := filepath.Join(projectDir, DirectoryName, "tickets")

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
