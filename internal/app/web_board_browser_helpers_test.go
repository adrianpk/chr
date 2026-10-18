package app

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// DOM insertion precedes HTMX listener installation and attribute settling.
// Wait for the observable framework boundary before a synthetic user action.
func webBoardBrowserReady(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	webBoardBrowserWait(t, browser, `!document.querySelector('.htmx-settling, .htmx-swapping, .htmx-request')`)
}

func webBoardBrowserClick(t *testing.T, browser *webWorkflowBrowser, selector string) {
	t.Helper()
	browser.wait(t, selector)
	webBoardBrowserReady(t, browser)
	browser.click(t, selector)
}

func webBoardBrowserSet(t *testing.T, browser *webWorkflowBrowser, selector, value string) {
	t.Helper()
	webBoardBrowserReady(t, browser)
	browser.evaluate(t, `(()=>{const e=document.querySelector(`+strconv.Quote(selector)+`);e.value=`+strconv.Quote(value)+`;e.dispatchEvent(new Event('input',{bubbles:true}));return 'set';})()`)
}

func webBoardBrowserWait(t *testing.T, browser *webWorkflowBrowser, expression string) {
	t.Helper()
	browser.evaluate(t, `(async()=>{for(let n=0;n<200;n++){if(`+expression+`)return 'ready';await new Promise(r=>setTimeout(r,50));}throw Error('condition timeout: '+`+strconv.Quote(expression)+`);})()`)
}

func webBoardBrowserDrag(t *testing.T, browser *webWorkflowBrowser, status string) {
	t.Helper()
	webBoardBrowserReady(t, browser)
	browser.evaluate(t, `(()=>{const card=document.querySelector('a[data-ticket-key="CHR-1"]'),column=document.querySelector('[data-board-column="'+`+strconv.Quote(status)+`+'"]'),data=new DataTransfer();card.dispatchEvent(new DragEvent('dragstart',{bubbles:true,dataTransfer:data}));column.dispatchEvent(new DragEvent('drop',{bubbles:true,cancelable:true,dataTransfer:data}));return 'dropped';})()`)
}

func webBoardBrowserViewport(t *testing.T, browser *webWorkflowBrowser, width, height int) {
	t.Helper()
	browser.call(t, "Emulation.setDeviceMetricsOverride", map[string]any{"width": width, "height": height, "deviceScaleFactor": 1, "mobile": false})
	browser.call(t, "Emulation.setPageScaleFactor", map[string]any{"pageScaleFactor": 1})
}

func webBoardBrowserCapture(t *testing.T, browser *webWorkflowBrowser, name string) {
	t.Helper()

	webBrowserCapture(t, browser, os.Getenv("CHR_WEB_BOARD_SCREENSHOTS"), name)
}

func webBrowserCapture(t *testing.T, browser *webWorkflowBrowser, directory, name string) {
	t.Helper()

	if directory == "" {
		return
	}
	// Capture after paint; the caller records viewport, DPR and browser zoom.
	browser.evaluate(t, `new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(()=>r('painted'))))`)

	result := browser.call(t, "Page.captureScreenshot", map[string]any{"format": "png", "captureBeyondViewport": false})

	var capture struct{ Data string }
	if err := json.Unmarshal(result, &capture); err != nil {
		t.Fatal(err)
	}

	data, err := base64.StdEncoding.DecodeString(capture.Data)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(directory, name+".png"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func webBoardBrowserFocus(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	browser.evaluate(t, `document.querySelector("button[form='ticket-editor-form']").focus(); 'focused'`)
	browser.call(t, "Input.dispatchKeyEvent", map[string]any{"type": "keyDown", "key": "Tab", "code": "Tab"})
	browser.call(t, "Input.dispatchKeyEvent", map[string]any{"type": "keyUp", "key": "Tab", "code": "Tab"})

	if got := browser.evaluate(t, `String(document.activeElement.matches('[data-ticket-maximize]'))`); got != "true" {
		t.Fatal("Tab escaped dialog", got)
	}

	browser.call(t, "Input.dispatchKeyEvent", map[string]any{"type": "keyDown", "key": "Tab", "code": "Tab", "modifiers": 8})
	browser.call(t, "Input.dispatchKeyEvent", map[string]any{"type": "keyUp", "key": "Tab", "code": "Tab", "modifiers": 8})

	if got := browser.evaluate(t, `String(document.activeElement.matches("button[form='ticket-editor-form']"))`); got != "true" {
		t.Fatal("Shift Tab escaped dialog", got)
	}
}
