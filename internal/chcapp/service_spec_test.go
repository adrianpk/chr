package chcapp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestServiceGenerateSpecUsesUseCaseContext(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeSpecGenerationUseCaseFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizDraftSpecResponse("response-spec", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateSpec returned error: %v", err)
	}

	assertGeneratedSpecResult(t, projectDir, result)

	request := model.requests[len(model.requests)-1]
	if request.Action != chc.QuizModelActionGenerateSpec {
		t.Fatalf("Action = %q, want spec generation", request.Action)
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
	})
	assertSpecGenerationContext(t, request)
}

func TestServiceGenerateSpecDoesNotWriteGeneratedFeature(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeSpecGenerationUseCaseFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizDraftSpecResponseWithFeature("response-spec", clock(), chc.AcceptanceFeatureGeneration{
				Status: chc.AcceptanceFeatureStatusGenerated,
				Body: `@spec:blog-spec @use-case:blog
Feature: Blog publishing

  Scenario: Editorial review is preserved
    Given an article draft exists
    When the editor publishes the article
    Then the editorial review status is preserved
`,
			}),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateSpec returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectDir, ".chc", "user", "acceptance", "blog.feature")); !os.IsNotExist(err) {
		t.Fatalf("feature file exists or stat failed with unexpected error: %v", err)
	}

	assertGeneratedSpecResult(t, projectDir, result)
}

func TestServiceGenerateSpecRejectsEmbeddedGherkinBody(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeSpecGenerationUseCaseFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizDraftSpecResponseWithBody("response-spec", clock(), "```gherkin\nFeature: Blog publishing\n```"),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err == nil {
		t.Fatal("GenerateSpec returned nil error")
	}
}

func writeSpecGenerationUseCaseFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeSpecGenerationQuizFixture(t, projectDir)
	writeSpecGenerationExcludedDocumentFixture(t, projectDir)
}

func writeSpecGenerationQuizFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "quizzes", "blog-quiz.md"), `---
status: "curated"
case_slug: "blog"
raw_path: ".chc/user/quizzes/blog-quiz-raw.md"
---

# Quiz Notes

## Decisions

- Preserve editorial review.
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "quizzes", "project-wide-quiz.md"), `---
status: "curated"
raw_path: ".chc/user/quizzes/project-wide-quiz-raw.md"
---

# Project-wide Quiz Notes

Project-wide quiz context that must not feed scoped spec generation.
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "quizzes", "blog-quiz-raw.md"), `---
status: "raw"
case_slug: "blog"
curated_path: ".chc/user/quizzes/blog-quiz.md"
---

# Raw Quiz Notes

Raw quiz turn that must not feed spec context.
`)
}

func writeSpecGenerationExcludedDocumentFixture(t *testing.T, projectDir string) {
	t.Helper()

	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case: "blog"
---

# Blog Plan
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "specs", "blog-prior-spec.md"), `---
title: "Prior Blog Spec"
status: "draft"
case: "blog"
---

# Prior Blog Spec

Prior spec content that must not feed initial spec generation.
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "dumps", "blog-context-raw.md"), `---
title: "Blog Context Raw Capture"
status: "raw"
type: "dump_raw"
curated: ".chc/user/dumps/blog-context.md"
case: "blog"
---

# Blog Context Raw Capture

Raw sibling body that must not feed spec context.
`)
}

func assertGeneratedSpecResult(t *testing.T, projectDir string, result SpecOperationResult) {
	t.Helper()

	if result.Message != "Spec generated." {
		t.Fatalf("Message = %q, want Spec generated.", result.Message)
	}

	spec := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "specs", "blog-spec.md"))
	for _, want := range []string{
		`status: "draft"`,
		`case_slug: "blog"`,
		"# Quiz Generated Spec",
		"Generated behavior from dump and quiz context.",
	} {
		if !strings.Contains(spec, want) {
			t.Fatalf("generated spec missing %q in:\n%s", want, spec)
		}
	}
}

func assertSpecGenerationContext(t *testing.T, request chc.QuizModelRequest) {
	t.Helper()

	for _, section := range request.ContextSections {
		if strings.Contains(section.Content, "Raw sibling body that must not feed spec context.") {
			t.Fatalf("spec context included raw dump sibling in section %#v", section)
		}

		if strings.Contains(section.Content, "Raw quiz turn that must not feed spec context.") {
			t.Fatalf("spec context included raw quiz sibling in section %#v", section)
		}

		if strings.Contains(section.Content, "Project-wide quiz context that must not feed scoped spec generation.") {
			t.Fatalf("spec context included project-wide quiz in scoped generation section %#v", section)
		}

		if strings.Contains(section.Authority, ".chc/user/plans/blog-plan.md") {
			t.Fatalf("spec context included plan section: %#v", section)
		}

		if strings.Contains(section.Content, "Prior spec content that must not feed initial spec generation.") {
			t.Fatalf("spec context included prior spec section: %#v", section)
		}
	}
}

func TestServiceGenerateSpecExcludesArchivedDocumentsFromContextAndProvenance(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "specs", "archive", "blog-spec-20260525.md"), `---
title: "Archived Blog Spec"
status: "draft"
case: "blog"
---

# Archived Blog Spec
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case: "blog"
---

# Blog Plan
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "plans", "archive", "blog-plan-20260525.md"), `---
title: "Archived Blog Plan"
status: "planned"
case: "blog"
---

# Archived Blog Plan
`)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizDraftSpecResponse("response-spec", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateSpec returned error: %v", err)
	}

	request := model.requests[len(model.requests)-1]
	for _, section := range request.ContextSections {
		if strings.Contains(section.Authority, "/archive/") {
			t.Fatalf("context section includes archived authority: %#v", section)
		}
	}

	spec := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "specs", "blog-spec.md"))
	if strings.Contains(spec, "/archive/") {
		t.Fatalf("generated provenance includes archived path:\n%s", spec)
	}
}

func TestServiceGenerateSpecUsesLongGenerationDeadline(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	model := &deadlineCapturingSpecModel{}
	operations := newTestService(t, projectDir)
	operations.quizModel = model

	_, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateSpec returned error: %v", err)
	}

	if !model.hasDeadline {
		t.Fatal("model context has no deadline, want long generation deadline")
	}

	if model.remaining < 4*time.Minute {
		t.Fatalf("generation deadline remaining = %s, want long generation timeout", model.remaining)
	}
}

func TestServiceListSpecsReturnsScopes(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "user", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec

Generated behavior.
`)

	operations := newTestService(t, projectDir)

	result, err := operations.ListSpecs()
	if err != nil {
		t.Fatalf("ListSpecs returned error: %v", err)
	}

	if len(result.Specs) != 2 {
		t.Fatalf("len(Specs) = %d, want project and blog scopes", len(result.Specs))
	}

	if result.Specs[1].Scope.CaseSlug != "blog" || result.Specs[1].Status != "draft" {
		t.Fatalf("blog spec = %#v, want draft blog scope", result.Specs[1])
	}

	if result.Specs[1].UseCaseTitle != "Blog Publishing" || result.Specs[1].Title != "Blog Spec" {
		t.Fatalf("blog spec titles = %#v, want use case and spec titles split", result.Specs[1])
	}
}

type deadlineCapturingSpecModel struct {
	hasDeadline bool
	remaining   time.Duration
}

func (m *deadlineCapturingSpecModel) InvokeQuizModel(
	ctx context.Context,
	request chc.QuizModelRequest,
) (chc.QuizModelResponse, error) {
	deadline, ok := ctx.Deadline()
	m.hasDeadline = ok
	m.remaining = time.Until(deadline)

	return serviceQuizDraftSpecResponse("response-spec", request.CreatedAt), nil
}
