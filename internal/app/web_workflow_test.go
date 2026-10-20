package app

import (
	"context"
	"html"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

// Web previews must preserve native identity, revision and all local Git/ops bytes.
func TestWebWorkflowPreview(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	handler := webWorkflowHandler(t, operations)
	selected := hierarchy.Tasks[0].Task

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	for _, tc := range []struct {
		name, query, identity string
		total                 int
	}{
		{"ticket", "extent=ticket&key=" + selected.Key, selected.Key, 1},
		{"slice", "extent=slice&key=" + selected.Key, selected.SliceID, 3},
		{"set", "extent=slice_set&key=" + selected.Key, selected.SetID, 6},
		{"case", "extent=use_case&identity=delivery", "delivery", 6},
		{"loose", "extent=loose&order=severity&project_area=ui", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkWebWorkflowPreview(t, operations, handler, hierarchy.Tasks[0], tc.query, tc.identity, tc.total, before)
		})
	}
}

// Confirmed HTTP admission and continuation execute real workers and retain host evidence.
func TestWebWorkflowStart(t *testing.T) {
	operations, record, _, provider := serviceWorkNativeFixture(t)
	handler := webWorkflowHandler(t, operations)
	preview := webWorkflowHTTP(t, handler, http.MethodGet, "/workflows?extent=ticket&key="+record.Task.Key, nil)
	fields := url.Values{"action": {"workflow.start"}, "confirm": {"true"}, "binding": {webWorkflowHidden(t, preview.Body.String(), "binding")}}
	result := webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)
	id := webWorkflowAttribute(t, result.Body.String(), "data-workflow-id")

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.ValidatedTasks != 1 || status.ConfirmedTickets != 0 || len(status.Evidence) != 1 || status.Continuation == nil || provider.creates != 0 {
		t.Fatal("HTTP activation lost evidence or claimed integration", status, err, result.Body.String())
	}

	fields = webWorkflowControl("continue", id, status.Revision)

	result = webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)
	if !strings.Contains(result.Body.String(), "data-write-status=\"ok\"") || !strings.Contains(result.Body.String(), "Candidate") {
		t.Fatal("continuation did not retain its candidate", result.Body.String())
	}

	webWorkflowRejectedControl(t, operations, handler, fields, webui.WriteStatusStale)
}

// Competing admission rejects the stale browser snapshot before any attempt or write.
func TestWebWorkflowOwnership(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	handler := webWorkflowHandler(t, operations)
	preview := webWorkflowHTTP(t, handler, http.MethodGet, "/workflows?extent=ticket&key="+hierarchy.Tasks[0].Task.Key, nil)
	id := workflowCLIManualStart(t, operations, hierarchy)
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	fields := url.Values{"action": {"workflow.start"}, "confirm": {"true"}, "binding": {webWorkflowHidden(t, preview.Body.String(), "binding")}}
	response := webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.TotalAttempts != 0 || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) || strings.Contains(response.Body.String(), "data-write-status=\"ok\"") {
		t.Fatal("stale admission changed work", response.Body.String(), err)
	}

	if !strings.Contains(response.Body.String(), "Workflow action failed") {
		t.Fatal("ownership conflict was hidden", response.Body.String())
	}
}

// Human-review controls display the actual wait, reject moved PR heads, and bind exact approval.
func TestWebWorkflowReview(t *testing.T) {
	operations, id, provider := serviceWorkflowReviewReady(t)
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n")

	if _, err := operations.ReviewWorkflowSlice(t.Context(), id, nil); err != nil {
		t.Fatal(err)
	}

	handler := webWorkflowHandler(t, operations)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	page := webWorkflowHTTP(t, handler, http.MethodGet, "/workflows?id="+id, nil)
	if !strings.Contains(page.Body.String(), "waiting_review") || !strings.Contains(page.Body.String(), status.ApprovalHash) || !strings.Contains(page.Body.String(), "operator") || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("review status changed work or lost approval wait", page.Body.String())
	}

	checkWebWorkflowApproval(t, operations, handler, status, provider)
}

// Invalid HTTP guards, pagination and confirmation payloads fail before retained state changes.
func TestWebWorkflowGuards(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	id := workflowCLIManualStart(t, operations, hierarchy)
	handler := webWorkflowHandler(t, operations)

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	for _, tc := range []struct {
		name, path string
		fields     url.Values
	}{
		{"unconfirmed", "/workflows/actions", url.Values{"action": {"workflow.cancel"}, "id": {id}, "revision": {"1"}}},
		{"missing revision", "/workflows/actions", url.Values{"action": {"workflow.pause"}, "id": {id}, "confirm": {"true"}}},
		{"unknown binding", "/workflows/actions", url.Values{"action": {"workflow.start"}, "confirm": {"true"}, "binding": {"{\"extra\":true}"}}},
		{"oversized body", "/workflows/actions", url.Values{"action": {"workflow.start"}, "confirm": {"true"}, "binding": {strings.Repeat("x", 33<<10)}}},
		{"negative page", "/workflows?offset=-1", nil},
		{"duplicate identity", "/workflows?id=" + id + "&id=other", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			method := http.MethodPost
			if tc.fields == nil {
				method = http.MethodGet
			}

			response := webWorkflowHTTP(t, handler, method, tc.path, tc.fields)
			if strings.Contains(response.Body.String(), "data-write-status=\"ok\"") || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
				t.Fatal("rejected request changed work", response.Body.String())
			}
		})
	}
}

func webWorkflowHandler(t *testing.T, operations service) http.Handler {
	t.Helper()

	provider := webViewProvider{projectContext: operations.projectContext, serviceFactory: func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }}

	handler, err := webui.NewHandler(webui.HandlerOptions{Provider: provider, Data: webui.ShellData{Title: "chr", RefreshSeconds: 30}})
	if err != nil {
		t.Fatal(err)
	}

	return handler
}

func webWorkflowHTTP(t *testing.T, handler http.Handler, method, path string, fields url.Values) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(fields.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	return response
}

func webWorkflowControl(action, id string, revision uint64) url.Values {
	return url.Values{"action": {"workflow." + action}, "id": {id}, "revision": {strconv.FormatUint(revision, 10)}, "confirm": {"true"}}
}

func webWorkflowHidden(t *testing.T, body, name string) string {
	t.Helper()

	matches := regexp.MustCompile(`name="` + regexp.QuoteMeta(name) + `" value="([^"]*)"`).FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatal("hidden field missing", name, body)
	}

	return html.UnescapeString(matches[1])
}

func webWorkflowAttribute(t *testing.T, body, name string) string {
	t.Helper()

	matches := regexp.MustCompile(regexp.QuoteMeta(name) + `="([^"]*)"`).FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatal("workflow attribute missing", name, body)
	}

	return html.UnescapeString(matches[1])
}

func webWorkflowRejectedControl(t *testing.T, operations service, handler http.Handler, fields url.Values, want string) {
	t.Helper()

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	response := webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)

	current, err := operations.WorkflowStatus(t.Context(), fields.Get("id"), WorkflowStatusRequest{Limit: 8})
	if err != nil || webWorkflowAttribute(t, response.Body.String(), "data-workflow-revision") != strconv.FormatUint(current.Revision, 10) || !strings.Contains(response.Body.String(), `data-write-status="`+want+`"`) || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("stale control did not preserve current state", response.Body.String(), err)
	}
}

func checkWebWorkflowPreview(t *testing.T, operations service, handler http.Handler, selected core.WorkflowRecord, query, identity string, total int, before map[string][32]byte) {
	t.Helper()

	response := webWorkflowHTTP(t, handler, http.MethodGet, "/workflows?"+query, nil)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "data-workflow-preview") {
		t.Fatal("preview failed", response.Body.String())
	}

	binding, err := decodeWebWorkflowBinding(webWorkflowHidden(t, response.Body.String(), "binding"))
	if err != nil || binding.Request.Selector.Identity != identity || len(binding.Request.SnapshotHash) != 64 {
		t.Fatal("selection changed at the HTTP boundary", binding, err)
	}

	checkWebWorkflowSelectedGuard(t, binding, selected, query)

	if !strings.Contains(response.Body.String(), strconv.Itoa(total)) || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("preview changed repository bytes")
	}
}

func checkWebWorkflowApproval(t *testing.T, operations service, handler http.Handler, status WorkflowWorkStatus, provider *serviceWorkflowReviewFixture) {
	t.Helper()

	fields := webWorkflowControl("approve", status.ID, status.Revision)
	fields.Set("binding_hash", status.ApprovalHash)

	original := provider.pr.HeadCommit
	provider.pr.HeadCommit = strings.Repeat("d", 40)

	response := webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)
	if !strings.Contains(response.Body.String(), "data-write-status=\"stale\"") || !strings.Contains(response.Body.String(), status.ApprovalHash) {
		t.Fatal("changed PR head accepted or hidden", response.Body.String())
	}

	provider.pr.HeadCommit = original

	current, err := operations.WorkflowStatus(t.Context(), status.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	fields.Set("revision", strconv.FormatUint(current.Revision, 10))

	response = webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)

	approved, err := operations.WorkflowStatus(t.Context(), status.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || approved.ReviewState != "authorized" || approved.ConfirmedTickets != 0 || !strings.Contains(response.Body.String(), "authorized") {
		t.Fatal("exact approval lost or falsely integrated", approved, err)
	}
}

func checkWebWorkflowSelectedGuard(t *testing.T, binding webWorkflowBinding, selected core.WorkflowRecord, query string) {
	t.Helper()

	if strings.Contains(query, "key=") && (binding.Selected == nil || binding.Selected.Key != selected.Task.Key || binding.Selected.Revision != selected.Revision) {
		t.Fatal("ticket revision guard lost")
	}
}

// Lifecycle forms compare the displayed revision and preserve live-owner exclusion.
func TestWebWorkflowLifecycle(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	id := workflowCLIManualStart(t, operations, hierarchy)
	handler := webWorkflowHandler(t, operations)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	var fields url.Values

	for _, action := range []string{"pause", "resume", "cancel"} {
		status, err = operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
		if err != nil {
			t.Fatal(err)
		}

		fields = webWorkflowControl(action, id, status.Revision)

		response := webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)
		if !strings.Contains(response.Body.String(), `data-write-status="ok"`) {
			t.Fatal("lifecycle failed", action, response.Body.String())
		}

		webWorkflowRejectedControl(t, operations, handler, fields, webui.WriteStatusStale)
	}

	status, err = operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.Status != "cancelled" {
		t.Fatal("cancellation not retained", status, err)
	}
}

// Existing case/ticket views expose exact native selection, and retained status
// remains observable if authority changes without granting new backlog execution.
func TestWebWorkflowViews(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)

	project := operations.projectContext.ProjectDir
	if _, err := core.NewCaseMetadataService(project).Save(core.CaseMetadata{Name: "Delivery", Slug: "delivery", Description: "Deliver the approved work."}); err != nil {
		t.Fatal(err)
	}

	handler := webWorkflowHandler(t, operations)
	for _, tc := range []struct{ name, path, marker string }{
		{"board", "/board", hierarchy.Tasks[0].Task.Key},
		{"ticket", "/board/ticket?key=" + hierarchy.Tasks[0].Task.Key, "data-workflow-ticket"},
		{"case", "/cases/detail?id=delivery", "data-workflow-case"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := webWorkflowHTTP(t, handler, http.MethodGet, tc.path, nil)
			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), tc.marker) {
				t.Fatal("native view controls unavailable", response.Body.String())
			}
		})
	}

	id := workflowCLIManualStart(t, operations, hierarchy)
	writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tracker_compat\n")

	before := workflowCLIImage(t, project)

	response := webWorkflowHTTP(t, handler, http.MethodGet, "/workflows?id="+id, nil)
	if !strings.Contains(response.Body.String(), id) || !maps.Equal(before, workflowCLIImage(t, project)) {
		t.Fatal("retained status hidden or mutated")
	}

	response = webWorkflowHTTP(t, handler, http.MethodGet, "/workflows?extent=loose", nil)
	if strings.Contains(response.Body.String(), "data-workflow-start") {
		t.Fatal("compat mode granted native admission")
	}
}

// A live descriptor owner prevents a second browser execution while exposing current status.
func TestWebWorkflowLiveOwner(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	id := workflowCLIManualStart(t, operations, hierarchy)
	handler := webWorkflowHandler(t, operations)

	claim, err := operations.ClaimWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.Close(); err != nil {
			t.Error(err)
		}
	})

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	fields := webWorkflowControl("continue", id, status.Revision)
	response := webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)

	current, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || current.TotalAttempts != status.TotalAttempts || current.ValidatedTasks != status.ValidatedTasks || current.ConfirmedTickets != 0 {
		t.Fatal("competing owner executed work", current, err)
	}

	if !strings.Contains(response.Body.String(), `data-write-status="blocked"`) || webWorkflowAttribute(t, response.Body.String(), "data-workflow-revision") != strconv.FormatUint(current.Revision, 10) {
		t.Fatal("ownership failure hid current checkpoint", response.Body.String())
	}
}
