package app

import (
	"net/http/httptest"
	"testing"
)

func webBoardBrowserNative(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()

	operations, hierarchy, _ := workflowCLIFixture(t)

	server := httptest.NewServer(webWorkflowHandler(t, operations))
	defer server.Close()

	key := hierarchy.Tasks[0].Task.Key

	webBoardBrowserViewport(t, browser, 1440, 1000)
	browser.navigate(t, server.URL+"/board")
	webBoardBrowserClick(t, browser, "a[data-ticket-key='"+key+"']")
	browser.wait(t, "#ticket-description[readonly]")
	webBoardBrowserWait(t, browser, "!document.querySelector('[name=use_case]') && Boolean(document.querySelector('.ticket-system-phase'))")
	webBoardBrowserSet(t, browser, "#ticket-title", "Descriptive native title")
	webBoardBrowserSet(t, browser, "[name='content']", "Native discussion remains independent.")
	webBoardBrowserClick(t, browser, ".ticket-comments__form button[type='submit']")
	browser.wait(t, ".ticket-comment")

	if got := browser.evaluate(t, "document.querySelector('#ticket-title').value"); got != "Descriptive native title" {
		t.Fatal("native comment discarded dirty title", got)
	}

	webBoardBrowserClick(t, browser, "button[form='ticket-editor-form']")
	browser.wait(t, "#modal-root:empty")

	document, err := operations.ReadTicketDocument(key)
	if err != nil || document.Ticket.Title != "Descriptive native title" || document.Fields["use_case"] != hierarchy.Tasks[0].Task.UseCase || document.Description == "" {
		t.Fatal("native editor lost metadata or immutable association", document, err)
	}
}
