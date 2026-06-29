package core

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestExecutionRunnerRunsHarnessAndMarksSucceeded(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())

	err := repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{resultPath: record.Artifacts.ResultPath}
	runner := NewExecutionRunner(repository, executor)
	runner.Now = fixedRunnerClock()

	got, err := runner.Run(t.Context(), RunExecutionOptions{
		Record:     record,
		Invocation: testRunnerInvocation(record),
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	assertSucceededRunnerRecord(t, got)

	if !slices.Equal(executor.commands, []string{"docker build", "docker run", "docker run"}) {
		t.Fatalf("commands = %#v, want build, run, and validation", executor.commands)
	}

	content, err := os.ReadFile(record.Artifacts.LogPath)
	if err != nil {
		t.Fatalf("read log artifact: %v", err)
	}

	if !strings.Contains(string(content), "stdout: docker build") {
		t.Fatalf("log content = %q, want build stdout", string(content))
	}

	validationContent, err := os.ReadFile(record.Artifacts.ValidationPath)
	if err != nil {
		t.Fatalf("read validation artifact: %v", err)
	}

	if !strings.Contains(string(validationContent), "stdout: docker run") {
		t.Fatalf("validation content = %q, want validation stdout", string(validationContent))
	}

	stored, err := repository.GetExecution(t.Context(), record.ID)
	if err != nil {
		t.Fatalf("GetExecution returned error: %v", err)
	}

	if stored.Status != ExecutionStatusSucceeded {
		t.Fatalf("stored Status = %q, want succeeded", stored.Status)
	}
}

func assertSucceededRunnerRecord(t *testing.T, record ExecutionRecord) {
	t.Helper()

	if record.Status != ExecutionStatusSucceeded {
		t.Fatalf("Status = %q, want succeeded", record.Status)
	}

	if record.StartedAt == nil {
		t.Fatal("StartedAt is nil, want timestamp")
	}

	if record.FinishedAt == nil {
		t.Fatal("FinishedAt is nil, want timestamp")
	}
}

func TestExecutionRunnerMarksFailedWhenBuildHarnessFails(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())

	err := repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{failAt: "build"}
	runner := NewExecutionRunner(repository, executor)
	runner.Now = fixedRunnerClock()

	got, err := runner.Run(t.Context(), RunExecutionOptions{
		Record:     record,
		Invocation: testRunnerInvocation(record),
	})
	if err == nil {
		t.Fatal("Run returned nil error")
	}

	if got.Status != ExecutionStatusFailed {
		t.Fatalf("Status = %q, want failed", got.Status)
	}

	if got.FailureReason == "" {
		t.Fatal("FailureReason is empty, want error")
	}

	if !strings.HasPrefix(got.FailureReason, "harness:") {
		t.Fatalf("FailureReason = %q, want harness prefix", got.FailureReason)
	}

	stored, err := repository.GetExecution(t.Context(), record.ID)
	if err != nil {
		t.Fatalf("GetExecution returned error: %v", err)
	}

	if stored.Status != ExecutionStatusFailed {
		t.Fatalf("stored Status = %q, want failed", stored.Status)
	}
}

func TestExecutionRunnerClassifiesWorkerCommandFailure(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())

	err := repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{failWorker: true}
	runner := NewExecutionRunner(repository, executor)
	runner.Now = fixedRunnerClock()

	got, err := runner.Run(t.Context(), RunExecutionOptions{
		Record:     record,
		Invocation: testRunnerInvocation(record),
	})
	if err == nil {
		t.Fatal("Run returned nil error")
	}

	if got.Status != ExecutionStatusFailed {
		t.Fatalf("Status = %q, want failed", got.Status)
	}

	if !strings.HasPrefix(got.FailureReason, "worker:") {
		t.Fatalf("FailureReason = %q, want worker prefix", got.FailureReason)
	}

	if !slices.Equal(executor.commands, []string{"docker build", "docker run"}) {
		t.Fatalf("commands = %#v, want build and worker run only", executor.commands)
	}
}

func TestExecutionRunnerMarksFailedWhenResultIsMissing(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())

	err := repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{}
	runner := NewExecutionRunner(repository, executor)
	runner.Now = fixedRunnerClock()

	got, err := runner.Run(t.Context(), RunExecutionOptions{
		Record:     record,
		Invocation: testRunnerInvocation(record),
	})
	if err == nil {
		t.Fatal("Run returned nil error")
	}

	if got.Status != ExecutionStatusFailed {
		t.Fatalf("Status = %q, want failed", got.Status)
	}

	if !strings.HasPrefix(got.FailureReason, "result_capture:") {
		t.Fatalf("FailureReason = %q, want result_capture prefix", got.FailureReason)
	}
}

func TestExecutionRunnerMarksFailedWhenResultIsMalformed(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())

	err := repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{
		resultPath:    record.Artifacts.ResultPath,
		resultContent: `{`,
	}
	runner := NewExecutionRunner(repository, executor)
	runner.Now = fixedRunnerClock()

	got, err := runner.Run(t.Context(), RunExecutionOptions{
		Record:     record,
		Invocation: testRunnerInvocation(record),
	})
	if err == nil {
		t.Fatal("Run returned nil error")
	}

	if got.Status != ExecutionStatusFailed {
		t.Fatalf("Status = %q, want failed", got.Status)
	}

	if !strings.HasPrefix(got.FailureReason, "result_capture:") {
		t.Fatalf("FailureReason = %q, want result_capture prefix", got.FailureReason)
	}
}

func TestExecutionRunnerClassifiesFailedResultAsValidationFailure(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())

	err := repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{
		resultPath:    record.Artifacts.ResultPath,
		resultContent: `{"execution_id":"exec-1","ticket_key":"CHR-1","status":"failed","summary":"tests failed","evidence_summary":{"strategy":"planned","produced":"none","validation":"failed"}}`,
	}
	runner := NewExecutionRunner(repository, executor)
	runner.Now = fixedRunnerClock()

	got, err := runner.Run(t.Context(), RunExecutionOptions{
		Record:     record,
		Invocation: testRunnerInvocation(record),
	})
	if err == nil {
		t.Fatal("Run returned nil error")
	}

	if got.Status != ExecutionStatusFailed {
		t.Fatalf("Status = %q, want failed", got.Status)
	}

	if !strings.HasPrefix(got.FailureReason, "validation:") {
		t.Fatalf("FailureReason = %q, want validation prefix", got.FailureReason)
	}
}

func TestExecutionRunnerClassifiesValidationCommandFailure(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())

	err := repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{
		resultPath:     record.Artifacts.ResultPath,
		failValidation: true,
	}
	runner := NewExecutionRunner(repository, executor)
	runner.Now = fixedRunnerClock()

	got, err := runner.Run(t.Context(), RunExecutionOptions{
		Record:     record,
		Invocation: testRunnerInvocation(record),
	})
	if err == nil {
		t.Fatal("Run returned nil error")
	}

	if got.Status != ExecutionStatusFailed {
		t.Fatalf("Status = %q, want failed", got.Status)
	}

	if !strings.HasPrefix(got.FailureReason, "validation:") {
		t.Fatalf("FailureReason = %q, want validation prefix", got.FailureReason)
	}

	content, err := os.ReadFile(record.Artifacts.ValidationPath)
	if err != nil {
		t.Fatalf("read validation artifact: %v", err)
	}

	if !strings.Contains(string(content), "stdout: docker run") {
		t.Fatalf("validation content = %q, want validation stdout", string(content))
	}
}

func TestExecutionRunnerDryRunDoesNotPersistOrExecute(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())

	err := repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{}
	runner := NewExecutionRunner(repository, executor)

	got, err := runner.Run(t.Context(), RunExecutionOptions{
		Record:     record,
		Invocation: testRunnerInvocation(record),
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got.Status != ExecutionStatusPending {
		t.Fatalf("Status = %q, want pending", got.Status)
	}

	if len(executor.commands) != 0 {
		t.Fatalf("commands = %#v, want none", executor.commands)
	}

	stored, err := repository.GetExecution(t.Context(), record.ID)
	if err != nil {
		t.Fatalf("GetExecution returned error: %v", err)
	}

	if stored.Status != ExecutionStatusPending {
		t.Fatalf("stored Status = %q, want pending", stored.Status)
	}
}

type fakeHarnessExecutor struct {
	commands       []string
	failAt         string
	failWorker     bool
	failValidation bool
	resultPath     string
	resultContent  string
}

func (e *fakeHarnessExecutor) Run(_ context.Context, command HarnessCommand, stdout, stderr io.Writer) error {
	name := command.Program + " " + command.Args[0]
	e.commands = append(e.commands, name)

	_, err := fmt.Fprintln(stdout, "stdout: "+name)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stderr, "stderr: "+name)
	if err != nil {
		return err
	}

	if e.failAt == command.Args[0] {
		return fmt.Errorf("failed %s", name)
	}

	if e.failWorker && isWorkerHarnessCommand(command) {
		return NewExecutionFailure(ExecutionFailureHarness, fmt.Errorf("failed %s", name))
	}

	if e.failValidation && isValidationHarnessCommand(command) {
		return NewExecutionFailure(ExecutionFailureHarness, fmt.Errorf("failed %s", name))
	}

	err = e.writeResultForWorkerRun(command)
	if err != nil {
		return err
	}

	return nil
}

func (e *fakeHarnessExecutor) writeResultForWorkerRun(command HarnessCommand) error {
	if command.Args[0] != "run" || isValidationHarnessCommand(command) || e.resultPath == "" {
		return nil
	}

	content := e.resultContent
	if content == "" {
		content = `{"execution_id":"exec-1","ticket_key":"CHR-1","status":"succeeded","summary":"done","evidence_summary":{"strategy":"planned","produced":"evidence produced","validation":"validation passed"}}`
	}

	return os.WriteFile(e.resultPath, []byte(content), 0o644)
}

func isWorkerHarnessCommand(command HarnessCommand) bool {
	return command.Args[0] == "run" && !isValidationHarnessCommand(command)
}

func isValidationHarnessCommand(command HarnessCommand) bool {
	return slices.Contains(command.Args, "chr-exec-1-validation")
}

func newTestExecutionRunnerRepository(t *testing.T) ExecutionRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	repository, err := NewExecutionRepositoryFromDB(t.Context(), db, testMigrationFS())
	if err != nil {
		t.Fatalf("NewExecutionRepositoryFromDB returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close execution repository: %v", err)
		}
	})

	return repository
}

func testRunnerExecutionRecord(dir string) ExecutionRecord {
	runDir := filepath.Join(dir, ".chr", "user", "runtime", "runs", "CHR-1", "exec-1")
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	return ExecutionRecord{
		ID:         "exec-1",
		ProjectID:  "project-1",
		TicketKey:  "CHR-1",
		TicketPath: filepath.Join(dir, ".chr", "user", "tickets", "CHR-1.md"),
		Status:     ExecutionStatusPending,
		Runtime: ExecutionRuntimeRef{
			Image:             "chr-agent-runtime:local",
			DockerfilePath:    filepath.Join(dir, ".chr", "user", "runtime", "agent", "Dockerfile"),
			ContainerWorkdir:  "/workspace",
			ValidationCommand: "make verify",
		},
		Artifacts: ExecutionArtifacts{
			RunDir:         runDir,
			ResultPath:     filepath.Join(runDir, "result.json"),
			LogPath:        filepath.Join(runDir, "runtime.log"),
			ValidationPath: filepath.Join(runDir, "validation.log"),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func testRunnerInvocation(record ExecutionRecord) HarnessInvocation {
	return HarnessInvocation{
		ExecutionID: record.ID,
		TicketKey:   record.TicketKey,
		Build: HarnessCommand{
			Program: "docker",
			Args:    []string{"build"},
		},
		Run: HarnessCommand{
			Program: "docker",
			Args:    []string{"run"},
		},
		Validation: HarnessCommand{
			Program: "docker",
			Args:    []string{"run", "--name", "chr-exec-1-validation"},
		},
		ResultPath: record.Artifacts.ResultPath,
		LogPath:    record.Artifacts.LogPath,
	}
}

func fixedRunnerClock() func() time.Time {
	ticks := []time.Time{
		time.Date(2026, 5, 19, 12, 1, 0, 0, time.UTC),
		time.Date(2026, 5, 19, 12, 2, 0, 0, time.UTC),
	}

	return func() time.Time {
		tick := ticks[0]
		if len(ticks) > 1 {
			ticks = ticks[1:]
		}

		return tick
	}
}
