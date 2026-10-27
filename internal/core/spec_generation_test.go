package core

import (
	"strings"
	"testing"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Valid workspace protocol references survive publication unchanged; legacy
// paths normalize once and unsupported namespaced paths remain rejectable.
func TestGeneratedSpecProtocolPaths(t *testing.T) {
	for _, item := range []struct {
		name, path, want string
		valid            bool
	}{
		{"canonical", ".chr/default/agents/ARCHITECTURE.md", ".chr/default/agents/ARCHITECTURE.md", true},
		{"legacy", ".chr/agents/ARCHITECTURE.md", ".chr/default/agents/ARCHITECTURE.md", true},
		{"unsupported", ".chr/default/invented/data.json", ".chr/default/invented/data.json", false},
	} {
		t.Run(item.name, func(t *testing.T) {
			body := "See `" + item.path + "`."

			got := NormalizeGeneratedSpecChrPathReferences(body)
			if got != "See `"+item.want+"`." || NormalizeGeneratedSpecChrPathReferences(got) != got {
				t.Fatal("normalization altered the namespace", got)
			}

			err := ValidateGeneratedSpec(GeneratedSpec{Scope: quizcore.QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"}, Title: "Newsletter requirements", Body: got, CreatedAt: time.Now().UTC()})
			if (err == nil) != item.valid {
				t.Fatal("incorrect protocol reference validation", err)
			}
		})
	}
}

func TestRenderGeneratedSpecRemovesDuplicateTitleHeading(t *testing.T) {
	spec := GeneratedSpec{
		Scope:            quizcore.QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"},
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
		Scope:            quizcore.QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"},
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

func TestValidateGeneratedSpecRejectsUnsupportedChrPathReference(t *testing.T) {
	spec := GeneratedSpec{
		Scope:     quizcore.QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"},
		Title:     "Newsletter Distribution",
		Body:      "Store newsletter drafts in .chr/newsletters/drafts.",
		CreatedAt: time.Date(2026, 5, 27, 18, 0, 0, 0, time.UTC),
	}

	err := ValidateGeneratedSpec(spec)
	if err == nil {
		t.Fatal("ValidateGeneratedSpec returned nil error")
	}

	if !strings.Contains(err.Error(), `.chr/newsletters/drafts`) {
		t.Fatalf("error = %q, want unsupported path reference", err)
	}
}

func TestValidateGeneratedSpecAllowsGherkinExample(t *testing.T) {
	spec := GeneratedSpec{
		Scope:     quizcore.QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"},
		Title:     "Newsletter Distribution",
		Body:      "Acceptance Criteria\n\n```gherkin\nFeature: Newsletter distribution\n```\n",
		CreatedAt: time.Date(2026, 5, 27, 18, 0, 0, 0, time.UTC),
	}

	err := ValidateGeneratedSpec(spec)
	if err != nil {
		t.Fatalf("ValidateGeneratedSpec returned error: %v", err)
	}
}

func TestValidateGeneratedSpecAllowsKnownChrPathReferences(t *testing.T) {
	spec := GeneratedSpec{
		Scope: quizcore.QuizScope{ProjectID: "project-1", CaseSlug: "newsletter"},
		Title: "Newsletter Distribution",
		Body: strings.Join([]string{
			"See .chr/default/project.md.",
			"See .chr/default/cases/newsletter.md, .chr/default/dumps/newsletter.md, and .chr/default/quizzes/newsletter.md.",
		}, "\n"),
		CreatedAt: time.Date(2026, 5, 27, 18, 0, 0, 0, time.UTC),
	}

	if err := ValidateGeneratedSpec(spec); err != nil {
		t.Fatalf("ValidateGeneratedSpec returned error: %v", err)
	}
}

func TestNormalizeGeneratedSpecChrPathReferencesKeepsKnownPaths(t *testing.T) {
	body := "See .chr/default/project.md and .chr/default/dumps/newsletter.md."

	got := NormalizeGeneratedSpecChrPathReferences(body)

	if got != body {
		t.Fatalf("normalized body = %q, want %q", got, body)
	}
}

func TestNormalizeGeneratedSpecChrPathReferencesMovesUnknownPathsOutOfChr(t *testing.T) {
	body := "Store drafts in .chr/newsletters/drafts/<post-slug>.md, then export them."

	got := NormalizeGeneratedSpecChrPathReferences(body)

	if strings.Contains(got, ".chr/newsletters") {
		t.Fatalf("normalized body still contains unsupported .chr path: %q", got)
	}

	if !strings.Contains(got, "newsletters/drafts/<post-slug>.md") {
		t.Fatalf("normalized body = %q, want application-owned path", got)
	}
}
