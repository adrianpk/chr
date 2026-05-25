//nolint:wsl_v5 // Service tests keep setup and assertions close to the behavior under test.
package chcapp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
	"forge.adrianpk.com/chc/internal/chctui"
)

//nolint:funlen // Integration-style service assertion keeps fixture and checks together.
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

//nolint:cyclop,funlen,wsl_v5 // End-to-end materialization assertions are intentionally grouped.
func TestServiceMaterializePlanTrackingCreatesTrackingAndTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case_slug: "blog"
---

# Blog Plan

## Slice 1

Implement planning foundation.
`)

	now := time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC)
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizPlanTrackingResponse("response-tracking", now),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = fixedServiceClock(now)

	result, err := operations.MaterializePlanTracking(PlanTrackingRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("MaterializePlanTracking returned error: %v", err)
	}

	if result.Message != "Created plan tracking and 1 tickets." {
		t.Fatalf("Message = %q, want created tracking message", result.Message)
	}

	if result.TrackingPath != ".chc/plans/blog-tracking.md" {
		t.Fatalf("TrackingPath = %q, want blog tracking path", result.TrackingPath)
	}

	tracking := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-tracking.md"))
	for _, want := range []string{
		`plan_path: ".chc/plans/blog-plan.md"`,
		"Track the generated plan slices through tickets.",
		"Slice 1: CHC-1 - Implement planning foundation",
	} {
		if !strings.Contains(tracking, want) {
			t.Fatalf("tracking missing %q in:\n%s", want, tracking)
		}
	}

	if len(result.Tickets) != 1 {
		t.Fatalf("Tickets = %#v, want one ticket", result.Tickets)
	}

	if result.Tickets[0].Status != "ready" {
		t.Fatalf("ticket status = %q, want ready", result.Tickets[0].Status)
	}

	ticket := readServiceWorkspaceFile(
		t,
		filepath.Join(projectDir, ".chc", "tickets", "blog", "CHC-1-Implement-planning-foundation.md"),
	)
	for _, want := range []string{
		`status: "ready"`,
		"Source plan: `.chc/plans/blog-plan.md`",
		"Tracking: `.chc/plans/blog-tracking.md`",
		"Slice: Slice 1",
	} {
		if !strings.Contains(ticket, want) {
			t.Fatalf("ticket missing %q in:\n%s", want, ticket)
		}
	}

	if len(model.requests) != 1 || model.requests[0].Action != chc.QuizModelActionMaterializePlanTracking {
		t.Fatalf("model requests = %#v, want plan tracking materialization", model.requests)
	}
}

func TestServiceMaterializePlanTrackingRequiresExistingPlan(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeServiceQuizFixture(t, projectDir)
	model := &fakeServiceQuizModel{}
	operations := newTestService(t, projectDir)
	operations.quizModel = model

	_, err := operations.MaterializePlanTracking(PlanTrackingRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err == nil || !strings.Contains(err.Error(), "plan is required before tracking can be created") {
		t.Fatalf("MaterializePlanTracking error = %v, want missing plan error", err)
	}

	if len(model.requests) != 0 {
		t.Fatalf("model requests = %#v, want no model request", model.requests)
	}
}

func TestServiceMaterializePlanTrackingRefusesExistingTracking(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case_slug: "blog"
---

# Blog Plan
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-tracking.md"), `---
title: "Blog Tracking"
status: "tracking"
case_slug: "blog"
---

# Blog Tracking
`)
	model := &fakeServiceQuizModel{}
	operations := newTestService(t, projectDir)
	operations.quizModel = model

	result, err := operations.MaterializePlanTracking(PlanTrackingRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("MaterializePlanTracking returned error: %v", err)
	}

	if result.Message != "Plan tracking already exists." {
		t.Fatalf("Message = %q, want existing tracking message", result.Message)
	}

	if len(model.requests) != 0 {
		t.Fatalf("model requests = %#v, want no model request", model.requests)
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

func TestServiceGeneratePlanUsesLongGenerationDeadline(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	model := &deadlineCapturingPlanModel{}
	operations := newTestService(t, projectDir)
	operations.quizModel = model

	_, err := operations.GeneratePlan(PlanGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GeneratePlan returned error: %v", err)
	}

	if !model.hasDeadline {
		t.Fatal("model context has no deadline, want long generation deadline")
	}

	if model.remaining < 4*time.Minute {
		t.Fatalf("generation deadline remaining = %s, want long generation timeout", model.remaining)
	}
}

func TestServiceGeneratePlanRefusesExistingCanonicalPlan(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: "Existing Blog Plan"
status: "planned"
case_slug: "blog"
---

# Existing Blog Plan

Manual plan content.
`)

	model := &fakeServiceQuizModel{}
	operations := newTestService(t, projectDir)
	operations.quizModel = model

	result, err := operations.GeneratePlan(PlanGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GeneratePlan returned error: %v", err)
	}

	if result.Message != "Plan already exists. Use regenerate plan to replace it." {
		t.Fatalf("Message = %q, want existing plan refusal", result.Message)
	}

	if result.Plan.RelativePath != ".chc/plans/blog-plan.md" {
		t.Fatalf("Plan.RelativePath = %q, want existing canonical path", result.Plan.RelativePath)
	}

	if len(model.requests) != 0 {
		t.Fatalf("model requests = %d, want no model invocation", len(model.requests))
	}
}

func TestTUIPlanGenerateReturnsExistingPlanForGeneratedScope(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []chc.QuizModelResponse{
			serviceQuizPlanResponse("response-plan", clock()),
		},
	}
	operations := newTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	executor := newTUIPlanActionExecutor(operations)
	result, err := executor.ExecutePlanAction(chctui.PlanActionRequest{
		Action: chctui.PlanActionGenerate,
		Scope:  chctui.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("ExecutePlanAction returned error: %v", err)
	}

	var generated chctui.PlanView
	for _, plan := range result.Plans {
		if plan.Scope.CaseSlug == "blog" {
			generated = plan

			break
		}
	}

	if generated.RelativePath != ".chc/plans/blog-plan.md" {
		t.Fatalf("generated RelativePath = %q, want canonical plan path in TUI result: %#v", generated.RelativePath, result.Plans)
	}

	if generated.Status == "missing" {
		t.Fatalf("generated Status = missing, want existing plan: %#v", generated)
	}
}

func TestTUIPlanGenerateReturnsRefusalMessageForExistingPlan(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chc", "plans", "blog-plan.md"), `---
title: "Existing Blog Plan"
status: "planned"
case_slug: "blog"
---

# Existing Blog Plan

Manual plan content.
`)

	model := &fakeServiceQuizModel{}
	operations := newTestService(t, projectDir)
	operations.quizModel = model

	executor := newTUIPlanActionExecutor(operations)
	result, err := executor.ExecutePlanAction(chctui.PlanActionRequest{
		Action: chctui.PlanActionGenerate,
		Scope:  chctui.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("ExecutePlanAction returned error: %v", err)
	}

	if result.Message != "Plan already exists. Use regenerate plan to replace it." {
		t.Fatalf("Message = %q, want existing plan refusal", result.Message)
	}

	if len(model.requests) != 0 {
		t.Fatalf("model requests = %d, want no model invocation", len(model.requests))
	}
}

func TestServiceGeneratePlanBlocksConcurrentSameScopeGeneration(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	model := newBlockingPlanModel()
	operations := newTestService(t, projectDir)
	operations.quizModel = model

	done := make(chan error, 1)
	go func() {
		_, err := operations.GeneratePlan(PlanGenerateRequest{
			Scope: chc.QuizScope{CaseSlug: "blog"},
		})
		done <- err
	}()

	<-model.started

	_, err := operations.GeneratePlan(PlanGenerateRequest{
		Scope: chc.QuizScope{CaseSlug: "blog"},
	})
	if err == nil || !strings.Contains(err.Error(), "plan generation already running for scope blog") {
		t.Fatalf("GeneratePlan error = %v, want guard error", err)
	}

	close(model.release)

	if err := <-done; err != nil {
		t.Fatalf("first GeneratePlan returned error: %v", err)
	}

	if model.requestCount() != 1 {
		t.Fatalf("model requests = %d, want one guarded invocation", model.requestCount())
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

type deadlineCapturingPlanModel struct {
	hasDeadline bool
	remaining   time.Duration
}

func (m *deadlineCapturingPlanModel) InvokeQuizModel(
	ctx context.Context,
	request chc.QuizModelRequest,
) (chc.QuizModelResponse, error) {
	deadline, ok := ctx.Deadline()
	m.hasDeadline = ok
	m.remaining = time.Until(deadline)

	return serviceQuizPlanResponse("response-plan", request.CreatedAt), nil
}

type blockingPlanModel struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	mu      sync.Mutex
	count   int
}

func newBlockingPlanModel() *blockingPlanModel {
	return &blockingPlanModel{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (m *blockingPlanModel) InvokeQuizModel(
	_ context.Context,
	request chc.QuizModelRequest,
) (chc.QuizModelResponse, error) {
	m.mu.Lock()
	m.count++
	m.mu.Unlock()

	m.once.Do(func() { close(m.started) })
	<-m.release

	return serviceQuizPlanResponse("response-plan", request.CreatedAt), nil
}

func (m *blockingPlanModel) requestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.count
}
