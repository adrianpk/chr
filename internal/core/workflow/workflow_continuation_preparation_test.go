package workflow

import (
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Preparation recovery requires explicit authority and an interrupted pre-input
// attempt. It cannot turn a failed gate, stop or ordinary poll into execution.
func TestWorkflowPreparationRetryAuthority(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*Invocation, *ContinuationRequest)
		allow  bool
	}{
		{"explicit preparation", func(*Invocation, *ContinuationRequest) {}, true},
		{"unbound", func(_ *Invocation, request *ContinuationRequest) { request.ExpectedRevision = 0 }, false},
		{"failed gate", func(run *Invocation, _ *ContinuationRequest) {
			run.Outcome = "integration_validation_failed"
		}, false},
		{"retained input", func(run *Invocation, _ *ContinuationRequest) { run.Attempts[0].InputRef = "input.json" }, false},
		{"stop", func(run *Invocation, _ *ContinuationRequest) { run.StopIntent = "pause" }, false},
		{"paused", func(run *Invocation, _ *ContinuationRequest) { run.Status = workflowstate.InvocationPaused }, false},
		{"legacy", func(run *Invocation, _ *ContinuationRequest) { run.Integration = nil }, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			run := Invocation{InvocationExecution: InvocationExecution{Status: workflowstate.InvocationBlocked, Outcome: "integration_preparation_failed", Attempts: []Attempt{{Purpose: workflowIntegrationValidationPurpose, Outcome: "interrupted"}}}, InvocationIntegration: InvocationIntegration{Integration: &workflowstate.IntegrationBinding{}}, InvocationPreparation: InvocationPreparation{Preparation: &PreparationBinding{State: "stale"}}}
			request := ContinuationRequest{ExpectedRevision: 1}
			test.change(&run, &request)

			if workflowContinuationPreparationRetry(run, request, 0) != test.allow {
				t.Fatal("preparation retry authority mismatch")
			}
		})
	}
}
