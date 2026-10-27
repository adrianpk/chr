package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// The actual modal retains its arguments and identity after a lost acknowledgement.
// An unchanged retry reconciles the committed capture; edited input becomes a new
// intent and selected cases remain independent captures at desktop/narrow widths.
func TestWebDumpCaptureBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned capture browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider := newWebDumpCaptureProvider(t)
	browser := newWebWorkflowBrowser(t, path)

	var loseAck atomic.Bool
	loseAck.Store(true)

	ids := make(chan string, 8)

	server := webDumpCaptureBrowserServer(t, provider, &loseAck, ids)
	defer server.Close()

	webBoardBrowserViewport(t, browser, 1440, 1000)
	browser.call(t, "Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": "window.confirm=()=>true"})
	browser.navigate(t, server.URL+"/dumps")
	browser.wait(t, "[data-modal-template='dump-create-modal']")
	browser.evaluate(t, `window.confirm=()=>true; window.captureTrace=[]; ["submit","htmx:beforeRequest","htmx:responseError","htmx:sendError","htmx:afterRequest"].forEach(name=>document.addEventListener(name,e=>{if(window.captureTrace.length<20)window.captureTrace.push({name,tag:e.detail?.elt?.tagName,target:e.detail?.target?.id,status:e.detail?.xhr?.status});})); 'ready'`)

	webDumpCaptureBrowserOpen(t, browser)
	webBoardBrowserSet(t, browser, "[data-dump-capture] [name='title']", "Project context")
	webBoardBrowserSet(t, browser, "[data-dump-capture] [name='content']", "Retained project context.")
	webDumpCaptureBrowserSubmit(t, browser, true)

	firstID := <-ids

	if browser.evaluate(t, "document.querySelector('[data-dump-capture] [name=content]').value") != "Retained project context." {
		t.Fatal("lost acknowledgement discarded capture input")
	}

	webDumpCaptureBrowserSubmit(t, browser, false)

	if secondID := <-ids; firstID == "" || secondID != firstID {
		t.Fatal("unchanged retry changed identity", firstID, secondID)
	}

	webDumpCaptureBrowserNamed(t, browser, &loseAck, ids)
	checkWebDumpCaptureBrowserInventory(t, provider.webViewProvider)
}

func webDumpCaptureBrowserServer(t *testing.T, provider webDumpCaptureProvider, loseAck *atomic.Bool, ids chan<- string) *httptest.Server {
	t.Helper()

	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/actions" || r.FormValue("action") != webActionDumpCreate {
			handler.ServeHTTP(w, r)

			return
		}

		ids <- r.FormValue("request_id")

		if !loseAck.Swap(false) {
			handler.ServeHTTP(w, r)

			return
		}

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, r)

		if recorder.Code != http.StatusOK {
			t.Error("fixture failed to commit before losing acknowledgement", recorder.Code)
		}

		http.Error(w, "lost acknowledgement", http.StatusServiceUnavailable)
	}))
}

func webDumpCaptureBrowserOpen(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	webBoardBrowserClick(t, browser, "[data-modal-template='dump-create-modal']")
	browser.wait(t, "[data-dump-capture]")

	if got := browser.evaluate(t, "String(document.querySelectorAll('#modal-root form').length)"); got != "1" {
		t.Fatal("New did not open a single form", got)
	}

	if got := browser.evaluate(t, `(()=>{const r=document.querySelector('#modal-root [role=dialog]').getBoundingClientRect();return String(r.left>=0&&r.right<=innerWidth&&r.top>=0&&r.height<=innerHeight);})()`); got != "true" {
		t.Fatal("capture dialog escaped the viewport", got)
	}

	if got := browser.evaluate(t, "Array.from(document.querySelector('[data-dump-capture] select').options).map(o=>o.textContent).join('|')"); got != "Project-wide|Named work" {
		t.Fatal("capture did not offer existing case names", got)
	}
}

func webDumpCaptureBrowserSubmit(t *testing.T, browser *webWorkflowBrowser, lost bool) {
	t.Helper()
	webBoardBrowserClick(t, browser, "[data-dump-capture] button[type='submit']")

	if lost {
		webDumpCaptureWait(t, browser, "document.querySelector('#dump-create-result')?.textContent.includes('Your context is retained')")

		return
	}

	webDumpCaptureWait(t, browser, "!document.querySelector('#modal-root [role=dialog]')")
	webBoardBrowserWait(t, browser, "document.activeElement.matches('[data-modal-template=\"dump-create-modal\"]')")
}

func webDumpCaptureBrowserNamed(t *testing.T, browser *webWorkflowBrowser, loseAck *atomic.Bool, ids <-chan string) {
	t.Helper()
	webBoardBrowserViewport(t, browser, 390, 844)
	webDumpCaptureBrowserOpen(t, browser)
	webBoardBrowserSet(t, browser, "[data-dump-capture] [name='title']", "Named context")
	webBoardBrowserSet(t, browser, "[data-dump-capture] [name='case_slug']", "named-work")
	webBoardBrowserSet(t, browser, "[data-dump-capture] [name='content']", "First named capture.")
	loseAck.Store(true)
	webDumpCaptureBrowserSubmit(t, browser, true)

	firstID := <-ids

	webBoardBrowserSet(t, browser, "[data-dump-capture] [name='content']", "Second named capture.")
	webDumpCaptureBrowserSubmit(t, browser, false)

	if secondID := <-ids; secondID == firstID || secondID == "" {
		t.Fatal("edited input reused a reserved capture identity", firstID, secondID)
	}
}

func checkWebDumpCaptureBrowserInventory(t *testing.T, provider webViewProvider) {
	t.Helper()

	observed, err := authoringcore.ObserveProjectDump(t.Context(), provider.projectContext.ProjectDir)
	if err != nil || len(observed.Candidates) != 1 || len(observed.Components) != 1 || len(observed.Scoped) != 2 {
		t.Fatal("browser retry or scope selection duplicated/collapsed captures", observed, err)
	}

	document, err := authoringcore.ReadDumpDocument(t.Context(), provider.projectContext.ProjectDir, observed.Destination)
	if err != nil || strings.Count(document.Content, "Retained project context.") != 1 {
		t.Fatal("browser duplicated an acknowledged capture", document, err)
	}
}

func webDumpCaptureWait(t *testing.T, browser *webWorkflowBrowser, expression string) {
	t.Helper()

	result := browser.evaluate(t, `(async()=>{for(let n=0;n<200;n++){if(`+expression+`)return 'ready';await new Promise(r=>setTimeout(r,50));}return JSON.stringify({condition:`+strconv.Quote(expression)+`,trace:window.captureTrace,modal:document.querySelector('#modal-root')?.innerText,id:document.querySelector('[data-dump-capture] [name=request_id]')?.value});})()`)
	if result != "ready" {
		t.Fatal("capture modal did not reach expected state", result)
	}
}
