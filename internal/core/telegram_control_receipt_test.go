package core

import (
	"errors"
	"testing"
)

func telegramControlJournalFixture(t *testing.T, action string) (*WorkflowStore, WorkflowDecisionControlRequest, WorkflowPullRequestProvider) {
	t.Helper()

	binding := WorkflowMutationBinding{RequestID: "control-receipt", Operator: "operator"}
	request := WorkflowDecisionControlRequest{Binding: binding, Action: action}

	var (
		store    *WorkflowStore
		provider WorkflowPullRequestProvider
	)

	switch action {
	case "approve":
		current, _, run, ports := workflowReviewFixture(t, ReviewWorkflowModeHumanReview)
		store, provider, request.ID = current, ports, run.ID

		review, err := store.ReviewWorkflowSlice(t.Context(), run.ID, provider, nil)
		if err != nil {
			t.Fatal(err)
		}

		request.ReviewHash = review.InputHash
	case "grant":
		current, entry := nativeBudgetFixture(t, nil)
		store, request.ID = current, entry.InvocationID
		request.Grant = &WorkflowIntegrationGrant{ID: binding.RequestID, Actor: binding.Operator, Kind: "submission", Amount: 1, Reason: "One bounded submission"}
	case "answer_decision":
		f, _, record := nativeDecisionFixture(t)
		store, provider, request.ID = f.store, f.provider, f.run.ID
		answer := nativeDecisionAnswer(t, store, f.run.ID, record)
		answer.Answer.ID = binding.RequestID
		request.Answer, request.DecisionHash, request.DecisionRevision = &answer.Answer, answer.DecisionHash, answer.DecisionRevision
	}

	inspection, err := store.InspectWorkflow(t.Context(), request.ID)
	if err != nil {
		t.Fatal(err)
	}

	request.ExpectedRevision = inspection.Invocation.Revision
	if inspection.Invocation.Review != nil {
		request.ReviewHash = inspection.Invocation.Review.InputHash
	}

	if observed := inspection.Integration; observed != nil && observed.Entry != nil {
		request.Integration = &WorkflowControlIntegrationGuard{EntryID: observed.Entry.ID, EntryRevision: observed.Entry.Revision, BudgetRevision: observed.Budget.Revision, CandidateHash: WorkflowIntegrationDisplayedCandidate(observed)}
	}

	return store, request, provider
}

// TestTelegramControlJournal interrupts a native transaction after its first
// replacement. Recovery commits its effect and exact request receipt once.
func TestTelegramControlJournal(t *testing.T) {
	for _, action := range []string{"approve", "grant", "answer_decision"} {
		t.Run(action, func(t *testing.T) { checkTelegramControlJournal(t, action) })
	}
}

func checkTelegramControlJournal(t *testing.T, action string) {
	t.Helper()

	store, request, provider := telegramControlJournalFixture(t, action)
	crash := errors.New("control journal interrupted")

	store.afterReplace = func(index int) error {
		if index == 0 {
			return crash
		}

		return nil
	}
	if _, err := store.ControlWorkflowDecision(t.Context(), request, provider, nil); !errors.Is(err, crash) {
		t.Fatal("native transaction was not interrupted", err)
	}

	restarted := NewWorkflowStore(store.layout, store.projectID)

	first, err := restarted.ControlWorkflowDecision(t.Context(), request, provider, nil)
	if err != nil || first.Revision != request.ExpectedRevision+1 || first.ReceiptRef == "" {
		t.Fatal("request receipt did not recover", first, err)
	}

	before := workflowTestRead(t, restarted, request.ID)

	second, err := restarted.ControlWorkflowDecision(t.Context(), request, provider, nil)
	if err != nil || second != first || workflowTestRead(t, restarted, request.ID).Revision != before.Revision {
		t.Fatal("replay repeated native control", second, err)
	}

	request.ExpectedRevision++
	if _, err := restarted.ControlWorkflowDecision(t.Context(), request, provider, nil); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("changed receipt payload accepted", err)
	}
}
