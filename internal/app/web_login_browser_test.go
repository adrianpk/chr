package app

import (
	"net/http/httptest"
	"os"
	"testing"

	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/webui"
)

// The real browser logs in, navigates without URL tokens and returns to login
// when an HTMX refresh loses its session. Narrow-screen layout remains usable.
func TestWebLoginBrowser(t *testing.T) {
	if testing.Short() {
		t.Skip("owned login browser integration is selected explicitly")
	}

	path := os.Getenv("CHR_WEB_BROWSER")
	if path == "" {
		t.Skip("set CHR_WEB_BROWSER to select the owned Chromium fixture")
	}

	credentials, err := webui.CreateLoginCredentials("operator", "example-only-password-2026")
	if err != nil {
		t.Fatal(err)
	}

	h, err := newWebServeMux(resolvedWebServeConfig{ProjectDir: t.TempDir(), Web: workspaceconfig.DefaultWebConfig(), Login: &credentials})
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(h)
	defer server.Close()

	browser := newWebWorkflowBrowser(t, path)
	webBoardBrowserViewport(t, browser, 390, 844)
	browser.navigate(t, server.URL+"/specs")
	browser.wait(t, ".login-card")
	webBoardBrowserSet(t, browser, "[name=username]", "operator")
	webBoardBrowserSet(t, browser, "[name=password]", "example-only-password-2026")
	browser.clickNavigation(t, ".login-card button")
	browser.wait(t, "form[action='/logout']")

	if browser.evaluate(t, "location.pathname") != "/specs" {
		t.Fatal("login lost the requested destination")
	}

	browser.navigate(t, server.URL+"/board")
	browser.wait(t, "form[action='/logout']")
	browser.call(t, "Network.clearBrowserCookies", map[string]any{})
	browser.evaluate(t, "htmx.ajax('GET', '/board/fragment', {target:'#workspace'}); 'requested'")
	browser.wait(t, ".login-card")

	if browser.evaluate(t, "String(document.documentElement.scrollWidth <= innerWidth)") != "true" {
		t.Fatal("mobile login overflows the viewport")
	}
}
