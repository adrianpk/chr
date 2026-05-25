package chcapp

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

//nolint:funlen // Integration-style service assertion keeps fixture and checks together.
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

func TestServiceGenerateSpecExcludesArchivedDocumentsFromContextAndProvenance(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "specs", "archive", "blog-spec-20260525.md"), `---
title: "Archived Blog Spec"
status: "draft"
case: "blog"
---

# Archived Blog Spec
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case: "blog"
---

# Blog Plan
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
			serviceQuizDraftSpecResponse("response-spec", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.GenerateSpec(SpecGenerateRequest{
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

	spec := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "specs", "blog-spec.md"))
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

	_, err := operations.GenerateSpec(SpecGenerateRequest{
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
