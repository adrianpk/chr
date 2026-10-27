package core

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramWorkReferenceFixture(t *testing.T, f *telegramConversationFixture, id string) TelegramWorkReference {
	t.Helper()

	start := f.next
	f.apply(t, TelegramConversationStart, func(input *TelegramConversationInput) { input.Project = "alpha" })

	input := telegramOperationActionInput(t, f, "native selection")
	input.Conversation = telegramConversationID(TelegramReceivedUpdate{Peer: input.Update.Peer, UpdateID: start})

	target, err := f.store.ResolveTelegramOperationTarget(t.Context(), input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReserveTelegramOperation(t.Context(), TelegramOperationRequest{Source: input.Update, Operator: input.Config.AllowedUsers[0].OperatorID, Target: target, OperationID: "work.admit", Arguments: json.RawMessage(`{"extent":"ticket"}`)}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.PrepareTelegramOperation(t.Context(), request, json.RawMessage(`{"native":"preview"}`), time.Now()); err != nil {
		t.Fatal(err)
	}

	action, err := f.store.OfferTelegramAction(t.Context(), TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{"start"}, Text: "Start the exact extent?", ExpiresAt: time.Now().Add(time.Hour)}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	answer := telegramOperationActionInput(t, f, "start")

	claim, err := f.store.ClaimTelegramAction(t.Context(), answer, TelegramActionAnswer{Token: action.Token, Choice: "start"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	bound := &TelegramInvocationBinding{ID: id, Revision: 1}
	if err := f.store.FinishTelegramAction(t.Context(), claim, TelegramActionResult{State: "confirmed", ReceiptRef: "workflow:" + id, Facts: json.RawMessage(`{"native":"admission"}`), Text: "Admitted", ConfigDigest: input.Config.Digest(), Invocation: bound}, time.Now()); err != nil {
		t.Fatal(err)
	}

	return TelegramWorkReference{Request: claim.Request, Invocation: *bound, ConversationRevision: target.ConversationRevision}
}

// TestTelegramWorkReferences proves bounded, peer-bound paging over confirmed
// admissions and monotonic observation without changing the original receipt.
func TestTelegramWorkReferences(t *testing.T) {
	f := newTelegramConversationFixture(t)
	for _, id := range []string{"run-a", "run-b", "run-c"} {
		telegramWorkReferenceFixture(t, f, id)
	}

	first, err := f.store.TelegramWorkReferences(t.Context(), 12345, "", 2)
	if err != nil || len(first) != 2 {
		t.Fatal("confirmed references not paged", first, err)
	}

	second, err := f.store.TelegramWorkReferences(t.Context(), 12345, first[1].Request.Target.ConversationID, 2)
	if err != nil || len(second) != 1 || second[0].Invocation.ID == first[1].Invocation.ID {
		t.Fatal("keyset continuation duplicated a binding", second, err)
	}

	checkTelegramWorkObservation(t, f, first[0])

	if _, err := f.store.TelegramWorkReferences(t.Context(), 12345, "", 17); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("unbounded reference page admitted", err)
	}

	foreign, err := f.store.TelegramWorkReferences(t.Context(), 54321, "", 2)
	if err != nil || len(foreign) != 0 {
		t.Fatal("another bot inherited work", foreign, err)
	}
}

func checkTelegramWorkObservation(t *testing.T, f *telegramConversationFixture, ref TelegramWorkReference) {
	t.Helper()

	before, err := readTelegramOperationID(t.Context(), f.store.db, ref.Request.ID)
	if err != nil {
		t.Fatal(err)
	}

	for _, revision := range []uint64{5, 3} {
		if err := f.store.ObserveTelegramWork(t.Context(), ref, revision, time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	ref.Invocation.ID = "another-invocation"
	if err := f.store.ValidateTelegramWorkReference(t.Context(), ref); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("rebound invocation accepted", err)
	}

	for _, c := range f.sessions(t) {
		if c.ID == ref.Request.Target.ConversationID && c.LastObservedRevision != 5 {
			t.Fatal("old observation regressed progress", c)
		}
	}

	checkTelegramWorkOriginalReceipt(t, f, before)
}

func checkTelegramWorkOriginalReceipt(t *testing.T, f *telegramConversationFixture, before TelegramOperationRequest) {
	t.Helper()

	after, err := readTelegramOperationID(t.Context(), f.store.db, before.ID)
	if err != nil || string(before.Result) != string(after.Result) || before.ReceiptRef != after.ReceiptRef {
		t.Fatal("observation replaced the admission receipt", err)
	}
}

// TestTelegramWorkContinuationGuard rejects revoked authority under the common
// acceptance lock, before a checkpoint, worker or other invocation change.
func TestTelegramWorkContinuationGuard(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request).Invocation
	calls := 0
	check := func(context.Context) error { calls++; return workflowstate.ErrConflict }

	_, err := store.ContinueWorkflow(t.Context(), run.ID, WorkflowContinuationRequest{ExpectedRevision: run.Revision, StepLimit: 1, AuthorizationCheck: check}, WorkflowContinuationPorts{})
	if !errors.Is(err, workflowstate.ErrConflict) || calls != 1 {
		t.Fatal("continuation authority was not checked", calls, err)
	}

	after, err := store.ReadWorkflow(t.Context(), run.ID)
	if err != nil || after.Revision != run.Revision || after.Continuation != nil || len(after.Attempts) != 0 {
		t.Fatal("revoked activation changed native work", after, err)
	}
}
