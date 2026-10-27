package core

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
)

// A reservation failure rolls back its source-owned auxiliary conversation;
// unretained source data cannot consume capacity or create confirmation authority.
func TestTelegramHarnessSelectionReservation(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationSelect, func(input *TelegramConversationInput) { input.Project = "alpha" })

	f.input.Update = telegramRetainedTestUpdate(t, 99, "/harness worker=beta")

	target, err := f.store.ResolveTelegramHarnessTarget(t.Context(), f.input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	request := TelegramOperationRequest{Source: f.input.Update, Operator: f.input.Config.AllowedUsers[0].OperatorID, Target: target, OperationID: TelegramHarnessSelectionOperation, Arguments: json.RawMessage(`{"worker":"beta"}`)}
	if _, err := f.store.ReserveTelegramHarnessOperation(t.Context(), f.input, request, json.RawMessage(`{"frozen":true}`), time.Now()); err == nil || len(f.sessions(t)) != 0 {
		t.Fatal("failed reservation leaked a confirmation context", err)
	}

	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{f.input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	retained, err := f.store.ReserveTelegramHarnessOperation(t.Context(), f.input, request, json.RawMessage(`{"frozen":true}`), time.Now())
	if err != nil || retained.Target.ConversationID != telegramConversationID(f.input.Update) || len(f.sessions(t)) != 1 {
		t.Fatal("source-owned confirmation was not reserved atomically", err)
	}

	if err := f.store.ValidateTelegramHarnessTarget(t.Context(), f.input, retained); err != nil {
		t.Fatal("retained target could not be revalidated", err)
	}
}

// A definitive local preparation rejection preserves its cause and leaves no
// published image or receipt, unlike an uncertain failure after publication.
func TestTelegramHarnessSelectionPreparation(t *testing.T) {
	project := t.TempDir()
	request := harnessSelectionTestRequest("unprepared", contenthash.Sum(nil), "alpha")
	options := harnessSelectionTestOptions(request)
	cause := errors.New("fixture local preparation rejected")
	options.Validate = func(context.Context, WorkspaceConfig) error { return cause }

	if _, err := ChangeHarnessSelections(t.Context(), project, options); !errors.Is(err, ErrHarnessSelectionPreparation) || !errors.Is(err, cause) {
		t.Fatal("prepublication rejection lost its identity", err)
	}

	if len(readHarnessSelectionTestConfig(t, project)) != 0 {
		t.Fatal("rejected preparation published configuration")
	}

	if _, found, err := readHarnessSelectionReceipt(project, request, ""); err != nil || found {
		t.Fatal("rejected preparation retained a receipt", err)
	}
}

// Expiry supersedes the pending confirmation and closes only its auxiliary
// conversation, retaining the original request for diagnostics and replay refusal.
func TestTelegramHarnessSelectionExpiry(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationSelect, func(input *TelegramConversationInput) { input.Project = "alpha" })

	f.input.Update = telegramRetainedTestUpdate(t, 99, "/harness worker=beta")
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{f.input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	target, err := f.store.ResolveTelegramHarnessTarget(t.Context(), f.input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReserveTelegramHarnessOperation(t.Context(), f.input, TelegramOperationRequest{Source: f.input.Update, Operator: f.input.Config.AllowedUsers[0].OperatorID, Target: target, OperationID: TelegramHarnessSelectionOperation, Arguments: json.RawMessage(`{"worker":"beta"}`)}, json.RawMessage(`{"frozen":true}`), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	action, err := f.store.OfferTelegramAction(t.Context(), TelegramActionOffer{Input: f.input, RequestID: request.ID, Choices: []string{"confirm", "cancel"}, Text: "Confirm harness defaults?", ExpiresAt: time.Now().Add(time.Minute)}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	input := f.input

	input.Update = telegramRetainedTestUpdate(t, 100, "/answer "+action.Token+" confirm")
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	claim, err := f.store.ClaimTelegramAction(t.Context(), input, TelegramActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now().Add(time.Hour))
	if err != nil || claim.Action.State != "expired" || f.sessions(t)[0].Phase != "failed" {
		t.Fatal("expiry left a reusable confirmation or live prompt context", err)
	}
}
