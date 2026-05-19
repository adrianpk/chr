package chc

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
	if err := repository.CreateExecution(t.Context(), record); err != nil {
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

	if !slices.Equal(executor.commands, []string{"docker build", "docker run"}) {
		t.Fatalf("commands = %#v, want build and run", executor.commands)
	}

	content, err := os.ReadFile(record.Artifacts.LogPath)
	if err != nil {
		t.Fatalf("read log artifact: %v", err)
	}

	if !strings.Contains(string(content), "stdout: docker build") {
		t.Fatalf("log content = %q, want build stdout", string(content))
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

func TestExecutionRunnerMarksFailedWhenHarnessFails(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())
	if err := repository.CreateExecution(t.Context(), record); err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{failAt: "run"}
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

func TestExecutionRunnerMarksFailedWhenResultIsMissing(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())
	if err := repository.CreateExecution(t.Context(), record); err != nil {
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
	if err := repository.CreateExecution(t.Context(), record); err != nil {
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
	if err := repository.CreateExecution(t.Context(), record); err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	executor := &fakeHarnessExecutor{
		resultPath:    record.Artifacts.ResultPath,
		resultContent: `{"execution_id":"exec-1","ticket_key":"CHC-1","status":"failed","summary":"tests failed"}`,
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

func TestExecutionRunnerDryRunDoesNotPersistOrExecute(t *testing.T) {
	repository := newTestExecutionRunnerRepository(t)

	record := testRunnerExecutionRecord(t.TempDir())
	if err := repository.CreateExecution(t.Context(), record); err != nil {
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
	commands      []string
	failAt        string
	resultPath    string
	resultContent string
}

func (e *fakeHarnessExecutor) Run(_ context.Context, command HarnessCommand, stdout, stderr io.Writer) error {
	name := command.Program + " " + command.Args[0]
	e.commands = append(e.commands, name)

	if _, err := fmt.Fprintln(stdout, "stdout: "+name); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(stderr, "stderr: "+name); err != nil {
		return err
	}

	if e.failAt == command.Args[0] {
		return fmt.Errorf("failed %s", name)
	}

	if command.Args[0] == "run" && e.resultPath != "" {
		content := e.resultContent
		if content == "" {
			content = `{"execution_id":"exec-1","ticket_key":"CHC-1","status":"succeeded","summary":"done"}`
		}

		if err := os.WriteFile(e.resultPath, []byte(content), 0o644); err != nil {
			return err
		}
	}

	return nil
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
		if err := repository.Close(); err != nil {
			t.Fatalf("close execution repository: %v", err)
		}
	})

	return repository
}

func testRunnerExecutionRecord(dir string) ExecutionRecord {
	runDir := filepath.Join(dir, ".chc", "runtime", "runs", "CHC-1", "exec-1")
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	return ExecutionRecord{
		ID:         "exec-1",
		ProjectID:  "project-1",
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(dir, ".chc", "tickets", "CHC-1.md"),
		Status:     ExecutionStatusPending,
		Runtime: ExecutionRuntimeRef{
			Image:             "chc-agent-runtime:local",
			DockerfilePath:    filepath.Join(dir, ".chc", "runtime", "agent", "Dockerfile"),
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
