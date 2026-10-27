package app

import (
	"os"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

type quizCatalogScenario struct {
	name, stop                     string
	answer, skip, counter, curated bool
}

// Actual confirmed entries decide membership, including partial and stopped
// sessions. Counters, curated documents and skipped entries cannot add cards.
func TestQuizCatalogMembership(t *testing.T) {
	cases := []quizCatalogScenario{
		{name: "absent"},
		{name: "active"},
		{name: "skip", skip: true},
		{name: "partial", answer: true},
		{name: "cancelled empty", stop: "cancel"},
		{name: "cancelled answered", stop: "cancel", answer: true},
		{name: "completed empty", stop: "done"},
		{name: "completed answered", stop: "done", answer: true},
		{name: "counter only", counter: true},
		{name: "curated only", curated: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { checkQuizCatalogScenario(t, tc) })
	}
}

func checkQuizCatalogScenario(t *testing.T, tc quizCatalogScenario) {
	t.Helper()

	provider, model := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	scope := operations.quizScope(quizcore.QuizScope{CaseSlug: "named-work"})
	if tc.name != "absent" && !tc.curated {
		seedQuizCatalogScope(t, operations, scope, tc.answer, tc.skip, tc.stop)
	}

	seedQuizCatalogMetadata(t, operations, scope, tc)

	requests := len(model.requests)

	catalog, err := operations.ReadQuizCatalog()

	wantItems, wantPending := 1, 2
	if tc.answer {
		wantItems, wantPending = 2, 1
	}

	checkQuizCatalogCounts(t, catalog, err, wantItems, wantPending)

	if catalog.Items[0].Title != "Project Quiz" || catalog.Pending[0].Scope.CaseSlug != "" {
		t.Fatal(catalog)
	}

	if tc.answer {
		checkQuizCatalogAnswer(t, catalog.Items[1])
	}

	if len(model.requests) != requests {
		t.Fatal("catalog generated questions")
	}
}

func seedQuizCatalogMetadata(t *testing.T, operations service, scope quizcore.QuizScope, tc quizCatalogScenario) {
	t.Helper()

	if tc.counter {
		state, err := operations.quizRepository().LoadRuntimeState(t.Context(), scope)
		if err != nil {
			t.Fatal(err)
		}

		state.AnswersRecorded = 1
		if _, err := operations.quizRepository().SaveRuntimeState(t.Context(), state); err != nil {
			t.Fatal(err)
		}
	}

	if tc.curated {
		_, err := operations.quizRepository().WriteCuratedNotes(t.Context(), quizcore.QuizCuratedNotes{Scope: scope, Status: "curated", Decisions: []string{"A document is not an answer."}, UpdatedAt: serviceQuizClock()().Format("2006-01-02T15:04:05Z")})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func seedQuizCatalogScope(t *testing.T, operations service, scope quizcore.QuizScope, answer, skip bool, stop string) {
	t.Helper()

	if _, err := operations.StartQuiz(QuizStartRequest{Scope: scope, Mode: quizcore.QuizModeFast}); err != nil {
		t.Fatal(err)
	}

	if answer {
		if _, err := operations.AnswerQuiz(QuizAnswerRequest{Scope: scope, Answer: "Keep confirmed context available."}); err != nil {
			t.Fatal(err)
		}
	}

	if skip {
		if _, err := operations.SkipQuiz(QuizSkipRequest{Scope: scope}); err != nil {
			t.Fatal(err)
		}
	}

	switch stop {
	case "cancel":
		if _, err := operations.CancelQuiz(QuizCancelRequest{Scope: scope}); err != nil {
			t.Fatal(err)
		}
	case "done":
		if _, err := operations.CompleteQuiz(QuizCompleteRequest{Scope: scope}); err != nil {
			t.Fatal(err)
		}
	}
}

// Retained answers survive a missing runtime, all-answered removes New Quiz,
// and a newly named case reintroduces only that pending destination.
func TestQuizCatalogNewCase(t *testing.T) {
	provider, _ := newWebQuizFixture(t)

	operations := provider.service(t.Context())
	for _, slug := range []string{"", "named-work"} {
		scope := operations.quizScope(quizcore.QuizScope{CaseSlug: slug})
		seedQuizCatalogScope(t, operations, scope, true, false, "done")

		paths, err := operations.quizRepository().Paths(scope)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.Remove(paths.RuntimePath); err != nil {
			t.Fatal(err)
		}
	}

	catalog, err := operations.ReadQuizCatalog()
	checkQuizCatalogCounts(t, catalog, err, 2, 0)

	if _, err := core.NewCaseMetadataService(operations.projectContext.ProjectDir).Save(core.CaseMetadata{SourcePath: core.CaseMetadataPath(operations.projectContext.ProjectDir, "new-case"), Name: "Newly named work"}); err != nil {
		t.Fatal(err)
	}

	catalog, err = operations.ReadQuizCatalog()
	checkQuizCatalogCounts(t, catalog, err, 2, 1)

	if catalog.Pending[0].Title != "Newly named work" {
		t.Fatal(catalog, err)
	}
}

func checkQuizCatalogCounts(t *testing.T, catalog QuizCatalog, err error, items, pending int) {
	t.Helper()

	if err != nil || len(catalog.Items) != items || len(catalog.Pending) != pending {
		t.Fatal(catalog, err)
	}
}

func checkQuizCatalogAnswer(t *testing.T, item QuizCatalogItem) {
	t.Helper()

	if item.AnswerCount != 1 || item.Preview == "" || item.SessionID == "" {
		t.Fatal(item)
	}
}

// With no cases or retained files, Project Quiz is both the sole card and the
// sole pending choice. Reading it leaves the runtime and answer files absent.
func TestQuizCatalogProjectOnly(t *testing.T) {
	provider, model := newWebQuizFixture(t)

	operations := provider.service(t.Context())
	if err := os.Remove(core.CaseMetadataPath(operations.projectContext.ProjectDir, "named-work")); err != nil {
		t.Fatal(err)
	}

	catalog, err := operations.ReadQuizCatalog()
	checkQuizCatalogCounts(t, catalog, err, 1, 1)

	if catalog.Items[0].Scope.CaseSlug != "" || catalog.Items[0].Preview != "" || len(model.requests) != 0 {
		t.Fatal(catalog)
	}

	paths, err := operations.quizRepository().Paths(operations.quizScope(quizcore.QuizScope{}))
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{paths.RawPath, paths.CuratedPath, paths.RuntimePath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("read allocated a quiz file", path, err)
		}
	}
}
