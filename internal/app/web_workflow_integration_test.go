package app

import (
	"maps"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/webui"
)

// TestWebIntegrationView reads actual prepared native evidence through HTTP,
// preserving its phase, source/result distinction and private ownership boundary.
func TestWebIntegrationView(t *testing.T) {
	operations, id := serviceIntegrationReady(t)
	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	page := webWorkflowHTTP(t, webWorkflowHandler(t, operations), http.MethodGet, "/workflows?id="+id, nil)
	body := page.Body.String()

	for _, field := range workflowIntegrationFields(status.Integration) {
		if !strings.Contains(body, "<dt>"+field.label+"</dt>") || !strings.Contains(body, field.value) {
			t.Fatal("Web omitted a shared retained integration fact", field.label)
		}
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("Web observation changed native authority")
	}

	for _, private := range []string{"claim_generation", "lease_until", "owner_token", "credential_identity", "owned_workspace_id"} {
		if strings.Contains(body, private) {
			t.Fatal("Web disclosed private ownership", private)
		}
	}
}

func webIntegrationForm(t *testing.T, body, action string) url.Values {
	t.Helper()

	fields := url.Values{"action": {"workflow." + action}, "confirm": {"true"}}
	for _, name := range []string{"id", "revision", "entry_id", "entry_revision", "budget_revision", "candidate_hash"} {
		fields.Set(name, webWorkflowHidden(t, body, name))
	}

	return fields
}

// TestWebIntegrationGrant submits the rendered guard through real HTTP, rejects
// each stale fence without writes, then preserves accounting on an explicit grant.
func TestWebIntegrationGrant(t *testing.T) {
	operations, id := serviceIntegrationReady(t)
	handler := webWorkflowHandler(t, operations)
	page := webWorkflowHTTP(t, handler, http.MethodGet, "/workflows?id="+id, nil)
	fields := webIntegrationForm(t, page.Body.String(), "grant")
	fields.Set("record_id", webWorkflowHidden(t, page.Body.String(), "record_id"))
	fields.Set("kind", "submission")
	fields.Set("amount", "1")
	fields.Set("actor", "operator")
	fields.Set("reason", "Authorize one additional bounded submission")

	before, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	checkWebIntegrationGrantGuards(t, operations, handler, fields, before)

	response := webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)
	if !strings.Contains(response.Body.String(), `data-write-status="ok"`) {
		t.Fatal("rendered bounded extension failed", response.Body.String())
	}

	after, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	checkWorkflowIntegrationExtension(t, before, after, err)
	webWorkflowRejectedControl(t, operations, handler, fields, webui.WriteStatusStale)
}

func checkWebIntegrationGrantGuards(t *testing.T, operations service, handler http.Handler, fields url.Values, shown WorkflowWorkStatus) {
	t.Helper()

	cases := []struct{ name, key, value string }{
		{"invocation", "revision", strconv.FormatUint(shown.Revision+1, 10)},
		{"entry", "entry_revision", strconv.FormatUint(shown.Integration.EntryRevision+1, 10)},
		{"budget", "budget_revision", strconv.FormatUint(shown.Integration.Budget.Revision+1, 10)},
		{"candidate", "candidate_hash", strings.Repeat("0", 64)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stale := maps.Clone(fields)
			stale.Set(tc.key, tc.value)
			webWorkflowRejectedControl(t, operations, handler, stale, webui.WriteStatusStale)
		})
	}
}

// TestWebIntegrationApproval binds native approval to the rendered candidate
// and guards, rejecting missing fences and replay without claiming publication.
func TestWebIntegrationApproval(t *testing.T) {
	operations, id := serviceIntegrationReady(t)
	handler := webWorkflowHandler(t, operations)
	page := webWorkflowHTTP(t, handler, http.MethodGet, "/workflows?id="+id, nil)
	fields := webIntegrationForm(t, page.Body.String(), "approve")
	fields.Set("binding_hash", webWorkflowHidden(t, page.Body.String(), "binding_hash"))

	missing := webWorkflowControl("approve", id, mustWebIntegrationRevision(t, fields.Get("revision")))
	missing.Set("binding_hash", fields.Get("binding_hash"))
	webWorkflowRejectedControl(t, operations, handler, missing, webui.WriteStatusStale)

	before, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	response := webWorkflowHTTP(t, handler, http.MethodPost, "/workflows/actions", fields)
	if !strings.Contains(response.Body.String(), `data-write-status="ok"`) {
		t.Fatal("rendered native approval failed", response.Body.String())
	}

	after, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	checkServiceIntegrationApproval(t, before, after, err)
	webWorkflowRejectedControl(t, operations, handler, fields, webui.WriteStatusStale)
}

func mustWebIntegrationRevision(t *testing.T, value string) uint64 {
	t.Helper()

	revision, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		t.Fatal(err)
	}

	return revision
}
