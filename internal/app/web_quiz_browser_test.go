package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// The owned browser exercises real forms, retries, stale observations, per-question
// draft recovery, mode switches, history and distinct discard/Done/Cancel flows.
func TestWebQuizBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned quiz browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider, model := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	started, err := operations.StartQuiz(QuizStartRequest{Mode: quizcore.ModeFast})
	if err != nil {
		t.Fatal(err)
	}

	ids := make(chan string, 8)

	server := webQuizBrowserServer(t, provider, ids)
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	webBoardBrowserViewport(t, browser, 1440, 1100)
	browser.call(t, "Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": "window.confirm=()=>true"})
	browser.navigate(t, server.URL+quizFocusURL(started.State.CuratedPath))
	browser.wait(t, "#quiz-answer")
	webQuizBrowserLostAck(t, browser, ids)
	webQuizBrowserDraftModes(t, browser, server.URL, started.State.CuratedPath)
	webQuizBrowserStale(t, browser, provider)
	webQuizBrowserHistory(t, browser, model)
	webQuizBrowserCapture(t, browser, "desktop-fast")
	webQuizBrowserMode(t, browser, "deep")
	webQuizBrowserCapture(t, browser, "desktop-deep")
	webBoardBrowserViewport(t, browser, 390, 844)
	webBoardBrowserWait(t, browser, "innerWidth===390")
	browser.evaluate(t, `new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(()=>r('painted'))))`)

	if browser.evaluate(t, "String(document.documentElement.scrollWidth<=innerWidth)") != "true" {
		t.Fatal("narrow quiz overflow", browser.evaluate(t, `JSON.stringify({width:innerWidth,scroll:document.documentElement.scrollWidth,body:document.body.scrollWidth,rail:document.querySelector('.top-nav-rail').scrollWidth,items:Array.from(document.querySelectorAll('body *')).map(e=>({tag:e.tagName,class:e.className,w:e.getBoundingClientRect().width,right:e.getBoundingClientRect().right})).filter(e=>e.right>innerWidth+1)})`))
	}

	webQuizBrowserCapture(t, browser, "narrow-deep")
	webQuizBrowserDone(t, browser, provider)
	webQuizBrowserCancel(t, browser, provider, server.URL)
}

func webQuizBrowserServer(t *testing.T, provider webViewProvider, ids chan<- string) *httptest.Server {
	t.Helper()

	handler := webQuizHandler(t, provider)

	var lose, failMode atomic.Bool
	lose.Store(true)
	failMode.Store(true)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.FormValue("action") == "quiz.mode" && failMode.Swap(false) {
			http.Error(w, "mode request unavailable", http.StatusServiceUnavailable)

			return
		}

		if r.Method != http.MethodPost || r.FormValue("action") != "quiz.answer" {
			handler.ServeHTTP(w, r)

			return
		}

		ids <- r.FormValue("request_id")

		if !lose.Swap(false) {
			handler.ServeHTTP(w, r)

			return
		}

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, r)

		if recorder.Code != http.StatusOK || recorder.Header().Get("HX-Retarget") != "#quiz-surface" {
			t.Error("answer did not commit before lost acknowledgement", recorder.Code, recorder.Body.String())
		}

		http.Error(w, "lost acknowledgement", http.StatusServiceUnavailable)
	}))
}

func webQuizBrowserLostAck(t *testing.T, browser *webWorkflowBrowser, ids <-chan string) {
	t.Helper()
	webBoardBrowserSet(t, browser, "#quiz-answer", "Keep confirmed answers durable across retries.")
	webBoardBrowserClick(t, browser, "[value='quiz.answer']")
	webBoardBrowserWait(t, browser, "document.querySelector('#quiz-result').textContent.includes('Your draft is retained')")

	first := <-ids

	webQuizBrowserValue(t, browser, "Keep confirmed answers durable across retries.")
	webBoardBrowserClick(t, browser, "[value='quiz.answer']")
	webBoardBrowserWait(t, browser, "document.querySelector('#quiz-surface').dataset.quizQuestion==='question-2'")

	if retry := <-ids; retry != first {
		t.Fatal("unchanged answer retry changed operation identity", first, retry)
	}

	webQuizBrowserValue(t, browser, "")
}

func webQuizBrowserDraftModes(t *testing.T, browser *webWorkflowBrowser, base, reference string) {
	t.Helper()
	webBoardBrowserSet(t, browser, "#quiz-answer", "An unsent answer must survive refresh and mode changes.")
	browser.navigate(t, base+quizFocusURL(reference))
	browser.wait(t, "#quiz-answer")
	webQuizBrowserValue(t, browser, "An unsent answer must survive refresh and mode changes.")

	webQuizBrowserModeFailure(t, browser)

	for _, mode := range []string{"deep", "fast"} {
		webQuizBrowserMode(t, browser, mode)
		webQuizBrowserValue(t, browser, "An unsent answer must survive refresh and mode changes.")
	}

	browser.evaluate(t, "window.quizDiscards=0;window.confirm=()=>{window.quizDiscards++;return false};'ready'")
	webBoardBrowserClick(t, browser, "[value='quiz.skip']")
	webBoardBrowserWait(t, browser, "window.quizDiscards===1")
	webQuizBrowserValue(t, browser, "An unsent answer must survive refresh and mode changes.")
	browser.evaluate(t, "window.confirm=()=>true;'ready'")
}

func webQuizBrowserMode(t *testing.T, browser *webWorkflowBrowser, mode string) {
	t.Helper()
	webQuizBrowserChangeMode(t, browser, mode)
	webBoardBrowserWait(t, browser, "document.querySelector('#quiz-mode input:checked').value==="+strconv.Quote(mode)+"&&document.querySelector('[data-quiz-write-status]')?.textContent.includes('mode switched')")
	webBoardBrowserReady(t, browser)
}

func webQuizBrowserStale(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider) {
	t.Helper()

	if _, err := provider.service(t.Context()).SkipQuiz(QuizSkipRequest{}); err != nil {
		t.Fatal(err)
	}

	webBoardBrowserClick(t, browser, "[value='quiz.answer']")
	webBoardBrowserWait(t, browser, "document.querySelector('[data-quiz-write-status]')?.dataset.quizWriteStatus==='stale'")
	webQuizBrowserValue(t, browser, "An unsent answer must survive refresh and mode changes.")
	webBoardBrowserClick(t, browser, "[data-quiz-reload]")
	webBoardBrowserWait(t, browser, "document.querySelector('#quiz-surface')?.dataset.quizQuestion==='question-3'")
	webQuizBrowserValue(t, browser, "")

	if browser.evaluate(t, "String(document.querySelector('[data-quiz-retained]').textContent.includes('An unsent answer'))") != "true" {
		t.Fatal("reload lost the old question draft")
	}

	webBoardBrowserClick(t, browser, "[data-quiz-retained] summary")
	webBoardBrowserClick(t, browser, "[data-quiz-retained] button")
	webBoardBrowserWait(t, browser, "!document.querySelector('[data-quiz-retained] details')")
}

func webQuizBrowserHistory(t *testing.T, browser *webWorkflowBrowser, model *fakeServiceQuizModel) {
	t.Helper()

	for index, answer := range []string{"Retry the same confirmed action without duplicate entries.", "Confirm the expected result against the retained state.", "Keep the current question and queued work across restart."} {
		webBoardBrowserSet(t, browser, "#quiz-answer", answer)
		webBoardBrowserClick(t, browser, "[value='quiz.answer']")
		webBoardBrowserWait(t, browser, "document.querySelectorAll('.quiz-history-card').length==="+strconv.Itoa(index+2))
		webBoardBrowserReady(t, browser)
	}

	webBoardBrowserSet(t, browser, "#quiz-answer", "Clarify the remaining boundary before finishing.")

	if len(model.requests) != 1 {
		t.Fatal("Fast regenerated a question", len(model.requests))
	}

	if browser.evaluate(t, "document.querySelector('.quiz-history-card p').textContent") != "Keep the current question and queued work across restart." {
		t.Fatal("answer history order changed")
	}
}

func webQuizBrowserCapture(t *testing.T, browser *webWorkflowBrowser, name string) {
	t.Helper()
	webQuizBrowserValue(t, browser, "Clarify the remaining boundary before finishing.")

	if browser.evaluate(t, "String(document.querySelectorAll('#quiz-answer').length===1&&!document.querySelector('.quiz-question textarea')&&document.querySelectorAll('.quiz-history-card').length===4)") != "true" {
		t.Fatal("focused question/history anatomy changed")
	}

	directory := os.Getenv("CHR_WEB_QUIZ_SCREENSHOTS")
	webBrowserCapture(t, browser, directory, name)

	if directory == "" {
		return
	}

	measurements := browser.evaluate(t, `JSON.stringify({width:innerWidth,height:innerHeight,dpr:devicePixelRatio,scrollWidth:document.documentElement.scrollWidth,mode:document.querySelector('#quiz-mode input:checked').value,history:document.querySelectorAll('.quiz-history-card').length,question:document.querySelector('.quiz-question p').textContent})`)
	if err := os.WriteFile(filepath.Join(directory, name+".json"), []byte(measurements+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func webQuizBrowserDone(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider) {
	t.Helper()
	browser.evaluate(t, "window.quizDiscards=0;window.confirm=()=>{window.quizDiscards++;return false};'ready'")
	webBoardBrowserClick(t, browser, "[value='quiz.done']")
	webBoardBrowserWait(t, browser, "window.quizDiscards===1")
	webQuizBrowserValue(t, browser, "Clarify the remaining boundary before finishing.")
	browser.evaluate(t, "window.confirm=()=>true;'ready'")
	webBoardBrowserClick(t, browser, "[value='quiz.done']")
	webBoardBrowserWait(t, browser, "location.pathname==='/quizzes'")

	state, err := provider.service(t.Context()).quizRepository().LoadRuntimeState(t.Context(), quizcore.Scope{ProjectID: provider.projectContext.ProjectID})
	if err != nil || state.Status != quizcore.RuntimeStatusDone || state.AnswersRecorded != 4 || state.SkipsRecorded != 1 {
		t.Fatal("Done submitted the unsent answer", state, err)
	}
}

func webQuizBrowserCancel(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider, base string) {
	t.Helper()

	started, err := provider.service(t.Context()).StartQuiz(QuizStartRequest{Scope: quizcore.Scope{CaseSlug: "named-work"}})
	if err != nil {
		t.Fatal(err)
	}

	browser.navigate(t, base+quizFocusURL(started.State.CuratedPath))
	browser.wait(t, "#quiz-answer")
	webBoardBrowserSet(t, browser, "#quiz-answer", "An unsent answer must not become a recorded answer on Cancel.")
	webBoardBrowserClick(t, browser, "[value='quiz.cancel']")
	webBoardBrowserWait(t, browser, "location.pathname==='/quizzes'")

	state, err := provider.service(t.Context()).quizRepository().LoadRuntimeState(t.Context(), quizcore.Scope{ProjectID: provider.projectContext.ProjectID, CaseSlug: "named-work"})
	if err != nil || state.Status != quizcore.RuntimeStatusCancelled || state.AnswersRecorded != 0 {
		t.Fatal("Cancel completed or submitted the quiz", state, err)
	}
}

func webQuizBrowserValue(t *testing.T, browser *webWorkflowBrowser, want string) {
	t.Helper()

	if got := browser.evaluate(t, "document.querySelector('#quiz-answer').value"); got != want {
		t.Fatal("question draft changed", got, want)
	}
}

func webQuizBrowserChangeMode(t *testing.T, browser *webWorkflowBrowser, mode string) {
	t.Helper()
	webBoardBrowserReady(t, browser)

	webBoardBrowserClick(t, browser, `label[for="quiz-mode-`+mode+`"]`)
}

func webQuizBrowserModeFailure(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	webQuizBrowserChangeMode(t, browser, "deep")
	webBoardBrowserWait(t, browser, "document.querySelector('#quiz-result').textContent.includes('Your draft is retained')&&document.querySelector('#quiz-mode input:checked').value==='fast'")
	webQuizBrowserValue(t, browser, "An unsent answer must survive refresh and mode changes.")
}
