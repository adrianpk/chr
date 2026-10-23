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
	reopenGuidedBrowser(t, browser, server.URL)
	assertGuidedBrowserText(t, browser, "0 delivered")
	assertGuidedBrowserText(t, browser, "The work needs review")
	guidedBrowserSubmit(t, browser, "pause", "")
	assertGuidedBrowserText(t, browser, "Paused")
	guidedBrowserSubmit(t, browser, "resume", "")
	host.wait.Wait()
	reopenGuidedBrowser(t, browser, server.URL)
	approveGuidedBrowserWork(t, browser, server.URL, provider, runURL)

	id := guidedBrowserRunID(t, provider, runURL)
	session := readWebGuidedSession(t, operations, id)
	publishGuidedWebResult(t, operations, host, publication, session.Guided.Work.Invocation.ID)
	webBoardBrowserViewport(t, browser, 390, 844)
	reopenGuidedBrowser(t, browser, server.URL)

	assertGuidedBrowserResult(t, browser)

	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_GUIDED_SCREENSHOTS"), "results-mobile")
	browser.click(t, `details:has(summary) summary`)

	if publication.creates != 1 || len(publication.submissions) != 1 {
		t.Fatal("browser journey repeated publication")
	}
}

func captureGuidedBrowserInput(t *testing.T, browser *webWorkflowBrowser, serverURL string, host *guidedWorkHost) string {
	t.Helper()
	browser.navigate(t, serverURL+"/cases")
	browser.clickNavigation(t, `a[href*="new=workflow"]`)
	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_GUIDED_SCREENSHOTS"), "name-desktop")
	assertGuidedBrowserName(t, browser)

	guidedBrowserSubmit(t, browser, "name", "Browser guided delivery")

	runURL := browser.evaluate(t, "location.href")
	assertGuidedBrowserTab(t, browser, "/dumps")

	if got := browser.evaluate(t, `String(new URL(location.href).searchParams.has("case"))`); got != "true" {
		t.Fatal("workflow lost its functional use-case reference", got)
	}

	guidedBrowserSubmit(t, browser, "append", "Deliver a bounded workflow from the original context.")
	browser.clickNavigation(t, `button[form="workflow-finish"]`)
	host.wait.Wait()
	browser.wait(t, `.quiz-response textarea`)
	assertGuidedBrowserTab(t, browser, "/quizzes")
	assertGuidedBrowserNavigation(t, browser)

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

func guidedBrowserRunID(t *testing.T, provider webViewProvider, location string) string {
	t.Helper()

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatal(err)
	}

	view, err := provider.CaseWorkflowView(t.Context(), parsed.Query().Get("case"), parsed.Query().Get("namespace"))
	if err != nil || view.RunID == "" {
		t.Fatal("browser did not retain its use-case context", err)
	}

	return view.RunID
}

func approveGuidedBrowserWork(t *testing.T, browser *webWorkflowBrowser, serverURL string, provider webViewProvider, runURL string) {
	t.Helper()

	session := readWebGuidedSession(t, provider.service(t.Context()), guidedBrowserRunID(t, provider, runURL))
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

func assertGuidedBrowserName(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()

	visible := browser.evaluate(t, "document.body.innerText")
	for _, extra := range []string{"New use case", "Name your use case", "What do you want", "← Cases", "Pause", "Cancel", "Development"} {
		if strings.Contains(visible, extra) {
			t.Fatal("redundant initial screen copy", extra)
		}
	}

	if count := browser.evaluate(t, `String(document.querySelectorAll('nav').length)`); count != "1" {
		t.Fatal("duplicated navigation", count)
	}
}

func assertGuidedBrowserTab(t *testing.T, browser *webWorkflowBrowser, path string) {
	t.Helper()

	if got := browser.evaluate(t, "location.pathname"); got != path {
		t.Fatal("workflow did not use the product screen", path, got)
	}

	if got := browser.evaluate(t, `document.querySelector('nav[aria-label="Primary"] .app-nav__link--active')?.textContent.trim().toLowerCase() || ''`); got != strings.TrimPrefix(path, "/") {
		t.Fatal("primary navigation did not show the current screen")
	}
}

func reopenGuidedBrowser(t *testing.T, browser *webWorkflowBrowser, serverURL string) {
	t.Helper()
	browser.navigate(t, serverURL+"/cases")
	browser.clickNavigation(t, `a.button[href*="case="]`)
	browser.wait(t, `.case-workflow`)
}

func assertGuidedBrowserNavigation(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	browser.clickNavigation(t, `nav a[href^="/dumps?"]`)
	assertGuidedBrowserTab(t, browser, "/dumps")
	assertGuidedBrowserText(t, browser, "Deliver a bounded workflow")

	if got := browser.evaluate(t, `String(document.querySelectorAll('.case-workflow textarea').length)`); got != "0" {
		t.Fatal("committed dump became editable")
	}

	browser.clickNavigation(t, `nav a[href^="/quizzes?"]`)
	browser.wait(t, `.quiz-response textarea`)
	assertGuidedBrowserTab(t, browser, "/quizzes")
}

func assertGuidedBrowserResult(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()

	for _, text := range []string{"All admitted work has confirmed delivery", "0 remaining", "View delivered changes", "dev"} {
		assertGuidedBrowserText(t, browser, text)
	}

	if visible := browser.evaluate(t, "document.body.innerText"); strings.Contains(visible, "Paused.") {
		t.Fatal("completed work was shown as paused")
	}

	if count := browser.evaluate(t, `String(document.querySelectorAll('.board-grid .ticket-card').length)`); count != "1" {
		t.Fatal("Board did not retain the selected use case", count)
	}
}
