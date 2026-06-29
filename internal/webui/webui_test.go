package webui

import (
	"net/http"
	"net/http/httptest"
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
