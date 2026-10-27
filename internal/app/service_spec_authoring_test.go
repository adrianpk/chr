package app

import (
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Concrete listing preserves independent identities in one scope while the
// existing latest-in-scope inventory remains available to generation adapters.
func TestSpecDocumentCatalog(t *testing.T) {
	provider, model := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	documents, err := operations.ListSpecDocuments()
	checkSpecDocumentCount(t, documents, err, 0)

	for _, status := range []string{"draft", "approved"} {
		if _, err := operations.CreateSpec(CreateSpecRequest{OperationID: status, Title: "Same scope spec", Body: "Actual opening text.", Status: status, CaseSlug: "named-work"}); err != nil {
			t.Fatal(err)
		}
	}

	if len(model.requests) != 0 {
		t.Fatal("manual creation invoked a model")
	}

	model.responses = []core.QuizModelResponse{serviceQuizDraftSpecResponse("generated", serviceQuizClock()())}

	if _, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{Scope: quizcore.Scope{CaseSlug: "named-work"}}); err != nil {
		t.Fatal(err)
	}

	documents, err = operations.ListSpecDocuments()
	checkSpecDocumentCount(t, documents, err, 3)

	latest, err := operations.ListSpecs()
	if err != nil || len(latest.Specs) != 2 {
		t.Fatal(latest, err)
	}
}

// Both project-wide and selected-case generation receive general context and
// optional matching discovery, without leaking another case or its quiz.
func TestSpecGenerationScopes(t *testing.T) {
	for _, slug := range []string{"", "named-work"} {
		t.Run("scope "+slug, func(t *testing.T) { checkSpecGenerationScope(t, slug) })
	}
}

func checkSpecGenerationScope(t *testing.T, slug string) {
	t.Helper()

	provider, model := newWebQuizFixture(t)
	operations := provider.service(t.Context())
	seedSpecGenerationScopes(t, operations)

	model.responses = []core.QuizModelResponse{serviceQuizDraftSpecResponse("generated", serviceQuizClock()())}

	if _, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{Scope: quizcore.Scope{CaseSlug: slug}}); err != nil {
		t.Fatal(err)
	}

	var content strings.Builder
	for _, section := range model.requests[0].ContextSections {
		content.WriteString(section.Content)
	}

	for _, want := range []string{"GENERAL_CONTEXT", "Workflow discovery"} {
		if !strings.Contains(content.String(), want) {
			t.Fatal("missing", want, content.String())
		}
	}

	assertSpecGenerationScope(t, slug, content.String())
}

func assertSpecGenerationScope(t *testing.T, slug, content string) {
	t.Helper()

	for _, token := range []string{"NAMED_DUMP", "NAMED_QUIZ", "PROJECT_QUIZ", "OTHER_DUMP", "OTHER_QUIZ"} {
		want := token == "PROJECT_QUIZ" && slug == "" || (token == "NAMED_DUMP" || token == "NAMED_QUIZ") && slug == "named-work"
		if strings.Contains(content, token) != want {
			t.Fatal("wrong scope context", slug, token, content)
		}
	}
}

func seedSpecGenerationScopes(t *testing.T, operations service) {
	t.Helper()

	operations.dumpCurator = fallbackDumpCurationProvider{Reason: "deterministic context fixture"}

	metadata := core.CaseMetadata{SourcePath: core.CaseMetadataPath(operations.projectContext.ProjectDir, "other"), Name: "Other case"}
	if _, err := core.NewCaseMetadataService(operations.projectContext.ProjectDir).Save(metadata); err != nil {
		t.Fatal(err)
	}

	for _, item := range []struct{ slug, body string }{{"", "GENERAL_CONTEXT"}, {"named-work", "NAMED_DUMP"}, {"other", "OTHER_DUMP"}} {
		request := DumpCreateRequest{OperationID: "dump-" + item.body, Namespace: dumpSpecTestNamespace(item.slug), Title: item.body, Content: item.body, Scope: DumpScope{CaseSlug: item.slug}}
		if _, err := operations.CreateDump(request); err != nil {
			t.Fatal(err)
		}
	}

	for _, item := range []struct{ slug, body string }{{"", "PROJECT_QUIZ"}, {"named-work", "NAMED_QUIZ"}, {"other", "OTHER_QUIZ"}} {
		notes := quizcore.CuratedNotes{Scope: operations.quizScope(quizcore.Scope{CaseSlug: item.slug}), Status: "curated", Decisions: []string{item.body}, UpdatedAt: serviceQuizClock()().Format("2006-01-02T15:04:05Z")}
		if _, err := operations.quizRepository().WriteCuratedNotes(t.Context(), notes); err != nil {
			t.Fatal(err)
		}
	}
}

func dumpSpecTestNamespace(slug string) string {
	if slug == "" {
		return ""
	}

	return "default"
}

func checkSpecDocumentCount(t *testing.T, documents []core.SpecDocument, err error, count int) {
	t.Helper()

	if err != nil || len(documents) != count {
		t.Fatal(documents, err)
	}
}
