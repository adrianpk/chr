package app

import (
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// TestGuidedUseCaseWebBrowser observes real Chromium form submission, navigation,
// native approval and final evidence against only owned local Git/provider state.
func TestGuidedUseCaseWebBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned Chromium journey is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider, host, publication := guidedWebJourneyFixture(t, false)
	operations := provider.service(t.Context())
	guidedBrowserSecondQuestion(t, operations)

	server := httptest.NewServer(guidedWebHandler(t, provider))
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	webBoardBrowserViewport(t, browser, 1440, 1000)

	runURL := captureGuidedBrowserInput(t, browser, server.URL, host)
	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_GUIDED_SCREENSHOTS"), "quiz-desktop")
	guidedBrowserSubmit(t, browser, "answer", "Keep the exact bounded behavior and review policy.")
	host.wait.Wait()
	browser.wait(t, `[data-quiz-question="question-2"]`)
	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_GUIDED_SCREENSHOTS"), "quiz-second-question")
	guidedBrowserSubmit(t, browser, "answer", "Preserve the existing interfaces and review the delivered change.")
	host.wait.Wait()
	browser.navigate(t, runURL)
	assertGuidedBrowserText(t, browser, "0 delivered")
	assertGuidedBrowserText(t, browser, "The work needs review")
	guidedBrowserSubmit(t, browser, "pause", "")
	assertGuidedBrowserText(t, browser, "Paused")
	guidedBrowserSubmit(t, browser, "resume", "")
	host.wait.Wait()
	browser.navigate(t, runURL)
	approveGuidedBrowserWork(t, browser, server.URL, provider, runURL)

	id := guidedBrowserRunID(t, runURL)
	session := readWebGuidedSession(t, operations, id)
	publishGuidedWebResult(t, operations, host, publication, session.Guided.Work.Invocation.ID)
	webBoardBrowserViewport(t, browser, 390, 844)
	browser.navigate(t, runURL)

	for _, text := range []string{"All admitted work has confirmed delivery", "0 remaining", "View delivered changes", "dev"} {
		assertGuidedBrowserText(t, browser, text)
	}

	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_GUIDED_SCREENSHOTS"), "results-mobile")
	browser.click(t, `details:has(summary) summary`)

	if publication.creates != 1 || len(publication.submissions) != 1 {
		t.Fatal("browser journey repeated publication")
	}
}

func captureGuidedBrowserInput(t *testing.T, browser *webWorkflowBrowser, serverURL string, host *guidedWorkHost) string {
	t.Helper()
	browser.navigate(t, serverURL+"/cases")
	browser.clickNavigation(t, `form:has(input[value="guided.start"]) button`)

	runURL := browser.evaluate(t, "location.href")
	guidedBrowserSubmit(t, browser, "name", "Browser guided delivery")

	if got := browser.evaluate(t, `String(new URL(location.href).searchParams.has("case"))`); got != "true" {
		t.Fatal("workflow lost its functional use-case reference", got)
	}

	guidedBrowserSubmit(t, browser, "append", "Deliver a bounded workflow from the original context.")
	browser.clickNavigation(t, `button[form="workflow-finish"]`)
	host.wait.Wait()
	browser.wait(t, `.quiz-response textarea`)

	return runURL
}

func guidedBrowserSubmit(t *testing.T, browser *webWorkflowBrowser, step, content string) {
	t.Helper()

	selector := `form:has(input[name="step"][value="` + step + `"])`
	browser.wait(t, selector)

	if content != "" {
		browser.evaluate(t, `(()=>{document.querySelector(`+strconv.Quote(selector)+`).querySelector('[name=content]').value=`+strconv.Quote(content)+`;return 'set';})()`)
	}

	before := browser.evaluate(t, `document.querySelector('.case-workflow').dataset.workflowState`)
	browser.clickNavigation(t, selector+" button[type=submit]")
	browser.wait(t, `.case-workflow:not([data-workflow-state="`+before+`"])`)
}

func assertGuidedBrowserText(t *testing.T, browser *webWorkflowBrowser, text string) {
	t.Helper()

	if got := browser.evaluate(t, "document.body.innerText"); !strings.Contains(got, text) {
		t.Fatal("browser omitted journey state", text, got)
	}
}

func guidedBrowserRunID(t *testing.T, location string) string {
	t.Helper()

	parsed, err := url.Parse(location)
	if err != nil || parsed.Query().Get("id") == "" {
		t.Fatal("browser did not retain the journey URL", err)
	}

	return parsed.Query().Get("id")
}

func approveGuidedBrowserWork(t *testing.T, browser *webWorkflowBrowser, serverURL string, provider webViewProvider, runURL string) {
	t.Helper()

	session := readWebGuidedSession(t, provider.service(t.Context()), guidedBrowserRunID(t, runURL))
	browser.navigate(t, serverURL+"/workflows?id="+session.Guided.Work.Invocation.ID)
	browser.click(t, `[data-workflow-action="approve"]`)
	browser.wait(t, `[data-write-status]`)

	if status := browser.evaluate(t, `document.querySelector("[data-write-status]").dataset.writeStatus`); status != "ok" {
		t.Fatal("native approval failed", status, browser.evaluate(t, "document.body.innerText"))
	}
}

func guidedBrowserSecondQuestion(t *testing.T, operations service) {
	t.Helper()

	model, ok := operations.quizModel.(*guidedWebWorkModel)
	if !ok {
		t.Fatal("missing owned quiz provider")
	}

	model.quiz.responses = []core.QuizModelResponse{model.quiz.responses[0], serviceQuizModelResponse("second", "question-2", operations.utcNow()), model.quiz.responses[1]}
}
