package workflow

import (
	"testing"
	"time"
)

// TestNativeDecisionRepair executes a current answered conflict through the real
// broker and host repair. Answering never replaces the remaining gate/review.
func TestNativeDecisionRepair(t *testing.T) {
	f, request, harness, state := nativeMergerAcceptanceFixture(t, "merger-answer", "merge")
	request.Input.Requirements[0].ConflictsWith = []string{"target-validation"}

	input, err := SealWorkflowMergerInput(request.Input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	questions, err := MergerRequiredDecisions(input)
	if err != nil {
		t.Fatal(err)
	}

	record, err := f.store.RecordWorkflowIntegrationDecisions(t.Context(), IntegrationDecisionRequest{IntegrationBudgetGuard: nativeBudgetGuard(t, f.store, f.run.ID), ID: "repair-interpretation", Input: input, Decisions: questions}, f.provider)
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), nativeDecisionAnswer(t, f.store, f.run.ID, record), f.provider)
	if err != nil {
		t.Fatal(err)
	}

	nativeDecisionRepairExecute(t, f, request, harness, resolved, state)
}

func nativeDecisionRepairExecute(t *testing.T, f workflowPreparationFixture, request MergerRequest, harness MergerHarness, resolved IntegrationDecisions, state MergerRepairState) {
	t.Helper()

	request = nativeDecisionRepairRequest(t, f, request, resolved, state)

	attempt, err := f.store.ExecuteWorkflowMergerRepair(t.Context(), request, harness, f.provider, f.git)
	if err != nil || attempt.Acceptance == nil || attempt.Acceptance.Outcome != "accepted_pending_validation" {
		t.Fatal("current answer did not reach host repair", attempt.Failure, err)
	}

	checkNativeMergerAcceptanceRetained(t, f, attempt)
}

func nativeDecisionRepairRequest(t *testing.T, f workflowPreparationFixture, request MergerRequest, resolved IntegrationDecisions, state MergerRepairState) MergerRequest {
	t.Helper()

	run, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := f.store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		t.Fatal(err)
	}

	request.Input = resolved.Input
	request.Input.CandidateRevision = entry.StateRevision
	request.Input.OperatorDecisionRef, request.Input.OperatorAnswers = entry.DecisionRef, resolved.Answers
	request.Input.Deadline = time.Now().UTC().Add(25 * time.Second)

	request.Input, err = SealWorkflowMergerInput(request.Input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	request.IntegrationBudgetGuard = nativeBudgetGuard(t, f.store, f.run.ID)

	ref, err := f.store.RetainWorkflowMergerCheckpoint(t.Context(), request.IntegrationBudgetGuard, state, f.git)
	if err != nil {
		t.Fatal("fresh host checkpoint was rejected", err)
	}

	request.Input.RepairCheckpointRef = &ref

	request.Input, err = SealWorkflowMergerInput(request.Input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	return request
}
