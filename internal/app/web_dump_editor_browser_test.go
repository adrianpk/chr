package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"sync/atomic"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// The actual single editor keeps pending input after lost acknowledgements,
// stale saves and catalog refreshes. Explicit reload discards only by choice;
// modal and full-page Save reopen the same document at desktop/narrow widths.
func TestWebDumpEditorBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned dump editor browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider := newWebDumpCaptureProvider(t)

	created, err := provider.operations.CreateDump(DumpCreateRequest{OperationID: "first", Title: "Project context", Content: "Original body."})
	if err != nil {
		t.Fatal(err)
	}

	ids := make(chan string, 8)

	server := webDumpEditorBrowserServer(t, provider, ids)
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	webBoardBrowserViewport(t, browser, 1440, 1000)
	browser.call(t, "Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": "window.confirm=()=>true"})
	browser.navigate(t, server.URL+"/dumps")
	webDumpEditorOpen(t, browser)
	webBoardBrowserSet(t, browser, "[data-dump-editor] textarea", "Corrected body.\n\nAdded in the same editor.")
	webBoardBrowserClick(t, browser, "[data-dump-editor] button[type=submit]")
	webBoardBrowserWait(t, browser, "document.querySelector('#dump-save-result')?.textContent.includes('Your edits are retained')")

	firstID := <-ids

	webDumpEditorValue(t, browser, "Corrected body.\n\nAdded in the same editor.")
	webDumpEditorSave(t, browser)

	if retryID := <-ids; retryID == "" || retryID != firstID {
		t.Fatal("unchanged Save retry changed identity", firstID, retryID)
	}

	webDumpEditorStale(t, browser, provider, created.RelativePath, ids)
	webDumpEditorPage(t, browser, server.URL, created.RelativePath)
}

func webDumpEditorBrowserServer(t *testing.T, provider webDumpCaptureProvider, ids chan<- string) *httptest.Server {
	t.Helper()

	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	var loseAck atomic.Bool
	loseAck.Store(true)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.FormValue("action") != webActionDumpSave {
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

		if recorder.Code != http.StatusOK || recorder.Header().Get("HX-Retarget") != "#modal-root" {
			t.Error("fixture did not save before losing acknowledgement", recorder.Code, recorder.Body.String())
		}

		http.Error(w, "lost acknowledgement", http.StatusServiceUnavailable)
	}))
}

func webDumpEditorOpen(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	webBoardBrowserClick(t, browser, ".dump-card-link")
	browser.wait(t, "[data-dump-editor]")

	if got := browser.evaluate(t, `String(document.querySelectorAll('#modal-root textarea').length===1&&!document.querySelector('#modal-root .artifact-content')&&!document.querySelector('#modal-root [value="dump.append"]'))`); got != "true" {
		t.Fatal("dump editor contains duplicate editing surfaces", got)
	}

	if got := browser.evaluate(t, `(()=>{const r=document.querySelector('.dump-detail-panel').getBoundingClientRect();return String(r.left>=0&&r.right<=innerWidth&&r.top>=0&&r.bottom<=innerHeight);})()`); got != "true" {
		t.Fatal("dump editor escaped the viewport", got)
	}
}

func webDumpEditorValue(t *testing.T, browser *webWorkflowBrowser, expected string) {
	t.Helper()

	if got := browser.evaluate(t, "document.querySelector('[data-dump-editor] textarea').value.trim()"); got != expected {
		t.Fatal("editor lost the pending/stored body", got, expected)
	}
}

func webDumpEditorSave(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	webBoardBrowserClick(t, browser, "[data-dump-editor] button[type=submit]")
	webBoardBrowserWait(t, browser, "!document.querySelector('#modal-root [role=dialog]')")
	webBoardBrowserWait(t, browser, "document.activeElement.matches('.dump-card-link')")
}

func webDumpEditorStale(t *testing.T, browser *webWorkflowBrowser, provider webDumpCaptureProvider, reference string, ids <-chan string) {
	t.Helper()
	webBoardBrowserViewport(t, browser, 390, 844)
	webDumpEditorOpen(t, browser)
	webDumpEditorValue(t, browser, "Corrected body.\n\nAdded in the same editor.")

	before, err := authoringcore.ReadDumpDocument(t.Context(), provider.projectContext.ProjectDir, reference)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := provider.operations.SaveDump(DumpSaveRequest{OperationID: "concurrent", ExpectedRevision: before.Revision, RelativePath: reference, Content: "Concurrent body."}); err != nil {
		t.Fatal(err)
	}

	webBoardBrowserSet(t, browser, "[data-dump-editor] textarea", "Pending replacement.")
	webBoardBrowserClick(t, browser, "[data-dump-editor] button[type=submit]")
	webBoardBrowserWait(t, browser, "document.querySelector('#dump-save-result [data-write-status=stale]')")

	staleID := <-ids

	webBoardBrowserSet(t, browser, "[data-dump-editor] textarea", "Edited pending replacement.")
	webBoardBrowserClick(t, browser, "[data-dump-editor] button[type=submit]")
	webBoardBrowserWait(t, browser, "document.querySelector('#dump-save-result [data-write-status=stale]')&&!document.querySelector('[data-dump-editor].htmx-request')")

	if changedID := <-ids; changedID == "" || changedID == staleID {
		t.Fatal("changed Save arguments reused a reserved identity", changedID, staleID)
	}

	browser.evaluate(t, `(async()=>{const w=document.querySelector('#workspace');await htmx.ajax('GET',w.getAttribute('hx-get'),{source:w,target:w,swap:'innerHTML'});return 'refreshed';})()`)
	webDumpEditorValue(t, browser, "Edited pending replacement.")
	webBoardBrowserClick(t, browser, "#dump-save-result button")

	state := browser.evaluate(t, `(async()=>{for(let n=0;n<200;n++){if(document.querySelector('[data-dump-editor] textarea')?.value.trim()==='Concurrent body.')return 'ready';await new Promise(r=>setTimeout(r,50));}return document.querySelector('#modal-root').innerHTML;})()`)
	if state != "ready" {
		t.Fatal("explicit reload did not load current body", state)
	}

	webBoardBrowserSet(t, browser, "[data-dump-editor] textarea", "Final modal body.")
	webDumpEditorSave(t, browser)
	<-ids
}

func webDumpEditorPage(t *testing.T, browser *webWorkflowBrowser, base, reference string) {
	t.Helper()
	browser.navigate(t, base+"/dumps/detail?id="+url.QueryEscape(reference))
	browser.wait(t, "[data-dump-editor]")
	webDumpEditorValue(t, browser, "Final modal body.")

	if got := browser.evaluate(t, "String(document.querySelector('#workspace').hasAttribute('hx-trigger'))"); got != "false" {
		t.Fatal("full-page editor can poll over pending input", got)
	}

	webBoardBrowserSet(t, browser, "[data-dump-editor] textarea", "Full-page revised body.")
	webBoardBrowserClick(t, browser, "[data-dump-editor] button[type=submit]")
	webBoardBrowserWait(t, browser, "document.querySelector('[data-dump-editor] textarea')?.value.trim()==="+strconv.Quote("Full-page revised body."))
	webBoardBrowserWait(t, browser, "!document.querySelector('[data-dump-editor].htmx-request')")
	browser.navigate(t, base+"/dumps/detail?id="+url.QueryEscape(reference))
	browser.wait(t, "[data-dump-editor]")
	webDumpEditorValue(t, browser, "Full-page revised body.")
}
