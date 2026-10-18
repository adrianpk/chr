package app

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/webui"
)

// The real browser protects dirty forms, guarded moves and reachable responsive
// controls through server-rendered HTTP and an isolated canonical ticket store.
func TestWebBoardBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned browser integration is explicitly selected")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	browser := newWebWorkflowBrowser(t, path)
	provider := webBoardBrowserFixture(t)
	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	webBoardBrowserViewport(t, browser, 1440, 1000)
	browser.navigate(t, server.URL+"/board")
	browser.wait(t, "[data-board-column='done']")
	webBoardBrowserCapture(t, browser, "desktop-board")
	browser.evaluate(t, `const grid=document.querySelector('.board-grid'); grid.scrollLeft=grid.scrollWidth; 'scrolled'`)
	webBoardBrowserCapture(t, browser, "desktop-board-right")
	browser.evaluate(t, `document.querySelector('.board-grid').scrollLeft=280; 'scrolled'`)
	webBoardBrowserEdit(t, browser, provider)
	webBoardBrowserMoves(t, browser, provider)
	webBoardBrowserConflict(t, browser, provider)
	webBoardBrowserNew(t, browser, provider)
	webBoardBrowserNarrow(t, browser)
	webBoardBrowserNative(t, browser)
}

func webBoardBrowserFixture(t *testing.T) webViewProvider {
	t.Helper()

	provider := webTicketEditorProvider(t, false)

	for i, status := range []string{"backlog", "ready_for_dev", "in_progress", "ready_for_review", "ready_for_merge", "done", "blocked"} {
		content := fmt.Sprintf("---\nkey: CHR-%d\ntitle: Investigate the complete workflow and preserve every recorded observation\nstatus: %s\npriority: high\nseverity: normal\ntype: task\nproject_area: ui\nuse_case: named-work\nassignee: Developer\nowner: Maintainer\nbranch: work/board\ncommits: []\ncustom:\n  retained: yes\n---\n# Description\n\nFull description with a second paragraph retained by the shared editor.\n\nDetails remain editable across maximize and restore.\n\n# Comments\n\n## observer · 2026-10-18 10:00\n\nOriginal discussion.\n", i+1, status)
		writeServiceWorkspaceFile(t, filepath.Join(provider.projectContext.ProjectDir, "ops", "default", "ticket", fmt.Sprintf("CHR-%d.md", i+1)), content)
	}

	return provider
}

func webBoardBrowserEdit(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider) {
	t.Helper()
	browser.click(t, "a[data-ticket-key='CHR-1']")
	browser.wait(t, "#ticket-description")
	webBoardBrowserCapture(t, browser, "desktop-edit")
	webBoardBrowserSet(t, browser, "#ticket-title", "Unsaved Board title")
	webBoardBrowserSet(t, browser, "#ticket-description", "Unsaved complete description.\n\nRetained paragraph.")
	webBoardBrowserSet(t, browser, "[name='assignee']", "Board developer")
	webBoardBrowserSet(t, browser, "[name='status']", "ready_for_dev")
	browser.click(t, "[data-ticket-maximize]")
	browser.wait(t, ".ticket-modal--maximized")
	webBoardBrowserCapture(t, browser, "desktop-maximized")
	browser.click(t, "[data-ticket-maximize]")
	webBoardBrowserWait(t, browser, "!document.querySelector('.ticket-modal--maximized')")
	webBoardBrowserCheckDirty(t, browser)
	browser.evaluate(t, `document.querySelector('.ticket-modal__body').scrollTop = 100000; 'scrolled'`)
	webBoardBrowserCapture(t, browser, "desktop-edit-lower")
	webBoardBrowserSet(t, browser, "[name='content']", "Independent comment while metadata is dirty.")
	browser.click(t, ".ticket-comments__form button[type='submit']")
	browser.wait(t, ".ticket-comment:nth-child(2)")
	webBoardBrowserCheckDirty(t, browser)

	saved := webBoardBrowserDetail(t, provider)
	if saved.Title == "Unsaved Board title" || len(saved.Comments) != 2 {
		t.Fatal("comment saved metadata or lost discussion", saved)
	}

	browser.click(t, "button[form='ticket-editor-form']")
	browser.wait(t, "#modal-root:empty")

	saved = webBoardBrowserDetail(t, provider)
	if saved.Title != "Unsaved Board title" || saved.Fields["assignee"] != "Board developer" || saved.Status != "ready_for_dev" || !strings.Contains(saved.Description, "Retained paragraph.") || len(saved.Comments) != 2 {
		t.Fatal("browser save lost edited fields or discussion", saved)
	}
}

func webBoardBrowserMoves(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider) {
	t.Helper()

	scrollPosition := browser.evaluate(t, "String(document.querySelector('.board-grid').scrollLeft)")
	browser.wait(t, "[data-board-column='ready_for_dev'] a[data-ticket-key='CHR-1']")

	if got := browser.evaluate(t, "document.activeElement.dataset.ticketKey || ''"); got != "CHR-1" {
		t.Fatal("save refresh lost focus trigger", got)
	}

	webBoardBrowserDrag(t, browser, "backlog")
	browser.wait(t, "[data-board-column='backlog'] a[data-ticket-key='CHR-1']")
	webBoardBrowserDrag(t, browser, "done")
	browser.wait(t, "#board-notice [data-write-status='validation_failed']")

	if got := browser.evaluate(t, "String(document.querySelector('.board-grid').scrollLeft)"); got != scrollPosition {
		t.Fatal("Board refresh lost horizontal position", scrollPosition, got)
	}

	if saved := webBoardBrowserDetail(t, provider); saved.Status != "backlog" {
		t.Fatal("rejected drag changed status", saved.Status)
	}

	if !strings.Contains(browser.evaluate(t, "document.querySelector('#board-notice').innerText"), "integration") {
		t.Fatal("drag rejection hid reason")
	}

	webBoardBrowserExternalEdit(t, provider, "External revision before drag")
	webBoardBrowserDrag(t, browser, "ready_for_dev")
	browser.wait(t, "#board-notice [data-write-status='stale']")
	browser.wait(t, "[data-board-column='backlog'] a[data-ticket-key='CHR-1']")

	if saved := webBoardBrowserDetail(t, provider); saved.Status != "backlog" {
		t.Fatal("stale drag changed status", saved.Status)
	}
}

func webBoardBrowserConflict(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider) {
	t.Helper()
	browser.click(t, "a[data-ticket-key='CHR-1']")
	browser.wait(t, "#ticket-title")
	webBoardBrowserSet(t, browser, "#ticket-title", "Retain stale input")
	webBoardBrowserSet(t, browser, "[name='content']", "Retain stale comment")
	webBoardBrowserExternalEdit(t, provider, "Current external title")
	browser.click(t, ".ticket-comments__form button[type='submit']")
	browser.wait(t, "#ticket-comment-result [data-write-status='stale']")

	if got := browser.evaluate(t, "document.querySelector('[name=content]').value"); got != "Retain stale comment" {
		t.Fatal("stale comment discarded input", got)
	}

	browser.click(t, "button[form='ticket-editor-form']")
	browser.wait(t, "#ticket-save-result [data-write-status='stale']")

	if got := browser.evaluate(t, "document.querySelector('#ticket-title').value"); got != "Retain stale input" {
		t.Fatal("stale save discarded input", got)
	}

	browser.click(t, "#ticket-save-result a")
	webBoardBrowserWait(t, browser, `document.querySelector('#ticket-title').value === 'Current external title'`)
	webBoardBrowserFocus(t, browser)
	browser.click(t, "[data-modal-close]")
	browser.wait(t, "#modal-root:empty")

	if got := browser.evaluate(t, "document.activeElement.dataset.ticketKey || ''"); got != "CHR-1" {
		t.Fatal("close lost focus trigger", got)
	}
}

func webBoardBrowserNew(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider) {
	t.Helper()
	browser.click(t, "a[href='/board/ticket?new=1']")
	browser.wait(t, "#ticket-title")

	if got := browser.evaluate(t, `String(document.querySelector('#ticket-title').value === '' && !document.querySelector('.ticket-comments'))`); got != "true" {
		t.Fatal("New inherited ticket or discussion")
	}

	webBoardBrowserCapture(t, browser, "desktop-new")
	webBoardBrowserSet(t, browser, "#ticket-title", "Created in shared modal")
	webBoardBrowserSet(t, browser, "#ticket-description", "Complete independent description.")
	webBoardBrowserSet(t, browser, "[name='use_case']", "named-work")
	browser.click(t, "button[form='ticket-editor-form']")
	browser.wait(t, "#modal-root:empty")
	browser.wait(t, "a[data-ticket-key='CHR-8']")

	created, err := provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: "CHR-8"})
	if err != nil || created.Title != "Created in shared modal" || created.UseCase != "named-work" || created.Slice != "" || len(created.Comments) != 0 {
		t.Fatal("New lost independent identity", created, err)
	}
}

func webBoardBrowserNarrow(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	webBoardBrowserViewport(t, browser, 390, 844)
	webBoardBrowserCapture(t, browser, "narrow-board")
	browser.click(t, "a[data-ticket-key='CHR-1']")
	browser.wait(t, "#ticket-description")
	webBoardBrowserCapture(t, browser, "narrow-edit")
	browser.evaluate(t, `document.querySelector('.ticket-modal__body').scrollTop = 100000; 'scrolled'`)
	webBoardBrowserCapture(t, browser, "narrow-edit-lower")
	webBoardBrowserWait(t, browser, `Array.from(document.querySelectorAll('[data-ticket-editor] input, [data-ticket-editor] textarea, [data-ticket-editor] select')).filter(e=>e.type!=='hidden').every(e=>e.getBoundingClientRect().right<=innerWidth)`)
	browser.click(t, "[data-ticket-maximize]")
	browser.wait(t, ".ticket-modal--maximized")
	webBoardBrowserCapture(t, browser, "narrow-maximized")
	browser.click(t, "[data-modal-close]")
	browser.wait(t, "#modal-root:empty")
	browser.click(t, "a[href='/board/ticket?new=1']")
	browser.wait(t, "#ticket-title")
	webBoardBrowserCapture(t, browser, "narrow-new")
	browser.click(t, "[data-modal-close]")
	browser.wait(t, "#modal-root:empty")
}

func webBoardBrowserDetail(t *testing.T, provider webViewProvider) webui.TicketDetailView {
	t.Helper()

	view, err := provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: "CHR-1"})
	if err != nil {
		t.Fatal(err)
	}

	return view
}

func webBoardBrowserExternalEdit(t *testing.T, provider webViewProvider, title string) {
	t.Helper()

	view := webBoardBrowserDetail(t, provider)

	result, err := provider.WriteAction(t.Context(), webui.WriteActionRequest{Action: "ticket.update", ID: view.Key, Revision: view.Revision, Confirm: true, Fields: map[string]string{"title": title}})
	if err != nil || result.Status != webui.WriteStatusOK {
		t.Fatal("external fixture edit", result, err)
	}
}

func webBoardBrowserCheckDirty(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()

	if got := browser.evaluate(t, `document.querySelector('#ticket-title').value + '|' + document.querySelector('#ticket-description').value + '|' + document.querySelector('[name="assignee"]').value`); got != "Unsaved Board title|Unsaved complete description.\n\nRetained paragraph.|Board developer" {
		t.Fatal("modal operation lost dirty input", got)
	}
}
