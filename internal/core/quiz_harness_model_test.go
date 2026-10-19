package core

import (
	"context"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/harness"
)

type fakeAgentRunner struct {
	Request harness.AgentRunRequest
	Result  harness.AgentRunResult
	Err     error
}

type fakeRoutingQuizModelAdapter struct {
	Request QuizModelRequest
	Called  bool
}

func (r *fakeAgentRunner) RunAgent(_ context.Context, request harness.AgentRunRequest) (harness.AgentRunResult, error) {
	r.Request = request

	result := r.Result
	result.Provider = "fixture-harness"

	return result, r.Err
}

func (a *fakeRoutingQuizModelAdapter) InvokeQuizModel(_ context.Context, request QuizModelRequest) (QuizModelResponse, error) {
	a.Request = request
	a.Called = true

	return QuizModelResponse{
		ID:        "response",
		RequestID: request.ID,
		Provider:  "fake",
		Status:    QuizModelResponseReady,
		Question:  validQuizQuestion(),
		CreatedAt: request.CreatedAt,
	}, nil
}

func TestHarnessQuizModelAdapterGeneratesSpecThroughAgentRunner(t *testing.T) {
	runner := &fakeAgentRunner{
		Result: harness.AgentRunResult{
			ID:     "spec-generation-blog",
			Output: []byte(`{"title":"Blog Spec","body":"Spec body.","status":"draft"}`),
		},
	}
	request := validCodexQuizModelRequest(QuizModelActionGenerateSpec)

	response, err := (HarnessQuizModelAdapter{
		ProjectDir: t.TempDir(),
		Runner:     runner,
	}).InvokeQuizModel(context.Background(), request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.Provider != "fixture-harness" {
		t.Fatalf("Provider = %q, want fixture-harness", response.Provider)
	}

	if response.DraftSpec.Title != "Blog Spec" {
		t.Fatalf("DraftSpec.Title = %q, want Blog Spec", response.DraftSpec.Title)
	}

	if runner.Request.OutputName != "output.json" {
		t.Fatalf("OutputName = %q, want output.json", runner.Request.OutputName)
	}

	if len(runner.Request.OutputSchema) == 0 {
		t.Fatal("OutputSchema is empty")
	}

	if !strings.Contains(runner.Request.Prompt, "spec_generation backend") {
		t.Fatalf("prompt does not identify spec generation backend:\n%s", runner.Request.Prompt)
	}

	assertCodexDecisionDiscoveryPrompt(t, runner.Request.Prompt, "the generated spec")
}

func TestHarnessQuizModelAdapterGeneratesPlanThroughAgentRunner(t *testing.T) {
	runner := &fakeAgentRunner{
		Result: harness.AgentRunResult{
			ID:     "plan-generation-blog",
			Output: []byte(`{"title":"Blog Plan","body":"Plan body.","status":"planned"}`),
		},
	}
	request := validCodexQuizModelRequest(QuizModelActionGeneratePlan)

	response, err := (HarnessQuizModelAdapter{
		ProjectDir: t.TempDir(),
		Runner:     runner,
	}).InvokeQuizModel(context.Background(), request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.Provider != "fixture-harness" {
		t.Fatalf("Provider = %q, want fixture-harness", response.Provider)
	}

	if response.Plan.Title != "Blog Plan" {
		t.Fatalf("Plan.Title = %q, want Blog Plan", response.Plan.Title)
	}

	if !strings.Contains(runner.Request.Prompt, "plan_generation backend") {
		t.Fatalf("prompt does not identify plan generation backend:\n%s", runner.Request.Prompt)
	}

	assertCodexDecisionDiscoveryPrompt(t, runner.Request.Prompt, "the generated plan")
}

func TestHarnessQuizModelAdapterGeneratesPlanTrackingThroughAgentRunner(t *testing.T) {
	runner := &fakeAgentRunner{
		Result: harness.AgentRunResult{
			ID: "plan-tracking-blog",
			Output: []byte(`{
				"title":"Blog Tracking",
				"body":"Track generated tickets.",
				"status":"tracking",
				"tickets":[{
					"slice_id":"Slice 1",
					"title":"Implement planning foundation",
					"synopsis":"Build the planning foundation slice.",
					"summary":"Implement the planning foundation.",
					"body":"Ticket body.",
					"priority":"normal",
					"severity":"normal",
					"type":"task",
					"tags":["plan"],
					"dependencies":[]
				}]
			}`),
		},
	}
	request := validCodexQuizModelRequest(QuizModelActionMaterializePlanTracking)

	response, err := (HarnessQuizModelAdapter{
		ProjectDir: t.TempDir(),
		Runner:     runner,
	}).InvokeQuizModel(context.Background(), request)
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if response.Provider != "fixture-harness" {
		t.Fatalf("Provider = %q, want fixture-harness", response.Provider)
	}

	if response.PlanTracking.Title != "Blog Tracking" {
		t.Fatalf("PlanTracking.Title = %q, want Blog Tracking", response.PlanTracking.Title)
	}

	if !strings.Contains(runner.Request.Prompt, "plan_tracking_materialization backend") {
		t.Fatalf("prompt does not identify plan tracking backend:\n%s", runner.Request.Prompt)
	}

	assertCodexDecisionDiscoveryPrompt(t, runner.Request.Prompt, "the generated tracking and tickets")
	assertCodexTestBasedTrackingPrompt(t, runner.Request.Prompt)
}

func TestCodexWorkspaceDiscoveryOmitsQuestionActions(t *testing.T) {
	if got := workspaceDiscoveryInstructions(QuizModelActionNextQuestion); got != "" {
		t.Fatalf("discovery instructions = %q, want empty", got)
	}
}

func TestActionRoutingQuizModelAdapterRoutesConfiguredActions(t *testing.T) {
	specAdapter := &fakeRoutingQuizModelAdapter{}
	defaultAdapter := &fakeRoutingQuizModelAdapter{}
	adapter := ActionRoutingQuizModelAdapter{
		ActionAdapters: map[QuizModelAction]QuizModelAdapter{
			QuizModelActionGenerateSpec: specAdapter,
		},
		DefaultAdapter: defaultAdapter,
	}

	_, err := adapter.InvokeQuizModel(context.Background(), validCodexQuizModelRequest(QuizModelActionGenerateSpec))
	if err != nil {
		t.Fatalf("InvokeQuizModel returned error: %v", err)
	}

	if !specAdapter.Called {
		t.Fatal("spec adapter was not called")
	}

	_, err = adapter.InvokeQuizModel(context.Background(), validCodexQuizModelRequest(QuizModelActionNextQuestion))
	if err != nil {
		t.Fatalf("InvokeQuizModel fallback returned error: %v", err)
	}

	if !defaultAdapter.Called {
		t.Fatal("default adapter was not called")
	}
}

func validCodexQuizModelRequest(action QuizModelAction) QuizModelRequest {
	request := validQuizModelRequest()
	request.Action = action
	request.Strategy = string(action)
	request.ID = string(action) + "-blog"

	request.Payload.OutputContract = NewQuizModelOutputContract(action, request.Mode)
	request.OutputContract = QuizModelOutputContract{
		RequiredFields: request.Payload.OutputContract.RequiredFields,
	}

	if action == QuizModelActionGeneratePlan {
		request.OutputContract = QuizModelOutputContract{RequiredFields: []string{"title", "body", "status"}}
	}

	return request
}

func assertCodexDecisionDiscoveryPrompt(t *testing.T, prompt, target string) {
	t.Helper()

	for _, want := range []string{
		"read-only operations",
		"Workspace discovery:",
		"search the workspace for architecture decision records",
		"ADRs, decision records, architecture decisions, RFCs",
		"Use accepted or active decisions as high-authority constraints for " + target,
		"Do not require configured decision-document paths",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func assertCodexTestBasedTrackingPrompt(t *testing.T, prompt string) {
	t.Helper()

	for _, want := range []string{
		"reproduce or observe the defect with a failing test",
		"Prefer table-driven tests for input/output behavior",
		"property-based, fuzz, race, or benchmark tests only when",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}
