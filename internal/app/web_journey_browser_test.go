package app

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// One browser navigates the persisted journey workspace, checks the shared
// card anatomy and keyboard/modal boundaries, and keeps dirty metadata through
// refresh and an independent comment without running an agent.
func TestWebConsolidatedBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned consolidated browser journey is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	j := completedWebJourney(t)

	server := httptest.NewServer(j.handler)
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	browser.call(t, "Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": "window.confirm=()=>true"})
	webBoardBrowserViewport(t, browser, 1280, 1000)
	browser.navigate(t, server.URL+"/cases")
	browser.wait(t, ".case-table")

	for _, view := range []string{"dumps", "quizzes", "specs", "plans", "board", "dash"} {
		webBoardBrowserViewport(t, browser, 1280, 1000)
		browser.clickNavigation(t, `.top-nav-rail a[href="/`+view+`"]`)
		browser.wait(t, "#workspace")
		webConsolidatedCapture(t, browser, view+"-desktop")
		webBoardBrowserViewport(t, browser, 390, 844)
		webConsolidatedCapture(t, browser, view+"-narrow")
	}

	webBoardBrowserViewport(t, browser, 1280, 1000)
	browser.clickNavigation(t, `.top-nav-rail a[href="/dumps"]`)
	webConsolidatedModal(t, browser, ".dump-card-link", "[data-dump-editor] textarea", "Pending dump text", "dump-editor")
	browser.clickNavigation(t, `.top-nav-rail a[href="/specs"]`)
	webConsolidatedModal(t, browser, ".document-card", "#spec-body", "Pending spec text", "spec-editor")
	browser.clickNavigation(t, `.top-nav-rail a[href="/board"]`)
	webConsolidatedModal(t, browser, `a[data-ticket-key="`+j.ticket+`"]`, `[name="owner"]`, "Pending ownership", "ticket-editor")
	webConsolidatedBoardComment(t, browser, j)
	j.reopen(t)
}

func webConsolidatedModal(t *testing.T, browser *webWorkflowBrowser, trigger, field, value, name string) {
	t.Helper()
	webBoardBrowserClick(t, browser, trigger)
	browser.wait(t, "#modal-root [role=dialog]")
	webBoardBrowserWait(t, browser, `document.querySelector('#modal-root [role=dialog]').contains(document.activeElement)`)
	webBoardBrowserSet(t, browser, field, value)
	browser.evaluate(t, `(async()=>{await htmx.ajax('GET',document.querySelector('#workspace').getAttribute('hx-get'),{target:'#workspace',swap:'innerHTML'});return 'refreshed';})()`)
	webBoardBrowserReady(t, browser)

	if browser.evaluate(t, `document.querySelector(`+strconv.Quote(field)+`).value`) != value {
		t.Fatal("refresh discarded dirty input", name)
	}

	webConsolidatedFocus(t, browser)
	webConsolidatedCapture(t, browser, name+"-desktop")
	webBoardBrowserViewport(t, browser, 390, 844)
	webConsolidatedCapture(t, browser, name+"-narrow")
	webQuizCatalogKey(t, browser, "Escape", 27)
	webBoardBrowserWait(t, browser, `!document.querySelector('#modal-root [role=dialog]')`)

	if browser.evaluate(t, `String(document.activeElement.matches(`+strconv.Quote(trigger)+`))`) != "true" {
		t.Fatal("closing the modal lost its trigger", name)
	}

	webBoardBrowserViewport(t, browser, 1280, 1000)
}

func webConsolidatedFocus(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	browser.evaluate(t, `(()=>{const d=document.querySelector('#modal-root [role=dialog]'),items=[...d.querySelectorAll('a[href],button,input:not([type=hidden]),textarea,select')].filter(e=>!e.disabled&&e.getClientRects().length);items[items.length-1].focus();return 'last';})()`)
	webQuizCatalogKey(t, browser, "Tab", 9)

	if browser.evaluate(t, `String(document.querySelector('#modal-root [role=dialog]').contains(document.activeElement))`) != "true" {
		t.Fatal("Tab escaped the modal")
	}

	if browser.evaluate(t, `String([...document.querySelectorAll('#modal-root input:not([type=hidden]),#modal-root textarea,#modal-root select')].every(e=>e.labels.length||e.getAttribute('aria-label')||e.getAttribute('aria-labelledby')))`) != "true" {
		t.Fatal("modal form control lacks an accessible label")
	}
}

func webConsolidatedCapture(t *testing.T, browser *webWorkflowBrowser, name string) {
	t.Helper()

	metrics := browser.evaluate(t, `(()=>{const r=e=>{const b=e.getBoundingClientRect();return {x:b.x,y:b.y,width:b.width,height:b.height}},cards=[...document.querySelectorAll('.dump-card,.quiz-card,.document-card')],dialog=document.querySelector('#modal-root [role=dialog]');return JSON.stringify({viewport:{width:innerWidth,height:innerHeight,dpr:devicePixelRatio,browserZoom:1,pageScale:visualViewport.scale},documentWidth:document.documentElement.scrollWidth,cards:cards.map(c=>({card:r(c),preview:r(c.querySelector('[class$=__preview]'))})),dialog:dialog?r(dialog):null,focus:document.activeElement.id});})()`)
	t.Log(name, metrics)

	if browser.evaluate(t, `String(document.documentElement.scrollWidth<=innerWidth&&devicePixelRatio===1&&visualViewport.scale===1)`) != "true" {
		t.Fatal("consolidated viewport overflow", metrics)
	}

	if browser.evaluate(t, `String([...document.querySelectorAll('.dump-card,.quiz-card,.document-card')].every(c=>{const p=c.querySelector('[class$=__preview]'),r=c.getBoundingClientRect(),s=getComputedStyle(p);return r.width===300&&r.height===218&&p.getBoundingClientRect().height===132&&s.overflowY==='hidden'&&s.backgroundColor==='rgb(247, 251, 247)'&&getComputedStyle(c).backgroundColor==='rgb(255, 255, 255)';}))`) != "true" {
		t.Fatal("shared fixed card anatomy changed", metrics)
	}

	if browser.evaluate(t, `(()=>{const card=document.querySelector('.dump-card,.quiz-card,.document-card');return String(!card||Math.abs(card.getBoundingClientRect().x-document.querySelector('h1').getBoundingClientRect().x)<0.1);})()`) != "true" {
		t.Fatal("catalog heading and cards have different left margins", metrics)
	}

	if browser.evaluate(t, `(()=>{const d=document.querySelector('#modal-root [role=dialog]');if(!d)return 'true';const r=d.getBoundingClientRect();return String(r.x>=0&&r.right<=innerWidth&&r.y>=0&&r.bottom<=innerHeight);})()`) != "true" {
		t.Fatal("modal escaped the viewport", metrics)
	}

	directory := os.Getenv("CHR_WEB_JOURNEY_SCREENSHOTS")
	if directory == "" {
		return
	}

	webBrowserCapture(t, browser, directory, name)

	if err := os.WriteFile(filepath.Join(directory, name+".json"), []byte(metrics+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
