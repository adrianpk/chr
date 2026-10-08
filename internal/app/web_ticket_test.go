package app

import (
	"os"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/webui"
)

// Ticket writes must preserve newer file contents until the browser reloads its revision.
func TestWebTicketRevision(t *testing.T) {
	for _, tt := range []struct {
		name     string
		action   string
		revision string
		status   string
	}{
		{name: "fresh edit", action: "ticket.update", revision: "fresh", status: webui.WriteStatusOK},
		{name: "stale edit", action: "ticket.update", revision: "stale", status: webui.WriteStatusStale},
		{name: "missing revision", action: "ticket.update", status: webui.WriteStatusValidation},
		{name: "fresh comment", action: "ticket.comment", revision: "fresh", status: webui.WriteStatusOK},
		{name: "stale comment", action: "ticket.comment", revision: "stale", status: webui.WriteStatusStale},
		{name: "missing comment revision", action: "ticket.comment", status: webui.WriteStatusValidation},
	} {
		t.Run(tt.name, func(t *testing.T) {
			provider, original := webTicketWriteFixture(t)

			current, err := provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: original.Key})
			if err != nil {
				t.Fatal(err)
			}

			revisions := map[string]string{"fresh": current.Revision, "stale": original.Revision}
			before := readWebTicketFile(t, provider, original.Key)

			result, err := provider.WriteAction(t.Context(), webui.WriteActionRequest{
				Action: tt.action, ID: original.Key, Revision: revisions[tt.revision], Confirm: true,
				Fields: map[string]string{"title": "Browser title", "synopsis": "Browser description.", "content": "Browser comment."},
			})
			if err != nil || result.Status != tt.status {
				t.Fatalf("WriteAction = %#v, %v; want %q", result, err, tt.status)
			}

			after := readWebTicketFile(t, provider, original.Key)
			assertWebTicketWrite(t, tt.action, result, before, after)
		})
	}
}

func webTicketWriteFixture(t *testing.T) (webViewProvider, webui.TicketDetailView) {
	t.Helper()

	projectDir := newTestWorkspace(t)
	writeWebServeBoardTestTicket(t, projectDir)

	provider := newWebServeBoardTestProvider(projectDir, newWebServeBoardTestEnv(t, t.TempDir()), newWebServeBoardProjectContext(t, projectDir))

	original, err := provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: "CHR-1"})
	if err != nil {
		t.Fatal(err)
	}

	result, err := provider.WriteAction(t.Context(), webui.WriteActionRequest{
		Action: "ticket.update", ID: original.Key, Revision: original.Revision, Confirm: true,
		Fields: map[string]string{"title": "Newer title", "synopsis": "Newer description."},
	})
	if err != nil || result.Status != webui.WriteStatusOK {
		t.Fatalf("newer edit = %#v, %v; want ok", result, err)
	}

	return provider, original
}

func readWebTicketFile(t *testing.T, provider webViewProvider, key string) string {
	t.Helper()

	ticket, err := provider.service(t.Context()).ticketByKey(key)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(ticket.Path)
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

func assertWebTicketWrite(t *testing.T, action string, result webui.WriteActionResponse, before, after string) {
	t.Helper()

	if result.Status != webui.WriteStatusOK {
		if after != before {
			t.Fatal("rejected write changed the ticket file")
		}

		if result.Status == webui.WriteStatusStale && (result.ReloadURL == "" || result.DiscardURL == "") {
			t.Fatal("stale write did not provide reload and discard actions")
		}

		return
	}

	want := `title: "Browser title"`
	if action == "ticket.comment" {
		want = "Browser comment."
	}

	if !strings.Contains(after, want) {
		t.Fatalf("saved ticket does not contain %q", want)
	}
}
