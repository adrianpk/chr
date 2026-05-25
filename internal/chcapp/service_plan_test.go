package chcapp

import (
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestServiceGeneratePlanUsesUseCaseContext(t *testing.T) {
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
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec

The implementation must preserve editorial review.
`)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizPlanResponse("response-plan", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GeneratePlan(PlanGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GeneratePlan returned error: %v", err)
	}

	if result.Message != "Plan generated." {
		t.Fatalf("Message = %q, want Plan generated.", result.Message)
	}

	plan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"))
	for _, want := range []string{
		`status: "planned"`,
		`case_slug: "blog"`,
		"# Generated Implementation Plan",
		"Slice 01 - Foundation",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("generated plan missing %q in:\n%s", want, plan)
		}
	}

	request := model.requests[len(model.requests)-1]
	if request.Action != chc.QuizModelActionGeneratePlan {
		t.Fatalf("Action = %q, want plan generation", request.Action)
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

func TestServiceListPlansReturnsScopes(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case: "blog"
---

# Blog Plan

Slice decomposition.
`)

	operations := newTestService(t, projectDir)

	result, err := operations.ListPlans()
	if err != nil {
		t.Fatalf("ListPlans returned error: %v", err)
	}

	if len(result.Plans) != 2 {
		t.Fatalf("len(Plans) = %d, want project and blog scopes", len(result.Plans))
	}

	if result.Plans[1].Scope.CaseSlug != "blog" || result.Plans[1].Status != "planned" {
		t.Fatalf("blog plan = %#v, want planned blog scope", result.Plans[1])
	}

	if result.Plans[1].UseCaseTitle != "Blog Publishing" || result.Plans[1].Title != "Blog Plan" {
		t.Fatalf("blog plan titles = %#v, want use case and plan titles split", result.Plans[1])
	}
}
