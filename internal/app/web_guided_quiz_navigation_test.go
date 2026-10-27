package app

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
)

// Fail one request before admission so the real browser must retain its exact
// draft and retry identity, then advance without an unsaved-change prompt.
func guidedQuizBrowserServer(t *testing.T, provider webViewProvider, ids chan<- string) *httptest.Server {
	t.Helper()

	handler := guidedWebHandler(t, provider)

	var failed atomic.Bool

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/actions" && r.FormValue("action") == "quiz.answer" {
			ids <- r.FormValue("request_id")

			if !failed.Swap(true) {
				http.Error(w, "answer request unavailable", http.StatusServiceUnavailable)

				return
			}
		}

		handler.ServeHTTP(w, r)
	}))
}

func observeGuidedQuizUnload(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	browser.call(t, "Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": `document.addEventListener('DOMContentLoaded',()=>window.addEventListener('beforeunload',event=>{
if(event.isTrusted && event.defaultPrevented)sessionStorage.setItem('test-guarded-unloads',String(Number(sessionStorage.getItem('test-guarded-unloads')||0)+1));
}));`})
}

func assertGuidedQuizFailedAnswer(t *testing.T, browser *webWorkflowBrowser, ids <-chan string) string {
	t.Helper()

	answer := "Keep the exact bounded behavior and review policy."
	webBoardBrowserSet(t, browser, "#quiz-answer", answer)
	browser.click(t, `[value="quiz.answer"]`)
	webBoardBrowserWait(t, browser, `document.querySelector('#quiz-result').textContent.includes('Your draft is retained')`)
	webQuizBrowserValue(t, browser, answer)

	if browser.evaluate(t, `String(document.querySelector('#quiz-surface').dataset.quizQuestion==='question-1' && document.querySelectorAll('.quiz-history-card').length===0)`) != "true" {
		t.Fatal("failed answer changed the question or history")
	}

	guarded := browser.evaluate(t, `(()=>{const event=new Event('beforeunload',{cancelable:true});window.dispatchEvent(event);return String(event.defaultPrevented)})()`)
	if guarded != "true" {
		t.Fatal("failed request disabled unsent-answer protection")
	}

	return <-ids
}

func assertGuidedQuizAcceptedAnswer(t *testing.T, browser *webWorkflowBrowser, first, retry string) {
	t.Helper()

	if first != retry {
		t.Fatal("retry changed its request identity")
	}

	if got := browser.evaluate(t, `sessionStorage.getItem('test-guarded-unloads')||'0'`); got != "0" {
		t.Fatal("accepted answer triggered an unsaved-change prompt", got)
	}

	answer := "Keep the exact bounded behavior and review policy."

	webBoardBrowserWait(t, browser, `document.querySelectorAll('.quiz-history-card').length===1`)

	if got := browser.evaluate(t, `String(document.querySelector('.quiz-history-card p').textContent===`+strconv.Quote(answer)+`)`); got != "true" {
		t.Fatal("accepted answer did not appear in history")
	}

	webQuizBrowserValue(t, browser, "")

	if count := browser.evaluate(t, `String(document.querySelectorAll('[data-quiz-retained] details').length)`); count != "0" {
		t.Fatal("accepted answer remained marked as unsent")
	}
}

func assertGuidedQuizHistory(t *testing.T, browser *webWorkflowBrowser, operations service, session authoringcore.AuthoringSession) {
	t.Helper()

	if got := browser.evaluate(t, `sessionStorage.getItem('test-guarded-unloads')||'0'`); got != "0" {
		t.Fatal("accepted quiz answers triggered a navigation warning", got)
	}

	observed := observeGuidedQuiz(t, operations, session)
	if len(observed.Entries) != 2 {
		t.Fatal("quiz history lost or duplicated an answer", observed.Entries)
	}

	expected := map[string]string{
		"question-1": "Keep the exact bounded behavior and review policy.",
		"question-2": "Preserve the existing interfaces and review the delivered change.",
	}
	for _, entry := range observed.Entries {
		if answer, ok := expected[entry.QuestionID]; !ok || entry.Answer != answer {
			t.Fatal("quiz answer was associated with the wrong question", entry)
		}

		delete(expected, entry.QuestionID)
	}
}
