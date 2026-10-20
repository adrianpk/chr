package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type planReadinessTestProvider func(context.Context, core.StructuredModelRequest) (json.RawMessage, error)

func (f planReadinessTestProvider) InvokeStructuredModel(ctx context.Context, request core.StructuredModelRequest) (json.RawMessage, error) {
	return f(ctx, request)
}

func sufficientPlanReadinessProvider() planReadinessTestProvider {
	return func(_ context.Context, request core.StructuredModelRequest) (json.RawMessage, error) {
		var input struct {
			ObservationHash string `json:"observation_hash"`
		}
		if err := json.Unmarshal(request.Input, &input); err != nil {
			return nil, err
		}

		return json.Marshal(map[string]any{"sufficient": true, "missing": []string{}, "observation_hash": input.ObservationHash})
	}
}

// Provider substitution is explicit: absent discovery never calls it, while
// semantic rejection, invalid output and provider failures cannot imply ready.
func TestPlanReadiness(t *testing.T) {
	for _, item := range []struct {
		name, output, status string
		present              bool
		providerError        error
	}{
		{name: "absent", status: "absent"},
		{name: "insufficient", present: true, output: `{"sufficient":false,"missing":["Define the expected result"],"observation_hash":"bound"}`, status: "insufficient"},
		{name: "sufficient short dump", present: true, output: `{"sufficient":true,"missing":[],"observation_hash":"bound"}`, status: "sufficient"},
		{name: "provider failure", present: true, providerError: errors.New("offline")},
		{name: "missing boolean", present: true, output: `{"missing":[],"observation_hash":"bound"}`},
		{name: "missing reasons", present: true, output: `{"sufficient":false,"missing":[],"observation_hash":"bound"}`},
		{name: "contradictory", present: true, output: `{"sufficient":true,"missing":["Unknown goal"],"observation_hash":"bound"}`},
		{name: "wrong observation", present: true, output: `{"sufficient":true,"missing":[],"observation_hash":"stale"}`},
	} {
		t.Run(item.name, func(t *testing.T) {
			operations := newTestService(t, t.TempDir())
			calls := 0
			operations.conversationModel = planReadinessTestProvider(func(_ context.Context, request core.StructuredModelRequest) (json.RawMessage, error) {
				calls++

				if request.Purpose != "plan_readiness" || !strings.Contains(string(request.Input), "Print hello") {
					t.Fatal("assessment lost its selected immutable input")
				}

				return json.RawMessage(item.output), item.providerError
			})
			observed := core.PlanSourceObservation{Hash: "bound", HasDiscovery: item.present, Inputs: []core.PlanSourceInput{{Kind: "dump", Content: "Print hello. Verify stdout."}}}

			decision, err := operations.assessPlanReadiness(t.Context(), observed)
			checkPlanReadinessDecision(t, decision, err, item.status)

			expectedCalls := 0
			if item.present {
				expectedCalls = 1
			}

			if calls != expectedCalls {
				t.Fatal("incorrect provider call count", calls)
			}
		})
	}
}

func checkPlanReadinessDecision(t *testing.T, decision PlanReadiness, err error, status string) {
	t.Helper()

	if status == "sufficient" {
		if err != nil || !decision.Sufficient {
			t.Fatal("sufficient context rejected", decision, err)
		}
	} else if err == nil || decision.Sufficient {
		t.Fatal("unready context accepted", decision, err)
	}

	if status != "" && decision.Status != status {
		t.Fatal("incorrect readiness state", decision, status)
	}
}

// Readiness keeps its observation binding and the authoring deadline instead of
// imposing a shorter deadline on the configured generation provider.
func TestPlanReadinessProviderBinding(t *testing.T) {
	operations := newTestService(t, t.TempDir())
	operations.conversationModel = planReadinessTestProvider(func(ctx context.Context, request core.StructuredModelRequest) (json.RawMessage, error) {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) < longGenerationTimeout-time.Second {
			t.Fatal("readiness shortened the authoring deadline")
		}

		if request.Timeout != longGenerationTimeout {
			t.Fatal("readiness did not propagate its authoring timeout")
		}

		return sufficientPlanReadinessProvider()(ctx, request)
	})

	decision, err := operations.assessPlanReadiness(t.Context(), core.PlanSourceObservation{Hash: "binding", HasDiscovery: true})
	if err != nil || decision.ObservationHash != "binding" || !decision.Sufficient {
		t.Fatal(decision, err)
	}
}

// The real spec writer preserves prose while its saved lifecycle reaches the
// semantic boundary separately; approval cannot waive genuine missing context.
func TestPlanSpecApproval(t *testing.T) {
	for _, item := range []struct {
		name, status, prose, missing string
	}{
		{"approved stale draft", "approved", "This is a draft awaiting user review.", ""},
		{"approved missing behavior", "approved", "This is a draft awaiting user review.", "Define the required error behavior"},
		{"draft claims approval", "draft", "This specification is approved and ready for implementation.", "Spec approval is required"},
	} {
		t.Run(item.name, func(t *testing.T) {
			operations, scope, generated := planApprovalFixture(t, item.status, item.prose)
			before := readServiceWorkspaceFile(t, generated.Path)
			operations.conversationModel = specApprovalTestProvider(t, item.status, item.prose, item.missing)

			_, err := operations.GeneratePlan(t.Context(), PlanGenerateRequest{Scope: scope})
			checkPlanApprovalResult(t, err, item.missing)

			if readServiceWorkspaceFile(t, generated.Path) != before {
				t.Fatal("plan readiness rewrote saved spec content")
			}
		})
	}
}

func planApprovalFixture(t *testing.T, status, prose string) (service, core.QuizScope, core.GeneratedSpecPaths) {
	t.Helper()

	operations, scope, _ := planPublicationFixture(t, "Print hello once. Verify stdout is hello.")
	project := operations.projectContext.ProjectDir

	generated, err := core.WriteGeneratedSpec(t.Context(), project, core.GeneratedSpec{Scope: scope, Title: "Greeting", Body: prose + "\n\nPrint hello once.", CreatedAt: operations.utcNow()})
	if err != nil {
		t.Fatal(err)
	}

	document, err := core.ReadSpecDocument(t.Context(), project, generated.RelativePath)
	if err != nil {
		t.Fatal(err)
	}

	_, err = core.MutateSpec(t.Context(), project, core.SpecMutationRequest{OperationID: "save-approval", Action: "save", Reference: document.Reference, ExpectedRevision: document.Revision, ProjectID: scope.ProjectID, CaseSlug: scope.CaseSlug, Title: document.Title, Body: document.Body, Status: status})
	if err != nil {
		t.Fatal(err)
	}

	return operations, scope, generated
}

func checkPlanApprovalResult(t *testing.T, err error, missing string) {
	t.Helper()

	if missing == "" {
		if err != nil {
			t.Fatal("approved spec rejected", err)
		}

		return
	}

	var rejected *PlanReadinessError
	if !errors.As(err, &rejected) || rejected.Decision.Status != "insufficient" {
		t.Fatal("missing approval or behavior was accepted", err)
	}
}

func specApprovalTestProvider(t *testing.T, status, prose, missing string) planReadinessTestProvider {
	t.Helper()

	return func(_ context.Context, request core.StructuredModelRequest) (json.RawMessage, error) {
		hash := checkSpecApprovalInput(t, request.Input, status, prose)

		reasons := []string{}
		if missing != "" {
			reasons = append(reasons, missing)
		}

		return json.Marshal(map[string]any{"sufficient": missing == "", "missing": reasons, "observation_hash": hash})
	}
}

func checkSpecApprovalInput(t *testing.T, data []byte, status, prose string) string {
	t.Helper()

	var input struct {
		Inputs []struct{ Kind, Status, Content string }
		Hash   string `json:"observation_hash"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}

	for _, source := range input.Inputs {
		if source.Kind != "spec" {
			continue
		}

		if source.Status != status || !strings.Contains(source.Content, prose) {
			t.Fatal("readiness lost canonical lifecycle or preserved prose", source.Status)
		}

		return input.Hash
	}

	t.Fatal("readiness omitted the selected spec")

	return ""
}
