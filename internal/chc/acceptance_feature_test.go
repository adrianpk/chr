package chc

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestAcceptanceFeatureLayoutUsesCanonicalPath(t *testing.T) {
	projectDir := t.TempDir()

	paths, err := AcceptanceFeatureLayout(projectDir, QuizScope{ProjectID: "project-1", CaseSlug: "article-authoring"})
	if err != nil {
		t.Fatalf("AcceptanceFeatureLayout returned error: %v", err)
	}

	if paths.ScopeID != "article-authoring" {
		t.Fatalf("scope id = %q, want article-authoring", paths.ScopeID)
	}

	if paths.RelativePath != ".chc/user/acceptance/article-authoring.feature" {
		t.Fatalf("relative path = %q", paths.RelativePath)
	}

	if paths.Path != filepath.Join(projectDir, ".chc", "user", "acceptance", "article-authoring.feature") {
		t.Fatalf("path = %q", paths.Path)
	}
}

func TestAcceptanceFeatureLayoutRejectsNestedCaseSlug(t *testing.T) {
	_, err := AcceptanceFeatureLayout(t.TempDir(), QuizScope{ProjectID: "project-1", CaseSlug: "../bad"})
	if err == nil {
		t.Fatal("AcceptanceFeatureLayout returned nil error")
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

func TestCheckAcceptanceFeatureArtifactAcceptsValidFeature(t *testing.T) {
	feature := validAcceptanceFeatureArtifact()
	feature.Body = `@spec:article-authoring-spec @use-case:article-authoring
Feature: Article authoring

  @requirement:publish-report
  Scenario: Publish report is written
    Given an existing article draft
    When the editor publishes the article
    Then a publication report is written
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

	if paths.RelativePath != ".chc/user/acceptance/article-authoring.feature" {
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

func assertAcceptanceFeatureFinding(t *testing.T, result AcceptanceFeatureCheckResult, code string) {
	t.Helper()

	if result.Valid {
		t.Fatalf("valid = true, want finding %q", code)
	}

	if !slices.ContainsFunc(result.Findings, func(finding AcceptanceFeatureCheckFinding) bool {
		return finding.Code == code
	}) {
		t.Fatalf("findings = %#v, want %q", result.Findings, code)
	}
}

func validAcceptanceFeatureArtifact() AcceptanceFeatureArtifact {
	return AcceptanceFeatureArtifact{
		Scope:       QuizScope{ProjectID: "project-1", CaseSlug: "article-authoring"},
		SpecSlug:    "article-authoring-spec",
		SpecPath:    ".chc/user/specs/article-authoring-spec.md",
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
`
}
