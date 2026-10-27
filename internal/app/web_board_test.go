package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/webui"
)

// Board reads must work after ops migration without requiring the obsolete directory.
func TestWebBoardSources(t *testing.T) {
	for _, tc := range []struct {
		name              string
		canonical, legacy bool
		want              int
	}{
		{name: "canonical", canonical: true, want: 1},
		{name: "legacy", legacy: true, want: 1},
		{name: "mixed", canonical: true, legacy: true, want: 2},
		{name: "empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project := newTestWorkspace(t)
			if tc.canonical {
				writeWebBoardCanonicalTicket(t, project)
			}

			if tc.legacy {
				writeWebServeBoardTestTicket(t, project)
			}

			provider := newWebServeBoardTestProvider(project, newWebServeBoardTestEnv(t, t.TempDir()), newWebServeBoardProjectContext(t, project))

			view, err := provider.BoardView(t.Context(), webui.BoardRequest{})
			if err != nil || view.TotalCount != tc.want || len(view.Columns) != 6 {
				t.Fatalf("board = %#v, %v; want %d tickets in six lanes", view, err, tc.want)
			}

			checkWebBoardRoutes(t, provider)

			_, err = os.Stat(workspacecore.UserPath(project, "tickets"))
			if !tc.legacy && !os.IsNotExist(err) {
				t.Fatal("read created the legacy ticket directory", err)
			}
		})
	}
}

func checkWebBoardRoutes(t *testing.T, provider webViewProvider) {
	t.Helper()

	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/board", "/board/fragment"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("HX-Request", "true")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "Native workflow tickets") {
			t.Fatalf("%s: %d %s", path, response.Code, response.Body.String())
		}
	}
}

// Continuation must render the requested page and keep polling on that page.
func TestWebBoardPaging(t *testing.T) {
	project := newTestWorkspace(t)

	for i := 1; i <= workflowstate.PageLimit+2; i++ {
		writeServiceWorkspaceFile(t, filepath.Join(project, "ops", "default", "ticket", fmt.Sprintf("CHR-%d.md", i)), fmt.Sprintf("---\nkey: CHR-%d\ntitle: Ticket %d\nstatus: backlog\n---\nBody.\n", i, i))
	}

	provider := newWebServeBoardTestProvider(project, newWebServeBoardTestEnv(t, t.TempDir()), newWebServeBoardProjectContext(t, project))

	first, err := provider.BoardView(t.Context(), webui.BoardRequest{})
	if err != nil || first.TotalCount != 130 || first.LoadedCount != 128 || first.NextURL != "/board?offset=128" {
		t.Fatal("first page lost continuation", first, err)
	}

	again, err := provider.BoardView(t.Context(), webui.BoardRequest{})
	if err != nil || again.Revision != first.Revision {
		t.Fatal("unchanged Board revision drifted", err)
	}

	checkWebBoardContinuation(t, provider, first.NextURL)
}

func checkWebBoardContinuation(t *testing.T, provider webViewProvider, next string) {
	t.Helper()

	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{next, "/board/fragment?offset=128"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("HX-Request", "true")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusOK || strings.Count(response.Body.String(), `class="ticket-card"`) != 2 || !strings.Contains(response.Body.String(), "2 loaded · 130 total") {
			t.Fatalf("second page = %d %s", response.Code, response.Body.String())
		}

		if path == next && !strings.Contains(response.Body.String(), `hx-get="/board/fragment?offset=128"`) {
			t.Fatal("refresh lost the current page", response.Body.String())
		}
	}
}

// Storage errors must remain unavailable responses rather than empty Boards.
func TestWebBoardFailure(t *testing.T) {
	project := newTestWorkspace(t)
	writeServiceWorkspaceFile(t, workspacecore.UserPath(project, "tickets"), "not a directory")

	provider := newWebServeBoardTestProvider(project, newWebServeBoardTestEnv(t, t.TempDir()), newWebServeBoardProjectContext(t, project))
	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/board", "/board/fragment"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("HX-Request", "true")

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("storage error became %d %s", response.Code, response.Body.String())
		}
	}
}

func writeWebBoardCanonicalTicket(t *testing.T, project string) {
	t.Helper()

	content, err := ticketcore.NewTicketRenderer().RenderTicket(ticketcore.Ticket{Key: "CHR-2", Title: "Canonical ticket", Synopsis: "Retained in ops.", Status: ticketcore.TicketStatusReadyForDev, Type: "task"})
	if err != nil {
		t.Fatal(err)
	}

	writeServiceWorkspaceFile(t, filepath.Join(project, "ops", "default", "ticket", "CHR-2.md"), content)
}

// Conditional polling returns no replacement for the same semantic observation.
func TestWebBoardConditional(t *testing.T) {
	provider := webTicketEditorProvider(t, false)
	writeWebBoardCanonicalTicket(t, provider.projectContext.ProjectDir)

	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	first := webWorkflowHTTP(t, handler, http.MethodGet, "/board/fragment", nil)
	request := httptest.NewRequest(http.MethodGet, "/board/fragment", nil)
	request.Header.Set("HX-Request", "true")
	request.Header.Set("If-None-Match", first.Header().Get("ETag"))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if first.Header().Get("ETag") == "" || response.Code != http.StatusNotModified || response.Body.Len() != 0 {
		t.Fatal("unchanged poll replaced Board", response.Code, response.Body.String())
	}

	writeServiceWorkspaceFile(t, filepath.Join(provider.projectContext.ProjectDir, "ops", "default", "ticket", "new.md"), "---\nkey: CHR-3\ntitle: Changed observation\nstatus: backlog\n---\nBody.\n")

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Header().Get("ETag") == first.Header().Get("ETag") || !strings.Contains(response.Body.String(), "Changed observation") {
		t.Fatal("changed poll suppressed new ticket", response.Code, response.Body.String())
	}
}
