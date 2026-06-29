package webui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"forge.adrianpk.com/chc/internal/webui/htmx"
	"forge.adrianpk.com/chc/internal/webui/ui"
)

func TestHandlerShellAndBootstrap(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{Data: testShellData()})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	for _, want := range []string{"ChrysoCoord", "project-1", "/assets/chc.css"} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Fatalf("body does not contain %q:\n%s", want, recorder.Body.String())
		}
	}

	fragment := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/bootstrap", nil)
	request.Header.Set(htmx.HeaderRequest, "true")
	handler.ServeHTTP(fragment, request)

	if fragment.Code != http.StatusOK {
		t.Fatalf("fragment status = %d, want %d", fragment.Code, http.StatusOK)
	}

	if strings.Contains(fragment.Body.String(), "<html") {
		t.Fatalf("fragment contains full html:\n%s", fragment.Body.String())
	}
}

func TestHandlerAssetsAndHealthz(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{Data: testShellData()})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	for path, want := range map[string]string{
		"/assets/chc.css":     ".app-shell",
		"/assets/htmx.min.js": "htmx",
		"/healthz":            "ok\n",
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}

		if !strings.Contains(recorder.Body.String(), want) {
			t.Fatalf("%s body does not contain %q", path, want)
		}
	}
}

func TestHandlerProjectBoardDashRoutes(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	for _, tt := range []struct {
		path string
		want []string
	}{
		{path: "/", want: []string{"Project", "Project directory", "/tmp/project"}},
		{path: "/project", want: []string{"Project", "Tickets", ".chc/user/tickets"}},
		{path: "/board", want: []string{"Board", "Ready For Dev", "CHC-1", "Ready work"}},
		{path: "/dash", want: []string{"Dash", "Readiness", "Queue", "worker_started"}},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", tt.path, recorder.Code, http.StatusOK)
		}

		for _, want := range tt.want {
			if !strings.Contains(recorder.Body.String(), want) {
				t.Fatalf("%s body does not contain %q:\n%s", tt.path, want, recorder.Body.String())
			}
		}
	}
}

func TestHandlerArtifactCatalogAndDetailRoutes(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	for _, tt := range []struct {
		path string
		want []string
	}{
		{path: "/cases", want: []string{"Cases", "Blog Authoring", "Case context"}},
		{path: "/dumps", want: []string{"Dumps", "Project Notes", ".chc/user/dumps/project-notes.md"}},
		{path: "/quizzes", want: []string{"Quizzes", "Project Quiz", "not_started"}},
		{path: "/specs", want: []string{"Specs", "Project Spec"}},
		{path: "/plans", want: []string{"Plans", "No plans found."}},
		{path: "/dumps/detail?id=.chc/user/dumps/project-notes.md", want: []string{"Project Notes", "Full dump content"}},
		{path: "/dumps/detail?id=missing.md", want: []string{"missing.md", "Artifact not found."}},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", tt.path, recorder.Code, http.StatusOK)
		}

		for _, want := range tt.want {
			if !strings.Contains(recorder.Body.String(), want) {
				t.Fatalf("%s body does not contain %q:\n%s", tt.path, want, recorder.Body.String())
			}
		}
	}
}

func TestHandlerBoardAndDashFragments(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	for _, path := range []string{"/project/fragment", "/board/fragment", "/dash/fragment"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set(htmx.HeaderRequest, "true")
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}

		if strings.Contains(recorder.Body.String(), "<html") {
			t.Fatalf("%s fragment contains full html:\n%s", path, recorder.Body.String())
		}
	}
}

func TestHandlerArtifactCatalogAndDetailFragments(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	for _, path := range []string{
		"/cases/fragment",
		"/dumps/fragment",
		"/quizzes/fragment",
		"/specs/fragment",
		"/plans/fragment",
		"/dumps/detail/fragment?id=.chc/user/dumps/project-notes.md",
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set(htmx.HeaderRequest, "true")
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}

		if strings.Contains(recorder.Body.String(), "<html") {
			t.Fatalf("%s fragment contains full html:\n%s", path, recorder.Body.String())
		}
	}
}

func TestHandlerWriteActionRedirectsAfterSuccess(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	form := url.Values{
		"action":  {"dump.save"},
		"id":      {".chc/user/dumps/project-notes.md"},
		"confirm": {"true"},
		"content": {"Updated content"},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}

	if got := recorder.Header().Get("Location"); got != "/dumps/detail?id=.chc%2Fuser%2Fdumps%2Fproject-notes.md" {
		t.Fatalf("Location = %q, want dump detail", got)
	}
}

func TestHandlerWriteActionRendersBlockedEnvelope(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	form := url.Values{"action": {"dump.save"}}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	for _, want := range []string{"Write action blocked.", "Action confirmation is required."} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Fatalf("body does not contain %q:\n%s", want, recorder.Body.String())
		}
	}
}

func testShellData() ShellData {
	return ShellData{
		Title:          "ChrysoCoord",
		ProjectDir:     "/tmp/project",
		ProjectID:      "project-1",
		URL:            "http://127.0.0.1:7331",
		Host:           "127.0.0.1",
		Port:           7331,
		RequireToken:   "auto",
		RefreshSeconds: 2,
		Navigation: []ui.NavLink{
			{Label: "Project", Href: "/", Active: true},
		},
	}
}

type testDataProvider struct{}

func (testDataProvider) ProjectView(context.Context) (ProjectView, error) {
	return ProjectView{
		ProjectID:      "project-1",
		ProjectDir:     "/tmp/project",
		URL:            "http://127.0.0.1:7331",
		Host:           "127.0.0.1",
		Port:           7331,
		RequireToken:   "auto",
		RefreshSeconds: 2,
		Paths: []ProjectPathView{
			{Label: "Tickets", Path: "/tmp/project/.chc/user/tickets"},
		},
	}, nil
}

func (testDataProvider) BoardView(context.Context) (BoardView, error) {
	return BoardView{
		ProjectID:      "project-1",
		TotalCount:     1,
		NormalCount:    1,
		AttentionCount: 0,
		Columns: []BoardColumnView{
			{
				ID:    "ready_for_dev",
				Title: "Ready For Dev",
				Count: 1,
				Tickets: []BoardTicketView{
					{Key: "CHC-1", Title: "Ready work", Status: "ready_for_dev", Priority: "high", Type: "feature"},
				},
			},
		},
	}, nil
}

func (testDataProvider) DashView(context.Context) (DashView, error) {
	return DashView{
		Readiness: WorkReadinessView{Status: "ready", WorkspaceMode: "worktree", CurrentBranch: "dev"},
		Queue: AgentQueueView{
			EligibleCount: 1,
			Tickets: []BoardTicketView{
				{Key: "CHC-1", Title: "Ready work", Status: "ready_for_dev", Priority: "high", Type: "feature"},
			},
		},
		Current: CurrentWorkView{Phase: "worker_started", Message: "worker started", TicketKey: "CHC-1"},
		RecentEvents: []AgentEventView{
			{Time: "12:00:00", Phase: "worker_started", Message: "worker started", TicketKey: "CHC-1"},
		},
	}, nil
}

func (testDataProvider) CatalogView(_ context.Context, family string) (ArtifactCatalogView, error) {
	catalog := ArtifactCatalogView{
		Family:       family,
		Title:        CatalogTitle(family),
		EmptyMessage: "No " + family + " found.",
		Items:        testCatalogItems(family),
	}

	return catalog, nil
}

func testCatalogItems(family string) []ArtifactCardView {
	switch family {
	case "cases":
		return []ArtifactCardView{
			{
				Family:       family,
				ID:           "blog",
				Title:        "Blog Authoring",
				Status:       "available",
				Scope:        "blog",
				Path:         ".chc/user/cases/blog.md",
				Preview:      "Case context",
				DetailURL:    "/cases/detail?id=blog",
				ContentState: "available",
			},
		}
	case "dumps":
		return []ArtifactCardView{
			{
				Family:       family,
				ID:           ".chc/user/dumps/project-notes.md",
				Title:        "Project Notes",
				Status:       "curated",
				Scope:        "Project-wide",
				Path:         ".chc/user/dumps/project-notes.md",
				Preview:      "Notes preview",
				DetailURL:    "/dumps/detail?id=.chc/user/dumps/project-notes.md",
				ContentState: "available",
			},
		}
	case "quizzes":
		return []ArtifactCardView{
			{
				Family:       family,
				ID:           "project",
				Title:        "Project Quiz",
				Status:       "not_started",
				Scope:        "Project-wide",
				Preview:      "No quiz yet.",
				DetailURL:    "/quizzes/detail?id=project",
				ContentState: "missing",
			},
		}
	case "specs":
		return []ArtifactCardView{
			{
				Family:       family,
				ID:           "project-spec",
				Title:        "Project Spec",
				Status:       "missing",
				Scope:        "Project-wide",
				DetailURL:    "/specs/detail?id=project-spec",
				ContentState: "missing",
			},
		}
	default:
		return nil
	}
}

func (testDataProvider) ArtifactDetailView(_ context.Context, request ArtifactDetailRequest) (ArtifactDetailView, error) {
	if request.ID != ".chc/user/dumps/project-notes.md" {
		return ArtifactDetailView{
			Family:  request.Family,
			Title:   request.ID,
			Path:    request.ID,
			Missing: true,
		}, nil
	}

	return ArtifactDetailView{
		Family:  request.Family,
		Title:   "Project Notes",
		Status:  "curated",
		Scope:   "Project-wide",
		Path:    request.ID,
		Content: "Full dump content",
	}, nil
}

func (testDataProvider) WriteAction(_ context.Context, request WriteActionRequest) (WriteActionResponse, error) {
	if !request.Confirm {
		return WriteActionResponse{
			Status:        WriteStatusBlocked,
			Summary:       "Write action blocked.",
			BlockedReason: "Action confirmation is required.",
		}, nil
	}

	return WriteActionResponse{
		Status:       WriteStatusOK,
		Summary:      "Saved dump.",
		ChangedFiles: []string{request.ID},
		RedirectURL:  "/dumps/detail?id=.chc%2Fuser%2Fdumps%2Fproject-notes.md",
	}, nil
}
