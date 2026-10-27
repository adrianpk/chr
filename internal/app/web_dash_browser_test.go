package app

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/webui"
)

type dashBrowserSource struct {
	webViewProvider
	mu          sync.Mutex
	observation WorkObservation
	unavailable bool
	reads       int
	writes      []webui.WriteActionRequest
	pending     chan struct{}
	release     sync.Once
}

func (p *dashBrowserSource) DashView(ctx context.Context) (webui.DashView, error) {
	return p.DashPageView(ctx, webui.DashRequest{})
}

func (p *dashBrowserSource) DashPageView(_ context.Context, request webui.DashRequest) (webui.DashView, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.reads++
	if p.unavailable {
		return webui.DashView{}, errors.New("temporary fixture outage")
	}

	observation := p.observation
	observation.ObservedAt = time.Now().UTC()
	finishWorkObservation(&observation, request.QueueOffset)

	return webDashObservation(observation, request.QueueOffset), nil
}

func (p *dashBrowserSource) WriteAction(ctx context.Context, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	p.mu.Lock()
	p.writes = append(p.writes, request)
	p.mu.Unlock()

	if request.Action == "work.preview" {
		select {
		case <-ctx.Done():
			return webui.WriteActionResponse{}, ctx.Err()
		case <-p.pending:
		}
	}

	return webui.WriteActionResponse{Status: webui.WriteStatusOK, Summary: "Control result retained", RedirectURL: "/dash"}, nil
}

func (p *dashBrowserSource) set(observation WorkObservation, unavailable bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.observation, p.unavailable = observation, unavailable
}

func (p *dashBrowserSource) writeCount() int { p.mu.Lock(); defer p.mu.Unlock(); return len(p.writes) }

// Scripted browser transitions exercise the real adapter, renderer and HTMX
// boundaries. Real native/legacy persistence is covered by the HTTP/race tests.
func TestWebDashBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned Dash browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	provider, _ := webPlanFixture(t)
	source := &dashBrowserSource{webViewProvider: provider, observation: dashBrowserObservation("queued"), pending: make(chan struct{})}
	config := provider.config
	config.Provider = source
	config.Web.RefreshIntervalSeconds = 1

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	t.Cleanup(func() { source.release.Do(func() { close(source.pending) }) })

	browser := newWebWorkflowBrowser(t, path)
	browser.call(t, "Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": "window.confirm=()=>true"})
	webBoardBrowserViewport(t, browser, 1280, 1000)
	browser.navigate(t, server.URL+"/dash")
	browser.wait(t, "[data-dash-observation]")
	webDashBrowserCapture(t, browser, "queued-desktop")
	webBoardBrowserViewport(t, browser, 390, 844)
	webDashBrowserCapture(t, browser, "queued-narrow")
	webDashBrowserActive(t, browser, source)
	webDashBrowserResults(t, browser, source, server.URL)
	webDashBrowserPolling(t, browser, source)
}

func dashBrowserObservation(state string) WorkObservation {
	at := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)

	observed := WorkObservation{Authority: workspaceconfig.WorkStateAuthorityTickets, Readiness: executioncore.WorkReadinessReport{Status: executioncore.WorkReadinessStatusReady}, Excluded: map[string]int{}}
	if state == "queued" {
		observed.Eligible = 2
		observed.Queue = []WorkQueueItem{{Key: "PRJ-1", Title: "Validate API", Scope: "Billing", Readiness: "ready"}, {Key: "PRJ-2", Title: "Apply metadata", Scope: "Accounts", Readiness: "ready"}}

		return observed
	}

	first := WorkContext{ID: "first:attempt", Kind: "native", InvocationID: "first", AttemptID: "attempt", TicketKey: "PRJ-1", Title: "Validate API", Scope: "Billing", Status: "running", Phase: "validating", UpdatedAt: at, Remaining: 1, Message: "Running go"}

	second := WorkContext{ID: "second:attempt", Kind: "native", InvocationID: "second", AttemptID: "attempt", TicketKey: "PRJ-2", Title: "Apply metadata", Scope: "Accounts", Status: "running", Phase: "working", UpdatedAt: at, Remaining: 1, Message: "Running worker"}
	if state == "running" {
		observed.Current = []WorkContext{first, second}
		for index := 3; index < 23; index++ {
			observed.Queue = append(observed.Queue, WorkQueueItem{Key: fmt.Sprintf("PRJ-%d", index), Title: "Dependent work", Readiness: "blocked", Reason: "Awaiting selected prerequisite"})
		}

		event := WorkEvent{ID: "second-event", ContextID: second.ID, TicketKey: second.TicketKey, Phase: "worker_started", Message: "Editing source", Time: at}
		observed.Events = []WorkEvent{event, event}

		return observed
	}

	first.Status, first.Phase, first.Validated, first.Confirmed, first.Remaining = "completed", "integrated", 1, 1, 0
	first.Message = "Confirmed delivery retained"
	second.Kind, second.Status, second.Phase, second.Message = "execution", "failed", "validation failed", "Validation reported an error"
	observed.Results = []WorkContext{first, second}

	return observed
}

func webDashBrowserRefresh(t *testing.T, browser *webWorkflowBrowser) {
	t.Helper()
	browser.evaluate(t, `(async()=>{try{await htmx.ajax('GET',document.querySelector('#workspace').getAttribute('hx-get'),{target:'#workspace',swap:'innerHTML'});}catch(e){}return 'refreshed';})()`)
	webBoardBrowserWait(t, browser, `!document.querySelector('#workspace.htmx-settling, #workspace.htmx-swapping, #workspace .htmx-settling, #workspace .htmx-swapping')`)
}

func webDashBrowserActive(t *testing.T, browser *webWorkflowBrowser, source *dashBrowserSource) {
	t.Helper()
	source.set(dashBrowserObservation("running"), false)
	webDashBrowserRefresh(t, browser)
	webBoardBrowserWait(t, browser, `document.querySelectorAll('.dash-current .dash-context').length===2`)
	webBoardBrowserViewport(t, browser, 1280, 1000)
	webDashBrowserCapture(t, browser, "running-desktop")
	webBoardBrowserViewport(t, browser, 390, 844)
	webDashBrowserCapture(t, browser, "running-narrow")

	if browser.evaluate(t, `String(document.querySelectorAll('.dash-event').length===1 && document.querySelector('[data-context-id="first:attempt"] .dash-phase').textContent==='validating' && document.querySelector('[data-context-id="second:attempt"] .dash-phase').textContent==='working')`) != "true" {
		t.Fatal("contexts or event deduplication changed")
	}

	browser.evaluate(t, `document.querySelector('.dash-current').retainedProbe=1; 'set'`)
	webDashBrowserRefresh(t, browser)

	if browser.evaluate(t, `String(document.querySelector('.dash-current').retainedProbe===1)`) != "true" {
		t.Fatal("unchanged observation replaced DOM")
	}

	webDashBrowserPosition(t, browser, source)
	webDashBrowserControls(t, browser, source)
}

func webDashBrowserPosition(t *testing.T, browser *webWorkflowBrowser, source *dashBrowserSource) {
	t.Helper()
	webBoardBrowserViewport(t, browser, 1280, 600)
	browser.evaluate(t, `(()=>{window.scrollTo(0,100);document.querySelector('.dash-queue__items').scrollTop=150;const link=document.querySelector('[data-context-id="first:attempt"] h3 a');link.focus({preventScroll:true});const range=document.createRange();range.selectNodeContents(link);getSelection().removeAllRanges();getSelection().addRange(range);return 'selected';})()`)

	changed := dashBrowserObservation("running")
	changed.Current[0].Phase = "validating output"
	source.set(changed, false)
	webDashBrowserRefresh(t, browser)

	if browser.evaluate(t, `String(getSelection().toString()==='Validate API' && document.querySelector('.dash-queue__items').scrollTop===150 && Math.abs(scrollY-100)<2 && !!document.activeElement.closest('[data-context-id="first:attempt"]'))`) != "true" {
		t.Fatal("changed refresh lost position or selection", browser.evaluate(t, `JSON.stringify({selection:getSelection().toString(),queue:document.querySelector('.dash-queue__items').scrollTop,y:scrollY,active:document.activeElement.outerHTML,phase:document.querySelector('[data-context-id="first:attempt"] .dash-phase').textContent})`))
	}
}

func webDashBrowserControls(t *testing.T, browser *webWorkflowBrowser, source *dashBrowserSource) {
	t.Helper()
	webBoardBrowserClick(t, browser, `.dash-controls form:has([value="work.preview"]) button`)
	webBoardBrowserWait(t, browser, `document.querySelector('.dash-controls form:has([value="work.preview"]) button').disabled`)

	changed := dashBrowserObservation("running")
	changed.Current[0].Phase = "quiet validation"
	source.set(changed, false)
	webDashBrowserRefresh(t, browser)

	if browser.evaluate(t, `String(!document.querySelector('.app-main').inert && document.querySelector('.dash-current').textContent.includes('quiet validation') && document.querySelector('.dash-controls form:has([value="work.preview"]) button').disabled)`) != "true" {
		t.Fatal("pending action hid monitoring or lost its form")
	}

	source.release.Do(func() { close(source.pending) })
	webBoardBrowserWait(t, browser, `document.querySelector('#dash-action-result').textContent.includes('Control result retained')`)
	webBoardBrowserClick(t, browser, `.dash-controls form:has([value="work.daemon.stop"]) button`)
	webBoardBrowserWait(t, browser, `!document.querySelector('.dash-controls .htmx-request')`)

	if source.writeCount() != 2 {
		t.Fatal("refresh invoked a control")
	}

	source.mu.Lock()
	defer source.mu.Unlock()

	if source.writes[0].Action != "work.preview" || source.writes[1].Action != "work.daemon.stop" || !source.writes[0].Confirm || !source.writes[1].Confirm {
		t.Fatal(source.writes)
	}
}

func webDashBrowserResults(t *testing.T, browser *webWorkflowBrowser, source *dashBrowserSource, base string) {
	t.Helper()
	source.set(dashBrowserObservation("completed"), false)
	webDashBrowserRefresh(t, browser)
	webBoardBrowserViewport(t, browser, 1280, 1000)
	browser.evaluate(t, `getSelection().removeAllRanges();window.scrollTo(0,0);'top'`)
	webDashBrowserCapture(t, browser, "completed-desktop")
	webBoardBrowserViewport(t, browser, 390, 844)
	webDashBrowserCapture(t, browser, "completed-narrow")

	if browser.evaluate(t, `String(!document.querySelector('.dash-current .dash-context') && document.querySelectorAll('.dash-recent .dash-context').length===2 && document.querySelector('.dash-recent').textContent.includes('confirmed tickets'))`) != "true" {
		t.Fatal("completed result disappeared without events")
	}

	source.set(dashBrowserObservation("completed"), true)
	webDashBrowserRefresh(t, browser)
	webBoardBrowserWait(t, browser, `document.querySelector('#dash-observation-status').dataset.state==='stale'`)
	webDashBrowserCapture(t, browser, "stale-narrow")
	source.set(dashBrowserObservation("completed"), false)
	webDashBrowserRefresh(t, browser)
	webBoardBrowserWait(t, browser, `document.querySelector('#dash-observation-status').dataset.state==='fresh'`)
	browser.navigate(t, base+"/dash")
	browser.wait(t, ".dash-recent .dash-context")

	if source.writeCount() != 2 {
		t.Fatal("reopen started work")
	}
}

// Hidden tabs pause scheduled observation; visible tabs resume without a control.
func webDashBrowserPolling(t *testing.T, browser *webWorkflowBrowser, source *dashBrowserSource) {
	t.Helper()
	browser.evaluate(t, `Object.defineProperty(document,'visibilityState',{configurable:true,value:'hidden'});'hidden'`)
	source.mu.Lock()
	before := source.reads
	source.mu.Unlock()
	browser.evaluate(t, `new Promise(resolve=>setTimeout(()=>resolve('waited'),1200))`)
	source.mu.Lock()
	after := source.reads
	source.mu.Unlock()

	if after != before {
		t.Fatal("hidden tab polled", before, after)
	}

	browser.evaluate(t, `delete document.visibilityState;'visible'`)

	changed := dashBrowserObservation("completed")
	changed.Results[0].Message = "Retained result refreshed automatically"
	source.set(changed, false)
	webBoardBrowserWait(t, browser, `document.querySelector('.dash-recent').textContent.includes('Retained result refreshed automatically')`)

	if source.writeCount() != 2 {
		t.Fatal("scheduled observation started work")
	}
}

func webDashBrowserCapture(t *testing.T, browser *webWorkflowBrowser, name string) {
	t.Helper()

	metrics := browser.evaluate(t, `(()=>{const r=e=>{const b=e.getBoundingClientRect();return {width:b.width,height:b.height,x:b.x,y:b.y}},layout=document.querySelector('.dash-layout');return JSON.stringify({viewport:{width:innerWidth,height:innerHeight,dpr:devicePixelRatio,browserZoom:1,pageScale:visualViewport.scale},documentWidth:document.documentElement.scrollWidth,columns:getComputedStyle(layout).gridTemplateColumns,queue:r(document.querySelector('.dash-queue')),current:r(document.querySelector('.dash-current')),recent:r(document.querySelector('.dash-recent')),contexts:[...document.querySelectorAll('.dash-current .dash-context')].map(e=>e.dataset.contextId),results:document.querySelectorAll('.dash-recent .dash-context').length,events:document.querySelectorAll('.dash-event').length,observation:document.querySelector('#dash-observation-status').dataset.state});})()`)
	t.Log(name, metrics)

	if browser.evaluate(t, `String(document.documentElement.scrollWidth<=innerWidth && (innerWidth<761 ? document.querySelector('.dash-current').getBoundingClientRect().y>document.querySelector('.dash-queue').getBoundingClientRect().y : document.querySelector('.dash-current').getBoundingClientRect().x>document.querySelector('.dash-queue').getBoundingClientRect().x))`) != "true" {
		t.Fatal("Dash geometry or overflow changed", metrics)
	}

	directory := os.Getenv("CHR_WEB_DASH_SCREENSHOTS")
	if directory == "" {
		return
	}

	webBrowserCapture(t, browser, directory, name)

	if err := os.WriteFile(filepath.Join(directory, name+".json"), []byte(metrics+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
