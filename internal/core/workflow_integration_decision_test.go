package core

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func nativeDecisionFixture(t *testing.T) (workflowPreparationFixture, WorkflowMergerRequest, WorkflowIntegrationDecisions) {
	t.Helper()

	f, request, _ := nativeMergerExecution(t, "decision-input")
	request.Input.Requirements[0].ConflictsWith = []string{"target-validation"}

	var err error

	request.Input, err = SealWorkflowMergerInput(request.Input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	decisions, err := WorkflowMergerRequiredDecisions(request.Input)
	if err != nil {
		t.Fatal(err)
	}

	record, err := f.store.RecordWorkflowIntegrationDecisions(t.Context(), WorkflowIntegrationDecisionRequest{WorkflowIntegrationBudgetGuard: request.WorkflowIntegrationBudgetGuard, ID: "intent-choice", Input: request.Input, Decisions: decisions}, f.provider)
	if err != nil {
		t.Fatal(err)
	}

	return f, request, record
}

func nativeDecisionAnswer(t *testing.T, store *WorkflowStore, id string, record WorkflowIntegrationDecisions) WorkflowIntegrationAnswerRequest {
	t.Helper()

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		t.Fatal(err)
	}

	question := record.Requests[0]

	return WorkflowIntegrationAnswerRequest{WorkflowIntegrationBudgetGuard: nativeBudgetGuard(t, store, id), DecisionRevision: record.Revision, DecisionHash: entry.DecisionRef.Hash, Answer: WorkflowIntegrationAnswer{ID: "answer-1", DecisionID: question.ID, Choice: question.Options[0], Action: "resolve", Actor: "operator", Authority: question.Authority, Reason: "Preserve both intents through the explicit compatible interpretation"}}
}

// TestNativeDecisionWait persists questions across restart and forbids a generic
// phase transition from treating silence or a timer as an operator answer.
func TestNativeDecisionWait(t *testing.T) {
	f, _, record := nativeDecisionFixture(t)
	store := NewWorkflowStore(f.store.layout, f.store.projectID)

	again, err := store.ReadWorkflowIntegrationDecisions(t.Context(), f.run.ID)
	if err != nil || !reflect.DeepEqual(record, again) || again.State != "waiting" {
		t.Fatal("question not retained", err)
	}

	run, err := store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.TransitionWorkflowIntegration(t.Context(), nativeEntryTransition(t, store, entry, "merging")); !errors.Is(err, ErrWorkflowOwned) {
		t.Fatal("silence resumed merger work", err)
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), f.run.ID)
	if err != nil || budget.UnitCounters[0].Merger != 0 {
		t.Fatal("asking a question consumed a model call", err)
	}
}

// TestNativeDecisionAnswer binds explicit answers to source/base and all current
// revisions; an exact retry preserves the original response and cannot reapprove.
func TestNativeDecisionAnswer(t *testing.T) {
	f, request, record := nativeDecisionFixture(t)
	answer := nativeDecisionAnswer(t, f.store, f.run.ID, record)

	completed, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider)
	if err != nil || completed.State != "resolved" || len(completed.Answers) != 1 {
		t.Fatal("explicit answer failed", err)
	}

	again, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider)
	if err != nil || !reflect.DeepEqual(completed, again) {
		t.Fatal("response retry mutated history", err)
	}

	nativeDecisionResolvedInput(t, f, request.Input, completed)
}

func nativeDecisionResolvedInput(t *testing.T, f workflowPreparationFixture, input WorkflowMergerInput, record WorkflowIntegrationDecisions) {
	t.Helper()

	run, err := f.store.readWorkflowInvocation(f.run.ID)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := f.store.readWorkflowIntegrationEntry(run, *run.Integration)
	if err != nil {
		t.Fatal(err)
	}

	input.CandidateRevision = entry.StateRevision
	input.OperatorDecisionRef = entry.DecisionRef
	input.OperatorAnswers = record.Answers

	input, err = SealWorkflowMergerInput(input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.CheckWorkflowMergerEvidence(t.Context(), input, time.Now().UTC()); err != nil {
		t.Fatal("answer evidence not bound", err)
	}

	decisions, err := WorkflowMergerRequiredDecisions(input)
	if err != nil || len(decisions) != 0 {
		t.Fatal("explicit interpretation was not available to merger", err)
	}

	input.OperatorAnswers[0].Actor = "forged-agent"

	input, err = SealWorkflowMergerInput(input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.CheckWorkflowMergerEvidence(t.Context(), input, time.Now().UTC()); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("model text replaced operator authority", err)
	}
}

// TestNativeDecisionGuards rejects stale hashes/revisions, wrong authority and
// changed external targets without appending an answer or changing budgets.
func TestNativeDecisionGuards(t *testing.T) {
	cases := []struct {
		name   string
		change func(*WorkflowIntegrationAnswerRequest, *workflowPreparationFixture)
	}{
		{"empty answer", func(r *WorkflowIntegrationAnswerRequest, _ *workflowPreparationFixture) {
			r.Answer = WorkflowIntegrationAnswer{}
		}},
		{"revision", func(r *WorkflowIntegrationAnswerRequest, _ *workflowPreparationFixture) { r.DecisionRevision++ }},
		{"hash", func(r *WorkflowIntegrationAnswerRequest, _ *workflowPreparationFixture) {
			r.DecisionHash = workflowHash([]byte("other"))
		}},
		{"authority", func(r *WorkflowIntegrationAnswerRequest, _ *workflowPreparationFixture) { r.Answer.Authority = "agent" }},
		{"source", func(_ *WorkflowIntegrationAnswerRequest, f *workflowPreparationFixture) {
			f.provider.pr.HeadCommit = workflowHash([]byte("changed-source"))[:40]
		}},
		{"scope choice", func(r *WorkflowIntegrationAnswerRequest, _ *workflowPreparationFixture) {
			r.Answer.Choice = "Approve a scoped requirement change"
		}},
		{"base", func(_ *WorkflowIntegrationAnswerRequest, f *workflowPreparationFixture) {
			f.provider.target = workflowHash([]byte("changed"))[:40]
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, _, record := nativeDecisionFixture(t)
			answer := nativeDecisionAnswer(t, f.store, f.run.ID, record)
			tc.change(&answer, &f)

			if _, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), answer, f.provider); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("invalid answer accepted", err)
			}

			again, err := f.store.ReadWorkflowIntegrationDecisions(t.Context(), f.run.ID)
			if err != nil || !reflect.DeepEqual(record, again) {
				t.Fatal("rejected answer mutated the question", err)
			}
		})
	}
}

// TestNativeDecisionContext prevents a sealed operator answer from approving
// different paths, requirements or a replacement gate under the same scope hash.
func TestNativeDecisionContext(t *testing.T) {
	f, request, record := nativeDecisionFixture(t)

	resolved, err := f.store.AnswerWorkflowIntegrationDecision(t.Context(), nativeDecisionAnswer(t, f.store, f.run.ID, record), f.provider)
	if err != nil {
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

	request.Input.OperatorDecisionRef, request.Input.OperatorAnswers = entry.DecisionRef, resolved.Answers
	request.Input.CandidateRevision = entry.StateRevision

	cases := []struct {
		name   string
		change func(*WorkflowMergerInput)
	}{
		{"paths", func(input *WorkflowMergerInput) { input.AllowedPaths = append(input.AllowedPaths, "other.txt") }},
		{"requirements", func(input *WorkflowMergerInput) { input.Requirements[0].ConflictsWith = nil }},
		{"gate", func(input *WorkflowMergerInput) { input.ValidationCommands = [][]string{{"true"}} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			copy, err := freezeWorkflowMergerRequest(request)
			if err != nil {
				t.Fatal(err)
			}

			tc.change(&copy.Input)

			input, err := SealWorkflowMergerInput(copy.Input, time.Now().UTC())
			if err != nil {
				t.Fatal(err)
			}

			if err := f.store.CheckWorkflowMergerEvidence(t.Context(), input, time.Now().UTC()); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("changed context reused answer", err)
			}
		})
	}
}
