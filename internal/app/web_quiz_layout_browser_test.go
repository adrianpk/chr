package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Opening a new quiz prepares its first question once, with visible loading and
// four actions. Reloading an active quiz cannot restart generation or its session.
func TestWebQuizLayoutBrowser(t *testing.T) {
	if testing.Short() || os.Getenv("CHR_WEB_BROWSER") == "" {
		t.Skip("select the owned Chromium fixture explicitly")
	}

	provider, model := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	paths, err := operations.quizRepository().Paths(operations.quizScope(quizcore.Scope{CaseSlug: "named-work"}))
	if err != nil {
		t.Fatal(err)
	}

	server, releaseStart := webQuizLayoutServer(t, provider)
	defer server.Close()
	defer releaseStart()

	browser := newWebWorkflowBrowser(t, os.Getenv("CHR_WEB_BROWSER"))
	webBoardBrowserViewport(t, browser, 1440, 1000)
	browser.navigate(t, server.URL+quizFocusURL(paths.CuratedRelativePath))
	browser.wait(t, ".quiz-panel")
	webBoardBrowserWait(t, browser, `document.querySelector('.quiz-panel').getAttribute('aria-busy')==='true'`)

	if got := browser.evaluate(t, `String(document.querySelectorAll('main h1').length===1 && !document.querySelector('.quiz-history') && !document.body.innerText.includes('Start the quiz') && !document.querySelector('[data-quiz-start]') && Array.from(document.querySelectorAll('.quiz-actions > *')).map(e=>e.textContent.trim()).join(',')==='Skip,Answer,Cancel,Done')`); got != "true" {
		t.Fatal("new quiz exposes a second start step instead of its four actions", got)
	}

	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_QUIZ_LAYOUT_SCREENSHOTS"), "initial")
	releaseStart()
	webBoardBrowserWait(t, browser, `document.querySelector('#quiz-surface').dataset.quizSession!=='' && document.querySelector('#quiz-answer:not(:disabled)')`)
	webQuizLayoutCheck(t, browser)
	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_QUIZ_LAYOUT_SCREENSHOTS"), "active-desktop")
	webQuizLayoutReload(t, browser, server.URL+quizFocusURL(paths.CuratedRelativePath), model)

	for _, width := range []int{768, 390, 320} {
		webBoardBrowserViewport(t, browser, width, 844)
		webBoardBrowserWait(t, browser, "innerWidth==="+strconv.Itoa(width))
		webQuizLayoutCheck(t, browser)
		webBrowserCapture(t, browser, os.Getenv("CHR_WEB_QUIZ_LAYOUT_SCREENSHOTS"), "active-"+strconv.Itoa(width))
	}
}

func webQuizLayoutReload(t *testing.T, browser *webWorkflowBrowser, location string, model *fakeServiceQuizModel) {
	t.Helper()

	before := browser.evaluate(t, "document.querySelector('#quiz-surface').dataset.quizSession")
	browser.navigate(t, location)
	browser.wait(t, "#quiz-answer:not(:disabled)")

	if got := browser.evaluate(t, "document.querySelector('#quiz-surface').dataset.quizSession"); got != before || len(model.requests) != 1 {
		t.Fatal("reload replaced the quiz or generated another first question", got, len(model.requests))
	}
}

func webQuizLayoutServer(t *testing.T, provider webViewProvider) (*httptest.Server, func()) {
	t.Helper()

	release := make(chan struct{})

	var once sync.Once

	handler := webQuizHandler(t, provider)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.FormValue("action") == "quiz.start" {
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}

		handler.ServeHTTP(w, r)
	}))

	return server, func() { once.Do(func() { close(release) }) }
}

func webQuizLayoutCheck(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()

	if got := browser.evaluate(t, `String(document.documentElement.scrollWidth<=innerWidth && document.querySelectorAll('.quiz-panel').length===1 && document.querySelector('.quiz-panel .quiz-question') && document.querySelector('.quiz-panel #quiz-answer') && document.querySelector('.quiz-panel .quiz-actions') && document.querySelector('#quiz-answer').getBoundingClientRect().height>=240 && Array.from(document.querySelectorAll('.quiz-actions button')).map(e=>e.textContent.trim()).join(',')==='Skip,Answer,Cancel,Done')`); got != "true" {
		t.Fatal("question, answer or action layout escaped its panel", browser.evaluate(t, `JSON.stringify({width:innerWidth,scroll:document.documentElement.scrollWidth,panel:document.querySelector('.quiz-panel')?.getBoundingClientRect(),answer:document.querySelector('#quiz-answer')?.getBoundingClientRect()})`))
	}
}

// Failed first-question requests retain one intent across both Retry and a full
// reload; the successful attempt admits one session and invokes the model once.
func TestWebQuizFirstRetryBrowser(t *testing.T) {
	if testing.Short() || os.Getenv("CHR_WEB_BROWSER") == "" {
		t.Skip("select the owned Chromium fixture explicitly")
	}

	provider, model := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	paths, err := operations.quizRepository().Paths(operations.quizScope(quizcore.Scope{}))
	if err != nil {
		t.Fatal(err)
	}

	ids := make(chan string, 3)

	server := webQuizFirstRetryServer(t, provider, ids)
	defer server.Close()

	browser := newWebWorkflowBrowser(t, os.Getenv("CHR_WEB_BROWSER"))
	location := server.URL + quizFocusURL(paths.CuratedRelativePath)
	browser.navigate(t, location)
	browser.wait(t, "[data-quiz-retry-start]:not([hidden])")

	first := <-ids

	webBoardBrowserClick(t, browser, "[data-quiz-retry-start]")

	second := <-ids

	browser.wait(t, "[data-quiz-retry-start]:not([hidden])")
	browser.navigate(t, location)
	browser.wait(t, "#quiz-answer:not(:disabled)")

	third := <-ids

	if first == "" || first != second || first != third || len(model.requests) != 1 {
		t.Fatal("retry or reload changed the initial intent or duplicated generation", first, second, third, len(model.requests))
	}
}

func webQuizFirstRetryServer(t *testing.T, provider webViewProvider, ids chan<- string) *httptest.Server {
	t.Helper()

	var starts atomic.Int32

	handler := webQuizHandler(t, provider)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.FormValue("action") == "quiz.start" {
			ids <- r.FormValue("request_id")

			if starts.Add(1) <= 2 {
				http.Error(w, "temporary first-question failure", http.StatusServiceUnavailable)

				return
			}
		}

		handler.ServeHTTP(w, r)
	}))
}
