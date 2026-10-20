package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// Pi performs real readiness and plan calls before the tracking publication owner.
func TestPiGenerationPlan(t *testing.T) {
	plan := serviceQuizPlanResponse("fixture", time.Now().UTC()).Plan
	payload := core.QuizModelPlanPayload{Title: plan.Title, Body: plan.Body, Status: "planned", Order: plan.Order}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	operations, driver := piGenerationFixture(t, string(body))
	project := operations.projectContext.ProjectDir
	writeServiceQuizFixture(t, project)
	harnessSelectionWrite(t, filepath.Join(driver, "plan-readiness"), "enabled")

	generated, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{OperationID: "pi-plan", Scope: core.QuizScope{CaseSlug: "blog"}})
	if err != nil {
		t.Fatal("Pi plan failed", err)
	}

	if _, err := os.Stat(filepath.Join(project, generated.Plan.RelativePath)); err != nil {
		t.Fatal(err)
	}

	writePiTrackingResponse(t, driver)

	tracking, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{OperationID: "pi-tracking", Scope: core.QuizScope{CaseSlug: "blog"}})
	if err != nil {
		t.Fatal("Pi tracking failed", err)
	}

	if _, err := os.Stat(filepath.Join(project, tracking.TrackingPath)); err != nil {
		t.Fatal(err)
	}

	checkPiGenerationMetadata(t, project, 2, 0)

	if len(piGenerationCalls(t, driver)) != 2 {
		t.Fatal("retained plan tracking started another model call")
	}
}

// Legacy tracking still dispatches through the selected Pi generation role.
func TestPiGenerationTracking(t *testing.T) {
	operations, driver := piGenerationFixture(t, `{}`)
	project := operations.projectContext.ProjectDir
	writeServiceQuizFixture(t, project)
	writeServiceWorkspaceFile(t, filepath.Join(project, ".chr", "default", "plans", "blog-plan.md"), "---\ntitle: Blog Plan\nstatus: planned\ncase_slug: blog\n---\n\n# Blog Plan\n\n## Slice 1\n\nImplement planning foundation.\n")

	path := canonicalServiceWorkspaceFixturePath(filepath.Join(project, ".chr", "default", "plans", "blog-plan.md"))

	relative, err := filepath.Rel(project, path)
	if err != nil {
		t.Fatal(err)
	}

	if err := core.EnsureOpsArtifactBodyMetadata(project, filepath.ToSlash(relative)); err != nil {
		t.Fatal(err)
	}

	writePiTrackingResponse(t, driver)

	result, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{OperationID: "pi-legacy-tracking", Scope: core.QuizScope{CaseSlug: "blog"}})
	if err != nil || len(result.Tickets) != 1 {
		t.Fatal("Pi legacy tracking failed", result.Message, err)
	}

	checkPiGenerationMetadata(t, project, 1, 0)
}

func writePiTrackingResponse(t *testing.T, driver string) {
	t.Helper()

	tracking := serviceQuizPlanTrackingResponse("fixture", time.Now().UTC()).PlanTracking
	ticket := tracking.Tickets[0]
	payload := core.QuizModelPlanTrackingPayload{Title: tracking.Title, Body: tracking.Body, Status: "tracking", Tickets: []core.QuizModelPlanTrackingTicketPayload{{
		SliceID: ticket.SliceID, Title: ticket.Title, Synopsis: ticket.Synopsis, Summary: ticket.Summary, Body: ticket.Body, Priority: ticket.Priority, Severity: ticket.Severity, Type: ticket.Type, Tags: ticket.Tags, Dependencies: []string{},
	}}}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(driver, "result.json"), string(data))
}
