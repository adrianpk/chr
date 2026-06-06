package chc

import (
	"path/filepath"
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
