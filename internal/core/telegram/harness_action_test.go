package telegram

import (
	"encoding/json"
	"testing"
	"time"
)

// A reservation failure rolls back its source-owned auxiliary conversation;
// unretained source data cannot consume capacity or create confirmation authority.
func TestTelegramHarnessSelectionReservation(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationSelect, func(input *ConversationInput) { input.Project = "alpha" })

	f.input.Update = telegramRetainedTestUpdate(t, 99, "/harness worker=beta")

	target, err := f.store.ResolveHarnessTarget(t.Context(), f.input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	request := OperationRequest{Source: f.input.Update, Operator: f.input.Config.AllowedUsers[0].OperatorID, Target: target, OperationID: HarnessSelectionOperation, Arguments: json.RawMessage(`{"worker":"beta"}`)}
	if _, err := f.store.ReserveHarnessOperation(t.Context(), f.input, request, json.RawMessage(`{"frozen":true}`), time.Now()); err == nil || len(f.sessions(t)) != 0 {
		t.Fatal("failed reservation leaked a confirmation context", err)
	}

	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{f.input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	retained, err := f.store.ReserveHarnessOperation(t.Context(), f.input, request, json.RawMessage(`{"frozen":true}`), time.Now())
	if err != nil || retained.Target.ConversationID != conversationID(f.input.Update) || len(f.sessions(t)) != 1 {
		t.Fatal("source-owned confirmation was not reserved atomically", err)
	}

	if err := f.store.ValidateHarnessTarget(t.Context(), f.input, retained); err != nil {
		t.Fatal("retained target could not be revalidated", err)
	}
}

// Expiry supersedes the pending confirmation and closes only its auxiliary
// conversation, retaining the original request for diagnostics and replay refusal.
func TestTelegramHarnessSelectionExpiry(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationSelect, func(input *ConversationInput) { input.Project = "alpha" })

	f.input.Update = telegramRetainedTestUpdate(t, 99, "/harness worker=beta")
	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{f.input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	target, err := f.store.ResolveHarnessTarget(t.Context(), f.input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReserveHarnessOperation(t.Context(), f.input, OperationRequest{Source: f.input.Update, Operator: f.input.Config.AllowedUsers[0].OperatorID, Target: target, OperationID: HarnessSelectionOperation, Arguments: json.RawMessage(`{"worker":"beta"}`)}, json.RawMessage(`{"frozen":true}`), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	action, err := f.store.OfferAction(t.Context(), ActionOffer{Input: f.input, RequestID: request.ID, Choices: []string{"confirm", "cancel"}, Text: "Confirm harness defaults?", ExpiresAt: time.Now().Add(time.Minute)}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	input := f.input

	input.Update = telegramRetainedTestUpdate(t, 100, "/answer "+action.Token+" confirm")
	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	claim, err := f.store.ClaimAction(t.Context(), input, ActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now().Add(time.Hour))
	if err != nil || claim.Action.State != "expired" || f.sessions(t)[0].Phase != "failed" {
		t.Fatal("expiry left a reusable confirmation or live prompt context", err)
	}
}
