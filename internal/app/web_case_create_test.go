package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

// An empty Cases page must let the user create a canonical case and immediately
// select it when capturing context. Invalid input cannot create a partial case.
func TestWebCaseCreate(t *testing.T) {
	provider := webCaseCreateProvider(t)
	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	page := webWorkflowHTTP(t, handler, http.MethodGet, "/cases", nil)
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), `data-modal-template="case-create-modal"`) {
		t.Fatal("empty Cases page has no creation entry point", page.Code)
	}

	for _, test := range []struct{ name, title, status string }{
		{"missing name", " ", webui.WriteStatusValidation},
		{"named case", "Task management", webui.WriteStatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			response, err := provider.WriteAction(t.Context(), webui.WriteActionRequest{
				Action: "case.create", Family: "cases", Confirm: true,
				Fields: map[string]string{"name": test.title, "description": "Manage personal tasks.", "context": "Keep task data local."},
			})
			if err != nil || response.Status != test.status {
				t.Fatal("unexpected creation result", response, err)
			}
		})
	}

	checkWebCaseCreated(t, provider)
}

func checkWebCaseCreated(t *testing.T, provider webViewProvider) {
	t.Helper()

	cases, err := core.NewCaseMetadataService(provider.projectContext.ProjectDir).LoadAll()
	if err != nil || len(cases) != 1 || cases[0].Slug != "task-management" || cases[0].Context != "Keep task data local." {
		t.Fatal("creation lost its canonical name, context or generated identity", cases, err)
	}

	dumps, err := provider.CatalogView(t.Context(), "dumps")
	if err != nil || len(dumps.Cases) != 1 || dumps.Cases[0].Name != "Task management" {
		t.Fatal("created case is unavailable to dump capture", dumps.Cases, err)
	}
}

func webCaseCreateProvider(t *testing.T) webViewProvider {
	t.Helper()

	project := newTestWorkspace(t)

	return newWebServeBoardTestProvider(project, newWebServeBoardTestEnv(t, t.TempDir()), newWebServeBoardProjectContext(t, project))
}

// The real browser opens New from an empty catalog, submits without a slug,
// observes the refreshed table and finds the new scope in the dump selector.
func TestWebCaseCreateBrowser(t *testing.T) {
	if testing.Short() || os.Getenv("CHR_WEB_BROWSER") == "" {
		t.Skip("select the owned Chromium fixture explicitly")
	}

	provider := webCaseCreateProvider(t)
	config := provider.config
	config.Provider = provider

	handler, err := newWebServeMux(config)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	browser := newWebWorkflowBrowser(t, os.Getenv("CHR_WEB_BROWSER"))
	browser.navigate(t, server.URL+"/cases")
	webBoardBrowserClick(t, browser, `[data-modal-template="case-create-modal"]`)
	browser.wait(t, "[data-case-create]")
	webBoardBrowserSet(t, browser, `[data-case-create] [name="name"]`, " ")
	webBoardBrowserSet(t, browser, `[data-case-create] [name="description"]`, "Manage personal tasks.")
	webBoardBrowserSet(t, browser, `[data-case-create] [name="context"]`, "Keep task data local.")
	webBoardBrowserClick(t, browser, `[data-case-create] button[type="submit"]`)
	browser.wait(t, `[data-case-create] [data-write-status="validation_failed"]`)

	if got := browser.evaluate(t, `document.querySelector('[data-case-create] [name="description"]').value`); got != "Manage personal tasks." {
		t.Fatal("validation discarded the new case input", got)
	}

	webBoardBrowserSet(t, browser, `[data-case-create] [name="name"]`, "Task management")
	webBoardBrowserClick(t, browser, `[data-case-create] button[type="submit"]`)
	webBoardBrowserWait(t, browser, `!document.querySelector('#modal-root [role="dialog"]') && document.querySelector('.case-table')?.textContent.includes('Task management')`)

	browser.navigate(t, server.URL+"/dumps")
	webBoardBrowserClick(t, browser, `[data-modal-template="dump-create-modal"]`)
	browser.wait(t, `[data-dump-capture] option[value="task-management"]`)

	if got := browser.evaluate(t, `document.querySelector('[data-dump-capture] option[value="task-management"]').textContent`); got != "Task management" {
		t.Fatal("dump selector lost the new case name", got)
	}
}
