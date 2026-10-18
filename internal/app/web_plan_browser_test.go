package app

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An owned browser verifies contextual actions, retained case input through
// failed/successful generation, selected quiz guidance and narrow geometry.
func TestWebPlanBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned plan browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider, _ := webPlanFixture(t)

	server := httptest.NewServer(webQuizHandler(t, provider))
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	webBoardBrowserViewport(t, browser, 1280, 1000)
	browser.navigate(t, server.URL+"/cases/detail?id=named-work")
	browser.wait(t, "[data-case-plan]")
	webPlanBrowserCapture(t, browser, "generate-desktop")
	webBoardBrowserSet(t, browser, ".case-edit-form [name='context']", "Pending case input stays here.")
	webBoardBrowserClick(t, browser, "[data-case-plan] button")
	webBoardBrowserWait(t, browser, `document.querySelector('[data-plan-write-status="validation_failed"]')`)
	webPlanBrowserCapture(t, browser, "missing-context")
	addWebPlanDiscovery(t, provider)
	webBoardBrowserClick(t, browser, "[data-case-plan] button")
	webBoardBrowserWait(t, browser, `document.querySelector('[data-case-plan] [name="action"]').value === 'plan.regenerate'`)

	if browser.evaluate(t, `document.querySelector('.case-edit-form [name="context"]').value`) != "Pending case input stays here." {
		t.Fatal("planning replaced pending case input")
	}

	webPlanBrowserCapture(t, browser, "regenerate-desktop")
	webBoardBrowserViewport(t, browser, 390, 844)
	webBoardBrowserWait(t, browser, "innerWidth===390")
	webPlanBrowserCapture(t, browser, "regenerate-narrow")

	if browser.evaluate(t, `String(document.documentElement.scrollWidth <= innerWidth)`) != "true" {
		t.Fatal("planning overflowed narrow viewport")
	}

	quizURL := webPlanDetail(t, provider).Plan.QuizURL
	if !strings.Contains(quizURL, "named-work-quiz") {
		t.Fatal("guidance lost selected case", quizURL)
	}
}

func webPlanBrowserCapture(t *testing.T, browser *webWorkflowBrowser, name string) {
	t.Helper()

	directory := os.Getenv("CHR_WEB_PLAN_SCREENSHOTS")
	if directory == "" {
		return
	}

	browser.evaluate(t, `document.querySelector('#case-plan-panel').scrollIntoView({block:'end'}); 'ready'`)
	webBrowserCapture(t, browser, directory, name)

	metrics := browser.evaluate(t, `JSON.stringify({viewport:{width:innerWidth,height:innerHeight,dpr:devicePixelRatio,pageScale:visualViewport.scale},documentWidth:document.documentElement.scrollWidth,panel:(()=>{const r=document.querySelector('#case-plan-panel').getBoundingClientRect();return {x:r.x,y:r.y,width:r.width,height:r.height}})(),action:document.querySelector('[data-case-plan] [name="action"]').value,bodyEditors:document.querySelectorAll('.case-edit-form textarea').length})`)
	if err := os.WriteFile(filepath.Join(directory, name+".json"), []byte(metrics+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
