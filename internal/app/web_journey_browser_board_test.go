package app

import (
	"reflect"
	"testing"

	"forge.adrianpk.com/chr/internal/webui"
)

func webConsolidatedBoardComment(t *testing.T, browser *webWorkflowBrowser, j *webJourney) {
	t.Helper()

	before := journeyMapFiles(t, j.provider.projectContext.ProjectDir)
	webBoardBrowserClick(t, browser, `a[data-ticket-key="`+j.ticket+`"]`)
	browser.wait(t, "#ticket-title")

	if browser.evaluate(t, `String(document.querySelector('#ticket-title').readOnly&&!document.querySelector('#ticket-title').name&&document.querySelector('#ticket-title-guard').textContent.includes('approved plan order'))`) != "true" {
		t.Fatal("bound title still offered an unsupported rename")
	}

	webBoardBrowserSet(t, browser, `[name="owner"]`, "Browser journey owner")
	webBoardBrowserSet(t, browser, `.ticket-comments__form [name="content"]`, "Independent journey comment")
	webBoardBrowserClick(t, browser, ".ticket-comments__form button")
	webBoardBrowserWait(t, browser, `document.querySelector('#ticket-discussion-content').textContent.includes('Independent journey comment')`)

	if browser.evaluate(t, `document.querySelector('[name=owner]').value`) != "Browser journey owner" {
		t.Fatal("comment discarded dirty ownership")
	}

	webBoardBrowserClick(t, browser, `button[form="ticket-editor-form"]`)
	webBoardBrowserWait(t, browser, `!document.querySelector('#modal-root [role=dialog]')`)

	detail, err := j.provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: j.ticket})
	if err != nil || detail.Fields["owner"] != "Browser journey owner" || len(detail.Comments) != 1 {
		t.Fatal(detail, err)
	}

	if !reflect.DeepEqual(before, journeyMapFiles(t, j.provider.projectContext.ProjectDir)) {
		t.Fatal("browser metadata or comment changed the order map")
	}
}
