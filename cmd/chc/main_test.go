package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

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
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
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
		"started_at: 2026-05-20T12:00:00Z",
		"heartbeat_at: 2026-05-20T12:01:00Z",
		"updated_at: 2026-05-20T12:01:00Z",
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
	startedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
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
		"stop_requested_at: 2026-05-20T12:01:00Z",
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
		"data_dir: " + filepath.Join(devRoot, "data", "chc"),
		"state_dir: " + filepath.Join(devRoot, "state", "chc"),
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

func TestDaemonStartAndStopRequireDryRun(t *testing.T) {
	devRoot := t.TempDir()

	for _, args := range [][]string{
		{"daemon", "start", "--dev-root", devRoot},
		{"daemon", "stop", "--dev-root", devRoot},
	} {
		err := run(t.Context(), args, &bytes.Buffer{}, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("run(%v) returned nil error", args)
		}

		if !strings.Contains(err.Error(), "requires --dry-run") {
			t.Fatalf("run(%v) error = %v, want dry-run requirement", args, err)
		}
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
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	if err := os.WriteFile(configPath, []byte("tickets:\n  key_prefix: PRJ\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := run(t.Context(), []string{"ticket", "new", "--dir", dir, "Do something"}, &bytes.Buffer{}, &bytes.Buffer{})
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
	if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755); err != nil {
		t.Fatalf("create runtime template directory: %v", err)
	}

	if err := os.WriteFile(dockerfilePath, []byte("FROM custom\n"), 0o644); err != nil {
		t.Fatalf("write runtime template: %v", err)
	}

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
		"dockerfile: " + filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"),
		"workdir: /workspace",
		"artifact_dir: " + filepath.Join(projectDir, ".chc", "runtime", "runs"),
		"mount: " + projectDir + " -> /workspace read_only=false",
		"check: ok workspace " + filepath.Join(projectDir, ".chc") + " ok",
		"check: ok dockerfile " + filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile") + " ok",
		"check: ok artifact_dir " + filepath.Join(projectDir, ".chc", "runtime", "runs") + " ok",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestRuntimeValidateFailsWhenRuntimeTemplateIsMissing(t *testing.T) {
	projectDir := newTestWorkspace(t)
	if err := os.MkdirAll(filepath.Join(projectDir, ".chc", "runtime", "runs"), 0o755); err != nil {
		t.Fatalf("create runtime runs dir: %v", err)
	}

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"validate",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("runtime validate returned nil error")
	}

	if !strings.Contains(stdout.String(), "check: fail dockerfile") {
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
		"run_artifact_dir: " + filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1", "exec-"),
		"result_path: " + filepath.Join(projectDir, ".chc", "runtime", "runs", "CHC-1", "exec-"),
		"check: ok dockerfile " + filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile") + " ok",
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

	if err := os.MkdirAll(filepath.Join(projectDir, ".chc", "runtime", "runs"), 0o755); err != nil {
		t.Fatalf("create runtime runs dir: %v", err)
	}

	addProject(t, devRoot, "example", projectDir)
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

	if !strings.Contains(stdout.String(), "check: fail dockerfile") {
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

	tickets := listProjectTickets(t, devRoot, projectDir)
	if len(tickets) != 1 || tickets[0].Status != "reviewing" {
		t.Fatalf("indexed tickets = %#v, want reviewing CHC-1", tickets)
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

	if err := os.MkdirAll(filepath.Dir(ticketPath), 0o755); err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	if err := os.WriteFile(ticketPath, []byte(content), 0o644); err != nil {
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
	if err := os.MkdirAll(filepath.Dir(badTicketPath), 0o755); err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	if err := os.WriteFile(badTicketPath, []byte("# Bad\n"), 0o644); err != nil {
		t.Fatalf("write bad ticket: %v", err)
	}

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

	if err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
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

	if err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
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
	if err := os.MkdirAll(filepath.Dir(badTicketPath), 0o755); err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	if err := os.WriteFile(badTicketPath, []byte("# Bad\n"), 0o644); err != nil {
		t.Fatalf("write bad ticket: %v", err)
	}

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
		"invalid: 1",
		"invalid_ticket: CHC-1-Bad.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestProjectAddRegistersProject(t *testing.T) {
	projectDir := t.TempDir()
	devRoot := t.TempDir()

	if err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755); err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
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

	if err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755); err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	err := run(t.Context(), []string{
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

	if err := os.RemoveAll(projectDir); err != nil {
		t.Fatalf("remove project dir: %v", err)
	}

	err := run(t.Context(), []string{
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

	if err := os.RemoveAll(filepath.Join(projectDir, ".chc")); err != nil {
		t.Fatalf("remove project workspace: %v", err)
	}

	err := run(t.Context(), []string{
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

	if err := os.RemoveAll(projectDir); err != nil {
		t.Fatalf("remove project dir: %v", err)
	}

	err := run(t.Context(), []string{
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

	if err := os.RemoveAll(filepath.Join(projectDir, ".chc")); err != nil {
		t.Fatalf("remove project workspace: %v", err)
	}

	err := run(t.Context(), []string{
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

func upsertDaemonState(t *testing.T, devRoot string, state chc.DaemonState) {
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

	defer func() {
		if err := repository.Close(); err != nil {
			t.Fatalf("close daemon state repository: %v", err)
		}
	}()

	if err := repository.UpsertDaemonState(t.Context(), state); err != nil {
		t.Fatalf("UpsertDaemonState returned error: %v", err)
	}
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
		if err := index.Close(); err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	}()

	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	ticket.ProjectID = projectID(projectDir)
	ticket.FileMTime = now
	ticket.FileSize = 100
	ticket.ContentHash = "hash"
	ticket.IndexedAt = now

	if err := index.UpsertTicket(t.Context(), ticket); err != nil {
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
		if err := index.Close(); err != nil {
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
		if err := repository.Close(); err != nil {
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
		if err := repository.Close(); err != nil {
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
		if err := repository.Close(); err != nil {
			t.Fatalf("close execution repository: %v", err)
		}
	}()

	record := testApplyResultExecutionRecord(projectDir, status, failureReason)
	if err := repository.CreateExecution(t.Context(), record); err != nil {
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

	if err := os.MkdirAll(record.Artifacts.RunDir, 0o755); err != nil {
		t.Fatalf("create execution run dir: %v", err)
	}

	content := fmt.Sprintf(
		`{"execution_id":%q,"ticket_key":%q,"status":%q,"summary":%q}`,
		record.ID,
		record.TicketKey,
		status,
		summary,
	)

	if err := os.WriteFile(record.Artifacts.ResultPath, []byte(content), 0o644); err != nil {
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

	if _, err := fmt.Fprintln(stdout, "stdout: "+name); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(stderr, "stderr: "+name); err != nil {
		return err
	}

	if e.failWorker && isFakeTicketStartWorkerCommand(command) {
		return chc.NewExecutionFailure(chc.ExecutionFailureHarness, fmt.Errorf("codex failed"))
	}

	if err := e.writeFakeTicketStartResult(command); err != nil {
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

	if configured, ok := e.resultStatuses[ticketKey]; ok {
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

	if err := os.MkdirAll(filepath.Dir(ticketPath), 0o755); err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	if err := os.WriteFile(ticketPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return ticketPath
}

func newTestWorkspace(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()

	if err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755); err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	return projectDir
}

func writeRuntimeTemplate(t *testing.T, projectDir string) {
	t.Helper()

	dockerfilePath := filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile")
	if err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755); err != nil {
		t.Fatalf("create runtime template directory: %v", err)
	}

	if err := os.WriteFile(dockerfilePath, []byte("FROM ubuntu:24.04\n"), 0o644); err != nil {
		t.Fatalf("write runtime template: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(projectDir, ".chc", "runtime", "runs"), 0o755); err != nil {
		t.Fatalf("create runtime runs directory: %v", err)
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

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	runFunc()
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat path %s: %v", path, err)
	}
}

func assertPathNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("path %s exists or returned unexpected stat error: %v", path, err)
	}
}
