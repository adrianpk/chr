package app

import (
	"os"
	"strconv"
	"testing"
)

// Cancellation is an explicit workflow stop. Declining its confirmation must
// leave the current input, control binding and read-only observation usable.
func assertGuidedCancelDeclined(t *testing.T, browser *webWorkflowBrowser, selector string) {
	t.Helper()
	browser.evaluate(t, `window.testOriginalConfirm=window.confirm;window.testCancelMessage='';window.confirm=message=>{window.testCancelMessage=message;return false};'ready'`)
	browser.click(t, selector)
	webBoardBrowserWait(t, browser, `window.testCancelMessage.includes('Cancel this workflow?') && window.testCancelMessage.includes('will be kept')`)

	if busy := browser.evaluate(t, `String(document.querySelector('.case-workflow').hasAttribute('aria-busy'))`); busy != "false" {
		t.Fatal("declined cancellation disabled the workflow")
	}

	browser.evaluate(t, `window.confirm=window.testOriginalConfirm;delete window.testOriginalConfirm;'restored'`)
}

func assertGuidedControlPanel(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	browser.wait(t, `.workflow-progress .ticket-editor-footer`)

	geometry := browser.evaluate(t, `(()=>{
const panel=document.querySelector('.workflow-progress'),footer=panel.querySelector('.ticket-editor-footer'),title=document.querySelector('.page-header h1');
const buttons=Array.from(footer.querySelectorAll('button,a.button'));
return String(Math.abs(panel.getBoundingClientRect().left-title.getBoundingClientRect().left)<1 && buttons.every(button=>panel.contains(button)) && !document.querySelector('.case-workflow > .workflow-controls'));
})()`)
	if geometry != "true" {
		t.Fatal("lifecycle actions lost their contextual panel or title alignment")
	}

	for _, action := range []string{"pause", "cancel"} {
		selector := `form:has(input[name="step"][value="` + action + `"])`
		if got := browser.evaluate(t, `String(Boolean(document.querySelector(`+strconv.Quote(selector)+`)?.closest('.workflow-progress .ticket-editor-footer')))`); got != "true" {
			t.Fatal("lifecycle action is detached from execution progress", action)
		}
	}

	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_GUIDED_SCREENSHOTS"), "execution-controls-desktop")
	webBoardBrowserViewport(t, browser, 390, 844)

	if overflow := browser.evaluate(t, `String(document.documentElement.scrollWidth>innerWidth)`); overflow != "false" {
		t.Fatal("workflow actions overflow a narrow viewport")
	}

	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_GUIDED_SCREENSHOTS"), "execution-controls-mobile")
	webBoardBrowserViewport(t, browser, 1440, 1000)
}
