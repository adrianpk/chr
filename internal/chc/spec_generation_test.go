package chc

import (
	"strings"
	"testing"
	"time"
)

func TestRenderGeneratedSpecRemovesDuplicateTitleHeading(t *testing.T) {
	spec := GeneratedSpec{
		Scope:            QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"},
		Title:            "Newsletter Distribution",
		Body:             "# Newsletter Distribution\n\nSummary\n\nPrepare local newsletter artifacts.",
		SourceRequestID:  "request-1",
		SourceResponseID: "response-1",
		CreatedAt:        time.Date(2026, 5, 27, 18, 0, 0, 0, time.UTC),
	}

	rendered := RenderGeneratedSpec(spec)

	if strings.Count(rendered, "# Newsletter Distribution") != 1 {
		t.Fatalf("rendered spec contains duplicate title heading:\n%s", rendered)
	}

	if !strings.Contains(rendered, "Summary\n\nPrepare local newsletter artifacts.") {
		t.Fatalf("rendered spec missing normalized body:\n%s", rendered)
	}
}

func TestRenderGeneratedSpecKeepsDifferentBodyHeading(t *testing.T) {
	spec := GeneratedSpec{
		Scope:            QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"},
		Title:            "Newsletter Distribution",
		Body:             "# Scope\n\nPrepare local newsletter artifacts.",
		SourceRequestID:  "request-1",
		SourceResponseID: "response-1",
		CreatedAt:        time.Date(2026, 5, 27, 18, 0, 0, 0, time.UTC),
	}

	rendered := RenderGeneratedSpec(spec)

	if !strings.Contains(rendered, "# Newsletter Distribution\n\n# Scope") {
		t.Fatalf("rendered spec removed a distinct body heading:\n%s", rendered)
	}
}

func TestValidateGeneratedSpecRejectsUnsupportedChrysoCoordPathReference(t *testing.T) {
	spec := GeneratedSpec{
		Scope:     QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"},
		Title:     "Newsletter Distribution",
		Body:      "Store newsletter drafts in .chc/newsletters/drafts.",
		CreatedAt: time.Date(2026, 5, 27, 18, 0, 0, 0, time.UTC),
	}

	err := ValidateGeneratedSpec(spec)
	if err == nil {
		t.Fatal("ValidateGeneratedSpec returned nil error")
	}

	if !strings.Contains(err.Error(), `.chc/newsletters/drafts`) {
		t.Fatalf("error = %q, want unsupported path reference", err)
	}
}

func TestValidateGeneratedSpecAllowsKnownChrysoCoordPathReferences(t *testing.T) {
	spec := GeneratedSpec{
		Scope: QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"},
		Title: "Newsletter Distribution",
		Body: strings.Join([]string{
			"See .chc/project.md.",
			"See .chc/cases/newsletter.md, .chc/dumps/newsletter.md, and .chc/quizzes/newsletter.md.",
		}, "\n"),
		CreatedAt: time.Date(2026, 5, 27, 18, 0, 0, 0, time.UTC),
	}

	if err := ValidateGeneratedSpec(spec); err != nil {
		t.Fatalf("ValidateGeneratedSpec returned error: %v", err)
	}
}
