package core

import (
	"strings"
	"testing"
	"time"
)

func deliverTelegramOperationAction(t *testing.T, f *telegramConversationFixture, token string) int64 {
	t.Helper()

	deliveries, err := f.store.DueDeliveries(t.Context(), 12345, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	var messageID int64

	for index, delivery := range deliveries {
		id := int64(index + 100)

		if err := f.store.ReserveDelivery(t.Context(), delivery.ID, time.Now()); err != nil {
			t.Fatal(err)
		}

		if err := f.store.DeliverySent(t.Context(), delivery.ID, id, time.Now()); err != nil {
			t.Fatal(err)
		}

		if delivery.ActionToken == token {
			messageID = id
		}
	}

	if messageID == 0 {
		t.Fatal("action prompt not delivered")
	}

	return messageID
}

// TestTelegramOperationActionReplies binds exact replies and callbacks to the same immutable action and peer.
func TestTelegramOperationActionReplies(t *testing.T) {
	f, _, action := telegramOperationActionFixture(t)
	messageID := deliverTelegramOperationAction(t, f, action.Token)
	input := telegramOperationActionInput(t, f, "confirm")
	input.Update.ReplyToMessageID = messageID

	reply, recognized, err := f.store.ResolveTelegramActionAnswer(t.Context(), input.Update)
	if err != nil || !recognized || reply.Token != action.Token || reply.Choice != "confirm" {
		t.Fatal("action reply unavailable", reply, recognized, err)
	}

	checkTelegramOperationCallbackAnswer(t, f, input, reply)
}

// TestTelegramOperationActionBounds prevents token, choice, lifetime and retained prompt expansion.
func TestTelegramOperationActionBounds(t *testing.T) {
	for _, kind := range []string{"lifetime", "choices", "duplicate", "text"} {
		t.Run(kind, func(t *testing.T) {
			f, offer, _ := telegramOperationActionFixture(t)

			switch kind {
			case "lifetime":
				offer.ExpiresAt = time.Now().Add(25 * time.Hour)
			case "choices":
				offer.Choices = make([]string, 9)
			case "duplicate":
				offer.Choices = []string{"confirm", "confirm"}
			case "text":
				offer.Text = strings.Repeat("x", 16385)
			}

			if _, err := f.store.OfferTelegramAction(t.Context(), offer, time.Now()); err == nil {
				t.Fatal("action bound exceeded", kind)
			}
		})
	}
}

func checkTelegramOperationCallbackAnswer(t *testing.T, f *telegramConversationFixture, input TelegramConversationInput, reply TelegramActionAnswer) {
	t.Helper()

	callback, err := TelegramActionCallbackData(reply)
	if err != nil || len(callback) > 64 {
		t.Fatal("callback binding exceeds transport", callback, err)
	}

	input.Update.Kind, input.Update.Text = "callback_query", callback

	answer, recognized, err := f.store.ResolveTelegramActionAnswer(t.Context(), input.Update)
	if err != nil || !recognized || answer != reply {
		t.Fatal("callback and reply diverged", answer, err)
	}

	input.Update.Peer.UserID++
	if _, err := f.store.ClaimTelegramAction(t.Context(), input, answer, time.Now()); err == nil {
		t.Fatal("action crossed peer identity")
	}
}

// TestTelegramOperationActionRevocation invalidates pending prompts when intake rejects revoked authority.
func TestTelegramOperationActionRevocation(t *testing.T) {
	f, _, action := telegramOperationActionFixture(t)
	deliverTelegramOperationAction(t, f, action.Token)

	update := telegramRetainedTestUpdate(t, f.next, "revoked input")

	update.Kind, update.Reason, update.Text = "rejected", "unauthorized_peer", ""
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	retained, err := readTelegramAction(t.Context(), f.store.db, action.Token, action.Peer)
	if err != nil || retained.State != "superseded" {
		t.Fatal("revoked prompt remained actionable", retained, err)
	}

	var active int
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_replies WHERE action_token=? AND invalidated_at IS NULL`, action.Token).Scan(&active); err != nil || active != 0 {
		t.Fatal("revoked reply remained active", active, err)
	}
}
