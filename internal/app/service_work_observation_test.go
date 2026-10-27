package app

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// The public service chooses native records even with an unrelated corrupt
// legacy store, retains separate persisted invocations, and never starts them.
func TestWorkObservationNative(t *testing.T) {
	operations, ids := nativeWorkObservationFixture(t)
	provider := webViewProvider{projectContext: operations.projectContext}
	before := webPlanFiles(t, provider)

	observed, err := operations.ObserveWork(t.Context(), 0)
	if err != nil || observed.Authority != workspaceconfig.WorkStateAuthorityTickets || len(observed.Current) != 2 {
		t.Fatal(observed, err)
	}

	checkObservedContextBindings(t, observed, ids)

	if !reflect.DeepEqual(before, webPlanFiles(t, provider)) {
		t.Fatal("observation changed retained state")
	}
}

// A quiet validation remains bound to its attempt; another invocation's latest
// event cannot overwrite it. Completed results survive without phase events.
func TestWorkObservationContexts(t *testing.T) {
	at := time.Now().UTC()
	tickets := map[string]ticketcore.Indexed{"PRJ-1": {IndexedIdentity: ticketcore.IndexedIdentity{Key: "PRJ-1"}, IndexedSummary: ticketcore.IndexedSummary{Title: "First"}}, "PRJ-2": {IndexedIdentity: ticketcore.IndexedIdentity{Key: "PRJ-2"}, IndexedSummary: ticketcore.IndexedSummary{Title: "Second"}}}
	records := map[string]core.WorkflowRecord{"PRJ-1": {Task: &core.WorkflowTask{Phase: "validating", AttemptID: "one"}}, "PRJ-2": {Task: &core.WorkflowTask{Phase: "working", AttemptID: "two"}}}
	result := WorkObservation{}

	for _, item := range []struct{ id, key, attempt string }{{"first", "PRJ-1", "one"}, {"second", "PRJ-2", "two"}} {
		run := core.WorkflowInvocation{WorkflowIdentity: core.WorkflowIdentity{ID: item.id, UpdatedAt: at}, WorkflowInvocationExecution: core.WorkflowInvocationExecution{Status: "running", Attempts: []core.WorkflowAttempt{{ID: item.attempt, TaskKey: item.key, Outcome: "running", StartedAt: at, Processes: []core.WorkflowProcess{{State: "running", Command: []string{"go", "test"}}}}}}}
		appendNativeWork(&result, run, tickets, records)
	}

	completed := core.WorkflowInvocation{WorkflowIdentity: core.WorkflowIdentity{ID: "complete", UpdatedAt: at}, WorkflowInvocationExecution: core.WorkflowInvocationExecution{Status: "completed", Outcome: "task_validated", Cursor: 1}, WorkflowInvocationInput: core.WorkflowInvocationInput{Members: []core.WorkflowMember{{Admission: core.WorkflowPreviewMember{Key: "PRJ-1"}}}}}
	appendNativeWork(&result, completed, tickets, records)

	if len(result.Current) != 2 || result.Current[0].Phase != "validating" || result.Current[0].Message != "Running go" || result.Current[1].Phase != "working" || len(result.Results) != 1 || result.Results[0].Confirmed != 0 || result.Results[0].Phase != "task_validated" {
		t.Fatal(result)
	}
}

// Pauses, blockers and explicit cancellation retain their lifecycle meaning
// without requiring a phase event or inventing confirmed delivery.
func TestWorkObservationLifecycle(t *testing.T) {
	for _, item := range []struct {
		name, status string
		terminal     bool
	}{{"blocked", "blocked", false}, {"stopped", "paused", false}, {"cancelled", "cancelled", true}} {
		t.Run(item.name, func(t *testing.T) {
			result := WorkObservation{}
			run := core.WorkflowInvocation{WorkflowIdentity: core.WorkflowIdentity{ID: item.name, UpdatedAt: time.Now().UTC()}, WorkflowInvocationExecution: core.WorkflowInvocationExecution{Status: item.status}}
			appendNativeWork(&result, run, nil, nil)

			contexts := result.Current
			if item.terminal {
				contexts = result.Results
			}

			if len(contexts) != 1 || contexts[0].Status != item.status || contexts[0].Phase != item.status || contexts[0].Confirmed != 0 {
				t.Fatal(result)
			}
		})
	}
}

// Legacy status records remain visible without an event phase, while event
// matching rejects newer messages from other tickets, executions or projects.
func TestWorkObservationLegacy(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "work:\n  state_authority: tracker_compat\n  mode: no_git_enforcement\n")

	observed, err := operations.ObserveWork(t.Context(), 0)
	if err != nil || len(observed.Current) != 0 || len(observed.Results) != 0 {
		t.Fatal(observed, err)
	}

	if observed.Authority != workspaceconfig.WorkStateAuthorityTrackerCompat {
		t.Fatal(observed.Authority)
	}
}

// Deduplication and context matching do not infer lifecycle from event silence.
func TestWorkObservationEventBinding(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	now := time.Now().UTC()
	current := WorkContext{ID: "execution:one", TicketKey: "PRJ-1", Status: "running", UpdatedAt: now}

	event := executioncore.AgentEvent{ProjectID: operations.projectContext.ProjectID, ExecutionID: "two", TicketKey: "PRJ-2", Timestamp: now.Add(time.Hour), Phase: executioncore.AgentEventPhaseWorkFailed, Message: "Different task"}
	if legacyWorkEventMatches(event, operations.projectContext.ProjectID, "one", "") {
		t.Fatal("mixed executions")
	}

	event.ExecutionID, event.TicketKey = "one", "PRJ-1"
	event.Phase = executioncore.AgentEventPhaseValidationStarted
	event.Timestamp = now.Add(-time.Minute)

	result := WorkObservation{}
	for range 2 {
		if err := appendObservedLegacyEvent(&result, &current, event); err != nil {
			t.Fatal(err)
		}
	}

	finishWorkObservation(&result, 0)

	if len(result.Events) != 1 || current.Status != "running" || current.Phase != "validation_started" {
		t.Fatal(result, current)
	}

	if legacyWorkEventMatches(event, "other-project", "one", "") {
		t.Fatal("mixed projects")
	}
}

func checkObservedContextBindings(t *testing.T, observed WorkObservation, ids []string) {
	t.Helper()

	for _, current := range observed.Current {
		if current.InvocationID != ids[0] && current.InvocationID != ids[1] {
			t.Fatal(current)
		}

		if current.Title == "" || current.TicketKey == "" {
			t.Fatal("lost ticket binding", current)
		}
	}
}

func nativeWorkObservationFixture(t *testing.T) (service, []string) {
	t.Helper()

	operations, independent, request := serviceWorkflowTicketFixture(t)
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n  max_active_workers: 2\n")

	var ids []string

	for _, key := range []string{independent.Task.Key, observedNativeGroupedKey(t, operations)} {
		request.RequestKey = "observe-" + key
		request.Selection.Selector.Identity = key

		start, err := operations.StartWorkflow(t.Context(), request)
		if err != nil {
			t.Fatal(err)
		}

		ids = append(ids, start.Invocation.ID)
		claimObservedWork(t, operations, start.Invocation.ID)
	}

	path := filepath.Join(operations.projectContext.ProjectDir, ".chr/default/tickets/broken.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("---\nkey: [invalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	return operations, ids
}

func observedNativeGroupedKey(t *testing.T, operations service) string {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := store.InspectWorkflowRecords(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, record := range snapshot.Records {
		if record.Task != nil && record.Task.SetID != "" {
			return record.Task.Key
		}
	}

	t.Fatal("grouped fixture ticket missing")

	return ""
}

// Persisted legacy results remain observable after the queue drains, without
// any event stream or latest phase, and keep actual loop counters separate.
func TestWorkObservationRetained(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "work:\n  state_authority: tracker_compat\n  mode: no_git_enforcement\n")
	seedObservedWorkResults(t, operations)

	observed, err := operations.ObserveWork(t.Context(), 0)
	if err != nil || len(observed.Results) != 3 || len(observed.Current) != 1 || len(observed.Queue) != 0 {
		t.Fatal(observed, err)
	}

	for _, result := range observed.Results {
		if result.Kind == "loop" && (result.Processed != 1 || result.Failed != 1) {
			t.Fatal(result)
		}
	}

	if len(observed.Events) != 0 {
		t.Fatal("invented events")
	}
}

func seedObservedWorkResults(t *testing.T, operations service) {
	t.Helper()

	repository, err := executioncore.NewRepository(t.Context(), operations.env, operations.assetsFS)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	for _, status := range []executioncore.Status{executioncore.StatusRunning, executioncore.StatusSucceeded, executioncore.StatusFailed} {
		record := executioncore.Record{ID: string(status), ProjectID: operations.projectContext.ProjectID, TicketKey: "PRJ-" + string(status), Status: status, CreatedAt: now, UpdatedAt: now}
		if err = repository.CreateExecution(t.Context(), record); err != nil {
			t.Fatal(err)
		}
	}

	if err = repository.Close(); err != nil {
		t.Fatal(err)
	}

	loops, err := executioncore.NewLoopRunRepository(t.Context(), operations.env, operations.assetsFS)
	if err != nil {
		t.Fatal(err)
	}

	record := executioncore.LoopRunRecord{ID: "retained-loop", ProjectID: operations.projectContext.ProjectID, Status: executioncore.LoopRunStatusSucceeded, EligibleCount: 2, ProcessedCount: 1, FailedCount: 1, CreatedAt: now, UpdatedAt: now}
	if err = loops.CreateLoopRun(t.Context(), record); err != nil {
		t.Fatal(err)
	}

	if err = loops.Close(); err != nil {
		t.Fatal(err)
	}
}

func claimObservedWork(t *testing.T, operations service, id string) {
	t.Helper()

	claim, err := operations.ClaimWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if claim == nil {
		t.Fatal("native fixture did not acquire its claim")
	}

	t.Cleanup(func() {
		if err := claim.Close(); err != nil {
			t.Error(err)
		}
	})
}
