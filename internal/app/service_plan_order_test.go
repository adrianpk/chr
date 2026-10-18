package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

// Both configured generation paths publish a complete unbound pair. Retrying
// the caller operation returns its retained body without another model call.
func TestServicePlanOrderPublication(t *testing.T) {
	for _, backend := range []string{"codex", "openai"} {
		t.Run(backend, func(t *testing.T) {
			t.Setenv("CHR_PLAN_GENERATION_BACKEND", backend)

			project := t.TempDir()
			writeServiceQuizFixture(t, project)

			operations := newPlanTestService(t, project)
			model := &fakeServiceQuizModel{}
			operations.quizModel = model
			request := PlanGenerateRequest{OperationID: "service-operation", Scope: core.QuizScope{CaseSlug: "blog"}}

			generated, err := operations.GeneratePlan(t.Context(), request)
			if err != nil {
				t.Fatal(err)
			}

			binding, err := core.ObservePlanRevision(t.Context(), project, operations.quizScope(request.Scope))
			if err != nil || binding.Revision != 1 {
				t.Fatal(binding, err)
			}

			calls := len(model.requests)

			replay, err := operations.GeneratePlan(t.Context(), request)
			if err != nil || len(model.requests) != calls || replay.Plan.Content != generated.Plan.Content {
				t.Fatal("retry called model or lost committed body", err)
			}
		})
	}
}

// Legacy captured body-only payloads remain decodable, but new generation
// cannot report success or publish anything without structured intended order.
func TestServicePlanRequiresOrder(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "codex")

	project := t.TempDir()
	writeServiceQuizFixture(t, project)

	operations := newPlanTestService(t, project)
	response := serviceQuizPlanResponse("legacy-output", serviceQuizClock()())
	response.Plan.Order = nil
	operations.quizModel = &fakeServiceQuizModel{responses: []core.QuizModelResponse{response}}

	_, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{Scope: core.QuizScope{CaseSlug: "blog"}})
	if err == nil || !strings.Contains(err.Error(), "ordered slices") {
		t.Fatal("body-only generation succeeded", err)
	}

	if _, exists, err := operations.existingGeneratedPlan(operations.quizScope(core.QuizScope{CaseSlug: "blog"})); err != nil || exists {
		t.Fatal("invalid output was published", exists, err)
	}
}

// The legacy action cannot allocate from a newly generated revision. It returns
// the retained project-relative companion without another model invocation.
func TestServicePlanOrderApprovalBoundary(t *testing.T) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "codex")

	project := t.TempDir()
	writeServiceQuizFixture(t, project)

	operations := newPlanTestService(t, project)
	model := &fakeServiceQuizModel{}
	operations.quizModel = model

	scope := core.QuizScope{CaseSlug: "blog"}
	if _, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{Scope: scope}); err != nil {
		t.Fatal(err)
	}

	calls := len(model.requests)

	result, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{Scope: scope})
	if err != nil || len(result.Tickets) != 0 || len(model.requests) != calls || !strings.HasPrefix(result.TrackingPath, "ops/default/tracker/") {
		t.Fatal("legacy action bypassed native approval", result, err)
	}

	if _, err := os.Stat(filepath.Join(project, result.TrackingPath)); err != nil {
		t.Fatal("companion reference is not project-relative", err)
	}

	if _, err := os.Stat(filepath.Join(project, "ops", "default", "ticket")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unapproved action allocated native tickets", err)
	}
}
