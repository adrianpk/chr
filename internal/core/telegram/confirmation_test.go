package telegram

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

func telegramConfirmationFixture(t *testing.T) (*telegramConversationFixture, TelegramActionOffer, TelegramAction) {
	t.Helper()

	return telegramOperationActionFixture(t, func(request TelegramOperationRequest, offer *TelegramActionOffer) {
		payload := []byte("# Complete scope\n\nOne exact record and its validation.\n")
		offer.Delivery = &tgmessage.DocumentResponse{
			Message:  tgmessage.Blocks{Family: "29", InitialText: "Create this ticket?", Content: "One exact record."},
			Document: tgmessage.Document{ProjectID: request.Target.ProjectID, Namespace: request.Target.Namespace, ProjectBindingDigest: request.Target.ProjectBindingDigest, SourceRef: "exports/confirmation.md", SourceHash: contenthash.Sum(payload), Filename: "confirmation.md", MediaType: "text/markdown", Payload: payload},
		}
	})
}

// TestTelegramPresentationConfirmationBinding freezes the caption, complete
// document and exact keyboard under the same original operation and guard.
func TestTelegramPresentationConfirmationBinding(t *testing.T) {
	f, offer, action := telegramConfirmationFixture(t)
	if action.GuardVersion != 1 || len(action.PresentationHash) != 64 {
		t.Fatal("new confirmation did not seal its presentation", action)
	}

	delivery := telegramConfirmationDelivery(t, f, action)
	if err := f.store.ValidateTelegramActionDelivery(t.Context(), delivery); err != nil {
		t.Fatal(err)
	}

	if ok, err := f.store.TryReserveDelivery(t.Context(), delivery.ID, time.Now()); err != nil || !ok {
		t.Fatal("could not reserve original delivery", err)
	}

	document, err := f.store.ReadDeliveryDocument(t.Context(), delivery)
	if err != nil || !bytes.Equal(document.Payload, offer.Delivery.Document.Payload) {
		t.Fatal("document changed before upload", err)
	}

	reopenTelegramConfirmation(t, f, offer, action)

	input := telegramOperationActionInput(t, f, "confirm")

	claim, err := f.store.ClaimTelegramAction(t.Context(), input, TelegramActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now())
	if err != nil || claim.Action.State != "applying" {
		t.Fatal("valid original confirmation was refused", claim, err)
	}

	checkTelegramOperationActionFinish(t, f, claim)
}

// TestTelegramPresentationConfirmationRollback refuses an incomplete document
// before committing any prompt, source acknowledgement or action authority.
func TestTelegramPresentationConfirmationRollback(t *testing.T) {
	for _, kind := range []string{"oversized", "wrong project", "quota"} {
		t.Run(kind, func(t *testing.T) {
			f := newTelegramConversationFixture(t)
			f.apply(t, TelegramConversationStart, func(input *TelegramConversationInput) { input.Project = "alpha" })

			input := telegramOperationActionInput(t, f, "prepare confirmation")
			input.Conversation = f.sessions(t)[0].ID

			target, err := f.store.ResolveTelegramOperationTarget(t.Context(), input, time.Now())
			if err != nil {
				t.Fatal(err)
			}

			request, err := f.store.ReserveTelegramOperation(t.Context(), TelegramOperationRequest{Source: input.Update, Operator: input.Config.AllowedUsers[0].OperatorID, Target: target, OperationID: "records.ticket.create", Arguments: json.RawMessage(`{"title":"One record"}`)}, time.Now())
			if err != nil {
				t.Fatal(err)
			}

			if err := f.store.PrepareTelegramOperation(t.Context(), request, json.RawMessage(`{"reserved_id":"PRJ-1"}`), time.Now()); err != nil {
				t.Fatal(err)
			}

			checkTelegramConfirmationRollback(t, f, request, input, kind)
		})
	}
}

func checkTelegramConfirmationRollback(t *testing.T, f *telegramConversationFixture, request TelegramOperationRequest, input TelegramConversationInput, kind string) {
	t.Helper()

	document := rejectedTelegramConfirmationDocument(t, f, request, kind)

	var before int
	if err := f.store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_document_snapshots").Scan(&before); err != nil {
		t.Fatal(err)
	}

	_, err := f.store.OfferTelegramAction(t.Context(), TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{"confirm"}, Delivery: &tgmessage.DocumentResponse{Message: tgmessage.Blocks{InitialText: "Confirm?"}, Document: document}, ExpiresAt: time.Now().Add(time.Hour)}, time.Now())
	if err == nil {
		t.Fatal("incomplete scope exposed a confirmation")
	}

	var actions, documents int
	if err := f.store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_actions WHERE request_id=?", request.ID).Scan(&actions); err != nil {
		t.Fatal(err)
	}

	if err := f.store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_document_snapshots").Scan(&documents); err != nil {
		t.Fatal(err)
	}

	var state string
	if err := f.store.db.QueryRowContext(t.Context(), "SELECT state FROM telegram_updates WHERE update_id=?", input.Update.UpdateID).Scan(&state); err != nil {
		t.Fatal(err)
	}

	if actions != 0 || documents != before || state != "pending" {
		t.Fatal("document refusal partially acknowledged or authorized the action", actions, documents, state)
	}
}

func telegramConfirmationDelivery(t *testing.T, f *telegramConversationFixture, action TelegramAction) TelegramReceiverDelivery {
	t.Helper()

	deliveries, err := f.store.DueDeliveries(t.Context(), 12345, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	var delivery TelegramReceiverDelivery

	for _, row := range deliveries {
		if row.ActionToken == action.Token {
			delivery = row
		}
	}

	if delivery.DocumentID == "" || delivery.Presentation == nil || delivery.Presentation.ActionRef != action.Token || delivery.Presentation.ActionLabels[0] != "Confirm change" {
		t.Fatal("caption, document and keyboard were separated", delivery)
	}

	return delivery
}

func reopenTelegramConfirmation(t *testing.T, f *telegramConversationFixture, offer TelegramActionOffer, action TelegramAction) {
	t.Helper()

	before := action

	if err := f.store.Close(); err != nil {
		t.Fatal(err)
	}

	f.store = openTelegramReceiverTestStore(t, f.env)

	again, err := f.store.OfferTelegramAction(t.Context(), offer, time.Now())
	if err != nil || again.Token != before.Token || again.GuardDigest != before.GuardDigest || again.PresentationHash != before.PresentationHash {
		t.Fatal("reopen or replay replaced the original image", again, err)
	}
}

func rejectedTelegramConfirmationDocument(t *testing.T, f *telegramConversationFixture, request TelegramOperationRequest, kind string) tgmessage.Document {
	t.Helper()

	document := tgmessage.Document{ProjectID: request.Target.ProjectID, Namespace: request.Target.Namespace, ProjectBindingDigest: request.Target.ProjectBindingDigest, SourceRef: "exports/scope.md", SourceHash: contenthash.Sum([]byte("scope")), Filename: "scope.md", MediaType: "text/markdown", Payload: []byte("scope")}

	switch kind {
	case "oversized":
		document.Payload = bytes.Repeat([]byte("x"), tgmessage.DocumentLimit+1)
	case "wrong project":
		document.ProjectID = "another-project"
	default:
		seedTelegramDocumentQuota(t, f.store, 32)
	}

	return document
}
