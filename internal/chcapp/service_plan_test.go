package chcapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestServiceRegeneratePlanArchivesExistingPlanBeforeReplace(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: "Existing Blog Plan"
status: "planned"
case_slug: "blog"
---

# Existing Blog Plan

Manual edits that must remain recoverable.
`)

	clock := fixedServiceClock(time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC))
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizPlanResponse("response-plan", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.RegeneratePlan(PlanRegenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("RegeneratePlan returned error: %v", err)
	}

	if result.Message != "Plan regenerated." {
		t.Fatalf("Message = %q, want Plan regenerated.", result.Message)
	}

	archivePath := filepath.Join(projectDir, ".chc", "plans", "archive", "blog-plan-20260525-111213000000014.md")
	archivedPlan := readServiceWorkspaceFile(t, archivePath)
	for _, want := range []string{
		`title: "Existing Blog Plan"`,
		`archived_from: ".chc/plans/blog-plan.md"`,
		`archived_at: "2026-05-25T11:12:13.000000014Z"`,
		"Manual edits that must remain recoverable.",
	} {
		if !strings.Contains(archivedPlan, want) {
			t.Fatalf("archived plan missing %q in:\n%s", want, archivedPlan)
		}
	}

	canonicalPlan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"))
	if strings.Contains(canonicalPlan, "Manual edits that must remain recoverable.") {
		t.Fatalf("canonical plan still contains old content:\n%s", canonicalPlan)
	}

	if !strings.Contains(canonicalPlan, "# Generated Implementation Plan") {
		t.Fatalf("canonical plan missing regenerated content:\n%s", canonicalPlan)
	}
}

func TestServiceRegeneratePlanWithoutExistingPlanCreatesCanonicalPlan(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := fixedServiceClock(time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC))
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizPlanResponse("response-plan", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.RegeneratePlan(PlanRegenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("RegeneratePlan returned error: %v", err)
	}

	plan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"))
	if !strings.Contains(plan, "# Generated Implementation Plan") {
		t.Fatalf("canonical plan missing generated content:\n%s", plan)
	}

	archiveRoot := filepath.Join(projectDir, ".chc", "plans", "archive")
	if _, err := os.Stat(archiveRoot); !os.IsNotExist(err) {
		t.Fatalf("archive root exists after missing-plan regenerate: %v", err)
	}
}

func TestServiceRegeneratePlanPreservesCanonicalPlanWhenArchiveFails(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: "Existing Blog Plan"
status: "planned"
case_slug: "blog"
---

# Existing Blog Plan

Manual edits that must remain recoverable.
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "archive"), "not a directory\n")

	clock := fixedServiceClock(time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC))
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizPlanResponse("response-plan", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.RegeneratePlan(PlanRegenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err == nil {
		t.Fatal("RegeneratePlan returned nil error, want archive failure")
	}

	canonicalPlan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"))
	if !strings.Contains(canonicalPlan, "Manual edits that must remain recoverable.") {
		t.Fatalf("canonical plan was not preserved:\n%s", canonicalPlan)
	}

	if len(model.requests) != 0 {
		t.Fatalf("model requests = %d, want no model invocation after archive failure", len(model.requests))
	}
}

func TestServiceGeneratePlanExcludesArchivedDocumentsFromContextAndProvenance(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "specs", "archive", "blog-spec-20260525.md"), `---
title: "Archived Blog Spec"
status: "draft"
case: "blog"
---

# Archived Blog Spec
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "archive", "blog-plan-20260525.md"), `---
title: "Archived Blog Plan"
status: "planned"
case: "blog"
---

# Archived Blog Plan
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

	_, err := operations.GeneratePlan(PlanGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GeneratePlan returned error: %v", err)
	}

	request := model.requests[len(model.requests)-1]
	for _, section := range request.ContextSections {
		if strings.Contains(section.Authority, "/archive/") {
			t.Fatalf("context section includes archived authority: %#v", section)
		}
	}

	plan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"))
	if strings.Contains(plan, "/archive/") {
		t.Fatalf("generated provenance includes archived path:\n%s", plan)
	}
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

func fixedServiceClock(now time.Time) func() time.Time {
	return func() time.Time {
		return now
	}
}
