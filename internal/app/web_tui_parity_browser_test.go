package app

import (
	"maps"
	"path/filepath"
	"reflect"
	"testing"
)

// A real periodic HTTP observation must retain a focused unsent answer while a
// peer saves an idea. Resize and declined cancellation preserve that same text.
func assertParityBrowserPolling(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider, runURL string) {
	t.Helper()

	operations := provider.service(t.Context())
	before := readWebGuidedSession(t, operations, guidedBrowserRunID(t, provider, runURL))
	browser.evaluate(t, `document.querySelector('.quiz-response textarea').focus(); 'focused'`)
	browser.call(t, "Input.insertText", map[string]any{"text": "BROWSER_UNSENT_BEFORE_POLL"})
	assertParityBrowserDraft(t, browser, "BROWSER_UNSENT_BEFORE_POLL")

	saved, err := operations.ChangeGuidedUseCase(t.Context(), guidedQuizCommand(before, "browser-parity-idea", "idea", "Keep browser polling evidence separate."))
	if err != nil {
		t.Fatal(err)
	}

	image := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	webBoardBrowserWait(t, browser, `document.querySelector('[data-workflow-feedback]')?.textContent.includes('Your unsent text is still here')`)
	assertParityBrowserDraft(t, browser, "BROWSER_UNSENT_BEFORE_POLL")
	webBoardBrowserViewport(t, browser, 390, 844)
	assertParityBrowserDraft(t, browser, "BROWSER_UNSENT_BEFORE_POLL")
	browser.call(t, "Input.insertText", map[string]any{"text": "_AFTER_RESIZE"})
	webBoardBrowserViewport(t, browser, 1440, 1000)
	assertParityBrowserDraft(t, browser, "BROWSER_UNSENT_BEFORE_POLL_AFTER_RESIZE")
	assertGuidedCancelDeclined(t, browser, `[value="quiz.cancel"]`)
	webQuizBrowserValue(t, browser, "BROWSER_UNSENT_BEFORE_POLL_AFTER_RESIZE")

	after := readWebGuidedSession(t, operations, before.ID)
	if after.Revision != saved.Revision || after.Guided.Stage != before.Guided.Stage || !reflect.DeepEqual(after.Guided.Quiz, before.Guided.Quiz) || !maps.Equal(image, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("polling or declined cancellation submitted input or advanced work")
	}

	// Clearing the unsent field and taking the offered Reload is explicit. The
	// rest of the existing journey keeps its request/retry/unload assertions.
	webBoardBrowserSet(t, browser, ".quiz-response textarea", "")
	browser.clickNavigation(t, `[data-workflow-feedback] a`)
	browser.wait(t, ".quiz-response textarea")
}

func assertParityBrowserDraft(t *testing.T, browser *webWorkflowBrowser, value string) {
	t.Helper()
	webQuizBrowserValue(t, browser, value)

	if browser.evaluate(t, `String(document.activeElement===document.querySelector('.quiz-response textarea') && location.pathname==='/quizzes')`) != "true" {
		t.Fatal("read or resize changed the selected view or input focus")
	}
}
