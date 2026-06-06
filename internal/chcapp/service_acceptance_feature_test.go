package chcapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestServiceGenerateAcceptanceFeatureWritesGeneratedFeature(t *testing.T) {
	projectDir := t.TempDir()
	writeAcceptanceFeatureServiceFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), chc.AcceptanceFeatureGeneration{
				Status: chc.AcceptanceFeatureStatusGenerated,
				Body: `@spec:blog-spec @use-case:blog
Feature: Blog publishing

  Scenario: Editorial review is preserved
    Given an article draft has editorial review status
    When the article is published
    Then the published article keeps the editorial review status
`,
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GenerateAcceptanceFeature(t.Context(), AcceptanceFeatureGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	if result.Status != chc.AcceptanceFeatureStatusGenerated {
		t.Fatalf("Status = %q, want generated", result.Status)
	}

	if result.Feature != ".chc/user/acceptance/blog.feature" {
		t.Fatalf("Feature = %q, want canonical feature path", result.Feature)
	}

	feature := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "acceptance", "blog.feature"))
	if !strings.Contains(feature, "Feature: Blog publishing") {
		t.Fatalf("feature missing body:\n%s", feature)
	}

	request := model.requests[len(model.requests)-1]
	if request.Action != chc.QuizModelActionGenerateAcceptanceFeature {
		t.Fatalf("Action = %q, want acceptance feature generation", request.Action)
	}

	assertQuizModelRequestSections(t, request, []string{
		"project",
		"quiz",
		"recent_entries",
		"current_question",
		"agent_protocol",
		"project_document",
		"use_case_document",
		"dump",
		"quiz",
		"spec",
	})
}

func TestServiceGenerateAcceptanceFeatureRejectsInvalidGeneratedFeature(t *testing.T) {
	projectDir := t.TempDir()
	writeAcceptanceFeatureServiceFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), chc.AcceptanceFeatureGeneration{
				Status: chc.AcceptanceFeatureStatusGenerated,
				Body: `@spec:blog-spec @use-case:blog
Feature: Blog publishing

  Scenario: Editorial review is preserved
    Given an article draft exists
    When the editor publishes the article
    Then it works correctly
`,
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GenerateAcceptanceFeature(t.Context(), AcceptanceFeatureGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	if result.Status != chc.AcceptanceFeatureStatusRejected {
		t.Fatalf("Status = %q, want rejected", result.Status)
	}

	if _, err := os.Stat(filepath.Join(projectDir, ".chc", "user", "acceptance", "blog.feature")); !os.IsNotExist(err) {
		t.Fatalf("feature file exists or stat failed with unexpected error: %v", err)
	}
}

func TestServiceGenerateAcceptanceFeatureReportsNeedsMoreInfo(t *testing.T) {
	projectDir := t.TempDir()
	writeAcceptanceFeatureServiceFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), chc.AcceptanceFeatureGeneration{
				Status:          chc.AcceptanceFeatureStatusNeedsMoreInfo,
				RejectionReason: "Missing observable failure path.",
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GenerateAcceptanceFeature(t.Context(), AcceptanceFeatureGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	if result.Status != chc.AcceptanceFeatureStatusNeedsMoreInfo {
		t.Fatalf("Status = %q, want needs_more_info", result.Status)
	}

	if !strings.Contains(result.RejectionReason, "failure path") {
		t.Fatalf("RejectionReason = %q", result.RejectionReason)
	}
}

func TestServiceGenerateAcceptanceFeatureReportsNotApplicable(t *testing.T) {
	projectDir := t.TempDir()
	writeAcceptanceFeatureServiceFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), chc.AcceptanceFeatureGeneration{
				Status:          chc.AcceptanceFeatureStatusNotApplicable,
				RejectionReason: "The topic is editorial guidance, not observable behavior.",
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GenerateAcceptanceFeature(t.Context(), AcceptanceFeatureGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	if result.Status != chc.AcceptanceFeatureStatusNotApplicable {
		t.Fatalf("Status = %q, want not_applicable", result.Status)
	}
}

func writeAcceptanceFeatureServiceFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeServiceQuizFixture(t, projectDir)
	writeSpecGenerationUseCaseFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case_slug: "blog"
---

# Blog Spec

Publishing must preserve editorial review status.
`)
}
