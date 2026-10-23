package app

import (
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
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

	server := httptest.NewServer(guidedWebHandler(t, provider))
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	webBoardBrowserViewport(t, browser, 1440, 1000)

	runURL := captureGuidedBrowserInput(t, browser, server.URL, host)
	browser.navigate(t, runURL)
	guidedBrowserSubmit(t, browser, "answer", "Keep the exact bounded behavior and review policy.")
	host.wait.Wait()
	browser.navigate(t, runURL)
	assertGuidedBrowserText(t, browser, "Confirmed delivery: 0")
	assertGuidedBrowserText(t, browser, "Review and approve the native work")
	guidedBrowserSubmit(t, browser, "pause", "")
	assertGuidedBrowserText(t, browser, "paused")
	guidedBrowserSubmit(t, browser, "resume", "")
	host.wait.Wait()
	browser.navigate(t, runURL)
	guidedBrowserSubmit(t, browser, "idea", "Add an audit export after separate review.")
	assertGuidedBrowserText(t, browser, "Pending review")
	approveGuidedBrowserWork(t, browser, server.URL, provider, runURL)

	operations := provider.service(t.Context())
	id := guidedBrowserRunID(t, runURL)
	session := readWebGuidedSession(t, operations, id)
	publishGuidedWebResult(t, operations, host, publication, session.Guided.Work.Invocation.ID)
	webBoardBrowserViewport(t, browser, 390, 844)
	browser.navigate(t, runURL)

	for _, text := range []string{"All admitted work has confirmed delivery", "Remaining: 0", "Confirmed set closures: 1", "Target branch: dev", "audit export", "Pending review"} {
		assertGuidedBrowserText(t, browser, text)
	}

	browser.click(t, `details:has(summary) summary`)

	if publication.creates != 1 || len(publication.submissions) != 1 {
		t.Fatal("browser journey repeated publication")
	}
}

func captureGuidedBrowserInput(t *testing.T, browser *webWorkflowBrowser, serverURL string, host *guidedWorkHost) string {
	t.Helper()
	browser.navigate(t, serverURL+"/guided")
	browser.clickNavigation(t, `form:has(input[value="guided.start"]) button`)

	runURL := browser.evaluate(t, "location.href")
	guidedBrowserSubmit(t, browser, "name", "Browser guided delivery")
	guidedBrowserSubmit(t, browser, "append", "Deliver a bounded workflow from the original context.")
	guidedBrowserSubmit(t, browser, "finish", "")
	host.wait.Wait()

	return runURL
}

func guidedBrowserSubmit(t *testing.T, browser *webWorkflowBrowser, step, content string) {
	t.Helper()

	selector := `form:has(input[name="step"][value="` + step + `"])`
	browser.wait(t, selector)

	if content != "" {
		browser.evaluate(t, `(()=>{document.querySelector(`+strconv.Quote(selector)+`).querySelector('[name=content]').value=`+strconv.Quote(content)+`;return 'set';})()`)
	}

	browser.clickNavigation(t, selector+" button[type=submit]")
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
	browser.wait(t, `[data-write-status="ok"]`)
}
