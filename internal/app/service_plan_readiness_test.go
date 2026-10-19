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
