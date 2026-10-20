package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// Named generation is reachable from an empty catalog, remains read-only until
// submission, and uses saved discovery through the shared generation service.
func TestWebSpecGeneration(t *testing.T) {
	provider, model := newWebQuizFixture(t)
	seedSpecGenerationScopes(t, provider.service(t.Context()))

	handler := webQuizHandler(t, provider)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/specs", nil))

	for _, want := range []string{`data-modal-template="spec-generate-modal"`, `id="spec-generate-form"`, `value="named-work"`, "Named work"} {
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), want) {
			t.Fatal("spec generation is unavailable", want, response.Body.String())
		}
	}

	if len(model.requests) != 0 {
		t.Fatal("catalog read invoked the provider")
	}

	if strings.Contains(response.Body.String(), `<option value="">Project-wide</option>`) {
		t.Fatal("project context was offered as a spec generation target")
	}

	model.responses = []core.QuizModelResponse{serviceQuizDraftSpecResponse("generated", serviceQuizClock()())}

	form := url.Values{"action": {"spec.generate"}, "family": {"specs"}, "surface": {"spec"}, "modal": {"true"}, "confirm": {"true"}, "case_slug": {"named-work"}, "request_id": {"generate-named-spec"}}

	response = webDumpEditorPost(handler, form)
	if response.Code != http.StatusOK || response.Header().Get("HX-Retarget") != "#modal-root" {
		t.Fatal(response.Code, response.Body.String())
	}

	if !strings.Contains(response.Body.String(), `id="spec-body"`) || !strings.Contains(response.Body.String(), "Generated behavior from dump and quiz context.") {
		t.Fatal("saved generation disappeared instead of opening its result", response.Body.String())
	}

	checkGeneratedWebSpec(t, provider, model)
}

// Project context and unknown cases cannot become Web generation targets,
// including direct submissions that bypass the named-case selector.
func TestWebSpecGenerationScope(t *testing.T) {
	for _, slug := range []string{"", " ", "missing-case"} {
		t.Run("scope "+slug, func(t *testing.T) {
			provider, model := newWebQuizFixture(t)
			form := url.Values{"action": {"spec.generate"}, "family": {"specs"}, "surface": {"spec"}, "modal": {"true"}, "confirm": {"true"}, "case_slug": {slug}, "request_id": {"invalid-scope"}}

			response := webDumpEditorPost(webQuizHandler(t, provider), form)
			if !strings.Contains(response.Body.String(), `data-spec-write-status="validation_failed"`) || len(model.requests) != 0 {
				t.Fatal("invalid scope invoked generation", response.Body.String())
			}

			documents, err := provider.service(t.Context()).ListSpecDocuments()
			if err != nil || len(documents) != 0 {
				t.Fatal("invalid scope created an artifact", documents, err)
			}
		})
	}
}

func checkGeneratedWebSpec(t *testing.T, provider webViewProvider, model *fakeServiceQuizModel) {
	t.Helper()

	documents, err := provider.service(t.Context()).ListSpecDocuments()
	if err != nil || len(documents) != 1 || documents[0].CaseSlug != "named-work" || len(model.requests) != 1 {
		t.Fatal("named spec was not generated", documents, err)
	}

	var context strings.Builder
	for _, section := range model.requests[0].ContextSections {
		context.WriteString(section.Content)
	}

	assertSpecGenerationScope(t, "named-work", context.String())
}

// The actual modal selects a named case, reports pending generation, suppresses
// repeated clicks, then shows the saved card without requiring a page reload.
func TestWebSpecGenerationBrowser(t *testing.T) {
	if testing.Short() || os.Getenv("CHR_WEB_BROWSER") == "" {
		t.Skip("select the owned Chromium fixture explicitly")
	}

	provider, model := newWebQuizFixture(t)
	model.responses = []core.QuizModelResponse{serviceQuizDraftSpecResponse("generated", serviceQuizClock()())}
	gate := make(chan struct{})
	release := sync.OnceFunc(func() { close(gate) })
	handler := webQuizHandler(t, provider)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.FormValue("action") == "spec.generate" {
			select {
			case <-gate:
			case <-r.Context().Done():
				return
			}
		}

		handler.ServeHTTP(w, r)
	}))
	defer server.Close()
	defer release()

	browser := newWebWorkflowBrowser(t, os.Getenv("CHR_WEB_BROWSER"))
	browser.navigate(t, server.URL+"/specs")
	browser.wait(t, `[data-modal-template="spec-generate-modal"]`)
	webBoardBrowserClick(t, browser, `[data-modal-template="spec-generate-modal"]`)
	browser.wait(t, "#spec-generate-form")
	webBoardBrowserSet(t, browser, "#spec-generate-case", "named-work")
	webBoardBrowserClick(t, browser, "#spec-generate-form button[type='submit']")
	webBoardBrowserWait(t, browser, `document.querySelector('#spec-result').textContent.includes('Generating specification') && document.querySelector('#spec-generate-case').disabled`)

	if got := browser.evaluate(t, `String(document.querySelector('#spec-generate-form [name="request_id"]').value.startsWith('web-spec-') && document.querySelector('#spec-generate-form button[type="submit"]').disabled)`); got != "true" {
		t.Fatal("generation has no retained identity or pending protection", got)
	}

	release()
	browser.wait(t, "#spec-body")
	checkSpecGenerationDialog(t, browser, "#spec-body")
	webBoardBrowserWait(t, browser, `document.querySelector('.document-card')`)

	if len(model.requests) != 1 || model.requests[0].Scope.CaseSlug != "named-work" {
		t.Fatal("generation lost selected scope", model.requests)
	}
}

func checkSpecGenerationDialog(t *testing.T, browser *webWorkflowBrowser, selector string) {
	t.Helper()
	webQuizCatalogKey(t, browser, "Escape", 27)
	browser.evaluate(t, `document.querySelector('#modal-root [data-modal-backdrop]').click(); 'clicked'`)

	if got := browser.evaluate(t, `String(!!document.querySelector(`+strconv.Quote(selector)+`))`); got != "true" {
		t.Fatal("screenshot dismissal hid the generation result", got)
	}

	webBoardBrowserClick(t, browser, "#modal-root [data-modal-close]")
	webBoardBrowserWait(t, browser, `!document.querySelector('#modal-root [role="dialog"]')`)
}

// A failed publication remains visible for inspection and capture, keeps the
// selected case, and cannot claim that a saved spec exists.
func TestWebSpecGenerationErrorBrowser(t *testing.T) {
	if testing.Short() || os.Getenv("CHR_WEB_BROWSER") == "" {
		t.Skip("select the owned Chromium fixture explicitly")
	}

	provider, model := newWebQuizFixture(t)
	response := serviceQuizDraftSpecResponse("invalid-path", serviceQuizClock()())
	response.DraftSpec.Body = "Store app data in .chr/default/invented/data.json."
	model.responses = []core.QuizModelResponse{response}

	server := httptest.NewServer(webQuizHandler(t, provider))
	defer server.Close()

	browser := newWebWorkflowBrowser(t, os.Getenv("CHR_WEB_BROWSER"))
	browser.navigate(t, server.URL+"/specs")
	browser.wait(t, `[data-modal-template="spec-generate-modal"]`)
	webBoardBrowserClick(t, browser, `[data-modal-template="spec-generate-modal"]`)
	browser.wait(t, "#spec-generate-form")
	webBoardBrowserSet(t, browser, "#spec-generate-case", "named-work")
	webBoardBrowserClick(t, browser, "#spec-generate-form button[type='submit']")
	browser.wait(t, `[data-spec-write-status="validation_failed"]`)

	if got := browser.evaluate(t, `document.querySelector('#spec-generate-case').value`); got != "named-work" {
		t.Fatal("publication failure lost the selected scope", got)
	}

	checkSpecGenerationDialog(t, browser, `[data-spec-write-status="validation_failed"]`)

	documents, err := provider.service(t.Context()).ListSpecDocuments()
	if err != nil || len(documents) != 0 || len(model.requests) != 1 {
		t.Fatal("failed publication created a document or retried generation", documents, err)
	}
}
