package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// Readiness, outline and slice detail are one real plan operation. A switch
// during its first HTTP call affects a subsequent independent Spec, not stages.
func TestHarnessGenerationBindingPlan(t *testing.T) {
	generationBindingPlanCase(t, harnessGenerationFixture)
}

func generationBindingPlanCase(t *testing.T, fixture func(*testing.T, string) (service, string)) {
	t.Setenv("CHR_PLAN_GENERATION_BACKEND", "openai")
	t.Setenv("CHR_OPENAI_MODEL", "api-model-alpha")

	operations, _ := fixture(t, `{"title":"Bounded Spec","body":"Preserve the selected operation scope.","status":"draft"}`)
	project := operations.projectContext.ProjectDir
	writeServiceQuizFixture(t, project)
	switchGenerationFixture(t, operations, "")

	var calls atomic.Int64

	switched := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			switched <- changeGenerationFixture(r.Context(), operations, "writing")
		}

		harnessGenerationPlanHTTP(t, w, r)
	}))
	t.Cleanup(server.Close)

	api := core.OpenAIQuizModelAdapter{ProjectDir: project, APIKey: "fixture", BaseURL: server.URL, Client: server.Client()}
	operations.quizModel = api
	provider := core.NewConfiguredStructuredModel(true)
	provider.API = api
	operations.conversationModel = provider

	scope := quizcore.Scope{CaseSlug: "blog"}
	if _, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{OperationID: "freeze", Scope: scope}); err != nil {
		t.Fatal("later plan stages rebound after the switch", err)
	}

	if err := <-switched; err != nil || calls.Load() != 3 {
		t.Fatal("expected readiness, outline and one detail", calls.Load(), err)
	}

	if _, err := operations.GenerateSpec(t.Context(), SpecGenerateRequest{OperationID: "fresh", Scope: scope}); err != nil {
		t.Fatal("independent request missed current profile", err)
	}

	checkGenerationFixtureModels(t, project, map[string]string{"plan-freeze": "api-model-alpha", "spec-fresh": "fixture-model"})

	if records := generationFixtureMetadata(t, project); len(records) != 4 {
		t.Fatal("nested model calls did not retain separate call evidence", len(records))
	}
}

func harnessGenerationPlanHTTP(t *testing.T, w http.ResponseWriter, request *http.Request) {
	t.Helper()

	var input struct {
		Model string
		Input string
		Text  struct{ Format struct{ Name string } }
	}
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		t.Error(err)

		return
	}

	if input.Model != "api-model-alpha" {
		t.Error("plan stage rebound its API model", input.Model)
	}

	output := harnessGenerationPlanPayload(t, input.Text.Format.Name, input.Input)
	if err := json.NewEncoder(w).Encode(map[string]string{"output_text": output}); err != nil {
		t.Error(err)
	}
}

func harnessGenerationPlanPayload(t *testing.T, name, input string) string {
	t.Helper()

	switch name {
	case "plan_readiness":
		var payload struct {
			ObservationHash string `json:"observation_hash"`
		}
		if err := json.Unmarshal([]byte(input), &payload); err != nil {
			t.Error(err)
		}

		body, err := json.Marshal(map[string]any{"sufficient": true, "missing": []string{}, "observation_hash": payload.ObservationHash})
		if err != nil {
			t.Error(err)
		}

		return string(body)
	case "generated_plan_outline":
		return `{"title":"Bounded Plan","status":"planned","summary":"Deliver the specified foundation.","assumptions":[],"open_questions":[],"risks":[],"validation_strategy":"Run focused checks.","slices":[{"id":"Slice 1","short_name":"Foundation","objective":"Preserve the bounded scope.","rationale":"Deliver one coherent capability.","dependencies":[],"branch_name":"feat/slice-1-foundation","pr_title":"feat(slice-1): foundation","report_path":"ops/default/report/slices/example/slice-1-foundation.md","complexity":"small","implementation_boundary":"Foundation only."}],"dependencies":[]}`
	case "generated_plan_slice_detail":
		return `{"slice_id":"Slice 1","short_name":"Foundation","objective":"Preserve the bounded scope.","implementation_notes":["Keep the original operation binding."],"tasks":[{"id":"T1.1","summary":"Implement foundation.","commit_hint":"feat: implement foundation"}],"acceptance_criteria":["The scope is retained."],"ticket_metadata":{"title":"Foundation","synopsis":"Deliver the foundation.","priority":"normal","severity":"normal","type":"task","tags":["plan"]},"branch_metadata":{"branch_name":"feat/slice-1-foundation"},"pr_expectations":["One PR against dev."],"validation_expectations":["Run focused checks."],"dependencies":[],"non_goals":["Unrelated work."]}`
	default:
		t.Error("unexpected generation family", name)

		return `{}`
	}
}
