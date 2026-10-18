package app

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/webui"
)

// The Board must expose the same plan-bound title guard that the order owner
// enforces; an offered rename cannot fail as if an unchanged form were stale.
func TestWebConsolidatedTitleGuard(t *testing.T) {
	if testing.Short() {
		t.Skip("real approved-plan fixture is selected explicitly")
	}

	j := completedWebJourney(t)

	detail, err := j.provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: j.ticket})
	if err != nil || detail.ReadOnly["title"] == "" {
		t.Fatal("bound title lacks its order ownership guard", detail.ReadOnly, err)
	}

	before := journeyMapFiles(t, j.provider.projectContext.ProjectDir)
	form := url.Values{"action": {"ticket.update"}, "id": {j.ticket}, "revision": {detail.Revision}, "confirm": {"true"}, "request_id": {detail.RequestID}, "title": {"Unapproved renamed order"}}

	response := webDumpEditorPost(j.handler, form)
	if !strings.Contains(response.Body.String(), "title is read-only") || strings.Contains(response.Body.String(), `data-write-status="stale"`) {
		t.Fatal("bound rename did not explain its guard", response.Body.String())
	}

	if !reflect.DeepEqual(before, journeyMapFiles(t, j.provider.projectContext.ProjectDir)) {
		t.Fatal("rejected title write changed the retained order")
	}
}

func completedWebJourney(t *testing.T) *webJourney {
	t.Helper()

	j := newWebJourney(t)
	j.discovery(t)
	j.quiz(t)
	j.authorSpecs(t)
	j.planAndBind(t)
	j.boardAndObserve(t)

	return j
}
