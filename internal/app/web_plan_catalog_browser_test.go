package app

import (
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

// An owned Chromium fixture measures short/long/unbroken cards, opens the exact
// complete read-only document and verifies that browsing leaves storage intact.
func TestWebPlanCatalogBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned plan catalog browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider, model := webPlanFixture(t)
	refs := seedWebPlanCatalog(t, provider)
	before := webPlanFiles(t, provider)

	server := httptest.NewServer(webQuizHandler(t, provider))
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	browser.navigate(t, server.URL+"/plans")
	browser.wait(t, ".document-card")

	for _, width := range []int{1280, 768, 390, 320} {
		webBoardBrowserViewport(t, browser, width, 900)
		webBoardBrowserWait(t, browser, "innerWidth==="+strconv.Itoa(width))
		webPlanCatalogCapture(t, browser, "catalog-"+strconv.Itoa(width))
	}

	webBoardBrowserViewport(t, browser, 1280, 1000)
	webPlanCatalogBrowserDetail(t, browser, refs[1], "Long plan")
	webPlanCatalogCapture(t, browser, "detail-desktop")
	webBoardBrowserViewport(t, browser, 390, 844)
	webPlanCatalogCapture(t, browser, "detail-narrow")
	webBoardBrowserClick(t, browser, `.detail-actions a[href="/plans"]`)
	browser.wait(t, ".document-card")
	webPlanCatalogBrowserDetail(t, browser, refs[2], "UnbrokenTitle")
	webPlanCatalogCapture(t, browser, "detail-unbroken")

	if !reflect.DeepEqual(before, webPlanFiles(t, provider)) || len(model.requests) != 0 {
		t.Fatal("browser mutated plans or invoked generation")
	}
}

func webPlanCatalogBrowserDetail(t *testing.T, browser *webWorkflowBrowser, reference, title string) {
	t.Helper()
	webBoardBrowserClick(t, browser, `.document-card[href="/plans/detail?id=`+url.QueryEscape(reference)+`"]`)
	browser.wait(t, ".plan-document")

	if browser.evaluate(t, `String(document.querySelector('.plan-document').textContent.includes(`+strconv.Quote("END "+title)+`) && !document.querySelector('.plan-document textarea, .plan-document form') && !document.querySelector('script:not([src])'))`) != "true" {
		t.Fatal("incomplete or editable plan detail")
	}
}

func webPlanCatalogCapture(t *testing.T, browser *webWorkflowBrowser, name string) {
	t.Helper()

	metrics := browser.evaluate(t, `(()=>{const r=e=>{const b=e.getBoundingClientRect();return {width:b.width,height:b.height,right:b.right}},cards=[...document.querySelectorAll('.document-card')],detail=document.querySelector('.plan-document');return JSON.stringify({viewport:{width:innerWidth,height:innerHeight,dpr:devicePixelRatio,browserZoom:1,pageScale:visualViewport.scale},documentWidth:document.documentElement.scrollWidth,cards:cards.map(c=>({card:r(c),preview:r(c.querySelector('.document-card__preview'))})),detail:detail?r(detail):null,bodyCharacters:detail?detail.textContent.length:0,editControls:detail?detail.querySelectorAll('textarea,input,form').length:0});})()`)
	t.Log(name, metrics)

	if browser.evaluate(t, `String([...document.querySelectorAll('.document-card')].every(c=>{const p=c.querySelector('.document-card__preview');return c.getBoundingClientRect().width===300&&c.getBoundingClientRect().height===218&&p.getBoundingClientRect().height===132&&getComputedStyle(p).overflowY==='hidden'&&getComputedStyle(p).backgroundColor==='rgb(247, 251, 247)'&&getComputedStyle(c).backgroundColor==='rgb(255, 255, 255)';}) && document.documentElement.scrollWidth<=innerWidth)`) != "true" {
		t.Fatal("plan geometry changed", metrics)
	}

	directory := os.Getenv("CHR_WEB_PLAN_CATALOG_SCREENSHOTS")
	if directory == "" {
		return
	}

	webBrowserCapture(t, browser, directory, name)

	if err := os.WriteFile(filepath.Join(directory, name+".json"), []byte(metrics+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
