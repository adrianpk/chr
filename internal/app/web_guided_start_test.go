package app

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// TestNamedGuidedWeb starts only on submission and replays the same name safely.
func TestNamedGuidedWeb(t *testing.T) {
	provider, model := guidedWebFixture(t)
	handler := guidedWebHandler(t, provider)
	operations := provider.service(t.Context())

	assertGuidedNameOpening(t, handler, operations)

	form := url.Values{"surface": {"guided"}, "action": {"guided.start"}, "step": {"name"}, "confirm": {"true"}, "namespace": {"default"}, "request_id": {"named-web"}, "content": {"Improve checkout"}}
	for range 2 {
		response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", form)
		if response.Code != http.StatusSeeOther || !strings.HasPrefix(response.Header().Get("Location"), "/dumps?") {
			t.Fatal("name did not lead into the existing Dumps screen", response.Code, response.Body.String())
		}
	}

	session := readWebGuidedSession(t, operations, "named-web")
	if session.Topic != "Improve checkout" || session.Guided.Stage != "dump" || len(model.requests) != 0 {
		t.Fatal("name replay changed scope or started generation")
	}

	form.Set("content", "A different change")

	response := guidedCaptureHTTP(t, handler, http.MethodPost, "/guided/actions", form)
	if response.Code == http.StatusSeeOther || readWebGuidedSession(t, operations, "named-web").Topic != "Improve checkout" {
		t.Fatal("replayed request replaced the original use case")
	}
}

func assertGuidedNameOpening(t *testing.T, handler http.Handler, operations service) {
	t.Helper()

	before, err := operations.InspectGuidedUseCase(t.Context(), "", "default")
	if err != nil {
		t.Fatal(err)
	}

	page := guidedCaptureHTTP(t, handler, http.MethodGet, "/cases?new=workflow", nil)

	after, err := operations.InspectGuidedUseCase(t.Context(), "", "default")
	if err != nil || page.Code != http.StatusOK || len(before.Runs) != len(after.Runs) {
		t.Fatal("opening the name form created work", err)
	}
}
