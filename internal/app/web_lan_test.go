package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/webui"
)

// One explicit flag must expose the UI without creating or supplying a token.
func TestWebLANCLI(t *testing.T) {
	t.Setenv(core.DefaultWebTokenEnv, "")

	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{"web", "serve", "--dir", projectDir, "--allow-lan", "--dry-run"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"bind: 0.0.0.0\n", "allow_lan: true\n", "token_required: false\n"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("missing effective policy %q:\n%s", want, stdout.String())
		}
	}
}

// Existing defaults and explicit token requirements must not silently select LAN access.
func TestWebLANOptions(t *testing.T) {
	projectDir := newTestWorkspace(t)

	tests := []struct {
		name string
		args []string
		want []string
		err  string
	}{
		{"default", nil, []string{"allow_lan: false\n", "bind: 127.0.0.1\n", "require_token: auto\n"}, ""},
		{"explicit false", []string{"--allow-lan=false"}, []string{"allow_lan: false\n", "bind: 127.0.0.1\n"}, ""},
		{"address override", []string{"--allow-lan", "--host", "192.0.2.10", "--port", "7444"}, []string{"bind: 192.0.2.10\n", "port: 7444\n", "token_required: false\n"}, ""},
		{"required token", []string{"--allow-lan", "--require-token=true"}, nil, "--allow-lan conflicts"},
		{"always token", []string{"--allow-lan", "--require-token=always"}, nil, "--allow-lan conflicts"},
		{"invalid token mode", []string{"--allow-lan", "--require-token=unknown"}, nil, "invalid web require_token mode"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := append([]string{"web", "serve", "--dir", projectDir, "--dry-run"}, test.args...)

			var stdout bytes.Buffer

			err := run(t.Context(), args, &stdout, &bytes.Buffer{})
			assertWebLANOutput(t, stdout.String(), err, test.want, test.err)
		})
	}
}

func assertWebLANOutput(t *testing.T, stdout string, failure error, expected []string, expectedError string) {
	t.Helper()

	if expectedError != "" {
		if failure == nil || !strings.Contains(failure.Error(), expectedError) {
			t.Fatalf("error = %v, want %s", failure, expectedError)
		}

		return
	}

	if failure != nil {
		t.Fatal(failure)
	}

	for _, want := range expected {
		if !strings.Contains(stdout, want) {
			t.Fatalf("missing %q in %s", want, stdout)
		}
	}
}

// A project can retain LAN opt-in while an explicit false flag restores the normal bind.
func TestWebLANProjectConfig(t *testing.T) {
	projectDir := newTestWorkspace(t)

	configPath := workspacecore.UserPath(projectDir, "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(configPath, []byte("web:\n  allow_lan: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name string
		args []string
		want []string
	}{
		{"inherited", nil, []string{"allow_lan: true\n", "bind: 0.0.0.0\n", "token_required: false\n"}},
		{"explicit false", []string{"--allow-lan=false"}, []string{"allow_lan: false\n", "bind: 127.0.0.1\n", "require_token: auto\n"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			args := append([]string{"web", "serve", "--dir", projectDir, "--dry-run"}, test.args...)

			var stdout bytes.Buffer

			err := run(t.Context(), args, &stdout, &bytes.Buffer{})
			assertWebLANOutput(t, stdout.String(), err, test.want, "")
		})
	}
}

// A configured stronger requirement is honored, and a selected network interface is retained.
func TestWebLANConfiguredPolicy(t *testing.T) {
	config := core.DefaultWebConfig()
	config.RequireToken = core.WebTokenModeAlways

	if _, err := applyWebLANAccess(config, true, false); err == nil {
		t.Fatal("LAN flag silently weakened configured token protection")
	}

	config.RequireToken = core.WebTokenModeAuto
	config.Host = "192.0.2.10"

	resolved, err := applyWebLANAccess(config, true, false)
	if err != nil || resolved.Host != config.Host || webTokenRequired(resolved) {
		t.Fatal("LAN flag discarded the configured network interface", resolved, err)
	}
}

func webLANHandler(t *testing.T, provider webui.DataProvider) http.Handler {
	t.Helper()

	config, err := applyWebLANAccess(core.DefaultWebConfig(), true, false)
	if err != nil {
		t.Fatal(err)
	}

	handler, err := newWebServeMux(resolvedWebServeConfig{ProjectDir: t.TempDir(), Web: config, Provider: provider})
	if err != nil {
		t.Fatal(err)
	}

	return handler
}

// The browser follows rendered links and fragments with no token, auth header or cookie jar.
func TestWebLANNavigation(t *testing.T) {
	t.Setenv(core.DefaultWebTokenEnv, "")

	server := httptest.NewServer(webLANHandler(t, nil))
	defer server.Close()

	body := webLANFetch(t, server, "/")
	links := regexp.MustCompile(`href="(/[^"#]*)"`).FindAllStringSubmatch(body, -1)

	visited := make(map[string]bool)
	for _, link := range links {
		if !visited[link[1]] {
			webLANFetch(t, server, link[1])

			visited[link[1]] = true
		}
	}

	if visited["/project"] || visited["/"] {
		t.Fatal("project diagnostics must not appear in rendered navigation")
	}

	for _, path := range []string{"/board", "/dash", "/cases", "/dumps", "/quizzes", "/specs", "/plans"} {
		if !visited[path] {
			t.Fatalf("rendered navigation omitted %s", path)
		}

		webLANFetch(t, server, path)
		webLANFetch(t, server, path+"/fragment")
	}

	webLANFetch(t, server, "/project")
	webLANFetch(t, server, "/project/fragment")
	webLANFetch(t, server, "/assets/chr-logo.png")
}

func webLANFetch(t *testing.T, server *httptest.Server, path string) string {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Host = "devbox:7331"
	if strings.HasSuffix(path, "/fragment") {
		request.Header.Set("HX-Request", "true")
	}

	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Error(err)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusOK || len(response.Cookies()) != 0 {
		t.Fatalf("%s: status %d, read %v; unexpected cookies: %v", path, response.StatusCode, err, response.Cookies())
	}

	return string(body)
}

// Token-free access still enforces origin, confirmation and stale-revision guards.
func TestWebLANWrites(t *testing.T) {
	provider, original := webTicketWriteFixture(t)

	current, err := provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: original.Key})
	if err != nil {
		t.Fatal(err)
	}

	handler := webLANHandler(t, provider)
	for _, test := range []struct {
		name, revision, origin string
		confirm, writes        bool
		status                 int
	}{
		{"cross origin", current.Revision, "http://other.example", true, false, http.StatusForbidden},
		{"unconfirmed", current.Revision, "http://devbox:7331", false, false, http.StatusOK},
		{"stale", original.Revision, "http://devbox:7331", true, false, http.StatusOK},
		{"confirmed", current.Revision, "http://devbox:7331", true, true, http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := readWebTicketFile(t, provider, original.Key)
			form := url.Values{"action": {"ticket.update"}, "id": {original.Key}, "revision": {test.revision}, "confirm": {fmt.Sprint(test.confirm)}, "title": {"LAN browser title"}, "synopsis": {"Direct network access."}}
			request := httptest.NewRequest(http.MethodPost, "http://devbox:7331/actions", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("HX-Request", "true")
			request.Header.Set("Origin", test.origin)

			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("HTTP status %d, want %d: %s", response.Code, test.status, response.Body.String())
			}

			after := readWebTicketFile(t, provider, original.Key)
			if test.writes && !strings.Contains(after, "LAN browser title") || !test.writes && after != before {
				t.Fatal("LAN access bypassed a business guard or blocked a valid write")
			}
		})
	}
}
