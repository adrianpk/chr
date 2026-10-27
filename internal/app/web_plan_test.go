package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/webui"
)

// Real HTTP submission enforces shared readiness, binds a named case/target,
// and updates only its planning panel.
func TestWebPlanContext(t *testing.T) {
	provider, model := webPlanFixture(t)
	handler := webQuizHandler(t, provider)

	detail := webPlanDetail(t, provider)
	if detail.Plan.Label != "Generate Plan" || detail.Plan.PlanURL != "" {
		t.Fatal(detail.Plan)
	}

	form := webPlanForm(detail)

	rejected := webDumpEditorPost(handler, form)
	if !strings.Contains(rejected.Body.String(), "absent") || len(model.requests) != 0 {
		t.Fatal("Web bypassed shared readiness", rejected.Body.String())
	}

	addWebPlanDiscovery(t, provider)

	created := webDumpEditorPost(handler, form)
	if created.Header().Get("HX-Retarget") != "#case-plan-panel" || !strings.Contains(created.Body.String(), "Regenerate Plan") {
		t.Fatal(created.Body.String(), created.Header())
	}

	checkWebPlanStale(t, provider, handler)

	checkWebPlanCatalog(t, handler)
}

func webPlanFixture(t *testing.T) (webViewProvider, *fakeServiceQuizModel) {
	t.Helper()

	provider, _ := newWebQuizFixture(t)
	operations := provider.service(t.Context())
	model := &fakeServiceQuizModel{}
	operations.quizModel, operations.conversationModel = model, sufficientPlanReadinessProvider()
	operations.dumpCurator = fallbackDumpCurationProvider{Reason: "deterministic planning fixture"}
	provider.serviceFactory = func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }

	return provider, model
}

func webPlanDetail(t *testing.T, provider webViewProvider) webui.ArtifactDetailView {
	t.Helper()

	detail, err := provider.ArtifactDetailView(t.Context(), webui.ArtifactDetailRequest{Family: "cases", ID: "named-work"})
	if err != nil || detail.Plan == nil || detail.Plan.BlockedReason != "" {
		t.Fatal(detail, err)
	}

	return detail
}

func webPlanForm(detail webui.ArtifactDetailView) url.Values {
	return url.Values{"action": {detail.Plan.Action}, "family": {"cases"}, "surface": {"case-plan"}, "modal": {"true"}, "id": {detail.ID}, "revision": {detail.Plan.Revision}, "request_id": {detail.Plan.OperationID}, "case_slug": {detail.ID}, "confirm": {"true"}}
}

func addWebPlanDiscovery(t *testing.T, provider webViewProvider) {
	t.Helper()

	if _, err := provider.service(t.Context()).CreateDump(DumpCreateRequest{OperationID: "plan-context", Title: "Greeting requirements", Content: "Print hello once. Verify stdout is hello.", Scope: DumpScope{CaseSlug: "named-work"}}); err != nil {
		t.Fatal(err)
	}
}

func checkWebPlanCatalog(t *testing.T, handler http.Handler) {
	t.Helper()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/plans", nil))

	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), `value="plan.generate"`) || strings.Contains(response.Body.String(), `value="plan.regenerate"`) {
		t.Fatal("Plans retained detached generation forms", response.Body.String())
	}
}

// Unbound forms cannot select a scope outside the displayed case/revision.
// GET observation invokes neither model boundary.
func TestWebPlanBindings(t *testing.T) {
	provider, model := webPlanFixture(t)
	for _, field := range []string{"case_slug", "id", "revision", "surface"} {
		form := webPlanForm(webPlanDetail(t, provider))
		form.Set(field, "")

		response := webDumpEditorPost(webQuizHandler(t, provider), form)
		if strings.Contains(response.Body.String(), "Plan generated.") || len(model.requests) != 0 {
			t.Fatal("unbound Web generation succeeded", field)
		}
	}

	checkWebPlanCatalog(t, webQuizHandler(t, provider))
}

// Native submission reports incomplete authoring on failure and redirects to
// the saved plan only after its tickets and order map are ready.
func TestWebPlanNative(t *testing.T) {
	for _, ready := range []bool{false, true} {
		provider, model := webPlanDeliveryFixture(t)
		model.fail = !ready

		form := webPlanDeliveryForm(t, provider)
		request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		response := httptest.NewRecorder()
		webQuizHandler(t, provider).ServeHTTP(response, request)

		if ready {
			if response.Code != http.StatusSeeOther || !strings.HasPrefix(response.Header().Get("Location"), "/plans/detail?id=") {
				t.Fatal(response.Code, response.Header())
			}
		} else if !strings.Contains(response.Body.String(), "ticket creation is incomplete") {
			t.Fatal(response.Body.String())
		}
	}
}

func checkWebPlanStale(t *testing.T, provider webViewProvider, handler http.Handler) {
	t.Helper()

	after := webPlanDetail(t, provider)
	stale := webPlanForm(after)

	operations := provider.service(t.Context())
	if _, err := operations.RegeneratePlan(t.Context(), PlanRegenerateRequest{Scope: quizcore.Scope{CaseSlug: "named-work"}, ExpectedRevision: after.Plan.Revision}); err != nil {
		t.Fatal(err)
	}

	response := webDumpEditorPost(handler, stale)
	if !strings.Contains(response.Body.String(), `data-plan-write-status="stale"`) || !strings.Contains(response.Body.String(), "Reload case") || response.Header().Get("HX-Retarget") != "" {
		t.Fatal("stale failure replaced the case panel", response.Body.String())
	}
}
