package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// Real catalog and focused HTTP reads never create sessions; pending choices
// carry the named scope, and active/cancelled selections keep their identity.
func TestWebQuizCatalogSelection(t *testing.T) {
	for _, stop := range []string{"", "cancel", "done"} {
		t.Run("session "+stop, func(t *testing.T) { checkWebQuizCatalogSelection(t, stop) })
	}
}

func checkWebQuizCatalogSelection(t *testing.T, stop string) {
	t.Helper()

	provider, model := newWebQuizFixture(t)
	operations := provider.service(t.Context())
	scope := operations.quizScope(core.QuizScope{CaseSlug: "named-work"})
	seedQuizCatalogScope(t, operations, scope, false, false, stop)

	before := quizCatalogObservation(t, operations, scope)

	requests := len(model.requests)

	catalog, err := provider.CatalogView(t.Context(), "quizzes")
	if err != nil || len(catalog.Items) != 1 || len(catalog.PendingQuizzes) != 2 {
		t.Fatal(catalog, err)
	}

	if catalog.PendingQuizzes[1].URL != quizFocusURL(before.State.CuratedPath) {
		t.Fatal(catalog.PendingQuizzes)
	}

	body := webQuizGET(t, webQuizHandler(t, provider), before.State.CuratedPath)

	want := map[string]string{"": `id="quiz-answer"`, "cancel": `value="quiz.resume"`, "done": "Start new session"}[stop]

	if !strings.Contains(body, want) || !strings.Contains(body, before.State.SessionID) {
		t.Fatal(body)
	}

	after := quizCatalogObservation(t, operations, scope)
	if !reflect.DeepEqual(before, after) || requests != len(model.requests) {
		t.Fatal("selection mutated the session")
	}
}

// A completed empty session stays pending until explicit Start creates a new
// session. Skipped entries from the previous session remain byte-for-byte.
func TestWebQuizCatalogRestart(t *testing.T) {
	provider, _ := newWebQuizFixture(t)
	operations := provider.service(t.Context())
	scope := operations.quizScope(core.QuizScope{CaseSlug: "named-work"})
	seedQuizCatalogScope(t, operations, scope, false, true, "done")

	before := quizCatalogObservation(t, operations, scope)

	paths, err := operations.quizRepository().Paths(scope)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(paths.RawPath)
	if err != nil {
		t.Fatal(err)
	}

	form := quizHTTPForm(before.State, "quiz.start", "explicit-restart")
	form.Set("mode", "fast")

	response := webDumpEditorPost(webQuizHandler(t, provider), form)
	if !strings.Contains(response.Body.String(), `data-ack-action="start"`) {
		t.Fatal(response.Body.String())
	}

	after := quizCatalogObservation(t, operations, scope)

	if after.State.SessionID == before.State.SessionID || len(after.Entries) != 1 || !reflect.DeepEqual(after.Entries[0], before.Entries[0]) {
		t.Fatal(before, after)
	}

	current, err := os.ReadFile(paths.RawPath)
	if err != nil || string(current) != string(raw) {
		t.Fatal("restart replaced retained history", err)
	}
}

// Catalog markup exposes only the specified card anatomy and named choices.
// The all-answered fragment removes New Quiz and its detached slug forms.
func TestWebQuizCatalogMarkup(t *testing.T) {
	provider, _ := newWebQuizFixture(t)
	handler := webQuizHandler(t, provider)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/quizzes", nil))

	body := response.Body.String()
	for _, want := range []string{`class="quiz-card"`, "No answers yet", "New Quiz", `aria-label="Pending quizzes"`} {
		if !strings.Contains(body, want) {
			t.Fatal(want, body)
		}
	}

	for _, unwanted := range []string{"Artifacts", "Case slug", "not_started", "quiz.start", "quiz.resume", "Start Quiz", "Resume Quiz", `class="kv-list"`} {
		if strings.Contains(body, unwanted) {
			t.Fatal(unwanted, body)
		}
	}

	operations := provider.service(t.Context())
	for _, slug := range []string{"", "named-work"} {
		seedQuizCatalogScope(t, operations, operations.quizScope(core.QuizScope{CaseSlug: slug}), true, false, "")
	}

	request := httptest.NewRequest(http.MethodGet, "/quizzes/fragment", nil)
	request.Header.Set("HX-Request", "true")

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "New Quiz") || strings.Count(response.Body.String(), `class="quiz-card"`) != 2 {
		t.Fatal(response.Code, response.Body.String())
	}
}

func quizCatalogObservation(t *testing.T, operations service, scope core.QuizScope) core.QuizObservation {
	t.Helper()

	observed, err := operations.quizRepository().Observe(t.Context(), scope)
	if err != nil {
		t.Fatal(err)
	}

	return observed
}
