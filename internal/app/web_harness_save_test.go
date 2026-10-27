package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/webui"
)

// The authenticated HTTP adapter publishes both roles once, replays the same
// receipt and restores legacy dispatch without restarting the handler.
func TestWebHarnessSelectionSave(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	marker := prepareWebHarnessProfiles(t, operations)
	handler, cookies := webHarnessAuthenticated(t, operations)
	form := webHarnessForm(t, handler, cookies)
	form.Set("worker", "beta")
	form.Set("generation", "pi-writing")
	webHarnessAssertWrite(t, handler, cookies, form, webui.WriteStatusOK)

	saved := webHarnessSnapshot(t, operations)
	if saved.Worker != "beta" || saved.Generation != "pi-writing" || saved.RequestID != form.Get("request_id") || saved.ChangedBy == "" {
		t.Fatal("the shared owner did not persist the exact choices and actor", saved)
	}

	webHarnessAssertWrite(t, handler, cookies, form, webui.WriteStatusOK)

	if replay := webHarnessSnapshot(t, operations); replay != saved {
		t.Fatal("receipt replay published another configuration")
	}

	webHarnessCheckProcess(t, operations, saved)
	form.Set("worker", "alpha")
	webHarnessAssertWrite(t, handler, cookies, form, webui.WriteStatusStale)

	webHarnessClearDefaults(t, handler, cookies, operations)

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("saving defaults launched a driver")
	}
}

func webHarnessClearDefaults(t *testing.T, handler http.Handler, cookies []*http.Cookie, operations service) {
	t.Helper()

	clear := webHarnessForm(t, handler, cookies)
	clear.Set("generation", "")
	webHarnessAssertWrite(t, handler, cookies, clear, webui.WriteStatusOK)

	if snapshot := webHarnessSnapshot(t, operations); snapshot.Worker != "beta" || snapshot.Generation != "" {
		t.Fatal("clearing generation changed the worker", snapshot)
	}

	clear = webHarnessForm(t, handler, cookies)
	clear.Set("worker", "")
	webHarnessAssertWrite(t, handler, cookies, clear, webui.WriteStatusOK)

	if snapshot := webHarnessSnapshot(t, operations); snapshot.Worker != "" || snapshot.Generation != "" {
		t.Fatal("clearing worker did not restore legacy dispatch", snapshot)
	}
}

// A setup failure after opening the chooser keeps the exact rejected input
// available for retry, even when fresh profile inspection cannot complete.
func TestWebHarnessSelectionSetup(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	prepareWebHarnessProfiles(t, operations)

	handler, cookies := webHarnessAuthenticated(t, operations)
	form := webHarnessForm(t, handler, cookies)
	form.Set("worker", "beta")

	before := webHarnessSnapshot(t, operations)
	harnessSelectionWrite(t, filepath.Join(operations.env.ConfigDir(), "harnesses.yml"), "profiles: [invalid\n")

	body := webHarnessAssertWrite(t, handler, cookies, form, webui.WriteStatusValidation)
	if !strings.Contains(body, `value="beta" selected`) || webWorkflowHidden(t, body, "revision") != form.Get("revision") || strings.Contains(body, operations.env.ConfigDir()) || webHarnessSnapshot(t, operations) != before {
		t.Fatal("setup rejection lost input, exposed private state or changed configuration")
	}
}

// The direct page keeps a normal POST fallback instead of opening a second
// copy of the form when HTMX is available on the surrounding shell.
func TestWebHarnessSelectionPage(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	prepareWebHarnessProfiles(t, operations)

	handler, cookies := webHarnessAuthenticated(t, operations)
	form := webHarnessForm(t, handler, cookies)
	form.Set("worker", "beta")

	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/project/harnesses", strings.NewReader(form.Encode()))
	request.Header.Set("Origin", "http://example.com")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `data-harness-write-status="ok"`) || strings.Contains(response.Body.String(), `hx-post="/project/harnesses"`) {
		t.Fatal("normal page could not save defaults", response.Code)
	}

	if webHarnessSnapshot(t, operations).Worker != "beta" {
		t.Fatal("normal page did not persist its selection")
	}
}

// Rejected browser input cannot change configuration or another project's
// identity. The form retains its attempted choice and original revision.
func TestWebHarnessSelectionReject(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	prepareWebHarnessProfiles(t, operations)

	handler, cookies := webHarnessAuthenticated(t, operations)

	before := webHarnessSnapshot(t, operations)
	for _, test := range []struct{ name, field, value, status string }{
		{"no change", "worker", webui.HarnessKeepProfile, webui.WriteStatusValidation},
		{"unknown profile", "worker", "missing-profile", webui.WriteStatusValidation},
		{"unprepared profile", "worker", "unavailable", webui.WriteStatusValidation},
		{"foreign project", "project_ref", "another-project", webui.WriteStatusBlocked},
		{"stale image", "revision", strings.Repeat("0", 64), webui.WriteStatusStale},
		{"confirmation", "confirm", "false", webui.WriteStatusBlocked},
	} {
		t.Run(test.name, func(t *testing.T) {
			form := webHarnessForm(t, handler, cookies)
			form.Set("worker", "beta")
			form.Set(test.field, test.value)

			body := webHarnessAssertWrite(t, handler, cookies, form, test.status)
			if webWorkflowHidden(t, body, "revision") != form.Get("revision") || webWorkflowHidden(t, body, "request_id") != form.Get("request_id") || !strings.Contains(body, `value="`+form.Get("worker")+`" selected`) {
				t.Fatal("rejection lost the attempted input or revision")
			}

			if after := webHarnessSnapshot(t, operations); after != before {
				t.Fatal("rejected form changed the persisted selections")
			}
		})
	}
}

// Existing login and same-origin guards run before the dedicated bounded form
// decoder; generic actions cannot bypass that decoder or accept private setup.
func TestWebHarnessSelectionGuards(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	prepareWebHarnessProfiles(t, operations)

	handler, cookies := webHarnessAuthenticated(t, operations)

	before := webHarnessSnapshot(t, operations)
	for _, test := range []struct {
		name, origin, field, path string
		authenticated             bool
		code                      int
	}{
		{"login", "http://example.com", "", "/project/harnesses", false, http.StatusUnauthorized},
		{"missing origin", "", "", "/project/harnesses", true, http.StatusForbidden},
		{"foreign origin", "https://foreign.example", "", "/project/harnesses", true, http.StatusForbidden},
		{"private field", "http://example.com", "credential_ref", "/project/harnesses", true, http.StatusBadRequest},
		{"duplicate", "http://example.com", "worker", "/project/harnesses", true, http.StatusBadRequest},
		{"oversized", "http://example.com", "oversized", "/project/harnesses", true, http.StatusBadRequest},
		{"generic action", "http://example.com", "", "/actions", true, http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			form := webHarnessForm(t, handler, cookies)
			form.Set("worker", "beta")

			webHarnessGuardField(form, test.field)

			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, test.path, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Origin", test.origin)

			if test.authenticated {
				for _, cookie := range cookies {
					request.AddCookie(cookie)
				}
			}

			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.code || webHarnessSnapshot(t, operations) != before {
				t.Fatal("guard accepted a rejected mutation", response.Code, response.Body.String())
			}
		})
	}
}

func webHarnessGuardField(form url.Values, field string) {
	if field == "" {
		return
	}

	value := "private-input"
	if field == "oversized" {
		value = strings.Repeat("x", 8193)
	}

	form.Add(field, value)
}

func webHarnessAuthenticated(t *testing.T, operations service) (http.Handler, []*http.Cookie) {
	t.Helper()

	credentials, err := webui.CreateLoginCredentials("operator", "example-only-password-2026")
	if err != nil {
		t.Fatal(err)
	}

	handler := webHarnessHandler(t, webHarnessProvider(operations), &credentials)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/login", strings.NewReader("username=operator&password=example-only-password-2026"))
	request.Header.Set("Origin", "http://example.com")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther {
		t.Fatal("login failed", response.Code)
	}

	result := response.Result()
	if err := result.Body.Close(); err != nil {
		t.Fatal(err)
	}

	return handler, result.Cookies()
}

func webHarnessForm(t *testing.T, handler http.Handler, cookies []*http.Cookie) url.Values {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/project/harnesses", nil)
	request.Header.Set("HX-Request", "true")

	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatal("chooser read failed", response.Code)
	}

	form := url.Values{"worker": {webui.HarnessKeepProfile}, "generation": {webui.HarnessKeepProfile}}
	for _, field := range []string{"action", "surface", "confirm", "revision", "project_ref", "request_id"} {
		form.Set(field, webWorkflowHidden(t, response.Body.String(), field))
	}

	return form
}

func webHarnessAssertWrite(t *testing.T, handler http.Handler, cookies []*http.Cookie, form url.Values, status string) string {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/project/harnesses", strings.NewReader(form.Encode()))
	request.Header.Set("Origin", "http://example.com")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")

	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `data-harness-write-status="`+status+`"`) {
		t.Fatal("unexpected save outcome", response.Code, response.Body.String())
	}

	return response.Body.String()
}

func webHarnessSnapshot(t *testing.T, operations service) harnesscore.SelectionSnapshot {
	t.Helper()

	config, err := workspaceconfig.Load(operations.projectContext.ProjectDir)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := config.HarnessSelections(operations.projectContext.ProjectID)
	if err != nil {
		t.Fatal(err)
	}

	return snapshot
}

func webHarnessCheckProcess(t *testing.T, operations service, expected harnesscore.SelectionSnapshot) {
	t.Helper()

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	command := exec.CommandContext(t.Context(), program, "-test.run=^TestWebHarnessSelectionProcess$")

	command.Env = append(os.Environ(), "CHR_WEB_HARNESS_PROBE="+operations.projectContext.ProjectDir, "CHR_WEB_HARNESS_PROJECT_ID="+operations.projectContext.ProjectID)

	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}

	var actual harnesscore.SelectionSnapshot
	if err := json.NewDecoder(strings.NewReader(string(output))).Decode(&actual); err != nil || actual != expected {
		t.Fatal("independent process did not observe the saved image", actual, err)
	}
}

// The parent launches the same binary as a separate process after the HTTP
// response; this reads persisted configuration without a shared service cache.
func TestWebHarnessSelectionProcess(t *testing.T) {
	project := os.Getenv("CHR_WEB_HARNESS_PROBE")
	if project == "" {
		t.Skip("owned persistence probe is launched by the save test")
	}

	config, err := workspaceconfig.Load(project)
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := config.HarnessSelections(os.Getenv("CHR_WEB_HARNESS_PROJECT_ID"))
	if err != nil {
		t.Fatal(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(snapshot); err != nil {
		t.Fatal(err)
	}
}
