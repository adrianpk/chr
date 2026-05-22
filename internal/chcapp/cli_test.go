package chcapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	return Run(ctx, args, strings.NewReader(""), stdout, stderr, testAssetsFS())
}

func testAssetsFS() fs.FS {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("resolve test asset path")
	}

	root := filepath.Join(filepath.Dir(file), "..", "..", "cmd", "chc")

	return os.DirFS(root)
}

func TestRootCommandDefaultsToTUIAction(t *testing.T) {
	cmd := newRootCommand(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if cmd.Action == nil {
		t.Fatal("root command Action is nil")
	}
}

func TestRootCommandIncludesTUICommand(t *testing.T) {
	cmd := newRootCommand(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})

	for _, child := range cmd.Commands {
		if child.Name == "tui" {
			return
		}
	}

	t.Fatal("root command does not include tui command")
}

func TestInitDryRunDoesNotWriteWorkspace(t *testing.T) {
	projectDir := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"init",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("init dry-run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "create .chc/agents") {
		t.Fatalf("stdout = %q, want planned .chc paths", stdout.String())
	}

	assertPathNotExists(t, filepath.Join(projectDir, ".chc"))
}

func TestInitPreservesExistingRootAgents(t *testing.T) {
	projectDir := t.TempDir()
	agentsPath := filepath.Join(projectDir, "AGENTS.md")

	err := os.WriteFile(agentsPath, []byte("custom agents\n"), 0o644)
	if err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{"init", "--dir", projectDir}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("init returned error: %v", err)
	}

	content := readFile(t, agentsPath)
	if content != "custom agents\n" {
		t.Fatalf("AGENTS.md = %q, want preserved custom content", content)
	}

	if !strings.Contains(stdout.String(), "skip existing AGENTS.md") {
		t.Fatalf("stdout = %q, want skipped AGENTS.md", stdout.String())
	}
}

func TestInitStandardProfileInstallsProtocolAssets(t *testing.T) {
	projectDir := t.TempDir()

	err := run(t.Context(), []string{
		"init",
		"--dir",
		projectDir,
		"--profile",
		"standard",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("init standard returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "agents", "CONVENTIONS.md"))
	assertPathExists(t, filepath.Join(projectDir, ".chc", "agents", "playbooks", "ticket-workflow.md"))
	assertPathExists(t, filepath.Join(projectDir, "AGENTS.md"))

	documenting := readTestFile(t, filepath.Join(projectDir, ".chc", "agents", "DOCUMENTING.md"))
	for _, unwanted := range []string{
		"ChrysoCoord Artifacts",
		"Use `.chc/` for project-local operational artifacts",
		"generated operational records",
	} {
		if strings.Contains(documenting, unwanted) {
			t.Fatalf("DOCUMENTING.md contains operational artifact rule %q:\n%s", unwanted, documenting)
		}
	}

	architecture := readTestFile(t, filepath.Join(projectDir, ".chc", "agents", "ARCHITECTURE.md"))
	if strings.Contains(architecture, ".chc/         ChrysoCoord project-local state") {
		t.Fatalf("ARCHITECTURE.md exposes .chc as application layout:\n%s", architecture)
	}

	invariants := readTestFile(t, filepath.Join(projectDir, ".chc", "agents", "INVARIANTS.md"))
	if strings.Contains(invariants, "Treat `.chc/` as project-local operational state") {
		t.Fatalf("INVARIANTS.md contains old .chc operational rule:\n%s", invariants)
	}
}

func TestInitRejectsFileTarget(t *testing.T) {
	projectDir := t.TempDir()
	targetPath := filepath.Join(projectDir, "not-a-dir")

	err := os.WriteFile(targetPath, []byte("content\n"), 0o644)
	if err != nil {
		t.Fatalf("write target file: %v", err)
	}

	err = run(t.Context(), []string{"init", "--dir", targetPath}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("init returned nil error")
	}

	if !strings.Contains(err.Error(), "project directory is not a directory") {
		t.Fatalf("init error = %v, want not a directory", err)
	}
}

func TestDaemonStatusPrintsStoppedWhenStateIsMissing(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopped",
		"state: missing",
		"data_dir: " + filepath.Join(devRoot, "data", "chc"),
		"state_dir: " + filepath.Join(devRoot, "state", "chc"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStatusPrintsStoredRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	heartbeatAt := startedAt.Add(time.Minute)

	upsertDaemonState(t, devRoot, chc.DaemonState{
		Status:        chc.DaemonStatusRunning,
		PID:           1234,
		StartedAt:     &startedAt,
		HeartbeatAt:   &heartbeatAt,
		FailureReason: "",
		UpdatedAt:     heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: running",
		"state: stored",
		"pid: 1234",
		"started_at:",
		"heartbeat_at:",
		"updated_at:",
		"data_dir: " + filepath.Join(devRoot, "data", "chc"),
		"state_dir: " + filepath.Join(devRoot, "state", "chc"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStatusPrintsStoredStoppingState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	stopRequestedAt := startedAt.Add(time.Minute)

	upsertDaemonState(t, devRoot, chc.DaemonState{
		Status:          chc.DaemonStatusStopping,
		PID:             1234,
		StartedAt:       &startedAt,
		StopRequestedAt: &stopRequestedAt,
		UpdatedAt:       stopRequestedAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopping",
		"state: stored",
		"pid: 1234",
		"stop_requested_at:",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStatusPrintsStoredStaleState(t *testing.T) {
	devRoot := t.TempDir()
	updatedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	upsertDaemonState(t, devRoot, chc.DaemonState{
		Status:        chc.DaemonStatusStale,
		PID:           1234,
		FailureReason: "missing process",
		UpdatedAt:     updatedAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stale",
		"state: stored",
		"pid: 1234",
		"failure_reason: missing process",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStatusRecoversStaleRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC().Add(-10 * time.Minute)
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, chc.DaemonState{
		Status:      chc.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stale",
		"state: stored",
		"pid: 1234",
		"failure_reason: daemon state is stale",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	state := getDaemonState(t, devRoot)
	if state.Status != chc.DaemonStatusStale {
		t.Fatalf("daemon state = %#v, want stale", state)
	}
}

func TestDaemonStartDryRunPrintsPlan(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon start returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: planned",
		"action: start",
		"dry_run: true",
		"allowed: true",
		"stale: false",
		"reason: daemon state is missing",
		"data_dir: " + filepath.Join(devRoot, "data", "chc"),
		"state_dir: " + filepath.Join(devRoot, "state", "chc"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStartDryRunRejectsFreshRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, chc.DaemonState{
		Status:      chc.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon start returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: planned",
		"action: start",
		"allowed: false",
		"stale: false",
		"reason: daemon already running with pid 1234",
		"current_status: running",
		"current_pid: 1234",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestAuthorSessionNewListAndShow(t *testing.T) {
	projectDir := newTestWorkspace(t)
	sessionID := createAuthorSessionForTest(t, projectDir)

	assertAuthorSessionListContains(t, projectDir, sessionID)
	assertAuthorSessionShowContains(t, projectDir, sessionID, sessionID)
	assertAuthorSessionShowContains(t, projectDir, "billing-sync", sessionID)
}

func createAuthorSessionForTest(t *testing.T, projectDir string) string {
	t.Helper()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"session",
		"new",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author session new returned error: %v", err)
	}

	return authorSessionIDFromOutput(t, stdout.String())
}

func assertAuthorSessionListContains(t *testing.T, projectDir, sessionID string) {
	t.Helper()

	var listOut bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"session",
		"list",
		"--dir",
		projectDir,
	}, &listOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author session list returned error: %v", err)
	}

	if !strings.Contains(listOut.String(), sessionID) {
		t.Fatalf("list output does not contain session id %q:\n%s", sessionID, listOut.String())
	}

	if !strings.Contains(listOut.String(), "id\tstatus\ttarget\tupdated_at\ttopic") {
		t.Fatalf("list output does not contain headers:\n%s", listOut.String())
	}

	if !strings.Contains(listOut.String(), "Z\tBilling sync") {
		t.Fatalf("list output does not contain updated time before topic:\n%s", listOut.String())
	}
}

func assertAuthorSessionShowContains(t *testing.T, projectDir, sessionHandle, expectedSessionID string) {
	t.Helper()

	var showOut bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"session",
		"show",
		"--dir",
		projectDir,
		sessionHandle,
	}, &showOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author session show returned error: %v", err)
	}

	for _, want := range []string{
		"session: " + expectedSessionID,
		"mode: markdown_direct",
		"target: spec",
		"topic: Billing sync",
		"updated_at:",
	} {
		if !strings.Contains(showOut.String(), want) {
			t.Fatalf("show output does not contain %q:\n%s", want, showOut.String())
		}
	}
}

func TestAuthorQuizNewDryRunPrintsEditableQuiz(t *testing.T) {
	projectDir := newTestWorkspace(t)
	installFakeAuthoringQuizProvider(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"quiz",
		"new",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
		"--question-count",
		"2",
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author quiz new dry-run returned error: %v", err)
	}

	for _, want := range []string{
		"quiz: planned",
		"dry_run: true",
		"mode: context",
		"provider: fake",
		"requested_questions: 2",
		"generated_questions: 3",
		"# Context-Aware Quiz",
		"Answer:",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestAuthorQuizNewWritesSessionAndQuiz(t *testing.T) {
	projectDir := newTestWorkspace(t)
	installFakeAuthoringQuizProvider(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"quiz",
		"new",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author quiz new returned error: %v", err)
	}

	sessionID := authorSessionIDFromOutput(t, stdout.String())
	sessionPath := filepath.Join(projectDir, ".chc", "authoring", "sessions", sessionID, "session.md")
	quizPath := filepath.Join(projectDir, ".chc", "quizzes", sessionID+"-billing-sync.md")

	for _, path := range []string{sessionPath, quizPath} {
		_, err = os.Stat(path)
		if err != nil {
			t.Fatalf("expected path %s: %v", path, err)
		}
	}
}

func TestAuthorPromoteQuizToDraftWritesDraftAndUpdatesSession(t *testing.T) {
	projectDir := newTestWorkspace(t)
	sessionID, quizPath := createAuthorQuizForTest(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"promote",
		"quiz-to-draft",
		"--dir",
		projectDir,
		"--quiz",
		quizPath,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author promote quiz-to-draft returned error: %v", err)
	}

	draftPath := filepath.Join(projectDir, ".chc", "drafts", sessionID+"-billing-sync.md")

	_, err = os.Stat(draftPath)
	if err != nil {
		t.Fatalf("expected draft path %s: %v", draftPath, err)
	}

	var showOut bytes.Buffer

	err = run(t.Context(), []string{
		"author",
		"session",
		"show",
		"--dir",
		projectDir,
		sessionID,
	}, &showOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author session show returned error: %v", err)
	}

	for _, want := range []string{
		"outputs: 2",
		"provenance: 2",
	} {
		if !strings.Contains(showOut.String(), want) {
			t.Fatalf("show output does not contain %q:\n%s", want, showOut.String())
		}
	}
}

func TestAuthorPromoteQuizToDraftDryRunPrintsDraft(t *testing.T) {
	projectDir := newTestWorkspace(t)
	sessionID, quizPath := createAuthorQuizForTest(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"promote",
		"quiz-to-draft",
		"--dir",
		projectDir,
		"--session",
		sessionID,
		"--quiz",
		quizPath,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author promote quiz-to-draft dry-run returned error: %v", err)
	}

	for _, want := range []string{
		"promotion: quiz-to-draft",
		"dry_run: true",
		"# Draft",
		"## Curated Answers",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestAuthorAssistedStartResumeAnswerStatusAndFinalize(t *testing.T) {
	projectDir := newTestWorkspace(t)
	sessionID := createAuthorAssistedSessionForTest(t, projectDir)
	question := askAuthorAssistedQuestionForTest(t, projectDir, sessionID)

	answerAuthorAssistedQuestionForTest(t, projectDir, sessionID, question)

	var statusOut bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"assisted",
		"status",
		"--dir",
		projectDir,
		"--session",
		sessionID,
	}, &statusOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted status returned error: %v", err)
	}

	for _, want := range []string{"assisted: status", "questions: 1", "answers: 1"} {
		if !strings.Contains(statusOut.String(), want) {
			t.Fatalf("status output missing %q:\n%s", want, statusOut.String())
		}
	}

	var finalizeOut bytes.Buffer

	err = run(t.Context(), []string{
		"author",
		"assisted",
		"finalize",
		"--dir",
		projectDir,
		"--session",
		sessionID,
		"--fake-summary",
		"Billing sync needs a durable spec.",
		"--target-kind",
		"spec",
		"--target-name",
		"billing sync spec",
	}, &finalizeOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted finalize returned error: %v", err)
	}

	for _, want := range []string{"assisted: finalized", "curated-summary.md", ".chc/specs/"} {
		if !strings.Contains(finalizeOut.String(), want) {
			t.Fatalf("finalize output missing %q:\n%s", want, finalizeOut.String())
		}
	}
}

func TestDaemonStartDryRunAllowsStaleRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC().Add(-10 * time.Minute)
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, chc.DaemonState{
		Status:      chc.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon start returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: planned",
		"action: start",
		"allowed: true",
		"stale: true",
		"reason: daemon state is stale",
		"current_status: running",
		"current_pid: 1234",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStopDryRunPrintsPlan(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"stop",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon stop returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: planned",
		"action: stop",
		"dry_run: true",
		"data_dir: " + filepath.Join(devRoot, "data", "chc"),
		"state_dir: " + filepath.Join(devRoot, "state", "chc"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStartRunsForegroundLoop(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	err := os.MkdirAll(filepath.Join(projectDir, ".chc", "tickets"), 0o755)
	if err != nil {
		t.Fatalf("create tickets directory: %v", err)
	}

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
		"--max-iterations",
		"1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon start returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopped",
		"iterations: 1",
		"stop_requested: false",
		"projects_loaded: 1",
		"projects_ready: 1",
		"projects_skipped: 0",
		"projects_refreshed: 1",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	state := getDaemonState(t, devRoot)
	if state.Status != chc.DaemonStatusStopped {
		t.Fatalf("daemon state = %#v, want stopped", state)
	}
}

func TestDaemonStartRejectsInvalidMaxIterations(t *testing.T) {
	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		t.TempDir(),
		"--max-iterations",
		"0",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("daemon start returned nil error")
	}

	if !strings.Contains(err.Error(), "max-iterations must be greater than zero") {
		t.Fatalf("daemon start error = %v, want max-iterations validation", err)
	}
}

func TestDaemonStartRejectsFreshRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, chc.DaemonState{
		Status:      chc.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("daemon start returned nil error")
	}

	if !strings.Contains(err.Error(), "daemon start blocked: daemon already running with pid 1234") {
		t.Fatalf("daemon start error = %v, want blocked running daemon", err)
	}
}

func TestDaemonStopRequestsRunningDaemonStop(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, chc.DaemonState{
		Status:      chc.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{"daemon", "stop", "--dev-root", devRoot}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon stop returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopping",
		"action: stop",
		"stop_requested_at:",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	state := getDaemonState(t, devRoot)
	if state.Status != chc.DaemonStatusStopping || state.StopRequestedAt == nil {
		t.Fatalf("daemon state = %#v, want stopping with stop request", state)
	}
}

func TestDaemonStopReportsStoppedWhenStateIsMissing(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{"daemon", "stop", "--dev-root", devRoot}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon stop returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopped",
		"action: stop",
		"state: missing",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStopRequestedHelperReadsStopRequest(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	stopRequestedAt := startedAt.Add(time.Minute)

	upsertDaemonState(t, devRoot, chc.DaemonState{
		Status:          chc.DaemonStatusStopping,
		PID:             1234,
		StartedAt:       &startedAt,
		StopRequestedAt: &stopRequestedAt,
		UpdatedAt:       stopRequestedAt,
	})

	repository := newDaemonStateRepositoryForTest(t, devRoot)

	stopRequested, err := daemonStopRequested(t.Context(), repository)
	if err != nil {
		t.Fatalf("daemonStopRequested returned error: %v", err)
	}

	if !stopRequested {
		t.Fatal("daemonStopRequested = false, want true")
	}
}

func TestTicketNewCreatesTicket(t *testing.T) {
	dir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"new",
		"--dir",
		dir,
		"--priority",
		"high",
		"--tag",
		"uxui",
		"Do something",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	wantPath := filepath.Join(dir, ".chc", "tickets", "CHC-1-Do-something.md")
	if !strings.Contains(stdout.String(), "ticket: "+wantPath) {
		t.Fatalf("stdout = %q, want ticket path %q", stdout.String(), wantPath)
	}

	content, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		`key: "CHC-1"`,
		`title: "Do something"`,
		`priority: "high"`,
		`tags: ["uxui"]`,
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestTicketNewUsesConfigKeyPrefix(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, ".chc", "config.yml")

	err := os.MkdirAll(filepath.Dir(configPath), 0o755)
	if err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	err = os.WriteFile(configPath, []byte("tickets:\n  key_prefix: PRJ\n"), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	err = run(t.Context(), []string{"ticket", "new", "--dir", dir, "Do something"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(dir, ".chc", "tickets", "PRJ-1-Do-something.md"))
}

func TestTicketNewUsesActiveProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	err = run(t.Context(), []string{
		"ticket",
		"new",
		"--dev-root",
		devRoot,
		"Active project ticket",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket new returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Active-project-ticket.md"))
}

func TestRuntimeInitCreatesTemplateWithExplicitDir(t *testing.T) {
	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"))

	if !strings.Contains(stdout.String(), "create .chc/runtime/agent/Dockerfile") {
		t.Fatalf("stdout = %q, want runtime Dockerfile create", stdout.String())
	}
}

func TestRuntimeInitDryRunDoesNotCreateTemplate(t *testing.T) {
	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	assertPathNotExists(t, filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"))

	if !strings.Contains(stdout.String(), "create .chc/runtime/agent/Dockerfile") {
		t.Fatalf("stdout = %q, want runtime Dockerfile create", stdout.String())
	}
}

func TestRuntimeInitSkipsExistingTemplate(t *testing.T) {
	projectDir := newTestWorkspace(t)

	dockerfilePath := filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile")

	err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755)
	if err != nil {
		t.Fatalf("create runtime template directory: %v", err)
	}

	err = os.WriteFile(dockerfilePath, []byte("FROM custom\n"), 0o644)
	if err != nil {
		t.Fatalf("write runtime template: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"runtime",
		"init",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("read runtime template: %v", err)
	}

	if string(content) != "FROM custom\n" {
		t.Fatalf("runtime template content = %q, want custom content unchanged", string(content))
	}

	if !strings.Contains(stdout.String(), "skip existing .chc/runtime/agent/Dockerfile") {
		t.Fatalf("stdout = %q, want runtime Dockerfile skip", stdout.String())
	}
}

func TestRuntimeInitUsesActiveProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dev-root",
		devRoot,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"))
}

func TestRuntimeInitFallsBackToCurrentDirectory(t *testing.T) {
	projectDir := newTestWorkspace(t)

	runInDirectory(t, projectDir, func() {
		err := run(t.Context(), []string{
			"runtime",
			"init",
			"--dev-root",
			t.TempDir(),
		}, &bytes.Buffer{}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("runtime init returned error: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"))
}

func TestRuntimeValidatePrintsRuntimePlan(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeRuntimeTemplate(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"validate",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime validate returned error: %v", err)
	}

	for _, want := range []string{
		"dry_run: true",
		"image: chc-agent-runtime:local",
		"dockerfile: .chc/runtime/agent/Dockerfile",
		"workdir: /workspace",
		"artifact_dir: .chc/runtime/runs",
		"mount: . -> /workspace read_only=false",
		"checks:",
		"ok    workspace    .chc",
		"ok    dockerfile   .chc/runtime/agent/Dockerfile",
		"ok    artifact_dir .chc/runtime/runs",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestRuntimeValidateFailsWhenRuntimeTemplateIsMissing(t *testing.T) {
	projectDir := newTestWorkspace(t)

	err := os.MkdirAll(filepath.Join(projectDir, ".chc", "runtime", "runs"), 0o755)
	if err != nil {
		t.Fatalf("create runtime runs dir: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"runtime",
		"validate",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("runtime validate returned nil error")
	}

	if !strings.Contains(stdout.String(), "fail  dockerfile") {
		t.Fatalf("stdout = %q, want failed Dockerfile check", stdout.String())
	}
}

func TestRuntimeValidateUsesActiveProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeRuntimeTemplate(t, projectDir)

	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := run(t.Context(), []string{
		"runtime",
		"validate",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime validate returned error: %v", err)
	}
}

func TestNotifyStatusPrintsNotificationReadiness(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeCLIWorkspaceConfig(t, projectDir, `notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"status",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify status returned error: %v", err)
	}

	for _, want := range []string{
		"notify: status",
		"project_id:",
		"project_dir: " + projectDir,
		"telegram_enabled: true",
		"telegram_status: ready",
		"telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN",
		"telegram_chat_id: 123456",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestNotifyStatusReportsInvalidTelegramConfig(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeCLIWorkspaceConfig(t, projectDir, `notifications:
  telegram_enabled: true
  telegram_chat_id: "123456"
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"status",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify status returned error: %v", err)
	}

	for _, want := range []string{
		"telegram_enabled: true",
		"telegram_status: invalid",
		"telegram_error: telegram bot token environment variable is required",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestNotifyTelegramTestDryRunPrintsPlan(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeCLIWorkspaceConfig(t, projectDir, `notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"test",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify telegram test dry-run returned error: %v", err)
	}

	for _, want := range []string{
		"telegram_test: planned",
		"dry_run: true",
		"bot_token_env: CHC_TELEGRAM_BOT_TOKEN",
		"chat_id: 123456",
		"message: ChrysoCoord Telegram notification test",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestNotifyTelegramTestRequiresDryRun(t *testing.T) {
	projectDir := newTestWorkspace(t)

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"test",
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("notify telegram test returned nil error")
	}

	if err.Error() != "telegram bot token is required" {
		t.Fatalf("notify telegram test error = %v", err)
	}
}

func TestNotifyTelegramTestSendsAndRecordsAudit(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	t.Setenv("CHC_TELEGRAM_BOT_TOKEN", "token-1")

	writeTelegramNotifyConfig(t, projectDir)

	gotPayload := installTestTelegramSender(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"test",
		"--dir",
		projectDir,
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify telegram test returned error: %v", err)
	}

	assertTelegramNotifyOutput(t, stdout.String())
	assertTelegramPayload(t, gotPayload)

	records := listNotificationEventsForTest(t, devRoot, projectID(projectDir))
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	if records[0].Event.Status != chc.NotificationStatusDelivered {
		t.Fatalf("notification status = %q, want delivered", records[0].Event.Status)
	}
}

func TestNotifyTelegramPollDryRunParsesFixture(t *testing.T) {
	devRoot := t.TempDir()
	fixturePath := writeTelegramUpdateFixture(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"poll",
		"--fixture",
		fixturePath,
		"--dry-run",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify telegram poll dry-run returned error: %v", err)
	}

	for _, want := range []string{"telegram_poll: parsed", "dry_run: true", "records: 2", "accepted", "rejected"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	records := listRemoteCommandRecordsForTest(t, devRoot)
	if len(records) != 0 {
		t.Fatalf("remote command records = %d, want dry-run persistence skipped", len(records))
	}
}

func TestNotifyTelegramPollPersistsFixtureRecords(t *testing.T) {
	devRoot := t.TempDir()
	fixturePath := writeTelegramUpdateFixture(t)

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"poll",
		"--fixture",
		fixturePath,
		"--dev-root",
		devRoot,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify telegram poll returned error: %v", err)
	}

	records := listRemoteCommandRecordsForTest(t, devRoot)
	if len(records) != 2 {
		t.Fatalf("remote command records = %d, want 2", len(records))
	}

	if records[0].Status != chc.RemoteCommandRecordAccepted || records[1].Status != chc.RemoteCommandRecordRejected {
		t.Fatalf("records = %#v, want accepted then rejected", records)
	}
}

type testTelegramPayload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func writeTelegramNotifyConfig(t *testing.T, projectDir string) {
	t.Helper()

	writeCLIWorkspaceConfig(t, projectDir, `notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)
}

func writeTelegramUpdateFixture(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "updates.json")
	content := `{"ok":true,"result":[
{"update_id":1,"message":{"message_id":10,"from":{"id":100,"username":"ops"},"chat":{"id":200},"date":1768000000,"text":"@example status"}},
{"update_id":2,"message":{"message_id":11,"from":{"id":101,"username":"ops"},"chat":{"id":201},"date":1768000001,"text":"merge dev"}}
]}`

	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write telegram update fixture: %v", err)
	}

	return path
}

func installTestTelegramSender(t *testing.T) *testTelegramPayload {
	t.Helper()

	gotPayload := &testTelegramPayload{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		err := json.NewDecoder(request.Body).Decode(gotPayload)
		if err != nil {
			t.Fatalf("decode telegram payload: %v", err)
		}

		_, err = writer.Write([]byte(`{"ok":true}`))
		if err != nil {
			t.Fatalf("write telegram response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	previousSender := notifyTelegramSender
	notifyTelegramSender = chc.TelegramSender{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	t.Cleanup(func() {
		notifyTelegramSender = previousSender
	})

	return gotPayload
}

func assertTelegramNotifyOutput(t *testing.T, output string) {
	t.Helper()

	for _, want := range []string{
		"telegram_test: sent",
		"dry_run: false",
		"notification_id: telegram-test-",
		"status_code: 200",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, output)
		}
	}
}

func assertTelegramPayload(t *testing.T, gotPayload *testTelegramPayload) {
	t.Helper()

	if gotPayload.ChatID != "123456" {
		t.Fatalf("telegram chat_id = %q, want 123456", gotPayload.ChatID)
	}

	if !strings.Contains(gotPayload.Text, "ChrysoCoord Telegram notification test") {
		t.Fatalf("telegram text = %q, want test message", gotPayload.Text)
	}
}

func TestTicketNewFallsBackToCurrentDirectory(t *testing.T) {
	projectDir := newTestWorkspace(t)

	runInDirectory(t, projectDir, func() {
		err := run(t.Context(), []string{
			"ticket",
			"new",
			"--dev-root",
			t.TempDir(),
			"Current project ticket",
		}, &bytes.Buffer{}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("ticket new returned error: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Current-project-ticket.md"))
}

func TestProjectContextResolverRejectsMissingContext(t *testing.T) {
	env, err := chcTestUserEnvironment(t)
	if err != nil {
		t.Fatalf("chcTestUserEnvironment returned error: %v", err)
	}

	_, err = newCommandProjectContextResolver(env).Resolve(t.Context(), commandProjectContextOptions{})
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}
}

func TestTicketNewRejectsMissingTitle(t *testing.T) {
	err := run(t.Context(), []string{"ticket", "new"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run returned nil error")
	}
}

func TestTicketListPrintsIndexedTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Status:      "open",
		Priority:    "high",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "onboarding",
		ProjectArea: "uxui",
		CreatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"list",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket list returned error: %v", err)
	}

	want := "CHC-1\tDo something\topen\thigh\tnormal\ttask\tonboarding\tCHC-1-Do-something.md"
	if !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout = %q, want ticket row %q", stdout.String(), want)
	}
}

func TestTicketLoopDryRunPrintsEligibleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketLoopDryRunTickets(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		"--priority",
		"high",
		"--project-area",
		"frontend",
		"--tag",
		"uxui",
		"--order",
		"priority",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	for _, want := range []string{
		"loop: planned",
		"dry_run: true",
		"order: priority",
		"eligible: 1",
		"ticket: CHC-1\tReady frontend\tready\thigh\tcritical\tfeature\tonboarding\tCHC-1-Ready.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if strings.Contains(stdout.String(), "CHC-2") || strings.Contains(stdout.String(), "CHC-3") {
		t.Fatalf("stdout contains ineligible tickets:\n%s", stdout.String())
	}
}

func writeTicketLoopDryRunTickets(t *testing.T, projectDir string) {
	t.Helper()

	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", testLoopTicket(
		"CHC-2", "Ready backend", "ready", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	))
	writeTicketMarkdownOnly(t, projectDir, "CHC-3-Open.md", testLoopTicket(
		"CHC-3", "Open frontend", "open", "critical", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))
}

func testLoopTicket(
	key, title, status, priority, severity, ticketType, useCase, projectArea string,
	tags []string,
) chc.Ticket {
	return chc.Ticket{
		Key:         key,
		Title:       title,
		Status:      status,
		Priority:    priority,
		Severity:    severity,
		Type:        ticketType,
		UseCase:     useCase,
		ProjectArea: projectArea,
		Tags:        tags,
		CreatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	}
}

func TestTicketLoopExecutesEligibleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	for _, want := range []string{
		"loop: loop-",
		"status: succeeded",
		"eligible: 1",
		"processed: 1",
		"skipped: 0",
		"blocked: 0",
		"failed: 0",
		"execution: exec-",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if !strings.Contains(strings.Join(executor.commands, ","), "docker build,docker run,docker run") {
		t.Fatalf("commands = %#v, want build, worker run, and validation run", executor.commands)
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 1 {
		t.Fatalf("execution records = %d, want 1", len(records))
	}

	loopRuns := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loopRuns) != 1 {
		t.Fatalf("loop runs = %d, want 1", len(loopRuns))
	}

	if loopRuns[0].Status != chc.LoopRunStatusSucceeded || loopRuns[0].ProcessedCount != 1 {
		t.Fatalf("loop run = %#v, want succeeded with one processed ticket", loopRuns[0])
	}
}

func TestTicketLoopHonorsMaxIterations(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", testLoopTicket(
		"CHC-2", "Ready backend", "ready", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--max-iterations",
		"1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	for _, want := range []string{
		"status: succeeded",
		"eligible: 2",
		"processed: 1",
		"skipped: 1",
		"execution: exec-",
		"CHC-1",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if strings.Contains(stdout.String(), "CHC-2") {
		t.Fatalf("stdout contains execution beyond max iterations:\n%s", stdout.String())
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-2")
	if len(records) != 0 {
		t.Fatalf("CHC-2 execution records = %d, want 0", len(records))
	}
}

func TestTicketLoopCompletesWithNoEligibleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Open.md", testLoopTicket(
		"CHC-1", "Open frontend", "open", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	for _, want := range []string{
		"status: succeeded",
		"eligible: 0",
		"processed: 0",
		"skipped: 0",
		"blocked: 0",
		"failed: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if len(executor.commands) != 0 {
		t.Fatalf("commands = %#v, want no ticket execution", executor.commands)
	}
}

func TestTicketLoopDispatchesConfiguredFinishedNotification(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	t.Setenv("CHC_TELEGRAM_BOT_TOKEN", "token-1")
	installFakeTicketStartHarnessExecutor(t, nil)

	gotPayload := installTestTelegramSender(t)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTelegramNotifyConfig(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	records := listNotificationEventsForTest(t, devRoot, projectID(projectDir))
	assertNotificationEvent(t, records, 0, chc.NotificationEventLoopFinished, chc.NotificationStatusDelivered)

	if gotPayload.ChatID != "123456" {
		t.Fatalf("telegram chat id = %q, want configured chat id", gotPayload.ChatID)
	}

	if !strings.Contains(gotPayload.Text, "Ticket loop loop-") {
		t.Fatalf("telegram text = %q, want loop notification", gotPayload.Text)
	}
}

func TestTicketLoopStopsOnExecutionFailure(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)
	executor.failWorker = true

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", testLoopTicket(
		"CHC-2", "Ready backend", "ready", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket loop returned nil error")
	}

	for _, want := range []string{
		"status: failed",
		"eligible: 2",
		"processed: 0",
		"skipped: 1",
		"blocked: 0",
		"failed: 1",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if strings.Contains(stdout.String(), "CHC-2") {
		t.Fatalf("stdout contains execution after failure:\n%s", stdout.String())
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-2")
	if len(records) != 0 {
		t.Fatalf("CHC-2 execution records = %d, want 0", len(records))
	}

	loopRuns := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loopRuns) != 1 {
		t.Fatalf("loop runs = %d, want 1", len(loopRuns))
	}

	if loopRuns[0].Status != chc.LoopRunStatusFailed || loopRuns[0].FailedCount != 1 {
		t.Fatalf("loop run = %#v, want failed with one failed ticket", loopRuns[0])
	}
}

func TestTicketLoopRecordsFailedNotification(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)
	executor.failWorker = true

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket loop returned nil error")
	}

	notificationRecords := listNotificationEventsForTest(t, devRoot, projectID(projectDir))
	assertNotificationEvent(t, notificationRecords, 0, chc.NotificationEventLoopFailed, chc.NotificationStatusSkipped)
}

func TestTicketLoopRecordsBlockedNotifications(t *testing.T) {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	record := chc.LoopRunRecord{
		ID:             "loop-1",
		ProjectID:      "project-1",
		Status:         chc.LoopRunStatusBlocked,
		EligibleCount:  1,
		ProcessedCount: 1,
		BlockedCount:   1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	event, ok := loopRunNotificationEvent(commandProjectContext{ProjectID: "project-1"}, record, now)
	if !ok {
		t.Fatal("loopRunNotificationEvent returned false")
	}

	if event.Type != chc.NotificationEventLoopBlocked {
		t.Fatalf("event type = %q, want loop_blocked", event.Type)
	}
}

func TestTicketShowPrintsIndexedTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Status:      "open",
		Priority:    "high",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "onboarding",
		ProjectArea: "uxui",
		CreatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"show",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket show returned error: %v", err)
	}

	for _, want := range []string{
		"key: CHC-1",
		"title: Do something",
		"status: open",
		"priority: high",
		"severity: normal",
		"type: task",
		"use_case: onboarding",
		"project_area: uxui",
		"path: CHC-1-Do-something.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestTicketStartDryRunPrintsPreflight(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket start returned error: %v", err)
	}

	for _, want := range []string{
		"execution: exec-",
		"ticket: CHC-1",
		"ticket_path: " + filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Do-something.md"),
		"status: pending",
		"dry_run: true",
		"run_artifact_dir: .chc/runtime/runs/CHC-1/exec-",
		"result_path: .chc/runtime/runs/CHC-1/exec-",
		"ok    dockerfile   .chc/runtime/agent/Dockerfile",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 0 {
		t.Fatalf("execution records = %d, want 0", len(records))
	}
}

func TestTicketStartCreatesPendingExecution(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket start returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "dry_run: false") {
		t.Fatalf("stdout = %q, want dry_run false", stdout.String())
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 1 {
		t.Fatalf("execution records = %d, want 1", len(records))
	}

	record := records[0]
	assertTicketStartSucceededExecution(t, record, projectDir)

	if !strings.Contains(strings.Join(executor.commands, ","), "docker build,docker run") {
		t.Fatalf("commands = %#v, want build and run", executor.commands)
	}

	logContent, err := os.ReadFile(record.Artifacts.LogPath)
	if err != nil {
		t.Fatalf("read execution log: %v", err)
	}

	if !strings.Contains(string(logContent), "stdout: docker build") {
		t.Fatalf("log content = %q, want harness output", string(logContent))
	}
}

func TestTicketStartPrintsHarnessFailure(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, fmt.Errorf("docker unavailable"))

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket start returned nil error")
	}

	for _, want := range []string{
		"status: failed",
		"failure_kind: harness",
		"failure_reason: harness: docker unavailable",
		"diagnostic_log_path: " + filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1"),
		"diagnostic_result_path: " + filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestTicketStartPrintsWorkerFailure(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)
	executor.failWorker = true

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket start returned nil error")
	}

	for _, want := range []string{
		"status: failed",
		"failure_kind: worker",
		"failure_reason: worker: codex failed",
		"diagnostic_log_path: " + filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1"),
		"diagnostic_result_path: " + filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1"),
		"diagnostic_validation_path: " + filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func assertTicketStartSucceededExecution(t *testing.T, record chc.ExecutionRecord, projectDir string) {
	t.Helper()

	if record.Status != chc.ExecutionStatusSucceeded {
		t.Fatalf("Status = %q, want succeeded", record.Status)
	}

	if record.TicketPath != filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Do-something.md") {
		t.Fatalf("TicketPath = %q, want ticket path", record.TicketPath)
	}

	if !strings.HasPrefix(record.ID, "exec-") {
		t.Fatalf("ID = %q, want exec prefix", record.ID)
	}

	if record.Artifacts.RunDir != filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1", record.ID) {
		t.Fatalf("Artifacts.RunDir = %q, want execution run dir", record.Artifacts.RunDir)
	}

	if record.Artifacts.ValidationPath != filepath.Join(record.Artifacts.RunDir, "validation.log") {
		t.Fatalf("Artifacts.ValidationPath = %q, want validation log path", record.Artifacts.ValidationPath)
	}
}

func TestTicketStartFailsWhenRuntimeTemplateIsMissing(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	err := os.MkdirAll(filepath.Join(projectDir, ".chc", "runtime", "runs"), 0o755)
	if err != nil {
		t.Fatalf("create runtime runs dir: %v", err)
	}

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket start returned nil error")
	}

	if !strings.Contains(stdout.String(), "fail  dockerfile") {
		t.Fatalf("stdout = %q, want failed Dockerfile check", stdout.String())
	}

	if !strings.Contains(stdout.String(), "failure_kind: runtime") {
		t.Fatalf("stdout = %q, want runtime failure kind", stdout.String())
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 0 {
		t.Fatalf("execution records = %d, want 0", len(records))
	}
}

func TestTicketApplyResultAppliesExecutionResult(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	addProject(t, devRoot, "example", projectDir)

	record := createApplyResultExecution(t, devRoot, projectDir, chc.ExecutionStatusSucceeded, "")
	writeApplyResultFile(t, record, chc.ExecutionResultStatusSucceeded, "ready for review")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		record.ID,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	for _, want := range []string{
		"execution: exec-apply",
		"ticket: CHC-1",
		"status: reviewing",
		"result_status: succeeded",
		"execution_report_path: " + filepath.Join(record.Artifacts.RunDir, "execution-report.md"),
		"dry_run: false",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	content := readFile(t, ticketPath)
	for _, want := range []string{
		`status: "reviewing"`,
		"Execution exec-apply finished with status succeeded.",
		"Summary: ready for review",
		"## ",
		"- command: make verify",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}

	assertPathExists(t, filepath.Join(record.Artifacts.RunDir, "result-application.md"))
	assertPathExists(t, filepath.Join(record.Artifacts.RunDir, "execution-report.md"))

	tickets := listProjectTickets(t, devRoot, projectDir)
	if len(tickets) != 1 || tickets[0].Status != "reviewing" {
		t.Fatalf("indexed tickets = %#v, want reviewing CHC-1", tickets)
	}
}

func TestTicketApplyResultRecordsBlockedNotification(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	addProject(t, devRoot, "example", projectDir)

	record := createApplyResultExecution(t, devRoot, projectDir, chc.ExecutionStatusFailed, "codex failed")
	writeApplyResultFile(t, record, chc.ExecutionResultStatusFailed, "blocked by compile error")

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		record.ID,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	records := listNotificationEventsForTest(t, devRoot, projectID(projectDir))
	assertNotificationEvent(t, records, 0, chc.NotificationEventTicketBlocked, chc.NotificationStatusSkipped)

	if records[0].Event.References.ExecutionID != record.ID {
		t.Fatalf("execution reference = %q, want %q", records[0].Event.References.ExecutionID, record.ID)
	}
}

func TestTicketApplyResultDryRunDoesNotMutateTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	addProject(t, devRoot, "example", projectDir)

	record := createApplyResultExecution(t, devRoot, projectDir, chc.ExecutionStatusSucceeded, "")
	writeApplyResultFile(t, record, chc.ExecutionResultStatusSucceeded, "ready for review")

	before := readFile(t, ticketPath)

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		record.ID,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	after := readFile(t, ticketPath)
	if after != before {
		t.Fatalf("dry-run mutated ticket:\n%s", after)
	}

	assertPathNotExists(t, filepath.Join(record.Artifacts.RunDir, "result-application.md"))
	assertPathNotExists(t, filepath.Join(record.Artifacts.RunDir, "execution-report.md"))
}

func TestReportListListsProjectReports(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"runtime",
		"runs",
		"CHC-1",
		"exec-1",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHC-1\n- status: succeeded\n")
	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"runtime",
		"runs",
		"loop-1",
		"loop-report.md",
	), "# Loop Report\n\n## Summary\n\n- loop: loop-1\n- status: blocked\n")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"report",
		"list",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("report list returned error: %v", err)
	}

	for _, want := range []string{
		"kind: execution path: .chc/runtime/runs/CHC-1/exec-1/execution-report.md ticket: CHC-1 execution: exec-1 loop: none status: succeeded",
		"kind: loop path: .chc/runtime/runs/loop-1/loop-report.md ticket: none execution: none loop: loop-1 status: blocked",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestReportListFiltersReports(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"runtime",
		"runs",
		"CHC-1",
		"exec-1",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHC-1\n- status: succeeded\n")
	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"runtime",
		"runs",
		"CHC-2",
		"exec-2",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-2\n- ticket: CHC-2\n- status: blocked\n")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"report",
		"list",
		"--dir",
		projectDir,
		"--kind",
		"execution",
		"--ticket",
		"CHC-2",
		"--status",
		"blocked",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("report list returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "execution: exec-2") {
		t.Fatalf("stdout does not contain exec-2:\n%s", stdout.String())
	}

	if strings.Contains(stdout.String(), "exec-1") {
		t.Fatalf("stdout contains filtered report:\n%s", stdout.String())
	}
}

func TestReportShowPrintsReportBody(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"runtime",
		"runs",
		"CHC-1",
		"exec-1",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHC-1\n- status: succeeded\n")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"report",
		"show",
		"--dir",
		projectDir,
		"--execution",
		"exec-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("report show returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "# Execution Report") {
		t.Fatalf("stdout does not contain report body:\n%s", stdout.String())
	}
}

func TestReportShowPrintsReportMetadata(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"runtime",
		"runs",
		"loop-1",
		"loop-report.md",
	), "# Loop Report\n\n## Summary\n\n- loop: loop-1\n- status: blocked\n")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"report",
		"show",
		"--dir",
		projectDir,
		"--loop",
		"loop-1",
		"--metadata",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("report show returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "kind: loop path: .chc/runtime/runs/loop-1/loop-report.md") {
		t.Fatalf("stdout does not contain report metadata:\n%s", stdout.String())
	}
}

func TestReportShowRejectsAmbiguousLookup(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"runtime",
		"runs",
		"CHC-1",
		"exec-1",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHC-1\n")
	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"runtime",
		"runs",
		"CHC-1",
		"exec-2",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-2\n- ticket: CHC-1\n")

	err := run(t.Context(), []string{
		"report",
		"show",
		"--dir",
		projectDir,
		"--ticket",
		"CHC-1",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("report show returned nil error")
	}

	if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("error = %v, want ambiguous lookup", err)
	}
}

func TestTicketShowRejectsMissingTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"show",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket show returned nil error")
	}
}

func TestTicketCommentAppendsComment(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"comment",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--author",
		"adrian",
		"CHC-1",
		"Looks ready.",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket comment returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "ticket: "+ticketPath) {
		t.Fatalf("stdout = %q, want ticket path", stdout.String())
	}

	updatedContent, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		"# Comments",
		"adrian",
		"\"Looks ready.\"",
	} {
		if !strings.Contains(string(updatedContent), want) {
			t.Fatalf("updated content does not contain %q:\n%s", want, string(updatedContent))
		}
	}
}

func writeIndexedTicketFile(t *testing.T, devRoot, projectDir string) string {
	t.Helper()

	ticketPath := filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Do-something.md")

	content, err := chc.NewTicketRenderer().RenderTicket(chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(ticketPath), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(ticketPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	indexTicket(t, devRoot, projectDir, chc.IndexedTicket{
		Key:          "CHC-1",
		Title:        "Do something",
		Status:       "open",
		RelativePath: filepath.Join(".chc", "tickets", "CHC-1-Do-something.md"),
		Path:         ticketPath,
	})

	return ticketPath
}

func TestTicketCommentRejectsMissingTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"comment",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"Looks ready.",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket comment returned nil error")
	}
}

func TestTicketStatusUpdatesTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"ready",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket status returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "ticket: CHC-1 ready") {
		t.Fatalf("stdout = %q, want status result", stdout.String())
	}

	updatedContent, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	if !strings.Contains(string(updatedContent), `status: "ready"`) {
		t.Fatalf("updated content does not contain ready status:\n%s", string(updatedContent))
	}
}

func TestTicketStatusRejectsInvalidStatus(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeIndexedTicketFile(t, devRoot, projectDir)
	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"unknown",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket status returned nil error")
	}
}

func TestTicketStatusRejectsMissingTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"ready",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket status returned nil error")
	}
}

func TestIndexRebuildIndexesTicketFiles(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})
	indexTicket(t, devRoot, projectDir, chc.IndexedTicket{
		Key:          "CHC-9",
		Title:        "Old",
		Status:       "open",
		RelativePath: "CHC-9-Old.md",
		Path:         filepath.Join(projectDir, ".chc", "tickets", "CHC-9-Old.md"),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	for _, want := range []string{
		"indexed: 1",
		"parse_errors: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}

	tickets := listProjectTickets(t, devRoot, projectDir)
	if len(tickets) != 1 {
		t.Fatalf("len(tickets) = %d, want 1", len(tickets))
	}

	if tickets[0].Key != "CHC-1" || tickets[0].Title != "Do something" {
		t.Fatalf("ticket = %#v, want rebuilt CHC-1", tickets[0])
	}
}

func TestIndexRebuildReportsParseErrors(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	badTicketPath := filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Bad.md")

	err := os.MkdirAll(filepath.Dir(badTicketPath), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(badTicketPath, []byte("# Bad\n"), 0o644)
	if err != nil {
		t.Fatalf("write bad ticket: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	for _, want := range []string{
		"indexed: 0",
		"parse_errors: 1",
		"parse_error: CHC-1-Bad.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsCleanIndex(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index status returned error: %v", err)
	}

	for _, want := range []string{
		"missing: 0",
		"stale: 0",
		"orphaned: 0",
		"invalid: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsMissingTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("index status returned nil error")
	}

	for _, want := range []string{
		"missing: 1",
		"missing_ticket: CHC-1-Do-something.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsStaleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", chc.Ticket{
		Key:       "CHC-1",
		Title:     "Do something else",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("index status returned nil error")
	}

	for _, want := range []string{
		"stale: 1",
		"stale_ticket: CHC-1-Do-something.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsInvalidTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	badTicketPath := filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Bad.md")

	err := os.MkdirAll(filepath.Dir(badTicketPath), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(badTicketPath, []byte("# Bad\n"), 0o644)
	if err != nil {
		t.Fatalf("write bad ticket: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("index status returned nil error")
	}

	for _, want := range []string{
		"invalid: 1",
		"invalid_ticket: CHC-1-Bad.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceDBStatusReportsPendingMigrationsWhenDatabaseIsMissing(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"db",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance db status returned error: %v", err)
	}

	for _, want := range []string{
		"database_exists: false",
		"schema_migrations_exists: false",
		"applied: 0",
		"pending: 8",
		"unknown: 0",
		"pending_migration: 20260518000001 create_projects",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceDBStatusReportsAppliedMigrations(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project add returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"db",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance db status returned error: %v", err)
	}

	for _, want := range []string{
		"database_exists: true",
		"schema_migrations_exists: true",
		"available: 8",
		"applied: 8",
		"pending: 0",
		"unknown: 0",
		"applied_migration: 20260518000001 create_projects",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceRegistryStatusReportsRegisteredProjectDiagnostics(t *testing.T) {
	projectDir := newTestWorkspace(t)
	noWorkspaceDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	addProject(t, devRoot, "no-workspace", noWorkspaceDir)

	err := os.RemoveAll(filepath.Join(noWorkspaceDir, ".chc"))
	if err != nil {
		t.Fatalf("RemoveAll returned error: %v", err)
	}

	err = run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	for _, want := range []string{
		"projects: 2",
		"project: example path: " + projectDir + " enabled: true active: true path_exists: true workspace_exists: true status: ok",
		"project: no-workspace path: " + noWorkspaceDir + " enabled: true active: false path_exists: true workspace_exists: false status: missing_workspace",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceRegistryStatusReportsMissingProjectPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "missing", projectDir)

	err := os.RemoveAll(projectDir)
	if err != nil {
		t.Fatalf("RemoveAll returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "status: missing_path") {
		t.Fatalf("stdout = %q, want missing path status", stdout.String())
	}
}

func TestMaintenanceRegistryDisableDisablesProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"registry",
		"disable",
		"--dev-root",
		devRoot,
		"example",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry disable returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "disabled: example path: "+projectDir) {
		t.Fatalf("stdout = %q, want disabled project", stdout.String())
	}

	var statusStdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &statusStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(statusStdout.String(), "enabled: false active: false path_exists: true workspace_exists: true status: disabled") {
		t.Fatalf("stdout = %q, want disabled status", statusStdout.String())
	}
}

func TestMaintenanceRegistryDisableDryRunDoesNotMutateProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"registry",
		"disable",
		"--dev-root",
		devRoot,
		"--dry-run",
		"example",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry disable returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "would_disable: example path: "+projectDir) {
		t.Fatalf("stdout = %q, want dry-run disable", stdout.String())
	}

	var statusStdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &statusStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(statusStdout.String(), "enabled: true active: false path_exists: true workspace_exists: true status: ok") {
		t.Fatalf("stdout = %q, want unchanged status", statusStdout.String())
	}
}

func TestMaintenanceRegistrySetPathUpdatesProjectPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	movedDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"registry",
		"set-path",
		"--dev-root",
		devRoot,
		"example",
		movedDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry set-path returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "updated_path: example from: "+projectDir+" to: "+movedDir) {
		t.Fatalf("stdout = %q, want updated path", stdout.String())
	}

	var statusStdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &statusStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(statusStdout.String(), "project: example path: "+movedDir) {
		t.Fatalf("stdout = %q, want moved path", statusStdout.String())
	}
}

func TestMaintenanceRegistrySetPathDryRunDoesNotMutateProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	movedDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"registry",
		"set-path",
		"--dev-root",
		devRoot,
		"--dry-run",
		"example",
		movedDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry set-path returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "would_update_path: example from: "+projectDir+" to: "+movedDir) {
		t.Fatalf("stdout = %q, want dry-run path update", stdout.String())
	}

	var statusStdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &statusStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(statusStdout.String(), "project: example path: "+projectDir) {
		t.Fatalf("stdout = %q, want original path", statusStdout.String())
	}
}

func TestMaintenanceExecutionArtifactsReportsMissingArtifacts(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "runtime", "runs", "exec-1")
	record := cliExecutionRecord(projectDir, runDir)

	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("create run dir: %v", err)
	}

	createCLIExecutionRecord(t, devRoot, record)

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"execution",
		"artifacts",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance execution artifacts returned error: %v", err)
	}

	for _, want := range []string{
		"executions: 1",
		"drifted: 1",
		"missing_artifacts: 4",
		"execution: exec-1 ticket: CHC-1 status: succeeded artifact_status: missing_artifacts",
		"artifact: exec-1 name: run_dir exists: true path: " + record.Artifacts.RunDir,
		"artifact: exec-1 name: result exists: false path: " + record.Artifacts.ResultPath,
		"artifact: exec-1 name: report exists: false path: " + chc.ExecutionReportPath(record),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceExecutionArtifactsReportsCompleteArtifacts(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "runtime", "runs", "exec-1")
	record := cliExecutionRecord(projectDir, runDir)
	writeCLIExecutionArtifacts(t, record)
	createCLIExecutionRecord(t, devRoot, record)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"execution",
		"artifacts",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance execution artifacts returned error: %v", err)
	}

	for _, want := range []string{
		"executions: 1",
		"drifted: 0",
		"missing_artifacts: 0",
		"execution: exec-1 ticket: CHC-1 status: succeeded artifact_status: ok",
		"artifact: exec-1 name: validation exists: true path: " + record.Artifacts.ValidationPath,
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceExecutionCleanupPreviewReportsUnreferencedArtifactDirs(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	artifactDir := filepath.Join(projectDir, ".chc", "runtime", "runs")
	referencedDir := filepath.Join(artifactDir, "CHC-1", "exec-1")
	unreferencedDir := filepath.Join(artifactDir, "CHC-2", "exec-2")
	record := cliExecutionRecord(projectDir, referencedDir)

	for _, dir := range []string{referencedDir, unreferencedDir} {
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			t.Fatalf("create runtime artifact dir %q: %v", dir, err)
		}
	}

	createCLIExecutionRecord(t, devRoot, record)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"execution",
		"cleanup-preview",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance execution cleanup-preview returned error: %v", err)
	}

	for _, want := range []string{
		"artifact_dir: " + artifactDir,
		"referenced_run_dirs: 1",
		"cleanup_candidates: 1",
		"mode: preview",
		"candidate: " + unreferencedDir,
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceExecutionCleanupPreviewDoesNotDeleteCandidates(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	unreferencedDir := filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-2", "exec-2")

	err := os.MkdirAll(unreferencedDir, 0o755)
	if err != nil {
		t.Fatalf("create runtime artifact dir: %v", err)
	}

	err = run(t.Context(), []string{
		"maintenance",
		"execution",
		"cleanup-preview",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance execution cleanup-preview returned error: %v", err)
	}

	assertPathExists(t, unreferencedDir)
}

func TestMaintenanceBundleWritesDiagnosticBundle(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	bundlePath := filepath.Join(projectDir, ".chc", "maintenance", "diagnostic-bundle.md")

	addProject(t, devRoot, "example", projectDir)
	writeCLIWorkspaceConfig(t, projectDir, `version: 1
notifications:
  telegram:
    bot_token_env: SUPER_SECRET_TOKEN_VALUE
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"bundle",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance bundle returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "bundle_path: "+bundlePath) {
		t.Fatalf("stdout = %q, want bundle path", stdout.String())
	}

	content := readFile(t, bundlePath)

	for _, want := range []string{
		"# Diagnostic Bundle",
		"project_dir: " + projectDir,
		"database_exists: true",
		"example path=" + projectDir,
		"workspace config contents omitted",
		"environment values omitted",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("bundle content does not contain %q:\n%s", want, content)
		}
	}

	if strings.Contains(content, "SUPER_SECRET_TOKEN_VALUE") {
		t.Fatalf("bundle content contains secret-like config value:\n%s", content)
	}
}

func TestProjectAddRegistersProject(t *testing.T) {
	projectDir := t.TempDir()
	devRoot := t.TempDir()

	err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755)
	if err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		"--name",
		"example",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want registered project", stdout.String())
	}

	assertPathExists(t, filepath.Join(devRoot, "data", "chc", "chc.db"))
}

func TestProjectAddRejectsProjectWithoutWorkspace(t *testing.T) {
	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		t.TempDir(),
		t.TempDir(),
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run returned nil error")
	}
}

func TestProjectListPrintsRegisteredProjects(t *testing.T) {
	projectDir := t.TempDir()
	devRoot := t.TempDir()

	err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755)
	if err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	err = run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		"--name",
		"example",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project add returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"list",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project list returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "example\t"+projectDir+"\ttrue\tproject-") {
		t.Fatalf("stdout = %q, want listed project", stdout.String())
	}
}

func TestProjectSelectByNameStoresCurrentProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var selectStdout bytes.Buffer

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &selectStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	if !strings.Contains(selectStdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want selected project", selectStdout.String())
	}

	var currentStdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"current",
		"--dev-root",
		devRoot,
	}, &currentStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project current returned error: %v", err)
	}

	if !strings.Contains(currentStdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want current project", currentStdout.String())
	}

	assertPathExists(t, filepath.Join(devRoot, "config", "chc", "active-project"))
}

func TestProjectSelectByPathStoresCurrentProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"current",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project current returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want current project", stdout.String())
	}
}

func TestProjectCurrentRejectsMissingSelection(t *testing.T) {
	err := run(t.Context(), []string{
		"project",
		"current",
		"--dev-root",
		t.TempDir(),
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project current returned nil error")
	}
}

func TestProjectSelectRejectsUnknownProject(t *testing.T) {
	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		t.TempDir(),
		"missing",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project select returned nil error")
	}
}

func TestProjectSelectRejectsMissingRegisteredPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := os.RemoveAll(projectDir)
	if err != nil {
		t.Fatalf("remove project dir: %v", err)
	}

	err = run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project select returned nil error")
	}
}

func TestProjectSelectRejectsRegisteredPathWithoutWorkspace(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := os.RemoveAll(filepath.Join(projectDir, ".chc"))
	if err != nil {
		t.Fatalf("remove project workspace: %v", err)
	}

	err = run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project select returned nil error")
	}
}

func TestTicketNewRejectsMissingActiveProjectPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := os.RemoveAll(projectDir)
	if err != nil {
		t.Fatalf("remove project dir: %v", err)
	}

	err = run(t.Context(), []string{
		"ticket",
		"new",
		"--dev-root",
		devRoot,
		"Do something",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket new returned nil error")
	}
}

func TestTicketNewRejectsActiveProjectWithoutWorkspace(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := os.RemoveAll(filepath.Join(projectDir, ".chc"))
	if err != nil {
		t.Fatalf("remove project workspace: %v", err)
	}

	err = run(t.Context(), []string{
		"ticket",
		"new",
		"--dev-root",
		devRoot,
		"Do something",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket new returned nil error")
	}
}

func addProject(t *testing.T, devRoot, name, projectDir string) {
	t.Helper()

	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		"--name",
		name,
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project add returned error: %v", err)
	}
}

func selectProject(t *testing.T, devRoot, selector string) {
	t.Helper()

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		selector,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}
}

func createCLIExecutionRecord(t *testing.T, devRoot string, record chc.ExecutionRecord) {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := chc.NewExecutionRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewExecutionRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close execution repository: %v", err)
		}
	})

	err = repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}
}

func cliExecutionRecord(projectDir, runDir string) chc.ExecutionRecord {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	finishedAt := now.Add(time.Minute)

	return chc.ExecutionRecord{
		ID:         "exec-1",
		ProjectID:  projectID(projectDir),
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(projectDir, ".chc", "tickets", "CHC-1.md"),
		Status:     chc.ExecutionStatusSucceeded,
		Runtime: chc.ExecutionRuntimeRef{
			Image:             "chc-agent-runtime:local",
			DockerfilePath:    filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"),
			ContainerWorkdir:  "/workspace",
			ValidationCommand: "make verify",
		},
		Artifacts: chc.ExecutionArtifacts{
			RunDir:         runDir,
			ResultPath:     filepath.Join(runDir, "result.json"),
			LogPath:        filepath.Join(runDir, "runtime.log"),
			ValidationPath: filepath.Join(runDir, "validation.log"),
		},
		CreatedAt:  now,
		StartedAt:  &now,
		FinishedAt: &finishedAt,
		UpdatedAt:  now.Add(time.Minute),
	}
}

func writeCLIExecutionArtifacts(t *testing.T, record chc.ExecutionRecord) {
	t.Helper()

	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("create run dir: %v", err)
	}

	for _, path := range []string{
		record.Artifacts.ResultPath,
		record.Artifacts.LogPath,
		record.Artifacts.ValidationPath,
		chc.ExecutionReportPath(record),
	} {
		err = os.WriteFile(path, []byte("artifact\n"), 0o644)
		if err != nil {
			t.Fatalf("write artifact %q: %v", path, err)
		}
	}
}

func upsertDaemonState(t *testing.T, devRoot string, state chc.DaemonState) {
	t.Helper()

	repository := newDaemonStateRepositoryForTest(t, devRoot)

	err := repository.UpsertDaemonState(t.Context(), state)
	if err != nil {
		t.Fatalf("UpsertDaemonState returned error: %v", err)
	}
}

func getDaemonState(t *testing.T, devRoot string) chc.DaemonState {
	t.Helper()

	repository := newDaemonStateRepositoryForTest(t, devRoot)

	state, err := repository.GetDaemonState(t.Context())
	if err != nil {
		t.Fatalf("GetDaemonState returned error: %v", err)
	}

	return state
}

func newDaemonStateRepositoryForTest(t *testing.T, devRoot string) chc.DaemonStateRepository {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := chc.NewDaemonStateRepository(t.Context(), env, assetsFS)
	if err != nil {
		t.Fatalf("NewDaemonStateRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close daemon state repository: %v", err)
		}
	})

	return repository
}

func indexTicket(t *testing.T, devRoot, projectDir string, ticket chc.IndexedTicket) {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	index, err := chc.NewTicketIndex(t.Context(), env, assetsFS)
	if err != nil {
		t.Fatalf("NewTicketIndex returned error: %v", err)
	}

	defer func() {
		err := index.Close()
		if err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	}()

	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	ticket.ProjectID = projectID(projectDir)
	ticket.FileMTime = now
	ticket.FileSize = 100
	ticket.ContentHash = "hash"
	ticket.IndexedAt = now

	err = index.UpsertTicket(t.Context(), ticket)
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}
}

func listProjectTickets(t *testing.T, devRoot, projectDir string) []chc.IndexedTicket {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	index, err := chc.NewTicketIndex(t.Context(), env, assetsFS)
	if err != nil {
		t.Fatalf("NewTicketIndex returned error: %v", err)
	}

	defer func() {
		err := index.Close()
		if err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	}()

	tickets, err := index.ListProjectTickets(t.Context(), projectID(projectDir))
	if err != nil {
		t.Fatalf("ListProjectTickets returned error: %v", err)
	}

	return tickets
}

func listTicketExecutions(t *testing.T, devRoot, projectID, ticketKey string) []chc.ExecutionRecord {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := chc.NewExecutionRepository(t.Context(), env, assetsFS)
	if err != nil {
		t.Fatalf("NewExecutionRepository returned error: %v", err)
	}

	defer func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close execution repository: %v", err)
		}
	}()

	records, err := repository.ListTicketExecutions(t.Context(), projectID, ticketKey)
	if err != nil {
		t.Fatalf("ListTicketExecutions returned error: %v", err)
	}

	return records
}

func listProjectLoopRuns(t *testing.T, devRoot, projectID string) []chc.LoopRunRecord {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := chc.NewLoopRunRepository(t.Context(), env, assetsFS)
	if err != nil {
		t.Fatalf("NewLoopRunRepository returned error: %v", err)
	}

	defer func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close loop run repository: %v", err)
		}
	}()

	records, err := repository.ListProjectLoopRuns(t.Context(), projectID)
	if err != nil {
		t.Fatalf("ListProjectLoopRuns returned error: %v", err)
	}

	return records
}

func createApplyResultExecution(
	t *testing.T,
	devRoot string,
	projectDir string,
	status chc.ExecutionStatus,
	failureReason string,
) chc.ExecutionRecord {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := chc.NewExecutionRepository(t.Context(), env, assetsFS)
	if err != nil {
		t.Fatalf("NewExecutionRepository returned error: %v", err)
	}

	defer func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close execution repository: %v", err)
		}
	}()

	record := testApplyResultExecutionRecord(projectDir, status, failureReason)

	err = repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}

	return record
}

func testApplyResultExecutionRecord(
	projectDir string,
	status chc.ExecutionStatus,
	failureReason string,
) chc.ExecutionRecord {
	runDir := filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1", "exec-apply")
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	return chc.ExecutionRecord{
		ID:            "exec-apply",
		ProjectID:     projectID(projectDir),
		TicketKey:     "CHC-1",
		TicketPath:    filepath.Join(projectDir, ".chc", "tickets", "CHC-1-Do-something.md"),
		Status:        status,
		FailureReason: failureReason,
		Runtime: chc.ExecutionRuntimeRef{
			Image:             "chc-runtime:latest",
			DockerfilePath:    filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"),
			ContainerWorkdir:  "/workspace",
			ValidationCommand: "make verify",
		},
		Artifacts: chc.ExecutionArtifacts{
			RunDir:         runDir,
			ResultPath:     filepath.Join(runDir, "result.json"),
			LogPath:        filepath.Join(runDir, "runtime.log"),
			ValidationPath: filepath.Join(runDir, "validation.log"),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func writeApplyResultFile(
	t *testing.T,
	record chc.ExecutionRecord,
	status chc.ExecutionResultStatus,
	summary string,
) {
	t.Helper()

	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("create execution run dir: %v", err)
	}

	content := fmt.Sprintf(
		`{"execution_id":%q,"ticket_key":%q,"status":%q,"summary":%q}`,
		record.ID,
		record.TicketKey,
		status,
		summary,
	)

	err = os.WriteFile(record.Artifacts.ResultPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write result file: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	return string(content)
}

func testTicket() chc.Ticket {
	return chc.Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Status:      "open",
		Priority:    "high",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "onboarding",
		ProjectArea: "uxui",
		CreatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	}
}

type fakeTicketStartHarnessExecutor struct {
	commands       []string
	err            error
	failWorker     bool
	resultStatuses map[string]chc.ExecutionResultStatus
}

func (e *fakeTicketStartHarnessExecutor) Run(
	_ context.Context,
	command chc.HarnessCommand,
	stdout io.Writer,
	stderr io.Writer,
) error {
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

	if e.failWorker && isFakeTicketStartWorkerCommand(command) {
		return chc.NewExecutionFailure(chc.ExecutionFailureHarness, fmt.Errorf("codex failed"))
	}

	err = e.writeFakeTicketStartResult(command)
	if err != nil {
		return err
	}

	return e.err
}

func isFakeTicketStartWorkerCommand(command chc.HarnessCommand) bool {
	return command.Args[0] == "run" && harnessEnvValue(command, "CHC_RESULT_PATH") != ""
}

func (e *fakeTicketStartHarnessExecutor) writeFakeTicketStartResult(command chc.HarnessCommand) error {
	if command.Args[0] != "run" {
		return nil
	}

	resultPath := harnessEnvValue(command, "CHC_RESULT_PATH")
	if resultPath == "" {
		return nil
	}

	status := chc.ExecutionResultStatusSucceeded
	ticketKey := harnessEnvValue(command, "CHC_TICKET_KEY")

	configured, ok := e.resultStatuses[ticketKey]
	if ok {
		status = configured
	}

	content := `{"execution_id":"` + harnessEnvValue(command, "CHC_EXECUTION_ID") +
		`","ticket_key":"` + ticketKey +
		`","status":"` + string(status) + `","summary":"done"}`

	return os.WriteFile(containerPathToProjectPath(command, resultPath), []byte(content), 0o644)
}

func harnessEnvValue(command chc.HarnessCommand, name string) string {
	prefix := name + "="
	for _, arg := range command.Args {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}

	return ""
}

func containerPathToProjectPath(command chc.HarnessCommand, path string) string {
	for index, arg := range command.Args {
		if arg == "--volume" && index+1 < len(command.Args) {
			parts := strings.SplitN(command.Args[index+1], ":", 2)
			if len(parts) == 2 && strings.HasPrefix(path, parts[1]) {
				return filepath.Join(parts[0], strings.TrimPrefix(path, parts[1]))
			}
		}
	}

	return path
}

func installFakeTicketStartHarnessExecutor(t *testing.T, err error) *fakeTicketStartHarnessExecutor {
	t.Helper()

	executor := &fakeTicketStartHarnessExecutor{err: err}
	previous := newTicketStartHarnessExecutor
	newTicketStartHarnessExecutor = func() chc.HarnessExecutor {
		return executor
	}

	t.Cleanup(func() {
		newTicketStartHarnessExecutor = previous
	})

	return executor
}

func writeTicketMarkdownOnly(t *testing.T, projectDir, relativePath string, ticket chc.Ticket) string {
	t.Helper()

	ticketPath := filepath.Join(projectDir, ".chc", "tickets", filepath.FromSlash(relativePath))

	content, err := chc.NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(ticketPath), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(ticketPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return ticketPath
}

func newTestWorkspace(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()

	err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755)
	if err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	return projectDir
}

func writeCLIWorkspaceConfig(t *testing.T, projectDir, content string) {
	t.Helper()

	configPath := filepath.Join(projectDir, ".chc", "config.yml")

	err := os.WriteFile(configPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write workspace config: %v", err)
	}
}

func writeCLIReport(t *testing.T, projectDir, relativePath, content string) {
	t.Helper()

	path := filepath.Join(projectDir, relativePath)

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create report directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write report: %v", err)
	}
}

func authorSessionIDFromOutput(t *testing.T, output string) string {
	t.Helper()

	for _, line := range strings.Split(output, "\n") {
		value, ok := strings.CutPrefix(line, "session: created ")
		if ok {
			return value
		}

		value, ok = strings.CutPrefix(line, "id: ")
		if ok {
			return value
		}

		value, ok = strings.CutPrefix(line, "session: ")
		if ok && (strings.HasPrefix(value, "auth-") || strings.HasPrefix(value, "author-")) {
			return value
		}
	}

	t.Fatalf("output does not contain session id:\n%s", output)

	return ""
}

func createAuthorQuizForTest(t *testing.T, projectDir string) (string, string) {
	t.Helper()

	installFakeAuthoringQuizProvider(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"quiz",
		"new",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author quiz new returned error: %v", err)
	}

	quizPath := authorQuizPathFromOutput(t, stdout.String())
	if !filepath.IsAbs(quizPath) {
		quizPath = filepath.Join(projectDir, quizPath)
	}

	return authorSessionIDFromOutput(t, stdout.String()), quizPath
}

func installFakeAuthoringQuizProvider(t *testing.T) {
	t.Helper()

	previous := newAuthoringQuizQuestionProvider
	newAuthoringQuizQuestionProvider = func() chc.AuthoringQuizQuestionProvider {
		return &chc.FakeAuthoringQuizQuestionProvider{
			Result: chc.AuthoringQuizGenerationResult{
				Provider: "fake",
				Strategy: chc.AuthoringQuizGenerationStrategyDeep,
				Questions: []chc.AuthoringQuizQuestion{
					{
						ID:                    "Q1",
						Text:                  "Which user workflow should the specification pin down first?",
						Tags:                  []string{"workflow", "specification"},
						FollowUpDepth:         1,
						CompletionRecommended: false,
					},
					{
						ID:                    "Q2",
						Text:                  "What project files or docs should be treated as authoritative context?",
						Tags:                  []string{"context", "documents"},
						FollowUpDepth:         1,
						CompletionRecommended: false,
					},
					{
						ID:                    "Q3",
						Text:                  "What validation evidence should prove the generated spec is useful?",
						Tags:                  []string{"validation", "evidence"},
						FollowUpDepth:         0,
						CompletionRecommended: true,
					},
				},
			},
		}
	}

	t.Cleanup(func() {
		newAuthoringQuizQuestionProvider = previous
	})
}

type authorAssistedQuestionOutput struct {
	TurnID     string
	TurnIndex  int
	QuestionID string
	Question   string
}

func createAuthorAssistedSessionForTest(t *testing.T, projectDir string) string {
	t.Helper()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"assisted",
		"start",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted start returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "assisted: started") {
		t.Fatalf("start output missing assisted marker:\n%s", stdout.String())
	}

	return authorSessionIDFromOutput(t, stdout.String())
}

func askAuthorAssistedQuestionForTest(
	t *testing.T,
	projectDir string,
	sessionID string,
) authorAssistedQuestionOutput {
	t.Helper()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"assisted",
		"resume",
		"--dir",
		projectDir,
		"--session",
		sessionID,
		"--fake-question",
		"What should the billing sync spec include?",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted resume returned error: %v", err)
	}

	return authorAssistedQuestionFromOutput(t, stdout.String())
}

func answerAuthorAssistedQuestionForTest(
	t *testing.T,
	projectDir string,
	sessionID string,
	question authorAssistedQuestionOutput,
) {
	t.Helper()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"assisted",
		"answer",
		"--dir",
		projectDir,
		"--session",
		sessionID,
		"--turn",
		question.TurnID,
		"--turn-index",
		fmt.Sprintf("%d", question.TurnIndex),
		"--question-id",
		question.QuestionID,
		"--question",
		question.Question,
		"--answer",
		"It should include retry behavior.",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted answer returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "assisted: answer") {
		t.Fatalf("answer output missing assisted marker:\n%s", stdout.String())
	}
}

func authorAssistedQuestionFromOutput(t *testing.T, output string) authorAssistedQuestionOutput {
	t.Helper()

	fields := authorOutputFields(output)

	turnIndex, err := parseAuthorOutputInt(fields["turn_index"])
	if err != nil {
		t.Fatalf("parse turn_index %q: %v", fields["turn_index"], err)
	}

	parsed := authorAssistedQuestionOutput{
		TurnID:     fields["turn"],
		TurnIndex:  turnIndex,
		QuestionID: fields["question_id"],
		Question:   fields["question"],
	}

	if parsed.TurnID == "" || parsed.TurnIndex == 0 || parsed.QuestionID == "" || parsed.Question == "" {
		t.Fatalf("invalid assisted question output:\n%s", output)
	}

	return parsed
}

func authorOutputFields(output string) map[string]string {
	fields := map[string]string{}

	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(line, ": ")
		if ok {
			fields[key] = value
		}
	}

	return fields
}

func parseAuthorOutputInt(value string) (int, error) {
	var parsed int

	_, err := fmt.Sscanf(value, "%d", &parsed)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}

func authorQuizPathFromOutput(t *testing.T, output string) string {
	t.Helper()

	var path string

	for _, line := range strings.Split(output, "\n") {
		value, ok := strings.CutPrefix(line, "path: ")
		if ok && strings.Contains(value, ".chc/quizzes/") {
			path = value
		}
	}

	if path != "" {
		return path
	}

	t.Fatalf("output does not contain quiz path:\n%s", output)

	return ""
}

func writeRuntimeTemplate(t *testing.T, projectDir string) {
	t.Helper()

	dockerfilePath := filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile")

	err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755)
	if err != nil {
		t.Fatalf("create runtime template directory: %v", err)
	}

	err = os.WriteFile(dockerfilePath, []byte("FROM ubuntu:24.04\n"), 0o644)
	if err != nil {
		t.Fatalf("write runtime template: %v", err)
	}

	err = os.MkdirAll(filepath.Join(projectDir, ".chc", "runtime", "runs"), 0o755)
	if err != nil {
		t.Fatalf("create runtime runs directory: %v", err)
	}
}

func listNotificationEventsForTest(
	t *testing.T,
	devRoot string,
	projectID string,
) []chc.NotificationAuditRecord {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := chc.NewNotificationEventRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewNotificationEventRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close notification repository: %v", err)
		}
	})

	records, err := repository.ListProjectNotificationEvents(t.Context(), projectID)
	if err != nil {
		t.Fatalf("ListProjectNotificationEvents returned error: %v", err)
	}

	return records
}

func listRemoteCommandRecordsForTest(t *testing.T, devRoot string) []chc.RemoteCommandRecord {
	t.Helper()

	env, err := chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := chc.NewRemoteCommandRecordRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewRemoteCommandRecordRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close remote command repository: %v", err)
		}
	})

	records, err := repository.ListRemoteCommandRecords(t.Context())
	if err != nil {
		t.Fatalf("ListRemoteCommandRecords returned error: %v", err)
	}

	return records
}

func assertNotificationEvent(
	t *testing.T,
	records []chc.NotificationAuditRecord,
	index int,
	eventType chc.NotificationEventType,
	status chc.NotificationStatus,
) {
	t.Helper()

	if len(records) <= index {
		t.Fatalf("notification records = %d, want index %d", len(records), index)
	}

	record := records[index]
	if record.Event.Type != eventType {
		t.Fatalf("notification type = %q, want %q", record.Event.Type, eventType)
	}

	if record.Event.Status != status {
		t.Fatalf("notification status = %q, want %q", record.Event.Status, status)
	}

	if record.Event.ProjectID == "" {
		t.Fatal("notification project id is empty")
	}
}

func chcTestUserEnvironment(t *testing.T) (chc.UserEnvironment, error) {
	t.Helper()

	return chc.NewUserEnvironment(chc.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: t.TempDir(),
	})
}

func runInDirectory(t *testing.T, dir string, runFunc func()) {
	t.Helper()

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	t.Cleanup(func() {
		err := os.Chdir(previousDir)
		if err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	runFunc()
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat path %s: %v", path, err)
	}
}

func assertPathNotExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatalf("path %s exists or returned unexpected stat error: %v", path, err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}

	return string(content)
}
