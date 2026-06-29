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

func TestHandlerTokenAccessPolicy(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data: testShellData(),
		Access: AccessPolicy{
			RequireToken: true,
			Token:        "secret-token",
			Origin:       "http://127.0.0.1:7331",
		},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/", nil))

	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want unauthorized", unauthorized.Code)
	}

	asset := httptest.NewRecorder()
	handler.ServeHTTP(asset, httptest.NewRequest(http.MethodGet, "/assets/chc.css", nil))

	if asset.Code != http.StatusOK {
		t.Fatalf("asset status = %d, want ok", asset.Code)
	}

	headerToken := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer secret-token")
	handler.ServeHTTP(headerToken, request)

	if headerToken.Code != http.StatusOK {
		t.Fatalf("header token status = %d, want ok", headerToken.Code)
	}

	queryToken := httptest.NewRecorder()
	handler.ServeHTTP(queryToken, httptest.NewRequest(http.MethodGet, "/?token=secret-token", nil))

	if queryToken.Code != http.StatusOK {
		t.Fatalf("query token status = %d, want ok", queryToken.Code)
	}
}

func TestHandlerOriginAccessPolicy(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data: testShellData(),
		Access: AccessPolicy{
			Origin:         "http://127.0.0.1:7331",
			AllowedOrigins: []string{"https://ops.example.test"},
		},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	blocked := httptest.NewRecorder()
	blockedRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	blockedRequest.Header.Set("Origin", "https://evil.example.test")
	handler.ServeHTTP(blocked, blockedRequest)

	if blocked.Code != http.StatusForbidden {
		t.Fatalf("blocked status = %d, want forbidden", blocked.Code)
	}

	sameOrigin := httptest.NewRecorder()
	sameOriginRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	sameOriginRequest.Header.Set("Origin", "http://127.0.0.1:7331")
	handler.ServeHTTP(sameOrigin, sameOriginRequest)

	if sameOrigin.Code != http.StatusOK {
		t.Fatalf("same-origin status = %d, want ok", sameOrigin.Code)
	}

	if got := sameOrigin.Header().Get("Access-Control-Allow-Origin"); got != "http://127.0.0.1:7331" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want same origin", got)
	}

	preflight := httptest.NewRecorder()
	preflightRequest := httptest.NewRequest(http.MethodOptions, "/actions", nil)
	preflightRequest.Header.Set("Origin", "https://ops.example.test")
	handler.ServeHTTP(preflight, preflightRequest)

	if preflight.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want no content", preflight.Code)
	}

	if got := preflight.Header().Get("Access-Control-Allow-Origin"); got != "https://ops.example.test" {
		t.Fatalf("preflight origin = %q, want allowed origin", got)
	}
}

func TestHandlerAllowsRequestHostOriginForWildcardBind(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data: testShellData(),
		Access: AccessPolicy{
			Origin: "http://0.0.0.0:7331",
		},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:7331/", nil)
	request.Header.Set("Origin", "http://127.0.0.1:7331")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want ok", recorder.Code)
	}
}

func TestHandlerRequestLoggingUsesSanitizedPath(t *testing.T) {
	var logs strings.Builder

	handler, err := NewHandler(HandlerOptions{
		Data:        testShellData(),
		LogRequests: &logs,
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?token=secret-token", nil))

	got := logs.String()
	for _, want := range []string{"web request:", "method=GET", "path=/", "status=200"} {
		if !strings.Contains(got, want) {
			t.Fatalf("logs do not contain %q:\n%s", want, got)
		}
	}

	for _, blocked := range []string{"secret-token", "token="} {
		if strings.Contains(got, blocked) {
			t.Fatalf("logs contain sensitive query data %q:\n%s", blocked, got)
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

func TestHandlerRendersGenerationAndWorkActionForms(t *testing.T) {
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
		{path: "/specs", want: []string{"spec.generate", "Generate Spec"}},
		{path: "/plans", want: []string{"plan.generate", "plan.regenerate", "plan.tracking.create"}},
		{path: "/quizzes", want: []string{"quiz.start", "quiz.resume", "Start Quiz"}},
		{path: "/dash", want: []string{"work.preview", "work.daemon.stop", "Preview Work"}},
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

func TestHandlerFullPagesConfigurePolling(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	for _, tt := range []struct {
		path      string
		refresh   string
		revision  string
		workspace string
	}{
		{path: "/", refresh: "/project/fragment", revision: "project-revision", workspace: `id="workspace"`},
		{path: "/board", refresh: "/board/fragment", revision: "board-revision", workspace: `id="workspace"`},
		{path: "/dash", refresh: "/dash/fragment", revision: "dash-revision", workspace: `id="workspace"`},
		{path: "/dumps", refresh: "/dumps/fragment", revision: "dumps-revision", workspace: `id="workspace"`},
		{
			path:      "/dumps/detail?id=.chc/user/dumps/project-notes.md",
			refresh:   "/dumps/detail/fragment?id=.chc%2Fuser%2Fdumps%2Fproject-notes.md",
			revision:  "detail-revision",
			workspace: `id="workspace"`,
		},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", tt.path, recorder.Code, http.StatusOK)
		}

		body := recorder.Body.String()
		for _, want := range []string{
			tt.workspace,
			`data-revision="` + tt.revision + `"`,
			`hx-get="` + tt.refresh + `"`,
			`hx-target="#workspace"`,
			`hx-swap="innerHTML"`,
			"every 2s",
			"document.visibilityState",
			"visible",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("%s body does not contain %q:\n%s", tt.path, want, body)
			}
		}
	}
}

func TestHandlerFragmentsCarryRevisionMarkers(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	for _, tt := range []struct {
		path     string
		revision string
	}{
		{path: "/project/fragment", revision: "project-revision"},
		{path: "/board/fragment", revision: "board-revision"},
		{path: "/dash/fragment", revision: "dash-revision"},
		{path: "/dumps/fragment", revision: "dumps-revision"},
		{path: "/dumps/detail/fragment?id=.chc/user/dumps/project-notes.md", revision: "detail-revision"},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, tt.path, nil)
		request.Header.Set(htmx.HeaderRequest, "true")
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", tt.path, recorder.Code, http.StatusOK)
		}

		body := recorder.Body.String()
		for _, want := range []string{`class="state-marker"`, `data-revision="` + tt.revision + `"`} {
			if !strings.Contains(body, want) {
				t.Fatalf("%s body does not contain %q:\n%s", tt.path, want, body)
			}
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

func TestHandlerWriteActionRendersStaleConflictActions(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	form := url.Values{
		"action":  {"dump.stale"},
		"id":      {".chc/user/dumps/project-notes.md"},
		"confirm": {"true"},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	for _, want := range []string{
		"Write action rejected stale data.",
		`href="/dumps/detail?id=.chc%2Fuser%2Fdumps%2Fproject-notes.md"`,
		">Reload</a>",
		`href="/dumps"`,
		">Discard</a>",
	} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Fatalf("body does not contain %q:\n%s", want, recorder.Body.String())
		}
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
		Revision: "project-revision",
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
		Revision: "board-revision",
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
		Revision: "dash-revision",
	}, nil
}

func (testDataProvider) CatalogView(_ context.Context, family string) (ArtifactCatalogView, error) {
	catalog := ArtifactCatalogView{
		Family:       family,
		Title:        CatalogTitle(family),
		EmptyMessage: "No " + family + " found.",
		Items:        testCatalogItems(family),
		Revision:     family + "-revision",
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
			Family:   request.Family,
			Title:    request.ID,
			Path:     request.ID,
			Missing:  true,
			Revision: "missing-revision",
		}, nil
	}

	return ArtifactDetailView{
		Family:   request.Family,
		Title:    "Project Notes",
		Status:   "curated",
		Scope:    "Project-wide",
		Path:     request.ID,
		Content:  "Full dump content",
		Revision: "detail-revision",
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

	if request.Action == "dump.stale" {
		return WriteActionResponse{
			Status:        WriteStatusStale,
			Summary:       "Write action rejected stale data.",
			BlockedReason: "The artifact changed after the page was loaded. Reload before saving.",
			ReloadURL:     "/dumps/detail?id=.chc%2Fuser%2Fdumps%2Fproject-notes.md",
			DiscardURL:    "/dumps",
		}, nil
	}

	return WriteActionResponse{
		Status:       WriteStatusOK,
		Summary:      "Saved dump.",
		ChangedFiles: []string{request.ID},
		RedirectURL:  "/dumps/detail?id=.chc%2Fuser%2Fdumps%2Fproject-notes.md",
	}, nil
}
