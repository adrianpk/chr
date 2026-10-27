package core

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

func TestAcceptanceFeatureLayoutUsesCanonicalPath(t *testing.T) {
	projectDir := t.TempDir()

	paths, err := AcceptanceFeatureLayout(projectDir, quizcore.Scope{ProjectID: "project-1", CaseSlug: "article-authoring"})
	if err != nil {
		t.Fatalf("AcceptanceFeatureLayout returned error: %v", err)
	}

	if paths.ScopeID != "article-authoring" {
		t.Fatalf("scope id = %q, want article-authoring", paths.ScopeID)
	}

	wantRelativePath := "ops/default/spec/00000-article-authoring-acceptance-feature/artifact.md"
	if paths.RelativePath != wantRelativePath {
		t.Fatalf("relative path = %q", paths.RelativePath)
	}

	if paths.Path != filepath.Join(projectDir, filepath.FromSlash(wantRelativePath)) {
		t.Fatalf("path = %q", paths.Path)
	}
}

func TestAcceptanceFeatureLayoutRejectsNestedCaseSlug(t *testing.T) {
	_, err := AcceptanceFeatureLayout(t.TempDir(), quizcore.Scope{ProjectID: "project-1", CaseSlug: "../bad"})
	if err == nil {
		t.Fatal("AcceptanceFeatureLayout returned nil error")
	}
}

func TestAcceptanceFeatureRoundLayoutUsesCanonicalPath(t *testing.T) {
	projectDir := t.TempDir()

	paths, err := AcceptanceFeatureRoundLayout(projectDir, quizcore.Scope{ProjectID: "project-1", CaseSlug: "article-authoring"})
	if err != nil {
		t.Fatalf("AcceptanceFeatureRoundLayout returned error: %v", err)
	}

	wantRelativePath := "ops/default/spec/00000-article-authoring-acceptance-feature-round/artifact.md"
	if paths.RelativePath != wantRelativePath {
		t.Fatalf("relative path = %q", paths.RelativePath)
	}

	if paths.Path != filepath.Join(projectDir, filepath.FromSlash(wantRelativePath)) {
		t.Fatalf("path = %q", paths.Path)
	}
}

func TestValidateAcceptanceFeatureArtifactRequiresTraceabilityInputs(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.SpecSlug = ""

	err := ValidateAcceptanceFeatureArtifact(feature)
	if err == nil {
		t.Fatal("ValidateAcceptanceFeatureArtifact returned nil error")
	}
}

func TestValidateAcceptanceFeatureGenerationRequiresPromptsForNeedsMoreInfo(t *testing.T) {
	_, err := generationcore.ValidateAcceptanceFeatureGeneration(generationcore.AcceptanceFeatureGeneration{
		Status:          generationcore.AcceptanceFeatureStatusNeedsMoreInfo,
		RejectionReason: "The failure behavior is missing.",
	})
	if err == nil {
		t.Fatal("ValidateAcceptanceFeatureGeneration returned nil error")
	}
}

func TestDecodeQuizModelFeaturePayloadKeepsClarificationPrompts(t *testing.T) {
	payload, err := DecodeQuizModelFeaturePayload([]byte(`{
		"status": "needs_more_info",
		"rejection_reason": "The failure behavior is missing.",
		"clarification_prompts": [
			{
				"id": "failure-path",
				"question": "What should happen when publishing fails?",
				"default_answer": "The article remains unpublished.",
				"reason": "The feature needs one observable failure path."
			}
		]
	}`))
	if err != nil {
		t.Fatalf("DecodeQuizModelFeaturePayload returned error: %v", err)
	}

	if len(payload.ClarificationPrompts) != 1 {
		t.Fatalf("ClarificationPrompts = %#v, want one prompt", payload.ClarificationPrompts)
	}

	if payload.ClarificationPrompts[0].DefaultAnswer != "The article remains unpublished." {
		t.Fatalf("DefaultAnswer = %q", payload.ClarificationPrompts[0].DefaultAnswer)
	}
}

func TestCheckAcceptanceFeatureArtifactAcceptsValidFeature(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `@spec:article-authoring-spec @use-case:article-authoring
Feature: Article authoring

  @requirement:publish-report
  Scenario: Publish report is written
    Given an existing article draft
    When the editor publishes the article
    Then a publication report is written

  @requirement:publish-report
  Scenario: Publishing dependency fails
    Given an existing article draft
    When the publishing dependency is unavailable
    Then the article remains unpublished
`

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		t.Fatalf("CheckAcceptanceFeatureArtifact returned error: %v", err)
	}

	if !result.Valid {
		t.Fatalf("valid = false, findings = %#v", result.Findings)
	}
}

func TestCheckAcceptanceFeatureArtifactAcceptsBackgroundSteps(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `@spec:article-authoring-spec
Feature: Article authoring

  Background:
    Given an existing article draft

  Scenario: Publish report is written
    When the editor publishes the article
    Then a publication report is written

  Scenario: Publishing dependency fails
    When the publishing dependency is unavailable
    Then the article remains unpublished
`

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		t.Fatalf("CheckAcceptanceFeatureArtifact returned error: %v", err)
	}

	if !result.Valid {
		t.Fatalf("valid = false, findings = %#v", result.Findings)
	}
}

func TestCheckAcceptanceFeatureArtifactRequiresGivenWhenThen(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `@spec:article-authoring-spec
Feature: Article authoring

  Scenario: Publish report is written
    Given an existing article draft
    Then a publication report is written
`

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		t.Fatalf("CheckAcceptanceFeatureArtifact returned error: %v", err)
	}

	assertAcceptanceFeatureFinding(t, result, "missing_when")
}

func TestCheckAcceptanceFeatureArtifactRejectsSingleBehaviorPath(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `@spec:article-authoring-spec
Feature: Article authoring

  Scenario: Publish report is written
    Given an existing article draft
    When the editor publishes the article
    Then a publication report is written
`

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		t.Fatalf("CheckAcceptanceFeatureArtifact returned error: %v", err)
	}

	assertAcceptanceFeatureFinding(t, result, "insufficient_behavior_paths")
}

func TestCheckAcceptanceFeatureArtifactRejectsMissingContrastPath(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `@spec:article-authoring-spec
Feature: Article authoring

  Scenario: Publish report is written
    Given an existing article draft
    When the editor publishes the article
    Then a publication report is written

  Scenario: Publish report is visible
    Given a published article exists
    When the editor views the publication report
    Then the publication report is visible
`

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		t.Fatalf("CheckAcceptanceFeatureArtifact returned error: %v", err)
	}

	assertAcceptanceFeatureFinding(t, result, "missing_contrast_path")
}

func TestCheckAcceptanceFeatureArtifactRejectsVagueSteps(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `@spec:article-authoring-spec
Feature: Article authoring

  Scenario: Publish report is written
    Given an existing article draft
    When the editor publishes the article
    Then it works correctly
`

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		t.Fatalf("CheckAcceptanceFeatureArtifact returned error: %v", err)
	}

	assertAcceptanceFeatureFinding(t, result, "vague_step")
}

func TestCheckAcceptanceFeatureArtifactRequiresTraceability(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `Feature: Article authoring

  Scenario: Publish report is written
    Given an existing article draft
    When the editor publishes the article
    Then a publication report is written
`

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		t.Fatalf("CheckAcceptanceFeatureArtifact returned error: %v", err)
	}

	assertAcceptanceFeatureFinding(t, result, "missing_traceability")
}

func TestCheckAcceptanceFeatureArtifactRejectsOpenQuestionScenarios(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `@spec:article-authoring-spec
Feature: Article authoring

  Scenario: Publish report is written
    Given an existing article draft
    When the editor publishes the article
    Then an open question decides which report is written
`

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		t.Fatalf("CheckAcceptanceFeatureArtifact returned error: %v", err)
	}

	assertAcceptanceFeatureFinding(t, result, "open_question")
}

func TestCheckAcceptanceFeatureArtifactRejectsImplementationDetails(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `@spec:article-authoring-spec
Feature: Article authoring

  Scenario: Publish report is written
    Given an existing database table contains an article draft
    When the editor publishes the article
    Then a publication report is written
`

	result, err := CheckAcceptanceFeatureArtifact(feature)
	if err != nil {
		t.Fatalf("CheckAcceptanceFeatureArtifact returned error: %v", err)
	}

	assertAcceptanceFeatureFinding(t, result, "implementation_detail")
}

func TestWriteAcceptanceFeatureWritesCheckedArtifact(t *testing.T) {
	projectDir := t.TempDir()
	feature := validAcceptanceFeatureArtifact()
	feature.Body = validAcceptanceFeatureBody()

	paths, result, err := WriteAcceptanceFeature(context.Background(), projectDir, feature)
	if err != nil {
		t.Fatalf("WriteAcceptanceFeature returned error: %v", err)
	}

	if !result.Valid {
		t.Fatalf("valid = false, findings = %#v", result.Findings)
	}

	if paths.RelativePath != "ops/default/spec/00000-article-authoring-acceptance-feature/artifact.md" {
		t.Fatalf("relative path = %q", paths.RelativePath)
	}

	content, err := os.ReadFile(paths.Path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if !strings.Contains(string(content), "Feature: Article authoring") {
		t.Fatalf("feature content = %q", string(content))
	}
}

func TestWriteAcceptanceFeatureRejectsInvalidArtifact(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `Feature: Article authoring

  Scenario: Publish report is written
    Given an existing article draft
    When the editor publishes the article
    Then it works correctly
`

	_, result, err := WriteAcceptanceFeature(context.Background(), t.TempDir(), feature)
	if err == nil {
		t.Fatal("WriteAcceptanceFeature returned nil error")
	}

	assertAcceptanceFeatureFinding(t, result, "vague_step")
}

func TestMergeAcceptanceFeatureRoundAnswersPreservesExistingAnswers(t *testing.T) {
	notes := AcceptanceFeatureRoundNotes{
		Scope:           quizcore.Scope{ProjectID: "project-1", CaseSlug: "followups"},
		Status:          generationcore.AcceptanceFeatureStatusNeedsMoreInfo,
		RejectionReason: "More queue behavior is needed.",
		Prompts: []generationcore.AcceptanceFeaturePrompt{
			{
				ID:            "time_source_default",
				Question:      "Confirm time source?",
				DefaultAnswer: "Use an injected clock.",
			},
			{
				ID:            "due_boundary_default",
				Question:      "Confirm due boundary?",
				DefaultAnswer: "Due at or before now.",
			},
		},
		SourceRequestID:  "request-1",
		SourceResponseID: "response-1",
		CreatedAt:        quizRuntimeTestTime(),
	}

	existing := `# Acceptance Feature Round

## Prompts

### time_source_default

- Question: Confirm time source?
- Default: Use an injected clock.
- Answer: Use an injected UTC clock.

### old_prompt

- Question: Old prompt?
- Default: Old default.
- Answer: Old answer.
`

	merged := MergeAcceptanceFeatureRoundAnswers(notes, existing)

	if merged.Prompts[0].Answer != "Use an injected UTC clock." {
		t.Fatalf("Answer = %q, want preserved answer", merged.Prompts[0].Answer)
	}

	if merged.Prompts[1].Answer != "" {
		t.Fatalf("new prompt Answer = %q, want empty", merged.Prompts[1].Answer)
	}
}

func TestWriteAcceptanceFeatureRoundNotesWritesDefaultPrompts(t *testing.T) {
	projectDir := t.TempDir()
	notes := AcceptanceFeatureRoundNotes{
		Scope:           quizcore.Scope{ProjectID: "project-1", CaseSlug: "article-authoring"},
		Status:          generationcore.AcceptanceFeatureStatusNeedsMoreInfo,
		RejectionReason: "The failure behavior is not explicit.",
		Prompts: []generationcore.AcceptanceFeaturePrompt{
			{
				ID:            "failure-behavior",
				Question:      "What should happen when publishing fails?",
				DefaultAnswer: "The article remains unpublished and the user sees the failure.",
				Reason:        "The feature needs one observable rejection path.",
			},
		},
		SourceRequestID:  "request-feature",
		SourceResponseID: "response-feature",
		CreatedAt:        quizRuntimeTestTime(),
	}

	paths, err := WriteAcceptanceFeatureRoundNotes(context.Background(), projectDir, notes)
	if err != nil {
		t.Fatalf("WriteAcceptanceFeatureRoundNotes returned error: %v", err)
	}

	content, err := os.ReadFile(paths.Path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	text := string(content)
	for _, want := range []string{
		`status: "needs_more_info"`,
		"### failure-behavior",
		"- Default: The article remains unpublished and the user sees the failure.",
		"- Answer:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("round notes missing %q in:\n%s", want, text)
		}
	}
}

func assertAcceptanceFeatureFinding(t *testing.T, result AcceptanceFeatureCheckResult, code string) {
	t.Helper()

	if result.Valid {
		t.Fatalf("valid = true, want finding %q", code)
	}

	found := slices.ContainsFunc(result.Findings, func(finding AcceptanceFeatureCheckFinding) bool {
		return finding.Code == code
	})

	if !found {
		t.Fatalf("findings = %#v, want %q", result.Findings, code)
	}
}

func validAcceptanceFeatureArtifact() AcceptanceFeatureArtifact {
	return AcceptanceFeatureArtifact{
		Scope:       quizcore.Scope{ProjectID: "project-1", CaseSlug: "article-authoring"},
		SpecSlug:    "article-authoring-spec",
		SpecPath:    ".chr/default/specs/article-authoring-spec.md",
		UseCaseSlug: "article-authoring",
		Body:        "Feature: Article authoring\n",
		CreatedAt:   quizRuntimeTestTime(),
	}
}

func validAcceptanceFeatureBody() string {
	return `@spec:article-authoring-spec @use-case:article-authoring
Feature: Article authoring

  @requirement:publish-report
  Scenario: Publish report is written
    Given an existing article draft
    When the editor publishes the article
    Then a publication report is written

  @requirement:publish-report
  Scenario: Publishing dependency fails
    Given an existing article draft
    When the publishing dependency is unavailable
    Then the article remains unpublished
`
}
