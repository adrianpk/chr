package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type planDeliveryTestModel struct {
	base  fakeServiceQuizModel
	fail  bool
	calls int
}

func (m *planDeliveryTestModel) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	if request.Strategy != quizmodel.StrategyPlanDelivery {
		return m.base.Invoke(ctx, request)
	}

	m.calls++
	if m.fail {
		return quizmodel.Response{}, errors.New("provider temporarily unavailable")
	}

	var input generationcore.GeneratedWorkflowPlan

	for _, section := range request.ContextSections {
		if section.Name == "planned_workflow" {
			if err := json.Unmarshal([]byte(section.Content), &input); err != nil {
				return quizmodel.Response{}, err
			}
		}
	}

	gate := workflowstate.Validation{Required: true, Commands: [][]string{{"go", "test", "./..."}}, TimeoutSeconds: 60}
	input.ClosingValidation = gate

	for i := range input.Slices {
		input.Slices[i].Contract, input.Slices[i].Validation = "Print the greeting once.", gate
		for j := range input.Slices[i].Tasks {
			task := &input.Slices[i].Tasks[j]
			task.Body, task.Outcome, task.Scope = "Print hello once and check stdout.", "Greeting printed.", "The selected greeting."
			task.Acceptance, task.Validation = []string{"stdout equals hello"}, gate
			task.Priority, task.Severity, task.Type = "normal", "normal", "task"
		}
	}

	return quizmodel.Response{ID: "delivery-response", RequestID: request.ID, Provider: "fixture", Status: quizmodel.ResponseReady, PlanTracking: generationcore.GeneratedPlanTracking{Workflow: &input}, CreatedAt: request.CreatedAt}, nil
}

func webPlanDeliveryFixture(t *testing.T) (webViewProvider, *planDeliveryTestModel) {
	t.Helper()
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "codex")

	provider, _ := webPlanFixture(t)
	operations := provider.service(t.Context())
	model := &planDeliveryTestModel{}
	operations.quizModel = model
	provider.serviceFactory = func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }

	serviceWorkflowDeliveryGit(t, operations.projectContext.ProjectDir, "init", "-q")
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "workflow:\n  version: 1\nticket:\n  key_prefix: PRJ\n")
	addWebPlanDiscovery(t, provider)

	if _, err := operations.CreateSpec(CreateSpecRequest{OperationID: "delivery-spec", Title: "Greeting behavior", Body: "Print hello once and verify stdout.", Status: "draft", CaseSlug: "named-work"}); err != nil {
		t.Fatal(err)
	}

	return provider, model
}

func webPlanDeliveryForm(t *testing.T, provider webViewProvider) url.Values {
	t.Helper()

	catalog, err := provider.CatalogView(t.Context(), "plans")
	if err != nil || len(catalog.PlanChoices) != 1 {
		t.Fatal(catalog, err)
	}

	choice := catalog.PlanChoices[0]

	return url.Values{"action": {"plan.generate"}, "family": {"plans"}, "surface": {"plan"}, "case_slug": {choice.CaseSlug}, "id": {choice.CaseSlug}, "revision": {choice.Revision}, "request_id": {"one-plan-operation"}, "confirm": {"true"}}
}

// One HTTP operation produces native tickets and a bound invisible order map.
// Repeated submission returns the same work without another model invocation.
func TestWebPlanDelivery(t *testing.T) {
	provider, model := webPlanDeliveryFixture(t)
	handler := webQuizHandler(t, provider)
	page := webPlanGet(handler, "/plans").Body.String()

	if !strings.Contains(page, `data-modal-template="plan-generate-modal"`) || !strings.Contains(page, "Named work") {
		t.Fatal("Plans has no discoverable generation action", page)
	}

	if strings.Contains(webPlanGet(handler, "/cases/detail?id=named-work").Body.String(), "data-case-plan") {
		t.Fatal("plan generation remained hidden in Cases")
	}

	form := webPlanDeliveryForm(t, provider)
	response := webDumpEditorPost(handler, form)

	if !strings.Contains(response.Body.String(), "Plan ready with 1 tickets in 1 slices") || !strings.Contains(response.Body.String(), "View Tickets") {
		t.Fatal(response.Body.String())
	}

	checkWebPlanDelivery(t, provider)
	checkWebPlanDeliveryRetries(t, provider, model, form)
}

func checkWebPlanDeliveryRetries(t *testing.T, provider webViewProvider, model *planDeliveryTestModel, form url.Values) {
	t.Helper()

	handler := webQuizHandler(t, provider)

	response := webDumpEditorPost(handler, form)
	if !strings.Contains(response.Body.String(), `data-plan-write-status="ok"`) || len(model.base.requests) != 1 || model.calls != 1 {
		t.Fatal("retry duplicated generation", response.Body.String(), len(model.base.requests), model.calls)
	}

	form = webPlanDeliveryForm(t, provider)
	form.Set("request_id", "reopened-modal")

	response = webDumpEditorPost(handler, form)
	if !strings.Contains(response.Body.String(), `data-plan-write-status="ok"`) || model.calls != 1 {
		t.Fatal("reopening allocated the same plan again", response.Body.String(), model.calls)
	}
}

// Forged or stale selections fail before provider work and ticket allocation.
func TestWebPlanDeliverySelection(t *testing.T) {
	for _, field := range []string{"case_slug", "id", "revision", "request_id"} {
		t.Run(field, func(t *testing.T) {
			provider, model := webPlanDeliveryFixture(t)
			form := webPlanDeliveryForm(t, provider)
			form.Set(field, "")

			response := webDumpEditorPost(webQuizHandler(t, provider), form)
			if !strings.Contains(response.Body.String(), `data-plan-write-status="validation_failed"`) || len(model.base.requests) != 0 || model.calls != 0 {
				t.Fatal("unbound selection reached a provider", response.Body.String())
			}
		})
	}
}

// Two cases may use the same model-local labels while materializing independent
// native sets. Their tickets remain distinct and each plan has one catalog card.
func TestWebPlanDeliveryCases(t *testing.T) {
	provider, _ := webPlanDeliveryFixture(t)

	operations := provider.service(t.Context())
	if _, err := core.NewCaseMetadataService(operations.projectContext.ProjectDir).Save(core.CaseMetadata{Name: "Another greeting", Slug: "another-greeting", Description: "Print another greeting."}); err != nil {
		t.Fatal(err)
	}

	if _, err := operations.CreateSpec(CreateSpecRequest{OperationID: "another-spec", Title: "Another greeting", Body: "Print hello once.", Status: "approved", CaseSlug: "another-greeting"}); err != nil {
		t.Fatal(err)
	}

	if _, err := operations.CreateDump(DumpCreateRequest{OperationID: "another-dump", Title: "Greeting", Content: "Print hello once; stdout must equal hello.", Scope: DumpScope{CaseSlug: "another-greeting"}}); err != nil {
		t.Fatal(err)
	}

	for _, slug := range []string{"named-work", "another-greeting"} {
		if _, err := operations.GeneratePlanDelivery(t.Context(), PlanGenerateRequest{OperationID: "generate-" + slug, Scope: quizcore.Scope{CaseSlug: slug}}); err != nil {
			t.Fatal(slug, err)
		}
	}

	catalog, err := provider.CatalogView(t.Context(), "plans")
	if err != nil || len(catalog.Items) != 2 {
		t.Fatal("independent cases collided or duplicated cards", catalog, err)
	}
}

func checkWebPlanDelivery(t *testing.T, provider webViewProvider) {
	t.Helper()

	operations := provider.service(t.Context())

	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: operations.projectContext.ProjectDir})
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := core.NewWorkflowStore(layout, operations.projectContext.ProjectID).Snapshot(t.Context())
	if err != nil || len(snapshot.Records) != 2 {
		t.Fatal("missing native plan and ticket", snapshot, err)
	}

	catalog, err := provider.CatalogView(t.Context(), "plans")
	if err != nil || len(catalog.Items) != 1 {
		t.Fatal("materialization duplicated plan cards", catalog, err)
	}

	board := webPlanGet(webQuizHandler(t, provider), "/board").Body.String()
	if !strings.Contains(board, "Implement foundation") {
		t.Fatal("generated ticket missing from Board", board)
	}

	binding, err := core.ObservePlanRevision(t.Context(), operations.projectContext.ProjectDir, operations.quizScope(quizcore.Scope{CaseSlug: "named-work"}))
	if err != nil {
		t.Fatal(err)
	}

	checkWebPlanInternalRoutes(t, webQuizHandler(t, provider), []string{"ops/default/" + binding.CurrentTrackerRef, "ops/default/" + binding.TrackerRef})
}

// A provider failure after plan publication is reported as incomplete. The same
// operation recovers its plan and finishes allocation without regenerating it.
func TestWebPlanDeliveryRecovery(t *testing.T) {
	provider, model := webPlanDeliveryFixture(t)
	model.fail = true
	form := webPlanDeliveryForm(t, provider)
	handler := webQuizHandler(t, provider)
	response := webDumpEditorPost(handler, form)

	if !strings.Contains(response.Body.String(), "ticket creation is incomplete") {
		t.Fatal(response.Body.String())
	}

	model.fail = false

	response = webDumpEditorPost(handler, form)
	if !strings.Contains(response.Body.String(), `data-plan-write-status="ok"`) || len(model.base.requests) != 1 {
		t.Fatal("partial recovery regenerated the plan", response.Body.String(), len(model.base.requests))
	}

	checkWebPlanDelivery(t, provider)
}

// Real form submission shows pending work, then exposes the plan and tickets
// together. The browser does not need a second approval or tracking action.
func TestWebPlanDeliveryBrowser(t *testing.T) {
	if testing.Short() || os.Getenv("CHR_WEB_BROWSER") == "" {
		t.Skip("select the owned Chromium fixture explicitly")
	}

	provider, _ := webPlanDeliveryFixture(t)

	server := httptest.NewServer(webQuizHandler(t, provider))
	defer server.Close()

	browser := newWebWorkflowBrowser(t, os.Getenv("CHR_WEB_BROWSER"))
	browser.navigate(t, server.URL+"/plans")
	webBoardBrowserClick(t, browser, `[data-modal-template="plan-generate-modal"]`)
	browser.wait(t, "#plan-generate-form")
	webBoardBrowserClick(t, browser, "#plan-generate-form button[type='submit']")
	browser.wait(t, `[data-plan-write-status="ok"]`)

	if browser.evaluate(t, `String(document.querySelector('#plan-result').textContent.includes('View Tickets')&&!document.querySelector('[name="action"][value="plan.tracking.create"]'))`) != "true" {
		t.Fatal("browser workflow requires a second tracking step")
	}

	checkWebPlanDelivery(t, provider)
	webBoardBrowserClick(t, browser, `#plan-result a[href="/board"]`)
	browser.wait(t, "[data-ticket-key]")

	if browser.evaluate(t, `String(document.body.textContent.includes('Implement foundation'))`) != "true" {
		t.Fatal("browser cannot open generated tickets")
	}

	if response := webPlanGet(webQuizHandler(t, provider), "/healthz"); response.Code != http.StatusOK {
		t.Fatal(response.Code)
	}
}
