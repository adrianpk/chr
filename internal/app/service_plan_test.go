//nolint:wsl_v5 // Service tests keep setup and assertions close to the behavior under test.
package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	"forge.adrianpk.com/chr/internal/tui"
)

//nolint:funlen // Integration-style service assertion keeps fixture and checks together.
func TestServiceGeneratePlanUsesUseCaseContext(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")

	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "quizzes", "blog-quiz.md"), `---
status: "curated"
case_slug: "blog"
---

# Quiz Notes

## Decisions

- Preserve editorial review.
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec

The implementation must preserve editorial review.
`)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: serviceQuizStagedPlanResponses(clock()),
	}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GeneratePlan returned error: %v", err)
	}

	if result.Message != "Plan generated." {
		t.Fatalf("Message = %q, want Plan generated.", result.Message)
	}

	plan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"))
	for _, want := range []string{
		`status: "planned"`,
		`case_slug: "blog"`,
		"# Generated Implementation Plan",
		"### Slice 1: Foundation",
		"## Model call provenance",
	} {
		if !strings.Contains(plan, want) {
			t.Fatalf("generated plan missing %q in:\n%s", want, plan)
		}
	}

	if len(model.requests) != 2 {
		t.Fatalf("model requests = %d, want outline and slice detail requests", len(model.requests))
	}

	if model.requests[0].Action != quizmodel.ActionGeneratePlanOutline {
		t.Fatalf("first action = %q, want plan outline generation", model.requests[0].Action)
	}

	if model.requests[1].Action != quizmodel.ActionGeneratePlanSliceDetail {
		t.Fatalf("second action = %q, want plan slice detail generation", model.requests[1].Action)
	}

	assertQuizModelRequestSections(t, model.requests[0], []string{
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

func TestServiceGeneratePlanCanUseSinglePassCodexBackend(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "codex")

	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec

The implementation must preserve editorial review.
`)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: []quizmodel.Response{
			serviceQuizPlanResponse("response-plan", clock()),
		},
	}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GeneratePlan returned error: %v", err)
	}

	if result.Message != "Plan generated." {
		t.Fatalf("Message = %q, want Plan generated.", result.Message)
	}

	if len(model.requests) != 1 {
		t.Fatalf("model requests = %d, want single plan generation request", len(model.requests))
	}

	if model.requests[0].Action != quizmodel.ActionGeneratePlan {
		t.Fatalf("action = %q, want plan generation", model.requests[0].Action)
	}

	plan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"))
	if !strings.Contains(plan, "Implement the generated plan in one pass.") {
		t.Fatalf("generated plan missing single-pass body in:\n%s", plan)
	}
}

func TestServiceBuildsPlanOutlineGenerationRequest(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec

The implementation must preserve editorial review.
`)

	operations := newPlanTestService(t, projectDir)
	scope := quizcore.Scope{ProjectID: operations.projectContext.ProjectID, CaseSlug: "blog"}
	state, err := operations.specGenerationRuntimeState(t.Context(), scope)
	if err != nil {
		t.Fatalf("specGenerationRuntimeState returned error: %v", err)
	}

	request, err := operations.newPlanOutlineGenerationModelRequest(t.Context(), state, planTestSources(t, operations, scope))
	if err != nil {
		t.Fatalf("newPlanOutlineGenerationModelRequest returned error: %v", err)
	}

	if request.Action != quizmodel.ActionGeneratePlanOutline {
		t.Fatalf("Action = %q, want plan outline generation", request.Action)
	}

	if request.Strategy != "plan_outline_generation" {
		t.Fatalf("Strategy = %q, want plan_outline_generation", request.Strategy)
	}

	if request.ID != "plan-outline-generation-blog" {
		t.Fatalf("ID = %q, want plan-outline-generation-blog", request.ID)
	}

	if request.Payload.OutputContract.RenderedTextField != "summary" {
		t.Fatalf("RenderedTextField = %q, want summary", request.Payload.OutputContract.RenderedTextField)
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
		"spec",
	})
}

func TestServiceBuildsPlanGenerationModelRequest(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec

The implementation must preserve editorial review.
`)

	operations := newPlanTestService(t, projectDir)
	scope := quizcore.Scope{ProjectID: operations.projectContext.ProjectID, CaseSlug: "blog"}
	state, err := operations.specGenerationRuntimeState(t.Context(), scope)
	if err != nil {
		t.Fatalf("specGenerationRuntimeState returned error: %v", err)
	}

	request, err := operations.newPlanGenerationModelRequest(t.Context(), state, planTestSources(t, operations, scope))
	if err != nil {
		t.Fatalf("newPlanGenerationModelRequest returned error: %v", err)
	}

	if request.Action != quizmodel.ActionGeneratePlan {
		t.Fatalf("Action = %q, want plan generation", request.Action)
	}

	if request.ID != "plan-generation-blog" {
		t.Fatalf("ID = %q, want plan-generation-blog", request.ID)
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
		"spec",
	})
}

func TestServiceBuildsPlanSliceDetailGenerationRequest(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec

The implementation must preserve editorial review.
`)

	operations := newPlanTestService(t, projectDir)
	scope := quizcore.Scope{ProjectID: operations.projectContext.ProjectID, CaseSlug: "blog"}
	state, err := operations.specGenerationRuntimeState(t.Context(), scope)
	if err != nil {
		t.Fatalf("specGenerationRuntimeState returned error: %v", err)
	}

	outline := validServiceGeneratedPlanOutline(scope, operations.now())
	request, err := operations.newPlanSliceDetailGenerationModelRequest(t.Context(), state, planTestSources(t, operations, scope), outline, "Slice 1")
	if err != nil {
		t.Fatalf("newPlanSliceDetailGenerationModelRequest returned error: %v", err)
	}

	if request.Action != quizmodel.ActionGeneratePlanSliceDetail {
		t.Fatalf("Action = %q, want plan slice detail generation", request.Action)
	}

	if request.Strategy != "plan_slice_detail_generation" {
		t.Fatalf("Strategy = %q, want plan_slice_detail_generation", request.Strategy)
	}

	if request.ID != "plan-slice-detail-generation-blog-slice-1" {
		t.Fatalf("ID = %q, want plan-slice-detail-generation-blog-slice-1", request.ID)
	}

	if request.Payload.OutputContract.RenderedTextField != "objective" {
		t.Fatalf("RenderedTextField = %q, want objective", request.Payload.OutputContract.RenderedTextField)
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
		"spec",
		"plan_outline",
		"target_slice",
	})
}

func TestServiceRegeneratePlanArchivesWithSuccessfulReplace(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")

	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"), `---
title: "Existing Blog Plan"
status: "planned"
case_slug: "blog"
---

# Existing Blog Plan

Manual edits that must remain recoverable.
`)

	clock := fixedServiceClock(time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC))
	model := &fakeServiceQuizModel{
		responses: serviceQuizStagedPlanResponses(clock()),
	}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	result, err := operations.RegeneratePlan(t.Context(), PlanRegenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("RegeneratePlan returned error: %v", err)
	}

	if result.Message != "Plan regenerated." {
		t.Fatalf("Message = %q, want Plan regenerated.", result.Message)
	}

	archivePath := filepath.Join(projectDir, filepath.FromSlash("ops/default/plan/00000-blog-plan-20260525-111213000000014/artifact.md"))
	archivedPlan := readServiceWorkspaceFile(t, archivePath)
	for _, want := range []string{
		`title: "Existing Blog Plan"`,
		`archived_from: "ops/default/plan/00000-blog-plan/artifact.md"`,
		`archived_at: "2026-05-25T11:12:13.000000014Z"`,
		"Manual edits that must remain recoverable.",
	} {
		if !strings.Contains(archivedPlan, want) {
			t.Fatalf("archived plan missing %q in:\n%s", want, archivedPlan)
		}
	}

	canonicalPlan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"))
	if strings.Contains(canonicalPlan, "Manual edits that must remain recoverable.") {
		t.Fatalf("canonical plan still contains old content:\n%s", canonicalPlan)
	}

	if !strings.Contains(canonicalPlan, "# Generated Implementation Plan") {
		t.Fatalf("canonical plan missing regenerated content:\n%s", canonicalPlan)
	}
}

func TestServiceRegeneratePlanWithoutExistingPlanCreatesCanonicalPlan(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")

	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := fixedServiceClock(time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC))
	model := &fakeServiceQuizModel{
		responses: serviceQuizStagedPlanResponses(clock()),
	}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.RegeneratePlan(t.Context(), PlanRegenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("RegeneratePlan returned error: %v", err)
	}

	plan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"))
	if !strings.Contains(plan, "# Generated Implementation Plan") {
		t.Fatalf("canonical plan missing generated content:\n%s", plan)
	}

	archiveRoot := filepath.Join(projectDir, ".chr", "default", "plans", "archive")
	if _, err := os.Stat(archiveRoot); !os.IsNotExist(err) {
		t.Fatalf("archive root exists after missing-plan regenerate: %v", err)
	}
}

func TestServiceGeneratePlanDoesNotWritePlanWhenSliceDetailFails(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")

	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	clock := serviceQuizClock()
	badDetail := serviceQuizPlanSliceDetailResponse("response-plan-slice-detail", clock())
	badDetail.SliceDetail.ShortName = "renamed slice"

	model := &fakeServiceQuizModel{
		responses: []quizmodel.Response{
			serviceQuizPlanOutlineResponse("response-plan-outline", clock()),
			badDetail,
		},
	}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err == nil {
		t.Fatal("GeneratePlan returned nil error, want slice detail validation failure")
	}

	planPath := filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md")
	if _, statErr := os.Stat(planPath); !os.IsNotExist(statErr) {
		t.Fatalf("plan path stat error = %v, want not exist", statErr)
	}
}

//nolint:cyclop,funlen,wsl_v5 // End-to-end materialization assertions are intentionally grouped.
func TestServiceMaterializePlanTrackingCreatesTrackingAndTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"), `---
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
		responses: []quizmodel.Response{
			serviceQuizPlanTrackingResponse("response-tracking", now),
		},
	}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model
	operations.now = fixedServiceClock(now)

	result, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("MaterializePlanTracking returned error: %v", err)
	}

	if result.Message != "Created plan tracking and 1 tickets." {
		t.Fatalf("Message = %q, want created tracking message", result.Message)
	}

	if result.TrackingPath != "ops/default/tracker/00000-blog-tracking/artifact.md" {
		t.Fatalf("TrackingPath = %q, want blog tracking path", result.TrackingPath)
	}

	tracking := readServiceWorkspaceFile(t, filepath.Join(projectDir, filepath.FromSlash("ops/default/tracker/00000-blog-tracking/artifact.md")))
	for _, want := range []string{
		`plan_path: "ops/default/plan/00000-blog-plan/artifact.md"`,
		"Track the generated plan slices through tickets.",
		"Slice 1: CHR-1 - Implement planning foundation",
	} {
		if !strings.Contains(tracking, want) {
			t.Fatalf("tracking missing %q in:\n%s", want, tracking)
		}
	}

	if len(result.Tickets) != 1 {
		t.Fatalf("Tickets = %#v, want one ticket", result.Tickets)
	}

	assertMaterializedPlanTrackingTicketMetadata(t, result.Tickets[0])

	ticket := readServiceWorkspaceFile(
		t,
		filepath.Join(projectDir, ".chr", "default", "tickets", "blog", "CHR-1-Implement-planning-foundation.md"),
	)
	assertMaterializedPlanTrackingTicketContent(t, ticket)

	if len(model.requests) != 1 || model.requests[0].Action != quizmodel.ActionMaterializePlanTracking {
		t.Fatalf("model requests = %#v, want plan tracking materialization", model.requests)
	}
}

func TestServiceMaterializePlanTrackingDoesNotWriteOnInvalidGeneratedTracking(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case_slug: "blog"
---

# Blog Plan
`)

	model := &fakeServiceQuizModel{
		responses: []quizmodel.Response{
			{
				ID:        "response-invalid-tracking",
				RequestID: "filled-by-fake",
				Provider:  "fake",
				Status:    quizmodel.ResponseReady,
				PlanTracking: generationcore.GeneratedPlanTracking{
					Title:     "Invalid Tracking",
					Body:      "Missing generated tickets.",
					CreatedAt: time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC),
				},
				CreatedAt: time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC),
			},
		},
	}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model

	_, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err == nil {
		t.Fatal("MaterializePlanTracking returned nil error")
	}

	assertServiceWorkspacePathMissing(t, projectDir, "ops/default/tracker/00000-blog-tracking/artifact.md")
	assertServiceWorkspacePathMissing(t, projectDir, ".chr/default/tickets/blog")
}

func TestServiceCreatePlanTrackingTicketsValidatesBatchBeforeWrites(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeServiceQuizFixture(t, projectDir)

	operations := newPlanTestService(t, projectDir)
	payloads := []generationcore.GeneratedPlanTrackingTicket{
		validPlanTrackingTicketPayload(),
		{
			SliceID:  "Slice 2",
			Title:    "Implement second slice",
			Synopsis: "Build the second slice.",
			Body:     "## Tasks\n\n- Implement the second slice.",
			Priority: "normal",
			Severity: "normal",
			Type:     "task",
		},
	}

	_, err := operations.createPlanTrackingTickets(
		quizcore.Scope{CaseSlug: "blog"},
		"ops/default/plan/00000-blog-plan/artifact.md",
		"ops/default/tracker/00000-blog-tracking/artifact.md",
		payloads,
	)
	if err == nil || !strings.Contains(err.Error(), "ticket summary is required") {
		t.Fatalf("createPlanTrackingTickets error = %v, want missing summary error", err)
	}

	assertServiceWorkspacePathMissing(t, projectDir, ".chr/default/tickets/blog")
}

func TestServiceCreatePlanTrackingTicketsRejectsGeneratedKeyAndPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeServiceQuizFixture(t, projectDir)

	operations := newPlanTestService(t, projectDir)
	payload := validPlanTrackingTicketPayload()
	payload.Key = "LLM-99"
	payload.Path = ".chr/default/tickets/blog/LLM-99-owned-by-model.md"

	_, err := operations.createPlanTrackingTickets(
		quizcore.Scope{CaseSlug: "blog"},
		"ops/default/plan/00000-blog-plan/artifact.md",
		"ops/default/tracker/00000-blog-tracking/artifact.md",
		[]generationcore.GeneratedPlanTrackingTicket{payload},
	)
	if err == nil || !strings.Contains(err.Error(), "generated ticket key must be empty") {
		t.Fatalf("createPlanTrackingTickets error = %v, want generated key error", err)
	}

	assertServiceWorkspacePathMissing(t, projectDir, ".chr/default/tickets/blog")
}

func validPlanTrackingTicketPayload() generationcore.GeneratedPlanTrackingTicket {
	return generationcore.GeneratedPlanTrackingTicket{
		SliceID:  "Slice 1",
		Title:    "Implement planning foundation",
		Synopsis: "Build the planning foundation slice.",
		Summary:  "Implement the foundational planning work.",
		Body:     "## Tasks\n\n- Implement the planning foundation.",
		Priority: "normal",
		Severity: "normal",
		Type:     "task",
		Tags:     []string{"plan-tracking"},
	}
}

func assertMaterializedPlanTrackingTicketContent(t *testing.T, ticket string) {
	t.Helper()

	for _, want := range []string{
		`status: "ready_for_dev"`,
		`synopsis: "Build the planning foundation slice."`,
		`source_kind: "plan_tracking"`,
		`source_ref: "ops/default/plan/00000-blog-plan/artifact.md"`,
		`slice: "Slice 1"`,
		"Source plan: `ops/default/plan/00000-blog-plan/artifact.md`",
		"Tracking: `ops/default/tracker/00000-blog-tracking/artifact.md`",
		"Slice: Slice 1",
	} {
		if !strings.Contains(ticket, want) {
			t.Fatalf("ticket missing %q in:\n%s", want, ticket)
		}
	}
}

func assertMaterializedPlanTrackingTicketMetadata(t *testing.T, ticket ticketcore.Indexed) {
	t.Helper()

	if ticket.Status != ticketcore.StatusReadyForDev {
		t.Fatalf("ticket status = %q, want ready_for_dev", ticket.Status)
	}

	if ticket.Synopsis != "Build the planning foundation slice." {
		t.Fatalf("ticket synopsis = %q, want generated synopsis", ticket.Synopsis)
	}

	sourceMetadataMatches := ticket.SourceKind == "plan_tracking" &&
		ticket.SourceRef == "ops/default/plan/00000-blog-plan/artifact.md" &&
		ticket.Slice == "Slice 1"

	if !sourceMetadataMatches {
		t.Fatalf("ticket source metadata = %#v", ticket)
	}
}

func TestServiceMaterializePlanTrackingRequiresExistingPlan(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeServiceQuizFixture(t, projectDir)
	model := &fakeServiceQuizModel{}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model

	_, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
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
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case_slug: "blog"
---

# Blog Plan
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, "ops", "default", "tracker", "00000-blog-tracking", "artifact.md"), `---
title: "Blog Tracking"
status: "tracking"
case_slug: "blog"
---

# Blog Tracking
`)
	model := &fakeServiceQuizModel{}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model

	result, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
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
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"), `---
title: "Existing Blog Plan"
status: "planned"
case_slug: "blog"
---

# Existing Blog Plan

Manual edits that must remain recoverable.
`)
	archiveBlocker := filepath.Join(projectDir, filepath.FromSlash("ops/default/plan/00000-blog-plan-20260525-111213000000014"))
	if err := os.MkdirAll(filepath.Dir(archiveBlocker), 0o755); err != nil {
		t.Fatalf("MkdirAll archive blocker returned error: %v", err)
	}
	if err := os.WriteFile(archiveBlocker, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("WriteFile archive blocker returned error: %v", err)
	}

	clock := fixedServiceClock(time.Date(2026, 5, 25, 11, 12, 13, 14, time.UTC))
	model := &fakeServiceQuizModel{
		responses: serviceQuizStagedPlanResponses(clock()),
	}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.RegeneratePlan(t.Context(), PlanRegenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err == nil {
		t.Fatal("RegeneratePlan returned nil error, want archive failure")
	}

	canonicalPlan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"))
	if !strings.Contains(canonicalPlan, "Manual edits that must remain recoverable.") {
		t.Fatalf("canonical plan was not preserved:\n%s", canonicalPlan)
	}

	if len(model.requests) != 2 {
		t.Fatalf("model requests = %d, want staged output before atomic publication checks", len(model.requests))
	}
}

func TestServiceGeneratePlanExcludesArchivedDocumentsFromContextAndProvenance(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")

	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "specs", "blog-spec.md"), `---
title: "Blog Spec"
status: "draft"
case: "blog"
---

# Blog Spec
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "specs", "archive", "blog-spec-20260525.md"), `---
title: "Archived Blog Spec"
status: "draft"
case: "blog"
---

# Archived Blog Spec
`)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "archive", "blog-plan-20260525.md"), `---
title: "Archived Blog Plan"
status: "planned"
case: "blog"
---

# Archived Blog Plan
`)

	clock := serviceQuizClock()
	model := &fakeServiceQuizModel{
		responses: serviceQuizStagedPlanResponses(clock()),
	}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model
	operations.now = clock

	_, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GeneratePlan returned error: %v", err)
	}

	request := model.requests[len(model.requests)-1]
	assertQuizModelRequestSections(t, request, []string{"spec"})

	for _, section := range request.ContextSections {
		if strings.Contains(section.Authority, "/archive/") {
			t.Fatalf("context section includes archived authority: %#v", section)
		}
	}

	plan := readServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"))
	if strings.Contains(plan, "/archive/") {
		t.Fatalf("generated provenance includes archived path:\n%s", plan)
	}
}

func TestServiceGeneratePlanUsesLongGenerationDeadline(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")

	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	model := &deadlineCapturingPlanModel{}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model

	_, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
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
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"), `---
title: "Existing Blog Plan"
status: "planned"
case_slug: "blog"
---

# Existing Blog Plan

Manual plan content.
`)

	model := &fakeServiceQuizModel{}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model

	result, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("GeneratePlan returned error: %v", err)
	}

	if result.Message != "Plan already exists. Use regenerate plan to replace it." {
		t.Fatalf("Message = %q, want existing plan refusal", result.Message)
	}

	if result.Plan.RelativePath != "ops/default/plan/00000-blog-plan/artifact.md" {
		t.Fatalf("Plan.RelativePath = %q, want existing canonical path", result.Plan.RelativePath)
	}

	if len(model.requests) != 0 {
		t.Fatalf("model requests = %d, want no model invocation", len(model.requests))
	}
}

// Combined preparation still returns the canonical generated document in its
// scoped terminal catalog, alongside the newly retained native tickets.
func TestTUIPlanGeneratedCatalog(t *testing.T) {
	provider, _ := webPlanDeliveryFixture(t)
	operations := provider.service(t.Context())
	result, err := newTUIPlanActionExecutor(operations).ExecutePlanAction(tuiPlanDeliveryRequest(t, operations, "catalog-plan"))
	if err != nil {
		t.Fatal(err)
	}
	layout, err := authoringcore.GeneratedPlanLayout(operations.projectContext.ProjectDir, operations.quizScope(quizcore.Scope{CaseSlug: "named-work"}))
	if err != nil {
		t.Fatal(err)
	}
	for _, plan := range result.Plans {
		if plan.Scope.CaseSlug == "named-work" {
			if plan.RelativePath != layout.RelativePath || plan.Status == "missing" || plan.Revision == "" {
				t.Fatal("incomplete generated catalog", plan)
			}

			return
		}
	}
	t.Fatal("generated case absent from catalog", result.Plans)
}

func TestTUIPlanRefreshRetainsExistingPlan(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"), `---
title: "Existing Blog Plan"
status: "planned"
case_slug: "blog"
---

# Existing Blog Plan

Manual plan content.
`)

	model := &fakeServiceQuizModel{}
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model

	executor := newTUIPlanActionExecutor(operations)
	result, err := executor.ExecutePlanAction(tui.PlanActionRequest{
		Action: tui.PlanActionRefresh,
		Scope:  tui.QuizScope{CaseSlug: "blog"},
	})
	if err != nil {
		t.Fatalf("ExecutePlanAction returned error: %v", err)
	}

	if result.Message != "Plans refreshed." {
		t.Fatalf("Message = %q, want read-only catalog refresh", result.Message)
	}

	if len(model.requests) != 0 {
		t.Fatalf("model requests = %d, want no model invocation", len(model.requests))
	}
}

func TestServiceGeneratePlanBlocksConcurrentSameScopeGeneration(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")

	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)

	model := newBlockingPlanModel()
	operations := newPlanTestService(t, projectDir)
	operations.quizModel = model

	done := make(chan error, 1)
	go func() {
		_, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{
			Scope: quizcore.Scope{CaseSlug: "blog"},
		})
		done <- err
	}()

	<-model.started

	_, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{
		Scope: quizcore.Scope{CaseSlug: "blog"},
	})
	if err == nil || !strings.Contains(err.Error(), "plan generation already running for scope blog") {
		t.Fatalf("GeneratePlan error = %v, want guard error", err)
	}

	close(model.release)

	if err := <-done; err != nil {
		t.Fatalf("first GeneratePlan returned error: %v", err)
	}

	if model.requestCount() != 2 {
		t.Fatalf("model requests = %d, want one guarded staged invocation", model.requestCount())
	}
}

func TestServiceListPlansReturnsScopes(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceQuizFixture(t, projectDir)
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "blog-plan.md"), `---
title: "Blog Plan"
status: "planned"
case: "blog"
---

# Blog Plan

Slice decomposition.
`)

	operations := newPlanTestService(t, projectDir)

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

func validServiceGeneratedPlanOutline(scope quizcore.Scope, createdAt time.Time) generationcore.GeneratedPlanOutline {
	return generationcore.GeneratedPlanOutline{
		Scope:              scope,
		Title:              "Staged Plan Generation",
		Status:             "planned",
		Summary:            "Generate an outline first.",
		Assumptions:        []string{"Existing plans remain valid."},
		OpenQuestions:      []string{"Should staged generation be default?"},
		Risks:              []string{"More model calls."},
		ValidationStrategy: "Validate schema and request construction.",
		Slices: []generationcore.GeneratedPlanOutlineSlice{
			{
				ID:                     "Slice 1",
				ShortName:              "outline generation contract",
				Objective:              "Add outline contracts.",
				Rationale:              "Keep the slice map coherent.",
				BranchName:             "feat/slice-1-outline-generation-contract",
				PRTitle:                "feat(slice-1): outline generation contract",
				ReportPath:             ".prj/reports/slices/staged-plan-generation/slice-1-outline-generation-contract.md",
				Complexity:             "medium",
				ImplementationBoundary: "Contracts only.",
			},
		},
		CreatedAt: createdAt,
	}
}

type deadlineCapturingPlanModel struct {
	hasDeadline bool
	remaining   time.Duration
}

func (m *deadlineCapturingPlanModel) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	deadline, ok := ctx.Deadline()
	m.hasDeadline = ok
	m.remaining = time.Until(deadline)

	if request.Action == quizmodel.ActionGeneratePlanOutline {
		return serviceQuizPlanOutlineResponse("response-plan-outline", request.CreatedAt), nil
	}

	return serviceQuizPlanSliceDetailResponse("response-plan-slice-detail", request.CreatedAt), nil
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

func (m *blockingPlanModel) Invoke(_ context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	m.mu.Lock()
	m.count++
	m.mu.Unlock()

	m.once.Do(func() { close(m.started) })
	<-m.release

	if request.Action == quizmodel.ActionGeneratePlanOutline {
		return serviceQuizPlanOutlineResponse("response-plan-outline", request.CreatedAt), nil
	}

	return serviceQuizPlanSliceDetailResponse("response-plan-slice-detail", request.CreatedAt), nil
}

func (m *blockingPlanModel) requestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.count
}

func newPlanTestService(t *testing.T, project string) service {
	t.Helper()
	// Canonical source/target observations require complete native envelopes.
	for _, kind := range []string{"spec", "plan"} {
		paths, err := filepath.Glob(filepath.Join(project, "ops", "*", kind, "*", "artifact.md"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			reference, err := filepath.Rel(project, path)
			if err != nil {
				t.Fatal(err)
			}
			if err := opscore.EnsureArtifactBodyMetadata(project, filepath.ToSlash(reference)); err != nil {
				t.Fatal(err)
			}
		}
	}
	operations := newTestService(t, project)
	operations.conversationModel = sufficientPlanReadinessProvider()

	return operations
}

func planTestSources(t *testing.T, operations service, scope quizcore.Scope) authoringcore.PlanSourceObservation {
	t.Helper()
	sources, err := operations.observePlanContext(t.Context(), scope)
	if err != nil {
		t.Fatal(err)
	}

	return sources
}
