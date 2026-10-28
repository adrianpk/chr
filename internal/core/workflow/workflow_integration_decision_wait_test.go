package workflow

import (
	"errors"
	"reflect"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestNativeDecisionRetry makes question creation idempotent and rejects reuse
// of its identifier with a different interpretation or stale source/base.
func TestNativeDecisionRetry(t *testing.T) {
	f, request, record := nativeDecisionFixture(t)
	create := IntegrationDecisionRequest{IntegrationBudgetGuard: request.IntegrationBudgetGuard, ID: record.ID, Input: request.Input, Decisions: record.Requests}

	again, err := f.store.RecordWorkflowIntegrationDecisions(t.Context(), create, f.provider)
	if err != nil || !reflect.DeepEqual(again, record) {
		t.Fatal("question retry changed history", err)
	}

	create.Decisions[0].Issue = "A different question"
	if _, err := f.store.RecordWorkflowIntegrationDecisions(t.Context(), create, f.provider); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("question identifier was reused", err)
	}
}

// TestNativeDecisionDefer retains an explicit deferral across restart. A later
// resolution appends history at its own revision rather than replacing the wait.
func TestNativeDecisionDefer(t *testing.T) {
	f, _, record := nativeDecisionFixture(t)
	answer := nativeDecisionAnswer(t, f.store, f.run.ID, record)
	answer.Answer.Action = "defer"

	waiting, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider)
	if err != nil || waiting.State != "waiting" || len(waiting.Answers) != 1 {
		t.Fatal("deferral approved work", err)
	}

	store := NewWorkflowStore(f.store.layout, f.store.projectID)

	again, err := store.ReadWorkflowIntegrationDecisions(t.Context(), f.run.ID)
	if err != nil || !reflect.DeepEqual(waiting, again) {
		t.Fatal("deferral was lost", err)
	}

	answer = nativeDecisionAnswer(t, store, f.run.ID, waiting)
	answer.Answer.ID = "answer-2"

	resolved, err := store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider)
	if err != nil || resolved.State != "resolved" || len(resolved.Answers) != 2 || resolved.Answers[0] != waiting.Answers[0] {
		t.Fatal("explicit resolution erased history", err)
	}
}

// TestNativeDecisionDriver persists a real bounded merger's question atomically
// with its stopped attempt and charged budget, without another agent call.
func TestNativeDecisionDriver(t *testing.T) {
	f, request, harness := nativeMergerExecution(t, "merger-decision")

	attempt, err := f.store.ExecuteWorkflowMerger(t.Context(), request, harness, f.provider)
	if err != nil || attempt.Proposal == nil || attempt.Proposal.Outcome != "needs_decision" {
		t.Fatal("driver question failed", attempt.Failure, err)
	}

	record, err := f.store.ReadWorkflowIntegrationDecisions(t.Context(), f.run.ID)
	if err != nil || record.State != "waiting" || len(record.Requests) != 1 {
		t.Fatal("driver question was not persisted", err)
	}

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.run.ID)
	if err != nil || budget.UnitCounters[0].Merger != 1 {
		t.Fatal("question changed consumed calls", err)
	}

	answer := nativeDecisionAnswer(t, f.store, f.run.ID, record)
	if _, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider); err != nil {
		t.Fatal("driver question could not be answered", err)
	}
}

// TestNativeDecisionBatch waits for every question and rejects an answer based
// on an older displayed revision even when it addresses an unanswered question.
func TestNativeDecisionBatch(t *testing.T) {
	f, request, _ := nativeMergerExecution(t, "decision-batch")
	request.Input.Requirements[0].ConflictsWith = []string{"target-validation"}
	other := request.Input.Requirements[1]
	other.ID, other.ConflictsWith = "target-extra", []string{"source-api"}
	request.Input.Requirements = append(request.Input.Requirements, other)

	input, err := SealWorkflowMergerInput(request.Input, request.Input.AdmittedAt)
	if err != nil {
		t.Fatal(err)
	}

	questions, err := MergerRequiredDecisions(input)
	if err != nil || len(questions) != 2 {
		t.Fatal("batch fixture lost questions", err)
	}

	record, err := f.store.RecordWorkflowIntegrationDecisions(t.Context(), IntegrationDecisionRequest{IntegrationBudgetGuard: request.IntegrationBudgetGuard, ID: "batch", Input: input, Decisions: questions}, f.provider)
	if err != nil {
		t.Fatal(err)
	}

	nativeDecisionBatchAnswers(t, f, record, questions)
}

func nativeDecisionBatchAnswers(t *testing.T, f workflowPreparationFixture, record IntegrationDecisions, questions []workflowstate.MergerDecision) {
	t.Helper()

	answer := nativeDecisionAnswer(t, f.store, f.run.ID, record)

	waiting, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider)
	if err != nil || waiting.State != "waiting" {
		t.Fatal("partial answer approved batch", err)
	}

	answer.Answer.ID, answer.Answer.DecisionID = "answer-2", questions[1].ID
	if _, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("old display approved unanswered question", err)
	}

	answer = nativeDecisionAnswer(t, f.store, f.run.ID, waiting)
	answer.Answer.ID, answer.Answer.DecisionID = "answer-2", questions[1].ID

	resolved, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider)
	if err != nil || resolved.State != "resolved" || len(resolved.Answers) != 2 {
		t.Fatal("complete explicit answers did not resolve batch", err)
	}
}

// TestNativeDecisionPhases permits normal validation after an answer while
// review still requires exact candidate evidence from the unchanged host gate.
func TestNativeDecisionPhases(t *testing.T) {
	f, _, record := nativeDecisionFixture(t)

	answer := nativeDecisionAnswer(t, f.store, f.run.ID, record)
	if _, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider); err != nil {
		t.Fatal(err)
	}

	run, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := f.store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		t.Fatal(err)
	}

	entry = nativeRoleTransition(t, f.store, entry, "validating")

	request := nativeEntryTransition(t, f.store, entry, "reviewing")

	current, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.guardWorkflowDecisionTransition(t.Context(), current, entry, request); err != nil {
		t.Fatal("resolved decision blocked normal review", err)
	}

	if _, err := f.store.TransitionWorkflowIntegration(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("answer skipped exact candidate validation", err)
	}
}
