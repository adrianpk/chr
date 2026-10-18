package app

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// One owned browser verifies bounded geometry, menu keyboard/pointer entry,
// retained session selection, all-answered removal and a newly added case.
func TestWebQuizCatalogBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned quiz catalog browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider := webQuizCatalogFixture(t)

	server := httptest.NewServer(webQuizHandler(t, provider))
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	webBoardBrowserViewport(t, browser, 1440, 1000)
	browser.navigate(t, server.URL+"/quizzes")
	browser.wait(t, ".quiz-card")
	webQuizCatalogCapture(t, browser, "desktop-closed", 4)
	webQuizCatalogMenu(t, browser)
	webQuizCatalogCapture(t, browser, "desktop-open", 4)
	webQuizCatalogContinue(t, browser, provider)
	browser.navigate(t, server.URL+"/quizzes")
	browser.wait(t, ".quiz-card")

	for _, width := range []int{768, 390, 320} {
		webBoardBrowserViewport(t, browser, width, 844)
		webBoardBrowserWait(t, browser, "innerWidth==="+strconv.Itoa(width))

		columns := 1
		if width == 768 {
			columns = 2
		}

		webBoardBrowserClick(t, browser, ".quiz-new summary")
		webBoardBrowserWait(t, browser, `document.querySelector(".quiz-new").open`)
		webQuizCatalogCapture(t, browser, "width-"+strconv.Itoa(width), columns)
		webBoardBrowserClick(t, browser, ".quiz-new summary")
		webBoardBrowserWait(t, browser, `!document.querySelector(".quiz-new").open`)
	}

	webBoardBrowserViewport(t, browser, 1440, 1000)
	webQuizCatalogAllAnswered(t, browser, provider, server.URL)
}

func webQuizCatalogFixture(t *testing.T) webViewProvider {
	t.Helper()

	provider, _ := newWebQuizFixture(t)
	operations := provider.service(t.Context())
	seedQuizCatalogScope(t, operations, operations.quizScope(core.QuizScope{CaseSlug: "named-work"}), true, false, "")

	for _, item := range []struct{ slug, title, answer string }{
		{"long-title", "Detailed research and retained decisions across several workflow boundaries", strings.Repeat("Opening context with useful decisions. ", 200)},
		{"unbroken-title", strings.Repeat("UnbrokenTitle", 12), strings.Repeat("UnbrokenContext", 200)},
	} {
		if _, err := core.NewCaseMetadataService(operations.projectContext.ProjectDir).Save(core.CaseMetadata{SourcePath: core.CaseMetadataPath(operations.projectContext.ProjectDir, item.slug), Name: item.title}); err != nil {
			t.Fatal(err)
		}

		scope := operations.quizScope(core.QuizScope{CaseSlug: item.slug})
		seedQuizCatalogScope(t, operations, scope, false, false, "")

		if _, err := operations.AnswerQuiz(QuizAnswerRequest{Scope: scope, Answer: item.answer}); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := core.NewCaseMetadataService(operations.projectContext.ProjectDir).Save(core.CaseMetadata{SourcePath: core.CaseMetadataPath(operations.projectContext.ProjectDir, "pending-work"), Name: "Pending named work"}); err != nil {
		t.Fatal(err)
	}

	seedQuizCatalogScope(t, operations, operations.quizScope(core.QuizScope{CaseSlug: "pending-work"}), false, false, "")

	return provider
}

func webQuizCatalogMenu(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	browser.call(t, "Page.bringToFront", map[string]any{})
	webBoardBrowserReady(t, browser)
	browser.evaluate(t, `document.querySelector('.quiz-new summary').focus(); 'focused'`)
	webQuizCatalogKey(t, browser, "Enter", 13)
	webBoardBrowserWait(t, browser, `document.querySelector('.quiz-new').open`)
	webQuizCatalogKey(t, browser, "Tab", 9)

	if browser.evaluate(t, `String(document.activeElement.matches('.quiz-new__choices a'))`) != "true" {
		t.Fatal("pending choices are not keyboard reachable")
	}

	webQuizCatalogKey(t, browser, "Escape", 27)

	if browser.evaluate(t, `String(!document.querySelector('.quiz-new').open && document.activeElement.matches('.quiz-new summary'))`) != "true" {
		t.Fatal("Escape did not close and focus New Quiz")
	}

	webBoardBrowserClick(t, browser, ".quiz-new summary")
	webBoardBrowserWait(t, browser, `document.querySelector(".quiz-new").open`)
	// A read refresh cannot discard the menu while the user selects a scope.
	if browser.evaluate(t, `(()=>{const e=new CustomEvent('htmx:beforeRequest',{bubbles:true,cancelable:true,detail:{elt:document.querySelector('#workspace')}});document.dispatchEvent(e);return String(e.defaultPrevented);})()`) != "true" {
		t.Fatal("poll replaced an open menu")
	}
}

func webQuizCatalogContinue(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider) {
	t.Helper()

	operations := provider.service(t.Context())
	scope := operations.quizScope(core.QuizScope{CaseSlug: "pending-work"})
	before := quizCatalogObservation(t, operations, scope)
	webBoardBrowserClick(t, browser, `.quiz-new__choices a[href*="pending-work-quiz"]`)
	browser.wait(t, "#quiz-answer")

	if got := browser.evaluate(t, `document.querySelector('#quiz-surface').dataset.quizSession`); got != before.State.SessionID {
		t.Fatal("pending choice replaced session", got, before.State.SessionID)
	}

	webBoardBrowserSet(t, browser, "#quiz-answer", "The named pending scope retains this confirmed answer.")
	webBoardBrowserClick(t, browser, `[value="quiz.answer"]`)
	webBoardBrowserWait(t, browser, `document.querySelectorAll('.quiz-history-card').length===1`)
	webBoardBrowserClick(t, browser, `[value="quiz.cancel"]`)
	browser.wait(t, ".quiz-card")

	if got := browser.evaluate(t, `String([...document.querySelectorAll('.quiz-card h2')].some(e=>e.textContent==='Pending named work'))`); got != "true" {
		t.Fatal("answered cancelled scope disappeared", got)
	}
}

func webQuizCatalogAllAnswered(t *testing.T, browser *webWorkflowBrowser, provider webViewProvider, baseURL string) {
	t.Helper()

	operations := provider.service(t.Context())
	seedQuizCatalogScope(t, operations, operations.quizScope(core.QuizScope{}), true, false, "done")
	browser.navigate(t, baseURL+"/quizzes")
	browser.wait(t, ".quiz-card")

	if got := browser.evaluate(t, `String(!document.querySelector('.quiz-new'))`); got != "true" {
		t.Fatal("New Quiz remained after all scopes were answered")
	}

	webQuizCatalogCapture(t, browser, "all-answered", 4)

	if _, err := core.NewCaseMetadataService(operations.projectContext.ProjectDir).Save(core.CaseMetadata{Slug: "new-case", Name: "Newly added case"}); err != nil {
		t.Fatal(err)
	}

	browser.navigate(t, baseURL+"/quizzes")
	browser.wait(t, ".quiz-new summary")

	if got := browser.evaluate(t, `JSON.stringify([...document.querySelectorAll('.quiz-new__choices a')].map(e=>e.textContent))`); got != `["Newly added case"]` {
		t.Fatal("new case did not become the only pending choice", got)
	}
}

func webQuizCatalogKey(t *testing.T, browser *webWorkflowBrowser, key string, keyCode int) {
	t.Helper()

	for _, kind := range []string{"keyDown", "keyUp"} {
		params := map[string]any{"type": kind, "key": key, "code": key, "windowsVirtualKeyCode": keyCode}
		if key == "Enter" && kind == "keyDown" {
			params["text"] = "\r"
		}

		browser.call(t, "Input.dispatchKeyEvent", params)
	}
}

func webQuizCatalogCapture(t *testing.T, browser *webWorkflowBrowser, name string, columns int) {
	t.Helper()

	metrics := browser.evaluate(t, `(()=>{const main=document.querySelector('.quiz-catalog-page'),grid=document.querySelector('.quiz-catalog__grid'),card=document.querySelector('.quiz-card'),preview=card.querySelector('.quiz-card__preview'),menu=document.querySelector('.quiz-new[open] .quiz-new__choices'),r=e=>{const b=e.getBoundingClientRect();return {x:b.x,y:b.y,width:b.width,height:b.height,right:b.right}};return JSON.stringify({innerWidth,innerHeight,dpr:devicePixelRatio,browserZoom:1,pageScale:visualViewport.scale,main:r(main),grid:getComputedStyle(grid).gridTemplateColumns,card:r(card),preview:r(preview),menu:menu?r(menu):null,cards:document.querySelectorAll('.quiz-card').length,pageWidth:document.documentElement.scrollWidth});})()`)
	t.Log(name, metrics)

	if got := browser.evaluate(t, `String([...document.querySelectorAll('.quiz-card')].every(c=>{const r=c.getBoundingClientRect(),p=c.querySelector('.quiz-card__preview'),s=getComputedStyle(p);return r.width===300&&r.height===218&&p.getBoundingClientRect().height===132&&r.right<=document.documentElement.clientWidth&&s.overflowY==='hidden'&&s.backgroundColor==='rgb(247, 251, 247)'&&getComputedStyle(c).backgroundColor==='rgb(255, 255, 255)';}))`); got != "true" {
		t.Fatal("fixed card geometry or preview failed", metrics)
	}

	if got := browser.evaluate(t, `String(document.documentElement.scrollWidth<=innerWidth && devicePixelRatio===1 && visualViewport.scale===1)`); got != "true" {
		t.Fatal("viewport overflow or unexpected zoom", metrics)
	}

	if got := browser.evaluate(t, `String(getComputedStyle(document.querySelector('.quiz-catalog__grid')).gridTemplateColumns.split(' ').length)`); got != strconv.Itoa(columns) {
		t.Fatal("unexpected track count", metrics)
	}

	if got := browser.evaluate(t, `(()=>{const m=document.querySelector('.quiz-new[open] .quiz-new__choices');if(!m)return 'true';const r=m.getBoundingClientRect();return String(r.width===232&&r.x>=0&&r.right<=innerWidth);})()`); got != "true" {
		t.Fatal("pending menu escaped viewport", metrics)
	}

	directory := os.Getenv("CHR_WEB_QUIZ_CATALOG_SCREENSHOTS")
	if directory == "" {
		return
	}

	webBrowserCapture(t, browser, directory, name)

	if err := os.WriteFile(filepath.Join(directory, name+".json"), []byte(metrics+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
