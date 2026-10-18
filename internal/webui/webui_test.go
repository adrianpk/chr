//nolint:wsl_v5 // UI tests keep request setup and response assertions adjacent.
package webui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/webui/htmx"
	"forge.adrianpk.com/chr/internal/webui/ui"
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

	for _, want := range []string{"Chryso", "project-1", "/assets/chr.css"} {
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

func TestHandlerShellCapitalizesProjectNameForDisplay(t *testing.T) {
	data := testShellData()
	data.ProjectName = "tictactoe"
	handler, err := NewHandler(HandlerOptions{Data: data})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), "Tictactoe") {
		t.Fatalf("body does not contain capitalized project name:\n%s", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), ">tictactoe<") {
		t.Fatalf("body contains raw lowercase project name:\n%s", recorder.Body.String())
	}
}

func TestHandlerAssetsAndHealthz(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{Data: testShellData()})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	for path, want := range map[string]string{
		"/assets/chr.css":      ".app-shell",
		"/assets/chr.js":       "openTemplateModal",
		"/assets/chr-logo.png": "",
		"/assets/favicon.svg":  "Chr",
		"/assets/htmx.min.js":  "htmx",
		"/healthz":             "ok\n",
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}

		if want != "" && !strings.Contains(recorder.Body.String(), want) {
			t.Fatalf("%s body does not contain %q", path, want)
		}
	}
}

func TestHandlerTicketModal(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	board := httptest.NewRecorder()
	handler.ServeHTTP(board, httptest.NewRequest(http.MethodGet, "/board", nil))

	if board.Code != http.StatusOK {
		t.Fatalf("board status = %d, want %d", board.Code, http.StatusOK)
	}

	if !strings.Contains(board.Body.String(), `hx-get="/board/ticket?key=CHR-1"`) {
		t.Fatalf("board does not contain ticket modal trigger:\n%s", board.Body.String())
	}

	detail := httptest.NewRecorder()
	handler.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/board/ticket?key=CHR-1", nil))

	if detail.Code != http.StatusOK {
		t.Fatalf("detail status = %d, want %d", detail.Code, http.StatusOK)
	}

	for _, want := range []string{
		"ticket-modal",
		"Ready work",
		"Full ticket description.",
		"A retained second paragraph.",
		"Comments",
		"Existing comment.",
		`name="action" value="ticket.update"`,
		`id="ticket-title" name="title" required`,
		`value="Ready work"`,
		`id="ticket-description" name="description" required`,
		`name="action" value="ticket.comment"`,
		`hx-post="/board/ticket/comment"`,
		`textarea name="content" required`,
	} {
		if !strings.Contains(detail.Body.String(), want) {
			t.Fatalf("detail body does not contain %q:\n%s", want, detail.Body.String())
		}
	}
	if strings.Contains(detail.Body.String(), `name="synopsis"`) {
		t.Fatal("modal exposed a competing synopsis editor")
	}
}

// Comment responses replace only discussion so dirty metadata remains in the DOM.
func TestHandlerTicketCommentRefreshesDiscussion(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	form := url.Values{
		"action":  {"ticket.comment"},
		"id":      {"CHR-1"},
		"confirm": {"true"},
		"content": {"Fresh modal comment."},
	}
	request := httptest.NewRequest(http.MethodPost, "/board/ticket/comment", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set(htmx.HeaderRequest, "true")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d:\n%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	if redirect := recorder.Header().Get(htmx.HeaderRedirect); redirect != "" {
		t.Fatalf("HX-Redirect = %q, want empty", redirect)
	}

	if strings.Contains(recorder.Body.String(), "ticket-edit-form") || strings.Contains(recorder.Body.String(), "ticket-modal") {
		t.Fatal("comment response would replace the editor")
	}

	for _, want := range []string{"data-ticket-commented", "Comments", "Existing comment."} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Fatalf("body does not contain %q:\n%s", want, recorder.Body.String())
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
	handler.ServeHTTP(asset, httptest.NewRequest(http.MethodGet, "/assets/chr.css", nil))

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

	hiddenDetails := []string{"Bind", "Refresh", "User workspace", ".chr/default/tickets", ".chr/default/runtime", "hx-trigger=", `href="/project"`}

	for _, tt := range []struct {
		path   string
		want   []string
		absent []string
	}{
		{path: "/", want: []string{"<h1>Project</h1>", "Project directory", "/tmp/project"}, absent: hiddenDetails},
		{path: "/project", want: []string{"<h1>Project</h1>", "Project name", "Project One", "Project directory", "/tmp/project", "Project ID", "project-1"}, absent: hiddenDetails},
		{path: "/board", want: []string{"<h1>Board</h1>", "Ready For Dev", "CHR-1", "Ready work"}},
		{path: "/dash", want: []string{"<h1>Dash</h1>", "Readiness", "Queue", "worker_started"}},
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

		for _, unwanted := range tt.absent {
			if strings.Contains(recorder.Body.String(), unwanted) {
				t.Fatalf("%s contains unwanted project detail %q", tt.path, unwanted)
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
		{path: "/cases", want: []string{"<h1>Cases</h1>", "case-table", "Blog Authoring", "Case context", "blog", "Edit", `hx-get="/cases/detail/fragment?id=blog"`}},
		{path: "/dumps", want: []string{"<h1>Dumps</h1>", "New context", "Scope", `data-modal-template="dump-create-modal"`, `id="dump-create-modal"`, "dump-card-link", "dump-card", "Project Notes", `hx-get="/dumps/detail/fragment?id=.chr/default/dumps/project-notes.md"`, "Project-wide", "Notes preview"}},
		{path: "/quizzes", want: []string{"<h1>Quizzes</h1>", "Project Quiz", "No answers yet"}},
		{path: "/specs", want: []string{"<h1>Specs</h1>", "Project Spec"}},
		{path: "/plans", want: []string{"<h1>Plans</h1>", "No plans found."}},
		{path: "/dumps/detail?id=.chr/default/dumps/project-notes.md", want: []string{"<h1>Project Notes</h1>", "Full dump content"}},
		{path: "/dumps/detail?id=missing.md", want: []string{"<h1>missing.md</h1>", "Artifact not found."}},
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

func TestHandlerCasesCatalogHidesSourcePath(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/cases", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if strings.Contains(recorder.Body.String(), ".chr/user/cases/blog.md") {
		t.Fatalf("cases catalog exposes source path:\n%s", recorder.Body.String())
	}
}

func TestHandlerDumpsCreateUsesModalTemplates(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/dumps", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := strings.Count(recorder.Body.String(), `name="action" value="dump.create"`); got != 1 {
		t.Fatalf("capture form count = %d, want one", got)
	}

	for _, unwanted := range []string{`href="#dump-create"`, `href="#dump-create-case"`, `id="dump-create-case-modal"`, "For Use Case", `<input name="case_slug"`} {
		if strings.Contains(recorder.Body.String(), unwanted) {
			t.Fatalf("dumps catalog contains inline create trigger %q:\n%s", unwanted, recorder.Body.String())
		}
	}
}

// TestHandlerDumpDetailFragmentUsesModal keeps editing in the catalog dialog.
func TestHandlerDumpDetailFragmentUsesModal(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/dumps/detail/fragment?id=.chr/default/dumps/project-notes.md", nil)
	request.Header.Set(htmx.HeaderRequest, "true")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	for _, want := range []string{
		"modal-backdrop",
		"dump-detail-panel",
		"Project Notes",
		"Full dump content",
		"dump.save",
		`name="id" value=".chr/default/dumps/project-notes.md"`,
		`name="revision" value="detail-revision"`,
		`hx-post="/actions"`,
		`name="surface" value="modal"`,
	} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Fatalf("body does not contain %q:\n%s", want, recorder.Body.String())
		}
	}

	for _, forbidden := range []string{"dump.append", "artifact-content", "Append note"} {
		if strings.Contains(recorder.Body.String(), forbidden) {
			t.Fatalf("dump editor contains duplicate surface %q", forbidden)
		}
	}
}

//nolint:cyclop // The test verifies the complete page and modal editing behavior together.
func TestHandlerCaseEditUsesSharedPanel(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/cases/detail?id=blog", nil))

	if page.Code != http.StatusOK {
		t.Fatalf("page status = %d, want %d", page.Code, http.StatusOK)
	}

	for _, want := range []string{"case-edit-panel", "Edit Case", "Blog Authoring", "Case context", "case.update"} {
		if !strings.Contains(page.Body.String(), want) {
			t.Fatalf("page body does not contain %q:\n%s", want, page.Body.String())
		}
	}

	for _, unwanted := range []string{".chr/user/cases/blog.md", "artifact-content"} {
		if strings.Contains(page.Body.String(), unwanted) {
			t.Fatalf("page body contains %q:\n%s", unwanted, page.Body.String())
		}
	}
	if strings.Contains(page.Body.String(), `name="surface" value="modal"`) {
		t.Fatalf("page body contains modal surface field:\n%s", page.Body.String())
	}

	modal := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/cases/detail/fragment?id=blog", nil)
	request.Header.Set(htmx.HeaderRequest, "true")
	handler.ServeHTTP(modal, request)

	if modal.Code != http.StatusOK {
		t.Fatalf("modal status = %d, want %d", modal.Code, http.StatusOK)
	}

	for _, want := range []string{
		"modal-backdrop",
		"case-edit-panel",
		"Edit Case",
		"Blog Authoring",
		"Case context",
		"case.update",
		`hx-post="/actions"`,
		`name="surface" value="modal"`,
	} {
		if !strings.Contains(modal.Body.String(), want) {
			t.Fatalf("modal body does not contain %q:\n%s", want, modal.Body.String())
		}
	}
}

func TestHandlerCaseModalSaveClosesModal(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	form := url.Values{
		"action":      {"case.update"},
		"family":      {"cases"},
		"id":          {"blog"},
		"confirm":     {"true"},
		"surface":     {"modal"},
		"name":        {"Updated Blog"},
		"description": {"Updated description."},
		"context":     {"Updated context."},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set(htmx.HeaderRequest, "true")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get(htmx.HeaderRetarget); got != "#modal-root" {
		t.Fatalf("HX-Retarget = %q, want #modal-root", got)
	}
	if got := recorder.Header().Get(htmx.HeaderReswap); got != string(htmx.SwapInner) {
		t.Fatalf("HX-Reswap = %q, want %q", got, htmx.SwapInner)
	}
	if got := recorder.Header().Get(htmx.HeaderRedirect); got != "" {
		t.Fatalf("HX-Redirect = %q, want empty", got)
	}
	if recorder.Body.String() != "" {
		t.Fatalf("body = %q, want empty", recorder.Body.String())
	}
}

// TestDumpModalWrites keeps successful dump creation, replacement and append operations in the catalog.
func TestDumpModalWrites(t *testing.T) {
	handler, err := NewHandler(HandlerOptions{
		Data:     testShellData(),
		Provider: testDataProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	for _, action := range []string{"dump.create", "dump.save", "dump.append"} {
		t.Run(action, func(t *testing.T) {
			form := url.Values{
				"action":  {action},
				"id":      {".chr/default/dumps/project-notes.md"},
				"confirm": {"true"},
				"surface": {"modal"},
				"content": {"Updated content"},
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set(htmx.HeaderRequest, "true")
			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if got := recorder.Header().Get(htmx.HeaderRetarget); got != "#modal-root" {
				t.Fatalf("HX-Retarget = %q, want #modal-root", got)
			}
			if got := recorder.Header().Get(htmx.HeaderRedirect); got != "" {
				t.Fatalf("HX-Redirect = %q, want empty", got)
			}
			if recorder.Body.String() != "" {
				t.Fatalf("body = %q, want empty", recorder.Body.String())
			}
		})
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
		{path: "/specs", want: []string{"New", `hx-get="/specs/detail/fragment"`, "document-card"}},
		{path: "/plans", want: []string{"plan.tracking.create"}},
		{path: "/quizzes", want: []string{"New Quiz", `aria-label="Pending quizzes"`, "/quizzes/detail?id=project"}},
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
		"/dumps/detail/fragment?id=.chr/default/dumps/project-notes.md",
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
		{path: "/board", refresh: "/board/fragment", revision: "board-revision", workspace: `id="workspace"`},
		{path: "/dash", refresh: "/dash/fragment", revision: "dash-revision", workspace: `id="workspace"`},
		{path: "/dumps", refresh: "/dumps/fragment", revision: "dumps-revision", workspace: `id="workspace"`},
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
		{path: "/dumps/detail/fragment?id=.chr/default/dumps/project-notes.md", revision: "detail-revision"},
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
		"id":      {".chr/default/dumps/project-notes.md"},
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

	if got := recorder.Header().Get("Location"); got != "/dumps/detail?id=.chr%2Fdefault%2Fdumps%2Fproject-notes.md" {
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
		"id":      {".chr/default/dumps/project-notes.md"},
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
		`href="/dumps/detail?id=.chr%2Fdefault%2Fdumps%2Fproject-notes.md"`,
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
		Title:          "Chryso",
		ProjectDir:     "/tmp/project",
		ProjectID:      "project-1",
		ProjectName:    "Project One",
		URL:            "http://127.0.0.1:7331",
		Host:           "127.0.0.1",
		Port:           7331,
		RequireToken:   "auto",
		RefreshSeconds: 2,
		Navigation: []ui.NavLink{
			{Label: "Cases", Href: "/cases"},
		},
	}
}

type testDataProvider struct{}

func (testDataProvider) ProjectView(context.Context) (ProjectView, error) {
	return ProjectView{
		ProjectID:  "project-1",
		ProjectDir: "/tmp/project",
		Revision:   "project-revision",
	}, nil
}

func (testDataProvider) BoardView(context.Context, BoardRequest) (BoardView, error) {
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
					{Key: "CHR-1", Title: "Ready work", Status: "ready_for_dev", Priority: "high", Type: "feature"},
				},
			},
		},
		Revision: "board-revision",
	}, nil
}

func (testDataProvider) TicketDetailView(_ context.Context, request TicketDetailRequest) (TicketDetailView, error) {
	return TicketDetailView{
		Key:          request.Key,
		Title:        "Ready work",
		Synopsis:     "Ready work synopsis.",
		Description:  "Full ticket description.\n\nA retained second paragraph.",
		Priority:     "high",
		Status:       "ready_for_dev",
		Type:         "feature",
		ProjectArea:  "backend",
		RelativePath: "CHR-1-ready-work.md",
		Comments: []TicketCommentView{
			{Author: "web", CreatedAt: "18 May 13:00", Body: "Existing comment."},
		},
		Revision: "ticket-detail-revision",
	}, nil
}

func (testDataProvider) DashView(context.Context) (DashView, error) {
	return DashView{
		Readiness: WorkReadinessView{Status: "ready", WorkspaceMode: "worktree", CurrentBranch: "dev"},
		Queue: AgentQueueView{
			EligibleCount: 1,
			Tickets: []BoardTicketView{
				{Key: "CHR-1", Title: "Ready work", Status: "ready_for_dev", Priority: "high", Type: "feature"},
			},
		},
		Current: CurrentWorkView{Phase: "worker_started", Message: "worker started", TicketKey: "CHR-1"},
		RecentEvents: []AgentEventView{
			{Time: "12:00:00", Phase: "worker_started", Message: "worker started", TicketKey: "CHR-1"},
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

	if family == "quizzes" {
		catalog.PendingQuizzes = []QuizScopeChoice{{Title: "Project Quiz", URL: "/quizzes/detail?id=project"}}
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
				Path:         ".chr/default/cases/blog.md",
				Preview:      "Case context",
				DetailURL:    "/cases/detail?id=blog",
				ContentState: "available",
			},
		}
	case "dumps":
		return []ArtifactCardView{
			{
				Family:       family,
				ID:           ".chr/default/dumps/project-notes.md",
				Title:        "Project Notes",
				DisplayTitle: "Project Notes",
				Status:       "curated",
				Scope:        "Project-wide",
				Meta:         "Project-wide",
				Path:         ".chr/default/dumps/project-notes.md",
				Preview:      "Notes preview",
				DetailURL:    "/dumps/detail?id=.chr/default/dumps/project-notes.md",
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
				Preview:      "",
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
	if request.Family == "cases" && request.ID == "blog" {
		return ArtifactDetailView{
			Family:      request.Family,
			ID:          request.ID,
			Title:       "Blog Authoring",
			Description: "Case context",
			Context:     "Draft blog posts with review.",
			Status:      "available",
			Scope:       "blog",
			Path:        ".chr/default/cases/blog.md",
			Revision:    "case-detail-revision",
		}, nil
	}

	if request.ID != ".chr/default/dumps/project-notes.md" {
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
		ID:       request.ID,
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
			ReloadURL:     "/dumps/detail?id=.chr%2Fdefault%2Fdumps%2Fproject-notes.md",
			DiscardURL:    "/dumps",
		}, nil
	}

	return WriteActionResponse{
		Status:       WriteStatusOK,
		Summary:      "Saved dump.",
		ChangedFiles: []string{request.ID},
		RedirectURL:  "/dumps/detail?id=.chr%2Fdefault%2Fdumps%2Fproject-notes.md",
	}, nil
}
