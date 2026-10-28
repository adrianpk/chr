package workflow

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeAgentHarness(t *testing.T, store *Store, id, mode string) AgentHarness {
	t.Helper()

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	command := []string{program, "-test.run=^TestNativeAgentProcess$", "--", mode}

	data, err := os.ReadFile(program)
	if err != nil {
		t.Fatal(err)
	}

	certificate := AgentHarnessCertificate{SchemaVersion: 1, Protocol: workflowAgentCallProtocol, Command: command, ExecutableHash: contenthash.Sum(data), CapabilityRevision: contenthash.Sum([]byte("fixture-whole-call/v1")), Roles: []string{"executor", "reviewer", "closing"}, Accounting: workflowTokenAccounting, EnforcesWholeCall: true, ReportsWholeCall: true}
	ref := retainNativeRecordEvidence(t, store, id, "agent-harness-"+mode, certificate)

	return AgentHarness{Command: command, Certificate: ref, InputTokens: 60, OutputTokens: 40}
}

// TestNativeAgentProcess models a trusted driver rather than a model response;
// all usage and role information comes from the immutable host call input.
func TestNativeAgentProcess(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-2] != "--" {
		return
	}

	input := workflowProcessInput(t)
	if input.AgentCall == nil {
		t.Fatal("driver did not receive its whole-call grant")
	}

	mode := os.Args[len(os.Args)-1]
	nativeAgentFaultMode(t, mode)

	result := nativeAgentModelResult(t, input)

	usedInput, usedOutput := int64(20), int64(10)
	response := AgentResponse{Protocol: workflowAgentCallProtocol, ReservationID: input.AgentCall.ReservationID, AttemptID: input.AttemptID, CapabilityRevision: input.AgentCall.Allowance.CapabilityRevision, WholeCall: true, Input: &usedInput, Output: &usedOutput, Result: result}

	nativeAgentUsageMode(t, mode, &response)

	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		t.Fatal(err)
	}

	os.Exit(0)
}

// TestNativeAgentSettlement preserves actual usage separately from model text
// and keeps the complete grant charged for missing or incompatible accounting.
func TestNativeAgentSettlement(t *testing.T) {
	for _, mode := range []string{"success", "missing", "partial", "overflow", "stale"} {
		t.Run(mode, func(t *testing.T) {
			store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
			request.Policy.AgentTokenBudget = 100
			run := workflowTestStart(t, store, request)
			harness := nativeAgentHarness(t, store, run.Invocation.ID, mode)

			claim := workflowTestClaim(t, store, run.Invocation.ID)
			defer func() { _ = claim.Close() }() //nolint:errcheck

			result, err := ExecuteTaskWithHarness(t.Context(), store, claim, harness)
			checkNativeAgentOutcome(t, result, err, mode)

			restarted := NewWorkflowStore(store.layout, store.projectID)

			budget, err := restarted.ReadWorkflowIntegrationBudget(t.Context(), run.Invocation.ID)
			if err != nil || len(budget.Reservations) != 1 {
				t.Fatal("call history missing", err)
			}

			checkNativeAgentCharge(t, budget, mode)
		})
	}
}

// TestNativeAgentCapability blocks output-only, mismatched-unit and role-unsafe
// drivers before reserving tokens or granting a process launch permit.
func TestNativeAgentCapability(t *testing.T) {
	cases := []struct {
		name   string
		change func(*AgentHarnessCertificate)
	}{
		{"output only", func(c *AgentHarnessCertificate) { c.EnforcesWholeCall = false }},
		{"no usage", func(c *AgentHarnessCertificate) { c.ReportsWholeCall = false }},
		{"units", func(c *AgentHarnessCertificate) { c.Accounting = "output-tokens/v1" }},
		{"role", func(c *AgentHarnessCertificate) { c.Roles = []string{"reviewer"} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
			request.Policy.AgentTokenBudget = 100
			run := workflowTestStart(t, store, request)
			harness := nativeAgentHarness(t, store, run.Invocation.ID, "success")

			var certificate AgentHarnessCertificate
			if err := store.readWorkflowNativeEvidence(run.Invocation.ID, harness.Certificate, &certificate); err != nil {
				t.Fatal(err)
			}

			tc.change(&certificate)

			harness.Certificate = retainNativeRecordEvidence(t, store, run.Invocation.ID, "invalid-agent-harness", certificate)

			claim := workflowTestClaim(t, store, run.Invocation.ID)
			defer func() { _ = claim.Close() }() //nolint:errcheck

			if _, err := ExecuteTaskWithHarness(t.Context(), store, claim, harness); !errors.Is(err, ErrWorkflowIntegrationBudget) {
				t.Fatal("uncertified call admitted", err)
			}

			retained := workflowTestRead(t, store, run.Invocation.ID)

			budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), run.Invocation.ID)
			if err != nil || len(budget.Reservations) != 0 || len(retained.Attempts[0].Processes) != 0 {
				t.Fatal("rejected call charged or launched", err)
			}
		})
	}
}

func nativeAgentModelResult(t *testing.T, input TaskInput) string {
	t.Helper()

	if input.Review == nil {
		workflowProcessWorker(t, input, filepath.Join(input.Lane.Worktree, input.TaskKey+".txt"), "worker")

		return "worker complete"
	}

	decision := workflowstate.ReviewDecision{BindingHash: input.ReviewHash, Result: workflowstate.AgenticReviewResult{Decision: workflowstate.IntegrationReviewDecisionPass, Summary: "Inspect the current candidate and validation evidence."}}

	data, err := json.Marshal(decision)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func checkNativeAgentCharge(t *testing.T, budget IntegrationBudget, mode string) {
	t.Helper()

	r := budget.Reservations[0]
	if mode == "success" {
		if budget.TokensSettled != 30 || budget.TokensReserved != 0 || r.State != "settled" {
			t.Fatal("usage not settled once")
		}

		return
	}

	if budget.TokensSettled != 0 || budget.TokensReserved != 100 || r.State != "unknown" {
		t.Fatal("uncertain usage refunded after restart")
	}
}

func nativeAgentUsageMode(t *testing.T, mode string, response *AgentResponse) {
	t.Helper()

	switch mode {
	case "missing":
		response.Input = nil
	case "partial":
		response.WholeCall = false
	case "overflow":
		*response.Input = math.MaxInt64
	case "stale":
		response.AttemptID = "another-call"
	case "success":
	default:
		t.Fatal("unknown driver mode")
	}
}

func checkNativeAgentOutcome(t *testing.T, result workflowstate.TaskResult, err error, mode string) {
	t.Helper()

	if mode == "success" {
		if err != nil || !result.Successful || result.Worker.Stdout != "worker complete" {
			t.Fatal("certified call failed", err)
		}

		return
	}

	if !errors.Is(err, ErrWorkflowIntegrationBudget) || result.Successful {
		t.Fatal("uncertain accounting accepted", err)
	}
}
