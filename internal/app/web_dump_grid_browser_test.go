package app

import (
	"encoding/json"
	"math"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Real browser zoom must expose the extra CSS viewport without changing the
// physical window, and fixed-size short/long cards must pack into that width.
func TestWebDumpGridBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned Dumps layout browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider := webDumpGridFixture(t)
	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	for _, zoom := range []float64{1, 0.9} {
		browser := webDumpGridBrowser(t, path, zoom)
		browser.navigate(t, server.URL+"/dumps")
		browser.wait(t, ".dump-card")

		metrics := webDumpGridGeometry(t, browser, zoom, 1280)
		webDumpGridCapture(t, browser, "desktop-"+strconv.Itoa(int(math.Round(zoom*100))), metrics)

		if zoom == 1 {
			for _, width := range []int{768, 390, 320} {
				webBoardBrowserViewport(t, browser, width, 844)
				webBoardBrowserWait(t, browser, "innerWidth==="+strconv.Itoa(width))

				metrics := webDumpGridGeometry(t, browser, 1, width)
				webDumpGridCapture(t, browser, "width-"+strconv.Itoa(width), metrics)
			}

			// Card count changes cannot stretch the surviving card.
			browser.call(t, "Emulation.clearDeviceMetricsOverride", map[string]any{})
			webBoardBrowserWait(t, browser, "innerWidth===1280")
			browser.evaluate(t, `document.querySelectorAll('.dump-card-link').forEach((e,i)=>{if(i)e.remove();}); 'one card'`)
			webDumpGridGeometry(t, browser, 1, 1280)
		}
	}
}

func webDumpGridFixture(t *testing.T) webDumpCaptureProvider {
	t.Helper()

	provider := newWebDumpCaptureProvider(t)
	for i, document := range []struct{ title, body string }{
		{"Project overview", "A short opening paragraph."},
		{"Detailed research notes and retained decisions across several topics", strings.Repeat("Long context to clip. ", 40)},
		{strings.Repeat("UnbrokenTitle", 12), strings.Repeat("UnbrokenContext", 40)},
		{"Next steps", "A second short paragraph."},
		{"Follow-up notes", "Keep related context together."},
	} {
		if _, err := provider.operations.CreateDump(DumpCreateRequest{OperationID: "grid-" + strconv.Itoa(i), Title: document.title, Content: document.body, Scope: DumpScope{CaseSlug: "named-work"}}); err != nil {
			t.Fatal(err)
		}
	}

	serviceDumpArtifact(t, provider.projectContext.ProjectDir, "default", "Empty capture", "---\ntitle: Empty capture\ncase: named-work\n---\n")

	return provider
}

func webDumpGridGeometry(t *testing.T, browser *webWorkflowBrowser, zoom float64, width int) string {
	t.Helper()

	metrics := browser.evaluate(t, `(()=>{const c=document.querySelector('.dump-card'),p=c.querySelector('.dump-card__preview'),r=c.getBoundingClientRect(),v=p.getBoundingClientRect();return JSON.stringify({innerWidth,innerHeight,outerWidth,outerHeight,dpr:devicePixelRatio,grid:getComputedStyle(document.querySelector('.dump-list')).gridTemplateColumns,width:document.querySelector('.dump-list').getBoundingClientRect().width,cardCount:document.querySelectorAll('.dump-card').length,cardWidth:r.width,cardHeight:r.height,previewWidth:v.width,previewHeight:v.height,cardBackground:getComputedStyle(c).backgroundColor,previewBackground:getComputedStyle(p).backgroundColor,pageOverflow:document.documentElement.scrollWidth>document.documentElement.clientWidth});})()`)
	t.Log("zoom", zoom, metrics)

	if got := browser.evaluate(t, `String([...document.querySelectorAll('.dump-card')].every(e=>{const r=e.getBoundingClientRect(),p=e.querySelector('.dump-card__preview').getBoundingClientRect();return Math.abs(r.width-300)<0.1&&Math.abs(r.height-218)<0.1&&Math.abs(p.height-132)<0.1&&r.right<=document.documentElement.clientWidth+0.1;}))`); got != "true" {
		t.Fatal("short, long or unbroken content changed fixed geometry or overflowed", metrics)
	}

	if got := browser.evaluate(t, `String(document.documentElement.scrollWidth<=document.documentElement.clientWidth)`); got != "true" {
		t.Fatal("Dumps layout overflows the viewport", metrics)
	}

	// Fractional zoom rounds borders; allow less than one physical pixel from
	// nominal preview width while requiring equal widths across all content.
	if got := browser.evaluate(t, `(()=>{const width=document.querySelector('.dump-card__preview').getBoundingClientRect().width;return String([...document.querySelectorAll('.dump-card')].every(e=>{const p=e.querySelector('.dump-card__preview'),s=getComputedStyle(p);return Math.abs(e.querySelector('h2').getBoundingClientRect().height-42)<0.1&&Math.abs(p.getBoundingClientRect().width-width)<0.05&&Math.abs(width-278)<1/devicePixelRatio&&s.overflowY==='hidden'&&getComputedStyle(p.firstElementChild).webkitLineClamp==='7'&&getComputedStyle(p.firstElementChild).overflowY==='hidden'&&s.backgroundColor==='rgb(247, 251, 247)'&&getComputedStyle(e).backgroundColor==='rgb(255, 255, 255)';}));})()`); got != "true" {
		t.Fatal("title or inset preview escaped the fixed card anatomy", metrics)
	}

	var measured struct{ DPR float64 }
	if err := json.Unmarshal([]byte(metrics), &measured); err != nil || math.Abs(measured.DPR-zoom) > 0.001 {
		t.Fatal("browser did not apply real zoom", measured, err)
	}

	columns := map[int]string{1280: "3", 768: "2", 390: "1", 320: "1"}[width]
	if zoom == 0.9 {
		columns = "4"
	}

	if got := browser.evaluate(t, `String(getComputedStyle(document.querySelector('.dump-list')).gridTemplateColumns.split(' ').length)`); got != columns {
		t.Fatal("grid did not pack fixed tracks into available width", got, columns, metrics)
	}

	return metrics
}

func webDumpGridCapture(t *testing.T, browser *webWorkflowBrowser, name, metrics string) {
	t.Helper()

	directory := os.Getenv("CHR_WEB_DUMP_GRID_SCREENSHOTS")
	if directory == "" {
		return
	}

	webBrowserCapture(t, browser, directory, name)

	if err := os.WriteFile(filepath.Join(directory, name+".json"), []byte(metrics+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func webDumpGridBrowser(t *testing.T, path string, zoom float64) *webWorkflowBrowser {
	t.Helper()

	profile := t.TempDir()
	// Chromium stores the default partition's real page zoom under key x.
	preferences := map[string]any{"partition": map[string]any{"default_zoom_level": map[string]float64{"x": math.Log(zoom) / math.Log(1.2)}}}

	data, err := json.Marshal(preferences)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(profile, "Default"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(profile, "Default", "Preferences"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	browser := newWebWorkflowBrowserProfile(t, path, profile)
	result := browser.call(t, "Browser.getWindowForTarget", map[string]any{})

	var window struct{ WindowID int }
	if err := json.Unmarshal(result, &window); err != nil {
		t.Fatal(err)
	}

	browser.call(t, "Browser.setWindowBounds", map[string]any{"windowId": window.WindowID, "bounds": map[string]any{"width": 1280, "height": 900}})

	return browser
}
