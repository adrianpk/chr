package telegram

import (
	"errors"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// TestTelegramControlReplyBinding keeps an early reply pending until the sender
// commits its exact prompt association, without interpreting it as capture text.
func TestTelegramControlReplyBinding(t *testing.T) {
	tests := []struct {
		name, operation, action string
		lane                    ProcessingLane
		control                 bool
	}{
		{"stop", "work.control", "pause", ProcessingUrgent, true},
		{"continuation", "work.control", "continue", ProcessingOrdinary, true},
		{"authoring", "authoring.quiz", "confirm", ProcessingOrdinary, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, action := telegramPendingControlLaneFixture(t, tt.operation, tt.action)
			delivery := telegramPendingActionDelivery(t, f.store, action.Token)
			input := telegramRetainedTestUpdate(t, f.next, tt.action)
			input.ReplyToMessageID = 777
			checkTelegramControlInputLane(t, f.store, input, ProcessingDecision, false)
			checkTelegramEarlyReply(t, f.store, input)

			if err := f.store.DeliverySent(t.Context(), delivery.ID, 777, time.Now()); err != nil {
				t.Fatal(err)
			}

			checkTelegramControlInputLane(t, f.store, input, tt.lane, tt.control)

			answerResolution, err := f.store.ResolveActionAnswer(t.Context(), input)

			answer, recognized := answerResolution.Answer, answerResolution.Recognized
			if err != nil || !recognized || answer.Token != action.Token || answer.Choice != tt.action {
				t.Fatal("definite prompt acknowledgement did not release the exact answer", answer, err)
			}
		})
	}
}

func telegramPendingActionDelivery(t *testing.T, store *ReceiverStore, token string) ReceiverDelivery {
	t.Helper()

	deliveries, err := store.DueDeliveries(t.Context(), 12345, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	for _, delivery := range deliveries {
		if delivery.ActionToken != token {
			continue
		}

		if err := store.ReserveDelivery(t.Context(), delivery.ID, time.Now()); err != nil {
			t.Fatal(err)
		}

		return delivery
	}

	t.Fatal("action delivery unavailable")

	return ReceiverDelivery{}
}

func checkTelegramEarlyReply(t *testing.T, store *ReceiverStore, update ReceivedUpdate) {
	t.Helper()

	answerResolution2, err := store.ResolveActionAnswer(t.Context(), update)

	recognized := answerResolution2.Recognized
	if !recognized || !errors.Is(err, ErrReplyPending) {
		t.Fatal("early reply escaped its pending prompt", recognized, err)
	}

	update.Peer.UserID++

	answerResolution3, answerResolution3Error := store.ResolveActionAnswer(t.Context(), update)

	recognized, err = answerResolution3.Recognized, answerResolution3Error
	if !recognized || !errors.Is(err, ErrActionUnavailable) {
		t.Fatal("another peer borrowed the sender's prompt", recognized, err)
	}
}

// TestTelegramControlReplyExhaustion cannot defer a reply forever after the
// fifth send loses its acknowledgement and its retained deadline has elapsed.
func TestTelegramControlReplyExhaustion(t *testing.T) {
	f, action := telegramPendingControlLaneFixture(t, "work.control", "pause")
	delivery := telegramPendingActionDelivery(t, f.store, action.Token)
	update := telegramRetainedTestUpdate(t, f.next, "pause")
	update.ReplyToMessageID = 777

	for _, test := range []struct {
		name    string
		delay   time.Duration
		failure error
	}{
		{"active", time.Minute, ErrReplyPending},
		{"exhausted", -time.Minute, ErrActionUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := f.store.db.ExecContext(t.Context(), `UPDATE telegram_outbox SET attempts=5,retry_at=? WHERE id=?`, database.FormatTime(time.Now().Add(test.delay)), delivery.ID); err != nil {
				t.Fatal(err)
			}

			answerResolution4, err := f.store.ResolveActionAnswer(t.Context(), update)

			recognized := answerResolution4.Recognized
			if !recognized || !errors.Is(err, test.failure) {
				t.Fatal("send deadline lost its reply boundary", recognized, err)
			}
		})
	}
}

// TestTelegramControlReplyPriority keeps unresolved early replies from hiding a
// real decision beyond the bounded scan, even while their sender remains active.
func TestTelegramControlReplyPriority(t *testing.T) {
	f, action := telegramPendingControlLaneFixture(t, "work.control", "resume")
	telegramPendingActionDelivery(t, f.store, action.Token)

	var batch []ReceivedUpdate

	for index := range 60 {
		update := telegramRetainedTestUpdate(t, f.next+int64(index), "early reply")
		update.ReplyToMessageID = 777 + int64(index)
		batch = append(batch, update)
	}

	decision := telegramRetainedTestUpdate(t, f.next+60, "tga:"+action.Token+":resume")
	decision.Kind = "callback_query"

	batch = append(batch, decision)
	if err := f.store.RetainBatch(t.Context(), 12345, batch, time.Now()); err != nil {
		t.Fatal(err)
	}

	inputs, err := f.store.PendingApplicationInputs(t.Context(), 12345, ProcessingDecision, time.Now())
	if err != nil || len(inputs) != 50 || inputs[0].UpdateID != decision.UpdateID {
		t.Fatal("early reply waiters hid an authorized decision", inputs, err)
	}
}
