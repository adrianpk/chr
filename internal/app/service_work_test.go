package app

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chc/internal/workspace"
)

func TestServiceStartAgentWorkTicketReusesTicketExecution(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkTicket("CHC-1")
	if err != nil {
		t.Fatalf("StartAgentWorkTicket returned error: %v", err)
	}

	if result.Execution == nil || result.Execution.TicketKey != "CHC-1" {
		t.Fatalf("Execution = %#v, want CHC-1 execution", result.Execution)
	}

	if !strings.Contains(result.Message, "Execution "+result.Execution.ID+" for ticket CHC-1 finished") {
		t.Fatalf("Message = %q, want execution summary", result.Message)
	}

	if !strings.Contains(result.Message, "Result applied to ticket") {
		t.Fatalf("Message = %q, want result application summary", result.Message)
	}

	if len(result.Tickets) != 1 || result.Tickets[0].Status != "reviewing" {
		t.Fatalf("Tickets = %#v, want reviewing ticket after result application", result.Tickets)
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 1 {
		t.Fatalf("execution records = %d, want 1", len(records))
	}

	loopRuns := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loopRuns) != 0 {
		t.Fatalf("loop runs = %d, want 0", len(loopRuns))
	}
}

func TestServiceStartAgentWorkTicketInitializesRuntime(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	_, err := operations.StartAgentWorkTicket("CHC-1")
	if err != nil {
		t.Fatalf("StartAgentWorkTicket returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"))
	assertPathExists(t, filepath.Join(projectDir, ".chc", "user", "runtime", "runs"))
}

func TestServiceStartAllReadyAgentWorkReusesLoopExecution(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-2", "Ready backend", "ready_for_dev", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAllReadyAgentWork()
	if err != nil {
		t.Fatalf("StartAllReadyAgentWork returned error: %v", err)
	}

	if result.LoopRun == nil {
		t.Fatal("LoopRun is nil")
	}

	loopMatches := result.LoopRun.Status == workspace.LoopRunStatusSucceeded &&
		result.LoopRun.EligibleCount == 2 &&
		result.LoopRun.ProcessedCount == 2
	if !loopMatches {
		t.Fatalf("LoopRun = %#v, want succeeded loop with two processed tickets", result.LoopRun)
	}

	if gotKeys := indexedAgentWorkTicketKeys(result.SelectedQueue); strings.Join(gotKeys, ",") != "CHC-1,CHC-2" {
		t.Fatalf("SelectedQueue keys = %#v, want oldest-first CHC-1, CHC-2", gotKeys)
	}

	for _, ticketKey := range []string{"CHC-1", "CHC-2"} {
		records := listTicketExecutions(t, devRoot, projectID(projectDir), ticketKey)
		if len(records) != 1 {
			t.Fatalf("%s execution records = %d, want 1", ticketKey, len(records))
		}
	}

	loopRuns := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loopRuns) != 1 {
		t.Fatalf("loop runs = %d, want 1", len(loopRuns))
	}
}

func TestServicePreviewAllReadyAgentWorkDoesNotExecute(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.PreviewAllReadyAgentWork()
	if err != nil {
		t.Fatalf("PreviewAllReadyAgentWork returned error: %v", err)
	}

	if result.EligibleCount != 1 || len(result.SelectedQueue) != 1 || result.SelectedQueue[0].Key != "CHC-1" {
		t.Fatalf("preview result = %#v, want one CHC-1 queued ticket", result)
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 0 {
		t.Fatalf("execution records = %d, want 0", len(records))
	}

	loopRuns := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loopRuns) != 0 {
		t.Fatalf("loop runs = %d, want 0", len(loopRuns))
	}
}

func TestServicePreviewAgentWorkLoopAppliesSelection(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-1", "Ready normal", "ready_for_dev", "normal", "normal", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-2", "Ready high", "ready_for_dev", "high", "critical", "bug", "maintenance", "backend", []string{"api"},
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.PreviewAgentWorkLoop(AgentWorkSelectionOptions{
		Filters: workspace.TicketLoopFilters{
			Priorities: []string{"high"},
		},
		Order: workspace.TicketLoopOrderPriority,
	})
	if err != nil {
		t.Fatalf("PreviewAgentWorkLoop returned error: %v", err)
	}

	if result.EligibleCount != 1 || len(result.SelectedQueue) != 1 || result.SelectedQueue[0].Key != "CHC-2" {
		t.Fatalf("preview result = %#v, want one high-priority CHC-2 queued ticket", result)
	}
}

func newAgentWorkTestService(t *testing.T, devRoot, projectDir string) service {
	t.Helper()

	env, err := workspace.NewUserEnvironment(workspace.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	return newService(t.Context(), env, commandProjectContext{
		ProjectDir: projectDir,
		ProjectID:  projectID(projectDir),
	}, testAssetsFS())
}

func indexedAgentWorkTicketKeys(tickets []workspace.IndexedTicket) []string {
	keys := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		keys = append(keys, ticket.Key)
	}

	return keys
}
