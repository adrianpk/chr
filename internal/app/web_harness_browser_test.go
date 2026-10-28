package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/webui"
)

// One authenticated browser switches both defaults while a claim remains live,
// retries a lost acknowledgement and rejects a stale form at wide/narrow sizes.
func TestWebHarnessSelectionBrowser(t *testing.T) {
	if testing.Short() || os.Getenv("CHR_WEB_BROWSER") == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned harness browser fixture")
	}

	operations, record, request := serviceWorkflowTicketFixture(t)
	marker := prepareWebHarnessProfiles(t, operations)
	before := claimWebHarnessInvocation(t, operations, request)
	handler, _ := webHarnessAuthenticated(t, operations)

	server := webHarnessBrowserServer(t, handler)
	defer server.Close()

	browser := newWebWorkflowBrowser(t, os.Getenv("CHR_WEB_BROWSER"))
	webBoardBrowserViewport(t, browser, 1280, 1000)
	browser.navigate(t, server.URL+"/workflows")
	browser.wait(t, ".login-card")
	webBoardBrowserSet(t, browser, "[name=username]", "operator")
	webBoardBrowserSet(t, browser, "[name=password]", "example-only-password-2026")
	browser.clickNavigation(t, ".login-card button")
	browser.wait(t, ".harness-context")
	webBoardBrowserClick(t, browser, ".harness-context")
	browser.wait(t, "#harness-worker")
	webHarnessBrowserLayout(t, browser, "desktop")
	webHarnessBrowserRetry(t, browser, operations)
	webHarnessCheckRetained(t, operations, before)
	webHarnessCheckNewInvocation(t, operations, record, request)
	webHarnessBrowserStale(t, browser, operations)

	webBoardBrowserViewport(t, browser, 390, 844)
	webBoardBrowserWait(t, browser, "innerWidth===390")
	webHarnessBrowserLayout(t, browser, "narrow")
	webBoardBrowserClick(t, browser, ".harness-actions a")
	browser.wait(t, "#harness-worker")
	webBoardBrowserSet(t, browser, "#harness-worker", "")
	webBoardBrowserSet(t, browser, "#harness-generation", "")
	webBoardBrowserClick(t, browser, "#harness-selection-form button[type=submit]")
	browser.wait(t, `[data-harness-write-status="ok"]`)

	if snapshot := webHarnessSnapshot(t, operations); snapshot.Worker != "" || snapshot.Generation != "" {
		t.Fatal("narrow-screen clear did not persist")
	}

	webHarnessCheckRetained(t, operations, before)
	webHarnessBrowserCancel(t, browser, operations)

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("browser selection launched a driver")
	}
}

func webHarnessBrowserServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	var lose atomic.Bool
	lose.Store(true)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/project/harnesses" || !lose.Swap(false) {
			handler.ServeHTTP(w, r)

			return
		}

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, r)

		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `data-harness-write-status="ok"`) {
			t.Error("save failed before acknowledgement was dropped", response.Code)
		}

		http.Error(w, "lost acknowledgement", http.StatusServiceUnavailable)
	}))
}

func webHarnessBrowserRetry(t *testing.T, browser *webWorkflowBrowser, operations service) {
	t.Helper()
	webBoardBrowserSet(t, browser, "#harness-worker", "beta")
	webBoardBrowserSet(t, browser, "#harness-generation", "pi-writing")

	id := browser.evaluate(t, `document.querySelector('#harness-selection-form [name=request_id]').value`)
	webBoardBrowserClick(t, browser, "#harness-selection-form button[type=submit]")
	webBoardBrowserReady(t, browser)

	if got := browser.evaluate(t, `document.querySelector('#harness-worker').value`); got != "beta" {
		t.Fatal("lost acknowledgement discarded input", got)
	}

	first := webHarnessSnapshot(t, operations)
	if first.Worker != "beta" || first.Generation != "pi-writing" || first.RequestID != id {
		t.Fatal("first browser save was not committed", first)
	}

	webBoardBrowserClick(t, browser, "#harness-selection-form button[type=submit]")
	browser.wait(t, `[data-harness-write-status="ok"]`)

	if replay := webHarnessSnapshot(t, operations); replay != first {
		t.Fatal("lost acknowledgement retry published twice")
	}

	webHarnessCheckProcess(t, operations, first)
	webBoardBrowserWait(t, browser, `document.querySelector('.harness-current').textContent.includes('beta') && document.querySelector('.harness-retained').textContent.includes('alpha')`)
}

func webHarnessBrowserStale(t *testing.T, browser *webWorkflowBrowser, operations service) {
	t.Helper()
	webBoardBrowserSet(t, browser, "#harness-worker", "alpha")

	id := browser.evaluate(t, `document.querySelector('#harness-selection-form [name=request_id]').value`)
	revision := browser.evaluate(t, `document.querySelector('#harness-selection-form [name=revision]').value`)
	// A separate authenticated request publishes while the browser retains its
	// original image. The next browser request must not silently rebase.
	other, cookies := webHarnessAuthenticated(t, operations)
	form := webHarnessForm(t, other, cookies)
	form.Set("generation", "beta")
	webHarnessAssertWrite(t, other, cookies, form, webui.WriteStatusOK)
	webBoardBrowserClick(t, browser, "#harness-selection-form button[type=submit]")
	browser.wait(t, `[data-harness-write-status="stale"]`)

	for selector, expected := range map[string]string{"#harness-worker": "alpha", "[name=request_id]": id, "[name=revision]": revision} {
		if got := browser.evaluate(t, `document.querySelector('#harness-selection-form').querySelector(`+strconv.Quote(selector)+`).value`); got != expected {
			t.Fatal("stale rejection lost input", selector, got)
		}
	}

	if snapshot := webHarnessSnapshot(t, operations); snapshot.Worker != "beta" || snapshot.Generation != "beta" {
		t.Fatal("stale browser changed the latest defaults")
	}
}

func webHarnessBrowserLayout(t *testing.T, browser *webWorkflowBrowser, name string) {
	t.Helper()
	webBoardBrowserReady(t, browser)

	if got := browser.evaluate(t, `String(document.documentElement.scrollWidth<=innerWidth && document.querySelector('.harness-modal').scrollWidth<=document.querySelector('.harness-modal').clientWidth)`); got != "true" {
		t.Fatal("harness dialog overflows", name)
	}

	if got := browser.evaluate(t, `String(!!document.querySelector('[aria-modal=true]') && !document.querySelector('input[type=password], input[name=credential_ref], input[name=command]'))`); got != "true" {
		t.Fatal("chooser exposed private setup or lost its dialog", name)
	}

	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_HARNESS_SCREENSHOTS"), name)
}

func webHarnessBrowserCancel(t *testing.T, browser *webWorkflowBrowser, operations service) {
	t.Helper()

	before := webHarnessSnapshot(t, operations)
	webBoardBrowserSet(t, browser, "#harness-worker", "alpha")
	webBoardBrowserClick(t, browser, "[data-modal-close]")
	webBoardBrowserWait(t, browser, `!document.querySelector('#harness-selection-form')`)

	if webHarnessSnapshot(t, operations) != before {
		t.Fatal("closing the chooser saved unsent input")
	}
}

func webHarnessCheckRetained(t *testing.T, operations service, before workflowcore.Invocation) {
	t.Helper()

	after, err := operations.ReadWorkflow(t.Context(), before.ID)
	if err != nil || after.Revision != before.Revision || after.Status != before.Status || *after.HarnessBinding != *before.HarnessBinding {
		t.Fatal("changing defaults changed claimed work", err)
	}
}

func webHarnessCheckNewInvocation(t *testing.T, operations service, record workflowstate.Record, request workflowcore.StartRequest) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	task := *record.Task

	task.Key = "PRJ-91"
	if _, err := store.Create(t.Context(), "Later independent task", workflowstate.Record{Kind: "ticket", Task: &task, Domain: record.Domain, Body: record.Body}); err != nil {
		t.Fatal(err)
	}

	request.RequestKey, request.Selection.Selector.Identity = "later-selection", task.Key

	started, err := operations.StartWorkflow(t.Context(), request)
	if err != nil || started.Invocation.HarnessBinding == nil || started.Invocation.HarnessBinding.Profile != "beta" {
		t.Fatal("new work did not retain the changed default", err)
	}
}
