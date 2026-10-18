package app

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

// Real HTTP observes two native contexts and retained legacy results without a
// TUI, a legacy Board dependency, an event phase or any lifecycle mutation.
func TestWebDashObservation(t *testing.T) {
	t.Run("native claims", func(t *testing.T) {
		operations, _ := nativeWorkObservationFixture(t)
		provider := observedDashProvider(t, operations)
		before := webPlanFiles(t, provider)

		view, err := provider.DashView(t.Context())
		if err != nil || len(view.Current) != 2 {
			t.Fatal(view, err)
		}

		response := webPlanGet(webQuizHandler(t, provider), "/dash")
		checkDashMarkup(t, response.Code, response.Body.String(), []string{"Current work", "Activity and results", "host-validated tasks", "confirmed tickets", "data-context-id"})

		if !reflect.DeepEqual(before, webPlanFiles(t, provider)) {
			t.Fatal("Dash mutated native claims")
		}
	})
	t.Run("retained legacy", func(t *testing.T) {
		operations := newTestService(t, newTestWorkspace(t))
		writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "work:\n  state_authority: tracker_compat\n  mode: no_git_enforcement\n")
		seedObservedWorkResults(t, operations)

		provider := observedDashProvider(t, operations)
		response := webPlanGet(webQuizHandler(t, provider), "/dash")
		checkDashMarkup(t, response.Code, response.Body.String(), []string{"Harness succeeded", "1 processed", "1 failed", "PRJ-running"})

		if strings.Contains(response.Body.String(), "No current work signal") {
			t.Fatal(response.Body.String())
		}
	})
}

func observedDashProvider(t *testing.T, operations service) webViewProvider {
	t.Helper()

	provider := newWebServeBoardTestProvider(operations.projectContext.ProjectDir, operations.env, operations.projectContext)
	provider.serviceFactory = func(ctx context.Context) service { copy := operations; copy.ctx = ctx; return copy }

	return provider
}

func checkDashMarkup(t *testing.T, status int, body string, wanted []string) {
	t.Helper()

	if status != 200 {
		t.Fatal(status, body)
	}

	for _, value := range wanted {
		if !strings.Contains(body, value) {
			t.Fatal(value, body)
		}
	}
}

// Observation timestamps cannot create new semantic revisions or mislabel a
// completed harness attempt as confirmed delivery.
func TestWebDashRevision(t *testing.T) {
	observation := WorkObservation{Current: []WorkContext{{ID: "one", Kind: "execution", TicketKey: "PRJ-1", Status: "running"}}}
	first := webDashObservation(observation, 0)
	observation.ObservedAt = observation.ObservedAt.Add(1000000000)

	second := webDashObservation(observation, 0)
	if first.Revision != second.Revision {
		t.Fatal("clock invalidated observation")
	}

	result := webWorkContext(WorkContext{Kind: "execution", Status: "succeeded", TicketKey: "PRJ-2"})
	if result.Status != "Harness succeeded" || result.Confirmed != 0 {
		t.Fatal(result)
	}
}
