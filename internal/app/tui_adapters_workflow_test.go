package app

import (
	"errors"
	"maps"
	"reflect"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

// TestTUIWorkflowPreview resolves native hierarchy IDs through the real adapter
// and HTTPS target reader while keeping all local Git and ops bytes unchanged.
func TestTUIWorkflowPreview(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	executor := newTUIWorkActionExecutor(operations)
	selected := hierarchy.Tasks[0]
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	cases := []struct {
		name, extent, identity string
		action                 tui.WorkActionID
	}{
		{"ticket", "ticket", selected.Task.Key, tui.WorkActionWorkflowTicket},
		{"slice", "slice", selected.Task.SliceID, tui.WorkActionWorkflowSlice},
		{"set", "slice_set", selected.Task.SetID, tui.WorkActionWorkflowSet},
		{"case", "use_case", selected.Task.UseCase, tui.WorkActionWorkflowCase},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := tui.WorkActionRequest{Action: tc.action, TicketKey: selected.Task.Key, Workflow: &tui.WorkflowActionRequest{Preview: true, Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: tc.extent}}}}
			checkTUIWorkflowPreview(t, operations, executor, request, selected, tc.identity, before)
		})
	}
}

// TestTUIWorkflowStale rejects a confirmation after a competing owner admits
// the ticket, without creating an attempt through the stale TUI action.
func TestTUIWorkflowStale(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	executor := newTUIWorkActionExecutor(operations)

	preview, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionWorkflowTicket, TicketKey: hierarchy.Tasks[0].Task.Key, Workflow: &tui.WorkflowActionRequest{Preview: true, Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket"}}}})
	if err != nil {
		t.Fatal(err)
	}

	id := workflowCLIManualStart(t, operations, hierarchy)

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	if _, err := executor.ExecuteWorkAction(preview.Confirmation.Request); !errors.Is(err, workflowstate.ErrConflict) && !errors.Is(err, core.ErrWorkflowOwned) {
		t.Fatal("stale confirmation started work", err)
	}

	current, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || current.TotalAttempts != 0 || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("stale confirmation changed retained work", err)
	}
}

// TestTUIWorkflowControls uses actual retained state, displaying a stale failure
// with current revision and preserving pause/resume/cancel lifecycle semantics.
func TestTUIWorkflowControls(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	id := workflowCLIManualStart(t, operations, hierarchy)
	executor := newTUIWorkActionExecutor(operations)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	request := tui.WorkActionRequest{Action: tui.WorkActionWorkflowPause, Workflow: &tui.WorkflowActionRequest{ID: id, ExpectedRevision: status.Revision, Preview: true}}
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	preview, err := executor.ExecuteWorkAction(request)
	if err != nil || preview.Confirmation == nil || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("control preview changed authorization", err)
	}

	paused, err := executor.ExecuteWorkAction(preview.Confirmation.Request)
	if err != nil || paused.Workflow == nil || paused.Workflow.Revision <= status.Revision {
		t.Fatal("pause failed", err)
	}

	checkTUIWorkflowStaleControl(t, operations, executor, preview.Confirmation.Request, paused.Workflow.Revision)

	checkTUIWorkflowStopLifecycle(t, executor, id, paused)
}

// TestTUIWorkflowStart executes one actual host-validated task checkpoint through
// the confirmed adapter, retaining progress without claiming delivery or a PR.
func TestTUIWorkflowStart(t *testing.T) {
	operations, record, _, provider := serviceWorkNativeFixture(t)
	executor := newTUIWorkActionExecutor(operations)

	preview, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionWorkflowTicket, TicketKey: record.Task.Key, Workflow: &tui.WorkflowActionRequest{Preview: true, Selection: workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket"}}}})
	if err != nil || preview.Confirmation == nil {
		t.Fatal("preview failed", err)
	}

	result, err := executor.ExecuteWorkAction(preview.Confirmation.Request)
	if err != nil || result.Workflow == nil || provider.creates != 0 {
		t.Fatal("confirmed adapter failed or published early", err)
	}

	checkTUIWorkflowTaskStatus(t, operations, result)
	checkTUIWorkflowContinuation(t, operations, executor, *result.Workflow)
}

// TestTUIWorkflowApproval preserves exact approval input across a changed PR
// head and returns current status alongside the rejected guarded action.
func TestTUIWorkflowApproval(t *testing.T) {
	operations, id, provider := serviceWorkflowReviewReady(t)

	review, err := operations.ReviewWorkflowSlice(t.Context(), id, nil)
	if err != nil {
		t.Fatal(err)
	}

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.NextActionOwner != "operator" {
		t.Fatal("human review wait lost its owner", err)
	}

	executor := newTUIWorkActionExecutor(operations)
	request := tui.WorkActionRequest{Action: tui.WorkActionWorkflowApprove, Workflow: &tui.WorkflowActionRequest{Preview: true, ID: id, ExpectedRevision: status.Revision, BindingHash: review.InputHash}}
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	preview, err := executor.ExecuteWorkAction(request)
	if err != nil || preview.Confirmation == nil || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("approval preview mutated work", err)
	}

	original := provider.pr.HeadCommit
	provider.pr.HeadCommit = strings.Repeat("d", 40)

	result, err := executor.ExecuteWorkAction(preview.Confirmation.Request)
	if !errors.Is(err, workflowstate.ErrConflict) || result.Workflow == nil || result.Workflow.ApprovalHash != review.InputHash {
		t.Fatal("moved PR head accepted or hidden", err)
	}

	provider.pr.HeadCommit = original

	checkTUIWorkflowApproval(t, operations, executor, *result.Workflow)
}

func checkTUIWorkflowPreview(t *testing.T, operations service, executor tuiWorkActionExecutor, request tui.WorkActionRequest, selected core.WorkflowRecord, identity string, before map[string][32]byte) {
	t.Helper()

	input := *request.Workflow

	result, err := executor.ExecuteWorkAction(request)
	if err != nil || result.Confirmation == nil {
		t.Fatal("native preview failed", err)
	}

	confirmed := result.Confirmation.Request.Workflow
	if confirmed.Selection.Selector.Identity != identity || confirmed.Selected.Key != selected.Task.Key || confirmed.Selected.Revision != selected.Revision || len(confirmed.Selection.SnapshotHash) != 64 || confirmed.Preview {
		t.Fatal("preview replaced selected identity or revision", confirmed)
	}

	if !reflect.DeepEqual(input, *request.Workflow) || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("preview changed its input, Git or operational bytes")
	}
}

func checkTUIWorkflowStaleControl(t *testing.T, operations service, executor tuiWorkActionExecutor, request tui.WorkActionRequest, revision uint64) {
	t.Helper()

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	stale, err := executor.ExecuteWorkAction(request)
	if !errors.Is(err, workflowstate.ErrConflict) || stale.Workflow == nil || stale.Workflow.Revision != revision || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("stale control hid current status or changed authorization", err)
	}
}

func checkTUIWorkflowTaskStatus(t *testing.T, operations service, result tui.WorkActionResult) {
	t.Helper()

	status, err := operations.WorkflowStatus(t.Context(), result.Workflow.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.ValidatedTasks != 1 || status.ConfirmedTickets != 0 || len(status.Evidence) != 1 || status.Continuation == nil {
		t.Fatal("task validation was confused with integration", status, err)
	}

	if !strings.Contains(strings.Join(result.Workflow.Lines, "\n"), "Evidence:") {
		t.Fatal("host evidence not displayed")
	}
}

func checkTUIWorkflowContinuation(t *testing.T, operations service, executor tuiWorkActionExecutor, view tui.WorkflowView) {
	t.Helper()

	request := tui.WorkActionRequest{Action: tui.WorkActionWorkflowContinue, Workflow: &tui.WorkflowActionRequest{Preview: true, ID: view.ID, ExpectedRevision: view.Revision}}

	preview, err := executor.ExecuteWorkAction(request)
	if err != nil || preview.Confirmation == nil {
		t.Fatal("continuation preview failed", err)
	}

	result, err := executor.ExecuteWorkAction(preview.Confirmation.Request)
	if err != nil || result.Workflow == nil {
		t.Fatal("guarded continuation failed", err)
	}

	status, err := operations.WorkflowStatus(t.Context(), view.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.CandidateHash == "" || status.ConfirmedTickets != 0 {
		t.Fatal("continuation lost candidate or fabricated integration", err)
	}

	checkTUIWorkflowStaleControl(t, operations, executor, preview.Confirmation.Request, result.Workflow.Revision)
}

func checkTUIWorkflowApproval(t *testing.T, operations service, executor tuiWorkActionExecutor, view tui.WorkflowView) {
	t.Helper()

	request := tui.WorkActionRequest{Action: tui.WorkActionWorkflowApprove, Workflow: &tui.WorkflowActionRequest{ID: view.ID, ExpectedRevision: view.Revision, BindingHash: view.ApprovalHash}}
	if _, err := executor.ExecuteWorkAction(request); err != nil {
		t.Fatal("exact approval failed", err)
	}

	status, err := operations.WorkflowStatus(t.Context(), view.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.ReviewState != "authorized" || status.ConfirmedTickets != 0 {
		t.Fatal("approval fabricated delivery or lost authorization", err)
	}
}

// TestTUIWorkflowRefresh reads the displayed identity on its selected retained
// page, without changing a checkpoint or substituting a different invocation.
func TestTUIWorkflowRefresh(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	id := workflowCLIManualStart(t, operations, hierarchy)
	executor := newTUIAgentWorkRefreshExecutor(operations)
	observability := tui.ObservabilityView{}
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	view, err := executor.refreshNativeWorkflow(tui.AgentWorkRefreshRequest{WorkflowID: id, WorkflowOffset: 1}, &observability)
	if err != nil || view == nil || view.ID != id || observability.WorkflowPage.Offset != 1 {
		t.Fatal("refresh substituted the retained selection", err)
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("dashboard refresh changed workflow or Git bytes")
	}
}

func checkTUIWorkflowStopLifecycle(t *testing.T, executor tuiWorkActionExecutor, id string, paused tui.WorkActionResult) {
	t.Helper()

	var err error

	for _, action := range []tui.WorkActionID{tui.WorkActionWorkflowResume, tui.WorkActionWorkflowCancel} {
		request := tui.WorkActionRequest{Action: action, Workflow: &tui.WorkflowActionRequest{ID: id, ExpectedRevision: paused.Workflow.Revision}}

		paused, err = executor.ExecuteWorkAction(request)
		if err != nil {
			t.Fatal("guarded lifecycle failed", action, err)
		}
	}

	if !strings.Contains(strings.Join(paused.Workflow.Lines, "\n"), "State: cancelled") {
		t.Fatal("cancellation not displayed", paused.Workflow.Lines)
	}
}
