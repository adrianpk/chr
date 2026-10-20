package app

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// Generated order maps remain internal even through direct generic URLs or a
// retained legacy tracking form. Rejected requests leave all files unchanged.
func TestWebPlanInternal(t *testing.T) {
	provider, model := webPlanFixture(t)
	addWebPlanDiscovery(t, provider)

	operations := provider.service(t.Context())
	if _, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{Scope: core.QuizScope{CaseSlug: "named-work"}}); err != nil {
		t.Fatal(err)
	}

	binding, err := core.ObservePlanRevision(t.Context(), provider.projectContext.ProjectDir, operations.quizScope(core.QuizScope{CaseSlug: "named-work"}))
	if err != nil {
		t.Fatal(err)
	}

	before := webPlanFiles(t, provider)
	requests := len(model.requests)
	handler := webQuizHandler(t, provider)
	checkWebPlanInternalRoutes(t, handler, []string{"ops/default/" + binding.CurrentTrackerRef, "ops/default/" + binding.TrackerRef})

	response := webPlanGet(handler, "/plans")
	checkWebPlanAbsent(t, response.Body.String(), []string{"Create Tracking", "plan.tracking.create", "Artifacts", "tracker/"})

	for _, action := range []string{"plan.tracking.create", "plan.save", "plan.append"} {
		response := webDumpEditorPost(handler, url.Values{"action": {action}, "family": {"plans"}, "case_slug": {"named-work"}, "confirm": {"true"}, "id": {"ops/default/" + binding.Reference}, "content": {"Attempted replacement"}})
		if !strings.Contains(response.Body.String(), "Unsupported web action") {
			t.Fatal(action, response.Body.String())
		}
	}

	if !reflect.DeepEqual(before, webPlanFiles(t, provider)) || len(model.requests) != requests {
		t.Fatal("tracker browsing or old form changed retained work")
	}
}

func checkWebPlanInternalRoutes(t *testing.T, handler http.Handler, references []string) {
	t.Helper()

	for _, reference := range references {
		for _, route := range []string{"/plans/detail", "/plans/detail/fragment", "/artifacts/detail", "/trackers/detail"} {
			response := webPlanGet(handler, route+"?family=plans&id="+url.QueryEscape(reference))
			if response.Code != http.StatusNotFound {
				t.Fatal(route, reference, response.Code, response.Body.String())
			}
		}
	}
}
