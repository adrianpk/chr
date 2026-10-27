package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

func TestServiceStartAgentWorkTicketReusesTicketExecution(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", testLoopTicket(
		"CHR-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkTicket("CHR-1")
	if err != nil {
		t.Fatalf("StartAgentWorkTicket returned error: %v", err)
	}

	if result.Execution == nil || result.Execution.TicketKey != "CHR-1" {
		t.Fatalf("Execution = %#v, want CHR-1 execution", result.Execution)
	}

	if !strings.Contains(result.Message, "Execution "+result.Execution.ID+" for ticket CHR-1 finished") {
		t.Fatalf("Message = %q, want execution summary", result.Message)
	}

	if !strings.Contains(result.Message, "Result applied to ticket") {
		t.Fatalf("Message = %q, want result application summary", result.Message)
	}

	if len(result.Tickets) != 1 || result.Tickets[0].Status != "reviewing" {
		t.Fatalf("Tickets = %#v, want reviewing ticket after result application", result.Tickets)
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHR-1")
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

	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", testLoopTicket(
		"CHR-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkTicket("CHR-1")
	if err != nil {
		t.Fatalf("StartAgentWorkTicket returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chr", "default", "runtime", "agent", "Dockerfile"))

	if result.Execution == nil {
		t.Fatal("Execution is nil")
	}

	assertPathExists(t, filepath.Join(result.Execution.Artifacts.RunDir, "artifact.md"))
}

func TestServiceStartAgentWorkTicketPreparesWorktreeWorkspace(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	preparer := installFakeTicketStartWorktreePreparer(t)

	writeRuntimeTemplate(t, projectDir)
	writeCLIWorkspaceConfig(t, projectDir, `work:
  workspace_mode: worktree
`)

	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", testLoopTicket(
		"CHR-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkTicket("CHR-1")
	if err != nil {
		t.Fatalf("StartAgentWorkTicket returned error: %v", err)
	}

	if len(preparer.results) != 1 {
		t.Fatalf("prepared worktrees = %d, want 1", len(preparer.results))
	}

	if result.Execution == nil {
		t.Fatal("Execution is nil")
	}

	assertPreparedTicketWorktreeWorkspace(t, ticketPath, preparer.results[0], *result.Execution)
}

func assertPreparedTicketWorktreeWorkspace(t *testing.T, ticketPath string, prepared executioncore.WorktreePrepareResult, execution core.ExecutionRecord) {
	t.Helper()

	if prepared.Plan.Branch != "chr/chr-1-ready-frontend" {
		t.Fatalf("prepared branch = %q, want derived branch", prepared.Plan.Branch)
	}

	if execution.Workspace.Mode != workspacecore.ExecutionWorktree {
		t.Fatalf("Workspace.Mode = %q, want worktree", execution.Workspace.Mode)
	}

	if execution.Workspace.ExecutionWorkspaceDir != prepared.Plan.WorktreePath {
		t.Fatalf("ExecutionWorkspaceDir = %q, want %q", execution.Workspace.ExecutionWorkspaceDir, prepared.Plan.WorktreePath)
	}

	if execution.Workspace.TicketBranch != prepared.Plan.Branch {
		t.Fatalf("TicketBranch = %q, want %q", execution.Workspace.TicketBranch, prepared.Plan.Branch)
	}

	if !strings.HasPrefix(execution.TicketPath, prepared.Plan.WorktreePath) {
		t.Fatalf("TicketPath = %q, want path inside worktree %q", execution.TicketPath, prepared.Plan.WorktreePath)
	}

	projectTicket := readFile(t, ticketPath)
	if strings.Contains(projectTicket, `branch: "chr/chr-1-ready-frontend"`) {
		t.Fatalf("project checkout ticket was mutated:\n%s", projectTicket)
	}

	worktreeTicket := readFile(t, execution.TicketPath)
	for _, want := range []string{
		`branch: "chr/chr-1-ready-frontend"`,
		`status: "reviewing"`,
	} {
		if !strings.Contains(worktreeTicket, want) {
			t.Fatalf("worktree ticket missing %q:\n%s", want, worktreeTicket)
		}
	}

	metadataPath, err := executioncore.WorktreeMetadataPath(prepared.Plan)
	if err != nil {
		t.Fatalf("WorktreeMetadataPath returned error: %v", err)
	}

	assertPathExists(t, metadataPath)
}

func TestServiceStartAgentWorkLoopPreparesOnlySelectedWorktree(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	preparer := installFakeTicketStartWorktreePreparer(t)

	writeRuntimeTemplate(t, projectDir)
	writeCLIWorkspaceConfig(t, projectDir, `work:
  workspace_mode: worktree
`)
	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHR-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))
	writeTicketMarkdownOnly(t, projectDir, "CHR-2-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHR-2", "Ready backend", "ready_for_dev", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkLoop(t.Context(), AgentWorkSelectionOptions{
		Statuses:      []string{ticketcore.StatusReadyForDev},
		Order:         workflowstate.TicketOrderCreatedAt,
		MaxIterations: 1,
	})
	if err != nil {
		t.Fatalf("StartAgentWorkLoop returned error: %v", err)
	}

	if result.EligibleCount != 2 {
		t.Fatalf("EligibleCount = %d, want 2", result.EligibleCount)
	}

	if len(result.SelectedQueue) != 1 || result.SelectedQueue[0].Key != "CHR-1" {
		t.Fatalf("SelectedQueue = %#v, want only CHR-1 selected", result.SelectedQueue)
	}

	if len(preparer.results) != 1 {
		t.Fatalf("prepared worktrees = %d, want only selected ticket", len(preparer.results))
	}

	if preparer.results[0].Plan.TicketKey != "CHR-1" {
		t.Fatalf("prepared ticket = %q, want CHR-1", preparer.results[0].Plan.TicketKey)
	}

	plan, err := executioncore.NewWorktreePlan(executioncore.WorktreePlanOptions{
		ProjectDir: projectDir,
		TicketKey:  "CHR-2",
		Branch:     "chr/chr-2-ready-backend",
	})
	if err != nil {
		t.Fatalf("NewWorktreePlan returned error: %v", err)
	}

	assertPathNotExists(t, plan.WorktreePath)
}

func TestServiceStartAgentWorkTicketReusesReviewRepairBranch(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	preparer := installFakeTicketStartWorktreePreparer(t)

	writeRuntimeTemplate(t, projectDir)
	writeCLIWorkspaceConfig(t, projectDir, `work:
  workspace_mode: worktree
`)
	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Repair.md", testReviewRepairTicket())

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkTicket("CHR-1")
	if err != nil {
		t.Fatalf("StartAgentWorkTicket returned error: %v", err)
	}

	if len(preparer.results) != 1 {
		t.Fatalf("prepared worktrees = %d, want 1", len(preparer.results))
	}

	if preparer.results[0].Plan.Branch != "feat/chr-1-review-repair" {
		t.Fatalf("prepared branch = %q, want existing repair branch", preparer.results[0].Plan.Branch)
	}

	if result.Execution == nil || result.Execution.Workspace.TicketBranch != "feat/chr-1-review-repair" {
		t.Fatalf("Execution = %#v, want repair branch workspace", result.Execution)
	}
}

func TestServiceStartAllReadyAgentWorkStopsBeforeSecondTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHR-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))
	writeTicketMarkdownOnly(t, projectDir, "CHR-2-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHR-2", "Ready backend", "ready_for_dev", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAllReadyAgentWork()
	if err != nil {
		t.Fatalf("StartAllReadyAgentWork returned error: %v", err)
	}

	if result.LoopRun == nil {
		t.Fatal("LoopRun is nil")
	}

	assertAgentWorkLoopRunBlockedAfterOneTicket(t, *result.LoopRun)

	if gotKeys := indexedAgentWorkTicketKeys(result.SelectedQueue); strings.Join(gotKeys, ",") != "CHR-1" {
		t.Fatalf("SelectedQueue keys = %#v, want oldest-first CHR-1 only", gotKeys)
	}

	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-1", 1)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-2", 0)

	loopRuns := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loopRuns) != 1 {
		t.Fatalf("loop runs = %d, want 1", len(loopRuns))
	}
}

func TestServiceStartAllReadyAgentWorkRunsAgenticReviewPass(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	runner := installFakeAgenticReviewRunner(t, core.AgenticReviewResult{
		Decision: core.IntegrationReviewDecisionPass,
		Summary:  "Implementation is correct and validation evidence is sufficient.",
		NonBlockingNotes: []workflowstate.ReviewFinding{{
			Category: core.AgenticReviewFindingCosmetic,
			Title:    "Optional wording cleanup",
			Detail:   "One help sentence could be clearer, but behavior is correct.",
		}},
	})

	writeRuntimeTemplate(t, projectDir)
	writeAgenticReviewLoopConfig(t, projectDir)

	runnerRemote, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)
	defer restoreRunner()

	mergeCommand := "tea api -X PUT /repos/adrianpk/chr/pulls/274/merge -d @-"
	runnerRemote.inputOutputs[mergeCommand] = []byte(`{"sha":"abc123"}`)

	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(
		testLoopTicketWithPullRequestRef("CHR-1", "Ready frontend"),
		time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	))
	writeTicketMarkdownOnly(t, projectDir, "CHR-2-Ready.md", withLoopTicketCreatedAt(testLoopTicketWithPullRequestRef(
		"CHR-2", "Ready backend",
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkLoop(t.Context(), AgentWorkSelectionOptions{MaxIterations: 2})
	if err != nil {
		t.Fatalf("StartAllReadyAgentWork returned error: %v", err)
	}

	assertAgentWorkLoopRunProcessedTwoTickets(t, *result.LoopRun)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-1", 1)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-2", 1)
	assertAgenticReviewRunnerCalls(t, runner, projectDir, []string{"CHR-1", "CHR-2"})
	assertAgentWorkTicketState(t, result.Tickets, "CHR-1", ticketcore.StatusDone)
	assertAgentWorkTicketState(t, result.Tickets, "CHR-2", ticketcore.StatusDone)
	assertAgenticReviewRunArtifact(t, projectDir, "CHR-1", result.Executions[0].ID)
	assertTicketContains(t, ticketPath, `review_state: "pass"`)
	assertTicketContains(t, ticketPath, `review_result: "pass"`)
	assertTicketContains(t, ticketPath, `merge_state: "merged"`)
	assertTicketContains(t, ticketPath, `merge_result: "merged"`)
	assertTicketContains(t, ticketPath, "Agentic review note (cosmetic): Optional wording cleanup")

	if len(result.Merges) != 2 {
		t.Fatalf("merges = %d, want 2", len(result.Merges))
	}
}

func TestServiceStartAllReadyAgentWorkStopsWhenMergePreflightBlocks(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)
	installFakeAgenticReviewRunner(t, core.AgenticReviewResult{
		Decision: core.IntegrationReviewDecisionPass,
		Summary:  "Implementation is correct and validation evidence is sufficient.",
	})

	writeRuntimeTemplate(t, projectDir)
	writeAgenticReviewLoopConfig(t, projectDir)

	runnerRemote, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)
	defer restoreRunner()

	runnerRemote.outputs["tea api /repos/adrianpk/chr/pulls/274"] = []byte(`{"state":"open","base":{"ref":"dev"},"head":{"ref":"feat/slice-2-pr-reference-workflow","sha":"abc123"},"mergeable":false}`)

	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(
		testLoopTicketWithPullRequestRef("CHR-1", "Ready frontend"),
		time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	))
	writeTicketMarkdownOnly(t, projectDir, "CHR-2-Ready.md", withLoopTicketCreatedAt(
		testLoopTicketWithPullRequestRef("CHR-2", "Ready backend"),
		time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC),
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkLoop(t.Context(), AgentWorkSelectionOptions{MaxIterations: 2})
	if err != nil {
		t.Fatalf("StartAgentWorkLoop returned error: %v", err)
	}

	assertAgentWorkLoopRunBlockedAfterOneTicket(t, *result.LoopRun)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-2", 0)
	assertAgentWorkTicketState(t, result.Tickets, "CHR-1", ticketcore.StatusReadyForMerge)
	assertTicketContains(t, ticketPath, `merge_state: "blocked"`)
	assertTicketContains(t, ticketPath, `merge_result: "blocked"`)
	assertTicketContains(t, ticketPath, `merge_blocker: "not_mergeable"`)

	if len(result.Merges) != 1 || result.Merges[0].MergeState != core.TicketMergeStateBlocked {
		t.Fatalf("merge result = %#v, want one blocked merge", result.Merges)
	}
}

func TestServiceStartAllReadyAgentWorkStopsWhenMergeProviderFails(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)
	installFakeAgenticReviewRunner(t, core.AgenticReviewResult{
		Decision: core.IntegrationReviewDecisionPass,
		Summary:  "Implementation is correct and validation evidence is sufficient.",
	})

	writeRuntimeTemplate(t, projectDir)
	writeAgenticReviewLoopConfig(t, projectDir)

	runnerRemote, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)
	defer restoreRunner()

	mergeCommand := "tea api -X PUT /repos/adrianpk/chr/pulls/274/merge -d @-"
	runnerRemote.inputErrs[mergeCommand] = errors.New("provider failed")

	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(
		testLoopTicketWithPullRequestRef("CHR-1", "Ready frontend"),
		time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	))
	writeTicketMarkdownOnly(t, projectDir, "CHR-2-Ready.md", withLoopTicketCreatedAt(
		testLoopTicketWithPullRequestRef("CHR-2", "Ready backend"),
		time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC),
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkLoop(t.Context(), AgentWorkSelectionOptions{MaxIterations: 2})
	if err != nil {
		t.Fatalf("StartAgentWorkLoop returned error: %v", err)
	}

	assertAgentWorkLoopRunBlockedAfterOneTicket(t, *result.LoopRun)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-2", 0)
	assertAgentWorkTicketState(t, result.Tickets, "CHR-1", ticketcore.StatusReadyForMerge)
	assertTicketContains(t, ticketPath, `merge_state: "blocked"`)
	assertTicketContains(t, ticketPath, `merge_result: "blocked"`)
	assertTicketContains(t, ticketPath, `merge_blocker: "provider failed"`)

	if len(result.Merges) != 1 || result.Merges[0].MergeState != core.TicketMergeStateBlocked {
		t.Fatalf("merge result = %#v, want one blocked merge", result.Merges)
	}
}

func TestServiceStartAllReadyAgentWorkTrustedAutoMergeMergesBeforeNextTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	runner := installFakeAgenticReviewRunner(t, core.AgenticReviewResult{
		Decision: core.IntegrationReviewDecisionNeedsChanges,
		Summary:  "This runner must not be called in trusted auto-merge mode.",
	})

	writeRuntimeTemplate(t, projectDir)
	writeTrustedAutoMergeLoopConfig(t, projectDir)

	runnerRemote, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)
	defer restoreRunner()

	mergeCommand := "tea api -X PUT /repos/adrianpk/chr/pulls/274/merge -d @-"
	runnerRemote.inputOutputs[mergeCommand] = []byte(`{"sha":"abc123"}`)

	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(
		testLoopTicketWithPullRequestRef("CHR-1", "Ready frontend"),
		time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	))
	writeTicketMarkdownOnly(t, projectDir, "CHR-2-Ready.md", withLoopTicketCreatedAt(
		testLoopTicketWithPullRequestRef("CHR-2", "Ready backend"),
		time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC),
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkLoop(t.Context(), AgentWorkSelectionOptions{MaxIterations: 2})
	if err != nil {
		t.Fatalf("StartAgentWorkLoop returned error: %v", err)
	}

	assertAgentWorkLoopRunProcessedTwoTickets(t, *result.LoopRun)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-1", 1)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-2", 1)
	assertAgentWorkTicketState(t, result.Tickets, "CHR-1", ticketcore.StatusDone)
	assertAgentWorkTicketState(t, result.Tickets, "CHR-2", ticketcore.StatusDone)

	if len(runner.reviews) != 0 {
		t.Fatalf("agentic review calls = %d, want 0", len(runner.reviews))
	}

	if len(result.Merges) != 2 {
		t.Fatalf("merges = %d, want 2", len(result.Merges))
	}
}

func TestServiceStartAllReadyAgentWorkRepairsAgenticReviewNeedsChangesBeforeNextTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	runner := installFakeAgenticReviewRunnerSequence(t, []core.AgenticReviewResult{
		agenticReviewNeedsChangesResult("Missing invalid input handling"),
		{
			Decision: core.IntegrationReviewDecisionPass,
			Summary:  "Repair is correct and validation evidence is sufficient.",
		},
		{
			Decision: core.IntegrationReviewDecisionPass,
			Summary:  "Implementation is correct and validation evidence is sufficient.",
		},
	})

	writeRuntimeTemplate(t, projectDir)
	writeAgenticReviewLoopConfig(t, projectDir)

	runnerRemote, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)
	defer restoreRunner()

	mergeCommand := "tea api -X PUT /repos/adrianpk/chr/pulls/274/merge -d @-"
	runnerRemote.inputOutputs[mergeCommand] = []byte(`{"sha":"abc123"}`)

	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(
		testLoopTicketWithPullRequestRef("CHR-1", "Ready frontend"),
		time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	))
	writeTicketMarkdownOnly(t, projectDir, "CHR-2-Ready.md", withLoopTicketCreatedAt(testLoopTicketWithPullRequestRef(
		"CHR-2", "Ready backend",
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAgentWorkLoop(t.Context(), AgentWorkSelectionOptions{MaxIterations: 2})
	if err != nil {
		t.Fatalf("StartAgentWorkLoop returned error: %v", err)
	}

	assertAgentWorkLoopRunProcessedTwoTickets(t, *result.LoopRun)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-1", 2)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-2", 1)
	assertAgenticReviewRunnerCalls(t, runner, projectDir, []string{"CHR-1", "CHR-1", "CHR-2"})
	assertAgentWorkTicketState(t, result.Tickets, "CHR-1", ticketcore.StatusDone)
	assertAgentWorkTicketState(t, result.Tickets, "CHR-2", ticketcore.StatusDone)
	assertTicketContains(t, ticketPath, `review_round: 2`)
	assertTicketContains(t, ticketPath, `review_state: "pass"`)
	assertTicketContains(t, ticketPath, `merge_state: "merged"`)
	assertRepairExecutionContextContains(t, result.Executions[1], "correctness: Missing invalid input handling")
	assertAgentWorkRepairEventPhases(t, projectDir, result)

	if len(result.Merges) != 2 {
		t.Fatalf("merges = %d, want 2", len(result.Merges))
	}
}

func TestServiceStartAllReadyAgentWorkBlocksAfterRepairReviewNeedsChanges(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	runner := installFakeAgenticReviewRunnerSequence(t, []core.AgenticReviewResult{
		agenticReviewNeedsChangesResult("Missing invalid input handling"),
		agenticReviewNeedsChangesResult("Still missing invalid input handling"),
	})

	writeRuntimeTemplate(t, projectDir)
	writeAgenticReviewLoopConfig(t, projectDir)

	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(
		testLoopTicketWithPullRequestRef("CHR-1", "Ready frontend"),
		time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	))
	writeTicketMarkdownOnly(t, projectDir, "CHR-2-Ready.md", withLoopTicketCreatedAt(testLoopTicketWithPullRequestRef(
		"CHR-2", "Ready backend",
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAllReadyAgentWork()
	if err != nil {
		t.Fatalf("StartAllReadyAgentWork returned error: %v", err)
	}

	assertAgentWorkLoopRunBlockedAfterOneTicket(t, *result.LoopRun)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-1", 2)
	assertTicketExecutionRecordCount(t, devRoot, projectDir, "CHR-2", 0)
	assertAgenticReviewRunnerCalls(t, runner, projectDir, []string{"CHR-1", "CHR-1"})
	assertAgentWorkTicketState(t, result.Tickets, "CHR-1", ticketcore.StatusReadyForDev)
	assertTicketContains(t, ticketPath, `review_round: 2`)
	assertTicketContains(t, ticketPath, `review_state: "changes_requested"`)
	assertTicketContains(t, ticketPath, `review_result: "needs_changes"`)
	assertTicketContains(t, ticketPath, "correctness: Still missing invalid input handling")

	if len(result.Merges) != 0 {
		t.Fatalf("merges = %d, want 0", len(result.Merges))
	}
}

func assertAgentWorkLoopRunBlockedAfterOneTicket(t *testing.T, record core.LoopRunRecord) {
	t.Helper()

	if record.Status != core.LoopRunStatusBlocked {
		t.Fatalf("LoopRun.Status = %q, want %q", record.Status, core.LoopRunStatusBlocked)
	}

	if record.EligibleCount != 2 {
		t.Fatalf("LoopRun.EligibleCount = %d, want 2", record.EligibleCount)
	}

	if record.ProcessedCount != 1 {
		t.Fatalf("LoopRun.ProcessedCount = %d, want 1", record.ProcessedCount)
	}

	if record.BlockedCount != 1 {
		t.Fatalf("LoopRun.BlockedCount = %d, want 1", record.BlockedCount)
	}
}

func assertAgentWorkLoopRunProcessedTwoTickets(t *testing.T, record core.LoopRunRecord) {
	t.Helper()

	if record.Status != core.LoopRunStatusSucceeded {
		t.Fatalf("LoopRun.Status = %q, want %q", record.Status, core.LoopRunStatusSucceeded)
	}

	if record.EligibleCount != 2 {
		t.Fatalf("LoopRun.EligibleCount = %d, want 2", record.EligibleCount)
	}

	if record.ProcessedCount != 2 {
		t.Fatalf("LoopRun.ProcessedCount = %d, want 2", record.ProcessedCount)
	}

	if record.BlockedCount != 0 {
		t.Fatalf("LoopRun.BlockedCount = %d, want 0", record.BlockedCount)
	}
}

func assertTicketExecutionRecordCount(t *testing.T, devRoot, projectDir, ticketKey string, want int) {
	t.Helper()

	records := listTicketExecutions(t, devRoot, projectID(projectDir), ticketKey)
	if len(records) != want {
		t.Fatalf("%s execution records = %d, want %d", ticketKey, len(records), want)
	}
}

func TestServiceStartAllReadyAgentWorkEmitsLifecycleEvents(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHR-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.StartAllReadyAgentWork()
	if err != nil {
		t.Fatalf("StartAllReadyAgentWork returned error: %v", err)
	}

	if result.LoopRun == nil {
		t.Fatal("LoopRun is nil")
	}

	assertRunAllAgentEventStreams(t, projectDir, result)
}

func assertRunAllAgentEventStreams(t *testing.T, projectDir string, result AgentWorkResult) {
	t.Helper()

	loopStream := readAgentEventStream(t, core.LoopRunEventStreamPath(*result.LoopRun))
	assertAgentEventPhases(t, loopStream.Events, []core.AgentEventPhase{
		core.AgentEventPhaseLoopStarted,
		core.AgentEventPhaseTicketSelected,
		core.AgentEventPhaseExecutionCreated,
		core.AgentEventPhaseRuntimePrepared,
		core.AgentEventPhaseWorkerStarted,
		core.AgentEventPhaseWorkerFinished,
		core.AgentEventPhaseValidationStarted,
		core.AgentEventPhaseValidationFinished,
		core.AgentEventPhaseResultApplied,
		core.AgentEventPhaseTicketStatusUpdated,
		core.AgentEventPhaseCycleRehydrated,
		core.AgentEventPhaseLoopFinished,
	})

	execution := result.Executions[0]
	executionStream := readAgentEventStream(t, core.ExecutionAgentEventStreamPath(execution))
	assertAgentEventPhases(t, executionStream.Events, []core.AgentEventPhase{
		core.AgentEventPhaseExecutionCreated,
		core.AgentEventPhaseRuntimePrepared,
		core.AgentEventPhaseWorkerStarted,
		core.AgentEventPhaseWorkerFinished,
		core.AgentEventPhaseValidationStarted,
		core.AgentEventPhaseValidationFinished,
		core.AgentEventPhaseResultApplied,
		core.AgentEventPhaseTicketStatusUpdated,
	})

	for _, event := range executionStream.Events {
		if event.LoopID != result.LoopRun.ID {
			t.Fatalf("event LoopID = %q, want %q", event.LoopID, result.LoopRun.ID)
		}
	}
}

func assertAgentWorkRepairEventPhases(t *testing.T, projectDir string, result AgentWorkResult) {
	t.Helper()

	loopStream := readAgentEventStream(t, core.LoopRunEventStreamPath(*result.LoopRun))
	assertAgentEventPhaseSubsequence(t, loopStream.Events, []core.AgentEventPhase{
		core.AgentEventPhaseReviewStarted,
		core.AgentEventPhaseReviewFinished,
		core.AgentEventPhaseRepairStarted,
		core.AgentEventPhaseRepairFinished,
		core.AgentEventPhaseMergeStarted,
		core.AgentEventPhaseMergeFinished,
	})
}

func TestServicePreviewAllReadyAgentWorkDoesNotExecute(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", testLoopTicket(
		"CHR-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.PreviewAllReadyAgentWork()
	if err != nil {
		t.Fatalf("PreviewAllReadyAgentWork returned error: %v", err)
	}

	if result.EligibleCount != 1 || len(result.SelectedQueue) != 1 || result.SelectedQueue[0].Key != "CHR-1" {
		t.Fatalf("preview result = %#v, want one CHR-1 queued ticket", result)
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHR-1")
	if len(records) != 0 {
		t.Fatalf("execution records = %d, want 0", len(records))
	}

	loopRuns := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loopRuns) != 0 {
		t.Fatalf("loop runs = %d, want 0", len(loopRuns))
	}
}

func TestServicePreviewAgentWorkLoopAcceptsWorktreeWorkspaceMode(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeCLIWorkspaceConfig(t, projectDir, `work:
  workspace_mode: worktree
`)
	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", testLoopTicket(
		"CHR-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.PreviewAgentWorkLoop(AgentWorkSelectionOptions{})
	if err != nil {
		t.Fatalf("PreviewAgentWorkLoop returned error: %v", err)
	}

	if result.EligibleCount != 1 || len(result.SelectedQueue) != 1 || result.SelectedQueue[0].Key != "CHR-1" {
		t.Fatalf("preview result = %#v, want one CHR-1 queued ticket", result)
	}
}

func TestServicePreviewAgentWorkLoopAppliesSelection(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHR-1", "Ready normal", "ready_for_dev", "normal", "normal", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))
	writeTicketMarkdownOnly(t, projectDir, "CHR-2-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHR-2", "Ready high", "ready_for_dev", "high", "critical", "bug", "maintenance", "backend", []string{"api"},
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	operations := newAgentWorkTestService(t, devRoot, projectDir)

	result, err := operations.PreviewAgentWorkLoop(AgentWorkSelectionOptions{
		Filters: workflowstate.TicketFilters{
			Priorities: []string{"high"},
		},
		Order: workflowstate.TicketOrderPriority,
	})
	if err != nil {
		t.Fatalf("PreviewAgentWorkLoop returned error: %v", err)
	}

	if result.EligibleCount != 1 || len(result.SelectedQueue) != 1 || result.SelectedQueue[0].Key != "CHR-2" {
		t.Fatalf("preview result = %#v, want one high-priority CHR-2 queued ticket", result)
	}
}

func newAgentWorkTestService(t *testing.T, devRoot, projectDir string) service {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	operations := newService(t.Context(), env, commandProjectContext{
		ProjectDir: projectDir,
		ProjectID:  projectID(projectDir),
	}, testAssetsFS())
	operations.generationBinder = testGenerationBinder

	return operations
}

type fakeServiceAgenticReviewRunner struct {
	result  core.AgenticReviewResult
	results []core.AgenticReviewResult
	reviews []core.AgenticReviewContext
}

func (r *fakeServiceAgenticReviewRunner) Review(_ context.Context, review core.AgenticReviewContext) (core.AgenticReviewArtifact, error) {
	result := r.resultForReview(len(r.reviews))
	r.reviews = append(r.reviews, review)

	runDir := core.AgenticReviewRunDir(review.ProjectDir, review.Ticket.Key, review.Execution.ID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return core.AgenticReviewArtifact{}, err
	}

	resultPath := filepath.Join(runDir, "result.json")

	resultBody, err := json.Marshal(result)
	if err != nil {
		return core.AgenticReviewArtifact{}, err
	}

	if err := os.WriteFile(resultPath, resultBody, 0o644); err != nil {
		return core.AgenticReviewArtifact{}, err
	}

	contextPath := filepath.Join(runDir, "context.md")
	if err := os.WriteFile(contextPath, []byte(core.RenderAgenticReviewContext(review)), 0o644); err != nil {
		return core.AgenticReviewArtifact{}, err
	}

	artifact := core.AgenticReviewArtifact{
		ContextPath: contextPath,
		ResultPath:  resultPath,
		SummaryPath: filepath.Join(runDir, "summary.md"),
		Run: harness.AgentRunResult{
			ID:     "review-" + review.Ticket.Key,
			RunDir: runDir,
		},
		Result: result,
	}
	if err := core.WriteAgenticReviewSummary(artifact); err != nil {
		return core.AgenticReviewArtifact{}, err
	}

	return artifact, nil
}

func (r *fakeServiceAgenticReviewRunner) resultForReview(index int) core.AgenticReviewResult {
	if len(r.results) == 0 {
		return r.result
	}

	if index >= len(r.results) {
		return r.results[len(r.results)-1]
	}

	return r.results[index]
}

func installFakeAgenticReviewRunner(t *testing.T, result core.AgenticReviewResult) *fakeServiceAgenticReviewRunner {
	t.Helper()

	runner := &fakeServiceAgenticReviewRunner{result: result}
	previous := newAgenticReviewRunner
	newAgenticReviewRunner = func() core.AgenticReviewRunner {
		return runner
	}

	t.Cleanup(func() {
		newAgenticReviewRunner = previous
	})

	return runner
}

func installFakeAgenticReviewRunnerSequence(t *testing.T, results []core.AgenticReviewResult) *fakeServiceAgenticReviewRunner {
	t.Helper()

	runner := &fakeServiceAgenticReviewRunner{results: results}
	previous := newAgenticReviewRunner
	newAgenticReviewRunner = func() core.AgenticReviewRunner {
		return runner
	}

	t.Cleanup(func() {
		newAgenticReviewRunner = previous
	})

	return runner
}

func writeAgenticReviewLoopConfig(t *testing.T, projectDir string) {
	t.Helper()

	writeCLIWorkspaceConfig(t, projectDir, `review:
  mode: agentic_review
  non_blocking_comment_limit: 1
`)
}

func writeTrustedAutoMergeLoopConfig(t *testing.T, projectDir string) {
	t.Helper()

	writeCLIWorkspaceConfig(t, projectDir, `review:
  mode: trusted_auto_merge
`)
}

func testLoopTicketWithPullRequestRef(key, title string) ticketcore.Ticket {
	ticket := testLoopTicket(key, title, ticketcore.StatusReadyForDev, "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"})
	ticket.Branch = "feat/slice-2-pr-reference-workflow"
	ticket.ReviewRefs = []ticketmetadata.ReviewRef{{
		Type:     "pull_request",
		Provider: "forgejo",
		URL:      "https://forge.example.com/adrianpk/chr/pulls/274",
		ID:       "274",
		Status:   "open",
	}}

	return ticket
}

func assertAgenticReviewRunnerCalls(t *testing.T, runner *fakeServiceAgenticReviewRunner, projectDir string, ticketKeys []string) {
	t.Helper()

	if len(runner.reviews) != len(ticketKeys) {
		t.Fatalf("agentic review calls = %d, want %d", len(runner.reviews), len(ticketKeys))
	}

	for index, ticketKey := range ticketKeys {
		review := runner.reviews[index]
		if review.ProjectDir != projectDir {
			t.Fatalf("review[%d] ProjectDir = %q, want %q", index, review.ProjectDir, projectDir)
		}

		if review.Ticket.Key != ticketKey {
			t.Fatalf("review[%d] ticket = %q, want %q", index, review.Ticket.Key, ticketKey)
		}
	}
}

func assertAgenticReviewRunArtifact(t *testing.T, projectDir, ticketKey, executionID string) {
	t.Helper()

	runDir := core.AgenticReviewRunDir(projectDir, ticketKey, executionID)
	assertPathExists(t, filepath.Join(runDir, "context.md"))
	assertPathExists(t, filepath.Join(runDir, "result.json"))
	assertPathExists(t, filepath.Join(runDir, "summary.md"))
}

func assertAgentWorkTicketState(t *testing.T, tickets []ticketcore.Indexed, ticketKey, status string) {
	t.Helper()

	for _, ticket := range tickets {
		if ticket.Key == ticketKey {
			if ticket.Status != status {
				t.Fatalf("%s status = %q, want %q", ticketKey, ticket.Status, status)
			}

			return
		}
	}

	t.Fatalf("ticket %s not found in %#v", ticketKey, tickets)
}

func assertTicketContains(t *testing.T, ticketPath, want string) {
	t.Helper()

	if content := readFile(t, ticketPath); !strings.Contains(content, want) {
		t.Fatalf("ticket %s missing %q:\n%s", ticketPath, want, content)
	}
}

func assertRepairExecutionContextContains(t *testing.T, execution core.ExecutionRecord, want string) {
	t.Helper()

	contextPath := filepath.Join(execution.Artifacts.RunDir, "files", "context.md")
	content := readFile(t, contextPath)

	for _, expected := range []string{
		"## Review Repair Handoff",
		want,
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("repair context missing %q:\n%s", expected, content)
		}
	}
}

func agenticReviewNeedsChangesResult(title string) core.AgenticReviewResult {
	return core.AgenticReviewResult{
		Decision: core.IntegrationReviewDecisionNeedsChanges,
		Summary:  "The implementation is missing a required behavior.",
		BlockingFindings: []workflowstate.ReviewFinding{{
			Category: core.AgenticReviewFindingCorrectness,
			Title:    title,
			Detail:   "The ticket requires invalid input recovery, but the implementation accepts bad moves.",
		}},
	}
}

func indexedAgentWorkTicketKeys(tickets []ticketcore.Indexed) []string {
	keys := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		keys = append(keys, ticket.Key)
	}

	return keys
}

func readAgentEventStream(t *testing.T, path string) core.AgentEventStream {
	t.Helper()

	stream, err := core.ReadAgentEventStream(t.Context(), path)
	if err != nil {
		t.Fatalf("ReadAgentEventStream returned error: %v", err)
	}

	if len(stream.Malformed) != 0 {
		t.Fatalf("Malformed events = %#v, want none", stream.Malformed)
	}

	return stream
}

func assertAgentEventPhases(t *testing.T, events []core.AgentEvent, want []core.AgentEventPhase) {
	t.Helper()

	got := make([]core.AgentEventPhase, 0, len(events))
	for _, event := range events {
		got = append(got, event.Phase)
	}

	if !agentEventPhasesEqual(got, want) {
		t.Fatalf("event phases = %#v, want %#v", got, want)
	}
}

func assertAgentEventPhaseSubsequence(t *testing.T, events []core.AgentEvent, want []core.AgentEventPhase) {
	t.Helper()

	next := 0
	for _, event := range events {
		if next < len(want) && event.Phase == want[next] {
			next++
		}
	}

	if next != len(want) {
		t.Fatalf("event phases do not contain subsequence %#v: %#v", want, agentEventPhases(events))
	}
}

func agentEventPhases(events []core.AgentEvent) []core.AgentEventPhase {
	phases := make([]core.AgentEventPhase, 0, len(events))
	for _, event := range events {
		phases = append(phases, event.Phase)
	}

	return phases
}

func agentEventPhasesEqual(left, right []core.AgentEventPhase) bool {
	if len(left) != len(right) {
		return false
	}

	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}

	return true
}
