package app

import (
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Chromium submits the rendered forms against real HTTP, Git workers and retained
// state. The repository smoke script enables this owned browser fixture explicitly.
func TestWebWorkflowBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("browser workflow integration runs in the nightly suite")
	}

	browserPath := os.Getenv("CHR_WEB_BROWSER")
	if browserPath == "" {
		t.Skip("run scripts/qa/web-ui-smoke.sh to enable Chromium workflow actions")
	}

	browser := newWebWorkflowBrowser(t, browserPath)
	t.Run("activation and continuation", func(t *testing.T) {
		webWorkflowBrowserActivation(t, browser)
	})
	t.Run("review wait and exact approval", func(t *testing.T) {
		webWorkflowBrowserReview(t, browser)
	})
	t.Run("competing admission", func(t *testing.T) {
		webWorkflowBrowserOwnership(t, browser)
	})
}

func webWorkflowBrowserReview(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()

	operations, id, provider := serviceWorkflowReviewReady(t)
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n")

	if _, err := operations.ReviewWorkflowSlice(t.Context(), id, nil); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(webWorkflowHandler(t, operations))
	defer server.Close()

	browser.navigate(t, server.URL+"/workflows?id="+id)
	browser.wait(t, `[data-workflow-status="waiting_review"]`)

	if !strings.Contains(browser.evaluate(t, "document.body.innerText"), "operator") {
		t.Fatal("browser hid review owner")
	}

	head := provider.pr.HeadCommit
	provider.pr.HeadCommit = strings.Repeat("e", 40)

	browser.click(t, `[data-workflow-action="approve"]`)
	browser.wait(t, `[data-write-status="stale"]`)

	provider.pr.HeadCommit = head

	browser.click(t, `[data-workflow-action="approve"]`)
	browser.wait(t, `[data-write-status="ok"]`)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.ReviewState != "authorized" || status.ConfirmedTickets != 0 {
		t.Fatal("browser exact approval lost binding", status, err)
	}

	if got := browser.evaluate(t, "document.body.innerText"); !strings.Contains(got, "authorized") {
		t.Fatalf("browser approval not displayed: %s", got)
	}
}

func webWorkflowBrowserActivation(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()

	operations, record, _, provider := serviceWorkNativeFixture(t)

	server := httptest.NewServer(webWorkflowHandler(t, operations))
	defer server.Close()

	page := webWorkflowHTTP(t, webWorkflowHandler(t, operations), "GET", "/board", nil)
	if page.Code != 200 || !strings.Contains(page.Body.String(), record.Task.Key) {
		t.Fatal("native board missing ticket", page.Code, page.Body.String())
	}

	browser.navigate(t, server.URL+"/board")
	browser.click(t, `a[href="/board/ticket?key=`+record.Task.Key+`"]`)
	browser.clickNavigation(t, "[data-workflow-ticket]")
	browser.wait(t, "[data-workflow-preview]")
	browser.click(t, "[data-workflow-start]")
	browser.wait(t, "[data-workflow-id]")

	id := browser.evaluate(t, `document.querySelector('[data-workflow-id]').dataset.workflowId`)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.ValidatedTasks != 1 || status.ConfirmedTickets != 0 || status.Continuation == nil || provider.creates != 0 {
		t.Fatal("browser did not execute a checkpoint", status, err)
	}

	webWorkflowBrowserContinue(t, browser, server.URL, operations, status)
}

func webWorkflowBrowserOwnership(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()

	operations, hierarchy, _ := workflowCLIFixture(t)

	server := httptest.NewServer(webWorkflowHandler(t, operations))
	defer server.Close()

	browser.navigate(t, server.URL+"/workflows?extent=ticket&key="+hierarchy.Tasks[0].Task.Key)
	browser.wait(t, "[data-workflow-preview]")

	id := workflowCLIManualStart(t, operations, hierarchy)
	browser.click(t, "[data-workflow-start]")
	browser.wait(t, `[data-write-status]:not([data-write-status="ok"])`)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil || status.TotalAttempts != 0 || !strings.Contains(browser.evaluate(t, "document.body.innerText"), "Workflow action failed") {
		t.Fatal("browser conflict started work or hid failure", status, err)
	}
}

func webWorkflowBrowserContinue(t *testing.T, browser *webWorkflowBrowser, serverURL string, operations service, status WorkflowWorkStatus) {
	t.Helper()
	browser.click(t, `[data-workflow-action="continue"]`)
	browser.wait(t, `[data-workflow-id][data-workflow-revision]:not([data-workflow-revision="`+strconv.FormatUint(status.Revision, 10)+`"])`)

	continued, err := operations.WorkflowStatus(t.Context(), status.ID, WorkflowStatusRequest{Limit: 8})
	if err != nil || continued.CandidateHash == "" || continued.ConfirmedTickets != 0 {
		t.Fatal("browser continuation lost candidate", continued, err)
	}

	browser.navigate(t, serverURL+"/workflows?id="+status.ID)
	browser.wait(t, "[data-workflow-id]")

	if _, err := operations.StopWorkflow(t.Context(), status.ID, "pause"); err != nil {
		t.Fatal(err)
	}

	browser.click(t, `[data-workflow-action="continue"]`)
	browser.wait(t, `[data-write-status="stale"]`)

	if !strings.Contains(browser.evaluate(t, "document.body.innerText"), "paused") {
		t.Fatal("browser hid current state after stale action")
	}
}
