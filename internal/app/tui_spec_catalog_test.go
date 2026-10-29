package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	"forge.adrianpk.com/chr/internal/tui"
)

// The production TUI catalog must contain every manual and generated document,
// including independent documents in one case and project-wide records, like Web.
func TestTUISpecCatalog(t *testing.T) {
	provider, model := newWebQuizFixture(t)

	operations := provider.service(t.Context())
	for _, item := range []struct{ id, slug string }{{"first", "named-work"}, {"second", "named-work"}, {"project", ""}} {
		if _, err := operations.CreateSpec(CreateSpecRequest{OperationID: item.id, Title: item.id, Body: "Actual spec body.", Status: "draft", CaseSlug: item.slug}); err != nil {
			t.Fatal(err)
		}
	}

	model.responses = []quizmodel.Response{serviceQuizDraftSpecResponse("catalog-generated", serviceQuizClock()())}

	if _, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{Scope: quizcore.Scope{CaseSlug: "named-work"}}); err != nil {
		t.Fatal(err)
	}

	terminal, err := loadTUISpecs(operations)
	if err != nil || len(terminal) != 4 {
		t.Fatal("terminal omitted canonical documents", terminal, err)
	}

	checkTUISpecCatalog(t, operations, terminal)
	checkTUISpecHTTP(t, provider, terminal)
}

func checkTUISpecHTTP(t *testing.T, provider webViewProvider, terminal []tui.SpecView) {
	t.Helper()

	response := httptest.NewRecorder()
	webQuizHandler(t, provider).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/specs", nil))

	if response.Code != http.StatusOK {
		t.Fatal(response.Code, response.Body.String())
	}

	for _, document := range terminal {
		if !strings.Contains(response.Body.String(), url.QueryEscape(document.RelativePath)) {
			t.Fatal("HTTP catalog omitted canonical reference", document.RelativePath)
		}
	}
}

func checkTUISpecCatalog(t *testing.T, operations service, terminal []tui.SpecView) {
	t.Helper()

	web, err := specCatalog(operations)
	if err != nil || len(web.Items) != len(terminal) {
		t.Fatal(web, err)
	}

	for index, document := range terminal {
		if document.RelativePath != web.Items[index].ID || document.Title != web.Items[index].Title {
			t.Fatal("Web and TUI enumerated different documents", terminal, web)
		}
	}
}
