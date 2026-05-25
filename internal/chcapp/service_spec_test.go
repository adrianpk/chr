package chcapp

import (
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestServiceGenerateSpecUsesUseCaseContext(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "quizzes", "blog-quiz.md"), `---
status: "curated"
case_slug: "blog"
---

# Quiz Notes

## Decisions

- Preserve editorial review.
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case: "blog"
---

# Blog Plan
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

	result, err := operations.GenerateSpec(SpecGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GenerateSpec returned error: %v", err)
	}

	if result.Message != "Spec generated." {
		t.Fatalf("Message = %q, want Spec generated.", result.Message)
	}

	spec := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "specs", "blog-spec.md"))
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
		"plan",
	})
}

func TestServiceListSpecsReturnsScopes(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "specs", "blog-spec.md"), `---
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
