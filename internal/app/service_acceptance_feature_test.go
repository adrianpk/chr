package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chc/internal/core"
)

func TestServiceGenerateAcceptanceFeatureWritesGeneratedFeature(t *testing.T) {
	projectDir := t.TempDir()
	writeAcceptanceFeatureServiceFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), core.AcceptanceFeatureGeneration{
				Status: core.AcceptanceFeatureStatusGenerated,
				Body:   validBlogAcceptanceFeatureBody(),
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GenerateAcceptanceFeature(t.Context(), AcceptanceFeatureGenerateRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	if result.Status != core.AcceptanceFeatureStatusGenerated {
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
	if request.Action != core.QuizModelActionGenerateAcceptanceFeature {
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
		responses: []core.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), core.AcceptanceFeatureGeneration{
				Status: core.AcceptanceFeatureStatusGenerated,
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
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	if result.Status != core.AcceptanceFeatureStatusRejected {
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
		responses: []core.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), core.AcceptanceFeatureGeneration{
				Status:          core.AcceptanceFeatureStatusNeedsMoreInfo,
				RejectionReason: "Missing observable failure path.",
				ClarificationPrompts: []core.AcceptanceFeaturePrompt{
					{
						ID:            "failure-path",
						Question:      "What should happen when publishing fails?",
						DefaultAnswer: "The article remains unpublished and the user sees a failure message.",
						Reason:        "The feature needs one observable failure path.",
					},
				},
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GenerateAcceptanceFeature(t.Context(), AcceptanceFeatureGenerateRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	if result.Status != core.AcceptanceFeatureStatusNeedsMoreInfo {
		t.Fatalf("Status = %q, want needs_more_info", result.Status)
	}

	if !strings.Contains(result.RejectionReason, "failure path") {
		t.Fatalf("RejectionReason = %q", result.RejectionReason)
	}

	if result.FeatureRound != ".chc/user/acceptance/blog-feature-round.md" {
		t.Fatalf("FeatureRound = %q, want feature round path", result.FeatureRound)
	}

	round := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "acceptance", "blog-feature-round.md"))
	for _, want := range []string{
		"### failure-path",
		"- Default: The article remains unpublished and the user sees a failure message.",
		"- Answer:",
	} {
		if !strings.Contains(round, want) {
			t.Fatalf("feature round notes missing %q in:\n%s", want, round)
		}
	}
}

func TestServiceGenerateAcceptanceFeaturePreservesExistingFeatureRoundAnswers(t *testing.T) {
	projectDir := t.TempDir()
	writeAcceptanceFeatureServiceFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "acceptance", "blog-feature-round.md"), existingBlogFeatureRoundAnswer())

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), core.AcceptanceFeatureGeneration{
				Status:          core.AcceptanceFeatureStatusNeedsMoreInfo,
				RejectionReason: "Missing observable failure path.",
				ClarificationPrompts: []core.AcceptanceFeaturePrompt{
					{
						ID:            "failure-path",
						Question:      "What should happen when publishing fails?",
						DefaultAnswer: "The article remains unpublished.",
					},
					{
						ID:            "retry-path",
						Question:      "What should happen after retry?",
						DefaultAnswer: "The publish can be attempted again.",
					},
				},
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.GenerateAcceptanceFeature(t.Context(), AcceptanceFeatureGenerateRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	round := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "acceptance", "blog-feature-round.md"))
	for _, want := range []string{
		"- Answer: The article remains unpublished and the command reports the failed dependency.",
		"### retry-path",
		"- Answer:",
	} {
		if !strings.Contains(round, want) {
			t.Fatalf("feature round notes missing %q in:\n%s", want, round)
		}
	}
}

func existingBlogFeatureRoundAnswer() string {
	return `---
title: "Acceptance Feature Round"
status: "needs_more_info"
type: "acceptance_feature_round"
case: "blog"
---

# Acceptance Feature Round

## Prompts

### failure-path

- Question: What should happen when publishing fails?
- Default: The article remains unpublished.
- Answer: The article remains unpublished and the command reports the failed dependency.
`
}

func TestServiceGenerateAcceptanceFeatureReportsNotApplicable(t *testing.T) {
	projectDir := t.TempDir()
	writeAcceptanceFeatureServiceFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), core.AcceptanceFeatureGeneration{
				Status:          core.AcceptanceFeatureStatusNotApplicable,
				RejectionReason: "The topic is editorial guidance, not observable behavior.",
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GenerateAcceptanceFeature(t.Context(), AcceptanceFeatureGenerateRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	if result.Status != core.AcceptanceFeatureStatusNotApplicable {
		t.Fatalf("Status = %q, want not_applicable", result.Status)
	}
}

func TestServiceGenerateAcceptanceFeatureUsesFeatureRoundNotesAsContext(t *testing.T) {
	projectDir := t.TempDir()
	writeAcceptanceFeatureServiceFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "acceptance", "blog-feature-round.md"), `---
title: "Acceptance Feature Round"
status: "needs_more_info"
type: "acceptance_feature_round"
case: "blog"
---

# Acceptance Feature Round

## Prompts

### failure-path

- Question: What should happen when publishing fails?
- Default: The article remains unpublished.
- Answer: The article remains unpublished and the command reports the failed dependency.
`)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []core.QuizModelResponse{
			serviceQuizAcceptanceFeatureResponse("response-feature", clock(), core.AcceptanceFeatureGeneration{
				Status: core.AcceptanceFeatureStatusGenerated,
				Body: `@spec:blog-spec @use-case:blog
Feature: Blog publishing

  Scenario: Publishing dependency fails
    Given an article draft has editorial review status
    When the publish dependency fails
    Then the article remains unpublished and the command reports the failed dependency

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

	_, err := operations.GenerateAcceptanceFeature(t.Context(), AcceptanceFeatureGenerateRequest{
		Scope: core.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateAcceptanceFeature returned error: %v", err)
	}

	request := model.requests[len(model.requests)-1]
	if !quizModelRequestHasSection(request, "acceptance_feature_round") {
		t.Fatalf("request sections = %#v, want acceptance feature round context", request.ContextSections)
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

func validBlogAcceptanceFeatureBody() string {
	return `@spec:blog-spec @use-case:blog
Feature: Blog publishing

  Scenario: Editorial review is preserved
    Given an article draft has editorial review status
    When the article is published
    Then the published article keeps the editorial review status

  Scenario: Publishing dependency fails
    Given an article draft has editorial review status
    When the publishing dependency is unavailable
    Then the article remains unpublished and the editor sees the failure
`
}

func quizModelRequestHasSection(request core.QuizModelRequest, name string) bool {
	for _, section := range request.ContextSections {
		if section.Name == name {
			return true
		}
	}

	return false
}
