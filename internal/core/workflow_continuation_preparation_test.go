package core

import "testing"

// Preparation recovery requires explicit authority and an interrupted pre-input
// attempt. It cannot turn a failed gate, stop or ordinary poll into execution.
func TestWorkflowPreparationRetryAuthority(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*WorkflowInvocation, *WorkflowContinuationRequest)
		allow  bool
	}{
		{"explicit preparation", func(*WorkflowInvocation, *WorkflowContinuationRequest) {}, true},
		{"unbound", func(_ *WorkflowInvocation, request *WorkflowContinuationRequest) { request.ExpectedRevision = 0 }, false},
		{"failed gate", func(run *WorkflowInvocation, _ *WorkflowContinuationRequest) {
			run.Outcome = "integration_validation_failed"
		}, false},
		{"retained input", func(run *WorkflowInvocation, _ *WorkflowContinuationRequest) { run.Attempts[0].InputRef = "input.json" }, false},
		{"stop", func(run *WorkflowInvocation, _ *WorkflowContinuationRequest) { run.StopIntent = "pause" }, false},
		{"paused", func(run *WorkflowInvocation, _ *WorkflowContinuationRequest) { run.Status = "paused" }, false},
		{"legacy", func(run *WorkflowInvocation, _ *WorkflowContinuationRequest) { run.Integration = nil }, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			run := WorkflowInvocation{WorkflowInvocationExecution: WorkflowInvocationExecution{Status: "blocked", Outcome: "integration_preparation_failed", Attempts: []WorkflowAttempt{{Purpose: workflowIntegrationValidationPurpose, Outcome: "interrupted"}}}, WorkflowInvocationIntegration: WorkflowInvocationIntegration{Integration: &WorkflowIntegrationBinding{}}, WorkflowInvocationPreparation: WorkflowInvocationPreparation{Preparation: &WorkflowPreparationBinding{State: "stale"}}}
			request := WorkflowContinuationRequest{ExpectedRevision: 1}
			test.change(&run, &request)

			if workflowContinuationPreparationRetry(run, request, 0) != test.allow {
				t.Fatal("preparation retry authority mismatch")
			}
		})
	}
}
