package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/tui"
)

func daemonWorkflowFixture(t *testing.T, output *bytes.Buffer) daemonWorkflowRunner {
	t.Helper()

	root := t.TempDir()
	repository := newDaemonStateRepositoryForTest(t, root)

	now := time.Now().UTC()
	if err := repository.UpsertDaemonState(t.Context(), core.DaemonState{Status: core.DaemonStatusRunning, PID: os.Getpid(), StartedAt: &now, HeartbeatAt: &now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}

	return daemonWorkflowRunner{stdout: output, repository: repository}
}

func daemonWorkflowStatus(t *testing.T, project, action, id string) (WorkflowWorkStatus, error) {
	t.Helper()

	var output bytes.Buffer

	err := run(t.Context(), []string{"daemon", "workflow", action, "--dir", project, "--json", id}, &output, &bytes.Buffer{})

	var status WorkflowWorkStatus
	if output.Len() > 0 {
		err = errors.Join(err, json.Unmarshal(output.Bytes(), &status))
	}

	return status, err
}

// Startup walks registered native projects without admitting any ready ticket,
// including native-only workspaces with no legacy ticket directory.
func TestDaemonWorkflowStartup(t *testing.T) {
	operations, _, _ := workflowCLIFixture(t)
	project, root := operations.projectContext.ProjectDir, t.TempDir()
	addProject(t, root, "native", project)

	before := workflowCLIImage(t, project)

	var output bytes.Buffer
	if err := run(t.Context(), []string{"daemon", "start", "--dev-root", root}, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	if !maps.Equal(before, workflowCLIImage(t, project)) || !strings.Contains(output.String(), "workflows_observed: 0") || !strings.Contains(output.String(), "projects_refreshed: 0") {
		t.Fatal("startup admitted backlog work or depended on legacy tickets", output.String())
	}
}

// An authorized ticket executes exactly one actual checkpoint per daemon
// activation. Other ready tasks, frozen membership and budgets stay unchanged.
func TestDaemonWorkflowContinue(t *testing.T) {
	operations, ticket, request, provider := serviceWorkNativeFixture(t)

	start, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	runner := daemonWorkflowFixture(t, &output)

	result := daemonForegroundLoopResult{}
	if err := runner.workflows(t.Context(), operations, &result); err != nil {
		t.Fatal(err)
	}

	current, err := operations.ReadWorkflow(t.Context(), start.Invocation.ID)
	if err != nil || result.WorkflowSteps != 1 || current.Cursor != 1 || current.ConfirmedCursor != 0 || len(current.Attempts) != 1 || provider.creates != 0 {
		t.Fatal("daemon exceeded one checkpoint or claimed delivery", current, result, err)
	}

	checkDaemonWorkflowSelection(t, start.Invocation, current, ticket.Task.Key, output.String())
}

func checkDaemonWorkflowSelection(t *testing.T, before, current core.WorkflowInvocation, key, output string) {
	t.Helper()

	if current.Selection.Selector.Identity != key || len(current.Members) != 1 || !reflect.DeepEqual(current.Members[0].Admission, before.Members[0].Admission) || current.Policy != before.Policy || current.Selection.Target != before.Selection.Target {
		t.Fatal("daemon changed retained selection, target or budgets")
	}

	if !strings.Contains(output, "task_validation: 1/1") || !strings.Contains(output, "confirmed_tickets: 0") {
		t.Fatal("daemon hid validation/integration distinction", output)
	}
}

// A live manual descriptor remains exclusive: daemon startup reports ownership
// without accepting a continuation, adding an attempt or changing native bytes.
func TestDaemonWorkflowLiveOwner(t *testing.T) {
	operations, _, request, _ := serviceWorkNativeFixture(t)

	start, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), start.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), 5*time.Second)
		defer cancel()

		if err := errors.Join(operations.CheckpointWorkflow(ctx, claim), claim.Close()); err != nil {
			t.Error(err)
		}
	})

	if _, err := operations.PrepareWorkflowManualTask(t.Context(), claim, nil); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	var output bytes.Buffer

	runner := daemonWorkflowFixture(t, &output)

	result := daemonForegroundLoopResult{}
	if err := runner.workflows(t.Context(), operations, &result); err != nil {
		t.Fatal(err)
	}

	if result.WorkflowSteps != 0 || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) || !strings.Contains(output.String(), "workflow_failure:") {
		t.Fatal("daemon bypassed a live manual owner", output.String())
	}
}

// Real daemon controls retain operator stops. Restart never clears those stops,
// and only an explicit resume can restore a pause; cancellation stays terminal.
func TestDaemonWorkflowStops(t *testing.T) {
	for _, intent := range []string{"pause", "cancel"} {
		t.Run(intent, func(t *testing.T) {
			checkDaemonWorkflowStops(t, intent)
		})
	}
}

func checkDaemonWorkflowStops(t *testing.T, intent string) {
	t.Helper()

	operations, hierarchy, _ := workflowCLIFixture(t)
	project, root := operations.projectContext.ProjectDir, t.TempDir()
	id := workflowCLIManualStart(t, operations, hierarchy)

	stopped, err := daemonWorkflowStatus(t, project, intent, id)
	if err != nil || stopped.StopIntent != intent {
		t.Fatal("daemon control lost explicit stop", stopped, err)
	}

	addProject(t, root, "stopped", project)

	before := workflowCLIImage(t, project)
	if err := run(t.Context(), []string{"daemon", "start", "--dev-root", root}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	if !maps.Equal(before, workflowCLIImage(t, project)) {
		t.Fatal("startup resumed an operator stop")
	}

	checkDaemonWorkflowResume(t, project, id, intent, before)
}

func checkDaemonWorkflowResume(t *testing.T, project, id, intent string, before map[string][32]byte) {
	t.Helper()

	resumed, err := daemonWorkflowStatus(t, project, "resume", id)
	if intent == "pause" && (err != nil || resumed.Status != "pending" || resumed.TotalAttempts != 0) {
		t.Fatal("explicit daemon resume failed", resumed, err)
	}

	if intent == "cancel" && (!errors.Is(err, core.ErrWorkflowOwned) || !maps.Equal(before, workflowCLIImage(t, project))) {
		t.Fatal("terminal cancellation resumed", err)
	}
}

// A retained human-review wait stays observational, preserving its exact
// approval binding and evidence instead of granting automatic approval.
func TestDaemonWorkflowReview(t *testing.T) {
	operations, id, _ := serviceWorkflowReviewReady(t)
	if _, err := operations.ReviewWorkflowSlice(t.Context(), id, nil); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	var output bytes.Buffer

	runner := daemonWorkflowFixture(t, &output)

	result := daemonForegroundLoopResult{}
	if err := runner.workflows(t.Context(), operations, &result); err != nil {
		t.Fatal(err)
	}

	if result.WorkflowsWaiting != 1 || result.WorkflowSteps != 0 || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("daemon bypassed a human review wait", result)
	}
}

// A supervised manual harness is host-validated once; actual CLI, daemon, TUI
// and Web reads show that same evidence without treating validation as delivery.
func TestWorkflowManualSurfaces(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	project := operations.projectContext.ProjectDir
	id := workflowCLIManualStart(t, operations, hierarchy)
	status, err := workflowCLIStatus(t, project, "reconcile", "--kind", "manual", "--harness", "sh", "--harness", "-c", "--harness", workflowCLIManualHarness, id)
	checkWorkflowCLIManualStatus(t, status, err)

	before := workflowCLIImage(t, project)

	observed, err := daemonWorkflowStatus(t, project, "status", id)
	if err != nil || !reflect.DeepEqual(status, observed) {
		t.Fatal("daemon status diverged from manual host evidence", observed, err)
	}

	executor := newTUIWorkActionExecutor(operations)

	view, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionWorkflowStatus, Workflow: &tui.WorkflowActionRequest{ID: id}})
	if err != nil || view.Workflow == nil || view.Workflow.Revision != status.Revision || !strings.Contains(strings.Join(view.Workflow.Lines, "\n"), "Host-validated tasks: 1/1; confirmed tickets: 0") {
		t.Fatal("TUI lost manual checkpoint evidence", view, err)
	}

	response := webWorkflowHTTP(t, webWorkflowHandler(t, operations), http.MethodGet, "/workflows?id="+id, nil)
	if !strings.Contains(response.Body.String(), status.Evidence[0].ResultHash) || !strings.Contains(response.Body.String(), "Host-validated tasks 1/1; confirmed tickets 0") || !maps.Equal(before, workflowCLIImage(t, project)) {
		t.Fatal("Web lost manual evidence or a status adapter wrote state", response.Body.String())
	}
}

func daemonWorkflowMarker(t *testing.T, marker string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(marker); err == nil {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("owned worker did not reach its barrier", filepath.Base(marker))
}
