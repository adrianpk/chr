package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// An owned browser creates approved specs, edits manual/generated identities,
// retains input through lost acknowledgements and stale writes, and measures
// fixed cards and the one-body modal at desktop and narrow widths.
func TestWebSpecBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned spec browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider, model, generated := webSpecBrowserFixture(t)
	ids := make(chan string, 8)

	server := webSpecBrowserServer(t, provider, ids)
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	browser.call(t, "Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": "window.confirm=()=>true"})
	webBoardBrowserViewport(t, browser, 1280, 1000)
	browser.navigate(t, server.URL+"/specs")
	browser.wait(t, ".document-card")
	webSpecBrowserCapture(t, browser, "catalog-desktop")
	webSpecBrowserCreate(t, browser, provider, true)

	reference := webSpecBrowserCreate(t, browser, provider, false)
	webSpecBrowserLostAck(t, browser, reference, ids)
	webSpecBrowserGenerated(t, browser, provider, generated)
	webSpecBrowserStale(t, browser, provider, reference)

	for _, width := range []int{768, 390, 320} {
		webBoardBrowserViewport(t, browser, width, 844)
		webBoardBrowserWait(t, browser, "innerWidth==="+strconv.Itoa(width))
		webSpecBrowserCapture(t, browser, "catalog-"+strconv.Itoa(width))
	}

	webBoardBrowserViewport(t, browser, 390, 844)
	webSpecBrowserOpen(t, browser, generated)
	webSpecBrowserCapture(t, browser, "editor-narrow")

	if len(model.requests) != 1 {
		t.Fatal("manual authoring invoked generation", len(model.requests))
	}
}

func webSpecBrowserFixture(t *testing.T) (webViewProvider, *fakeServiceQuizModel, string) {
	t.Helper()

	provider, model := newWebQuizFixture(t)

	operations := provider.service(t.Context())
	for index, item := range []struct{ title, body string }{{"Workflow boundaries", "First retained requirement."}, {"Detailed requirements across several independent workflow boundaries", strings.Repeat("Opening requirements with retained line structure.\n", 30)}, {strings.Repeat("UnbrokenTitle", 12), strings.Repeat("UnbrokenBody", 200)}} {
		if _, err := operations.CreateSpec(CreateSpecRequest{OperationID: "browser-seed-" + strconv.Itoa(index), Title: item.title, Body: item.body, Status: "draft"}); err != nil {
			t.Fatal(err)
		}
	}

	model.responses = []core.QuizModelResponse{serviceQuizDraftSpecResponse("generated", serviceQuizClock()())}

	generated, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{Scope: quizcore.Scope{CaseSlug: "named-work"}})
	if err != nil {
		t.Fatal(err)
	}

	return provider, model, generated.Spec.RelativePath
}

func webSpecBrowserServer(t *testing.T, provider webViewProvider, ids chan<- string) *httptest.Server {
	t.Helper()

	handler := webQuizHandler(t, provider)

	var lose atomic.Bool
	lose.Store(true)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.FormValue("action") != "spec.save" {
			handler.ServeHTTP(w, r)

			return
		}

		ids <- r.FormValue("request_id")

		if !lose.Swap(false) {
			handler.ServeHTTP(w, r)

			return
		}

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, r)

		if response.Code != http.StatusOK || response.Header().Get("HX-Retarget") != "#modal-root" {
			t.Error(response.Code, response.Body.String())
		}

		http.Error(w, "lost acknowledgement", http.StatusServiceUnavailable)
	}))
}

func webSpecBrowserCreate(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider, draft bool) string {
	t.Helper()
	webBoardBrowserClick(t, browser, `.page-header a[hx-get="/specs/detail/fragment"]`)
	browser.wait(t, "#spec-body")
	webBoardBrowserWait(t, browser, `document.activeElement.id==='spec-title'`)

	title, scope, status, capture := "Directly approved scope", "named-work", "approved", "new-approved"
	if draft {
		title, scope, status, capture = "Browser draft spec", "", "draft", "new-draft"
	}

	webBoardBrowserSet(t, browser, "#spec-title", title)
	webBoardBrowserSet(t, browser, "#spec-case", scope)
	webBoardBrowserSet(t, browser, "#spec-body", "Approved opening text.\n\nSecond requirement.")
	browser.evaluate(t, `document.querySelector('#spec-draft').checked=`+strconv.FormatBool(draft)+`; 'selected'`)
	webSpecBrowserCapture(t, browser, capture)
	webSpecBrowserFocus(t, browser)
	webBoardBrowserClick(t, browser, "#spec-editor-form button[type='submit']")
	webBoardBrowserWait(t, browser, `!document.querySelector('[data-spec-editor]') && [...document.querySelectorAll('.document-card h2')].some(e=>e.textContent===`+strconv.Quote(title)+`)`)

	documents, err := provider.service(t.Context()).ListSpecDocuments()
	if err != nil {
		t.Fatal(err)
	}

	for _, document := range documents {
		if document.Title == title {
			if document.Status != status || document.CaseSlug != scope {
				t.Fatal(document)
			}

			return document.Reference
		}
	}

	t.Fatal("created spec missing")

	return ""
}

func webSpecBrowserOpen(t *testing.T, browser *webWorkflowBrowser, reference string) {
	t.Helper()
	webBoardBrowserClick(t, browser, `.document-card[href="`+"/specs/detail?id="+url.QueryEscape(reference)+`"]`)
	browser.wait(t, "#spec-body")
	webBoardBrowserReady(t, browser)
}

func webSpecBrowserLostAck(t *testing.T, browser *webWorkflowBrowser, reference string, ids <-chan string) {
	t.Helper()
	webSpecBrowserOpen(t, browser, reference)
	webBoardBrowserSet(t, browser, "#spec-body", "Corrected manual requirements with a retained acknowledgement.")
	webBoardBrowserClick(t, browser, "#spec-editor-form button[type='submit']")
	webBoardBrowserWait(t, browser, `document.querySelector('#spec-result').textContent.includes('Your edits are retained')`)

	if got := browser.evaluate(t, `document.querySelector('#spec-body').value`); got != "Corrected manual requirements with a retained acknowledgement." {
		t.Fatal(got)
	}

	first := <-ids

	webBoardBrowserClick(t, browser, "#spec-editor-form button[type='submit']")
	webBoardBrowserWait(t, browser, `!document.querySelector('[data-spec-editor]')`)

	if second := <-ids; second != first {
		t.Fatal("retry changed operation", first, second)
	}
}

func webSpecBrowserGenerated(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider, reference string) {
	t.Helper()
	webSpecBrowserOpen(t, browser, reference)

	if got := browser.evaluate(t, `String(document.querySelector('#spec-case').disabled && !document.querySelector('#spec-body').value.includes('## Provenance'))`); got != "true" {
		t.Fatal("generated editor leaked managed content or changed scope")
	}

	webBoardBrowserSet(t, browser, "#spec-body", "Corrected generated requirements.\n\nAn omitted detail now remains explicit.")
	webSpecBrowserCapture(t, browser, "generated-editor")
	webBoardBrowserClick(t, browser, "#spec-editor-form button[type='submit']")
	webBoardBrowserWait(t, browser, `!document.querySelector('[data-spec-editor]')`)

	detail := webSpecDetail(t, provider, reference)
	if !strings.Contains(detail.Content, "omitted detail") || !detail.Spec.Draft || detail.Spec.CaseSlug != "named-work" {
		t.Fatal(detail)
	}
}

func webSpecBrowserStale(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider, reference string) {
	t.Helper()
	webSpecBrowserOpen(t, browser, reference)
	webBoardBrowserSet(t, browser, "#spec-body", "Pending edits must survive a stale response.")
	browser.evaluate(t, `(async()=>{const w=document.querySelector('#workspace');await htmx.ajax('GET','/specs/fragment',{target:w,swap:'innerHTML'});return 'refreshed';})()`)

	detail := webSpecDetail(t, provider, reference)

	_, err := provider.service(t.Context()).SaveSpec(SaveSpecRequest{OperationID: "another-client", Reference: reference, Revision: detail.Revision, Title: detail.Title, Body: "Another client retained update.", Status: "approved", CaseSlug: detail.Spec.CaseSlug})
	if err != nil {
		t.Fatal(err)
	}

	webBoardBrowserClick(t, browser, "#spec-editor-form button[type='submit']")
	webBoardBrowserWait(t, browser, `document.querySelector('[data-spec-write-status="stale"]')`)

	if got := browser.evaluate(t, `document.querySelector('#spec-body').value`); got != "Pending edits must survive a stale response." {
		t.Fatal(got)
	}

	webSpecBrowserCapture(t, browser, "stale-editor")
	webBoardBrowserClick(t, browser, "#spec-result a")
	webBoardBrowserWait(t, browser, `document.querySelector('#spec-body').value==='Another client retained update.'`)
	webBoardBrowserClick(t, browser, "[data-modal-close]")
	webBoardBrowserWait(t, browser, `!document.querySelector('[data-spec-editor]')`)
}

func webSpecBrowserCapture(t *testing.T, browser *webWorkflowBrowser, name string) {
	t.Helper()

	metrics := browser.evaluate(t, `(()=>{const r=e=>{const b=e.getBoundingClientRect();return {width:b.width,height:b.height,right:b.right}},cards=[...document.querySelectorAll('.document-card')],editor=document.querySelector('[data-spec-editor]');return JSON.stringify({innerWidth,innerHeight,dpr:devicePixelRatio,browserZoom:1,pageScale:visualViewport.scale,pageWidth:document.documentElement.scrollWidth,cards:cards.map(c=>({card:r(c),preview:r(c.querySelector('.document-card__preview'))})),editor:editor?r(editor):null,bodyEditors:editor?editor.querySelectorAll('textarea').length:0});})()`)
	t.Log(name, metrics)

	if got := browser.evaluate(t, `String([...document.querySelectorAll('.document-card')].every(c=>{const p=c.querySelector('.document-card__preview');return c.getBoundingClientRect().width===300&&c.getBoundingClientRect().height===218&&p.getBoundingClientRect().height===132&&getComputedStyle(p).overflowY==='hidden'&&getComputedStyle(p).backgroundColor==='rgb(247, 251, 247)'&&getComputedStyle(c).backgroundColor==='rgb(255, 255, 255)';}))`); got != "true" {
		t.Fatal("fixed card anatomy changed", metrics)
	}

	if got := browser.evaluate(t, `String(document.documentElement.scrollWidth<=innerWidth && (!document.querySelector('[data-spec-editor]') || document.querySelectorAll('[data-spec-editor] textarea').length===1))`); got != "true" {
		t.Fatal("overflow or duplicate body", metrics)
	}

	directory := os.Getenv("CHR_WEB_SPEC_SCREENSHOTS")
	if directory == "" {
		return
	}

	webBrowserCapture(t, browser, directory, name)

	if err := os.WriteFile(filepath.Join(directory, name+".json"), []byte(metrics+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func webSpecBrowserFocus(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	browser.evaluate(t, `document.querySelector('#spec-editor-form button').focus(); 'focused'`)

	for _, step := range []struct {
		mod      int
		selector string
	}{{0, "[data-modal-close]"}, {8, "#spec-editor-form button"}} {
		for _, kind := range []string{"keyDown", "keyUp"} {
			browser.call(t, "Input.dispatchKeyEvent", map[string]any{"type": kind, "key": "Tab", "code": "Tab", "windowsVirtualKeyCode": 9, "modifiers": step.mod})
		}

		if got := browser.evaluate(t, `String(document.activeElement.matches(`+strconv.Quote(step.selector)+`))`); got != "true" {
			t.Fatal("focus escaped the modal")
		}
	}
}
