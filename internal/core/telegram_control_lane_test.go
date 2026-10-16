package core

import (
	"encoding/json"
	"testing"
	"time"
)

func telegramPendingControlLaneFixture(t *testing.T, operation, action string) (*telegramConversationFixture, TelegramAction) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })

	input := telegramOperationActionInput(t, f, "control preview")
	input.Conversation = f.sessions(t)[0].ID

	target, err := f.store.ResolveTelegramOperationTarget(t.Context(), input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	args, err := json.Marshal(map[string]string{"action": action})
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReservePreparedTelegramOperation(t.Context(), TelegramOperationRequest{Source: input.Update, Operator: "maintainer", Target: target, OperationID: operation, Arguments: args}, json.RawMessage(`{}`), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	offer := TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{action}, Text: "Confirm the observed operation.", ExpiresAt: time.Now().Add(24 * time.Hour)}

	a, err := f.store.OfferTelegramAction(t.Context(), offer, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	return f, a
}

func telegramControlLaneFixture(t *testing.T, operation, action string) (*telegramConversationFixture, TelegramAction) {
	t.Helper()

	f, a := telegramPendingControlLaneFixture(t, operation, action)

	deliveries, err := f.store.DueDeliveries(t.Context(), 12345, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	for _, delivery := range deliveries {
		if err := f.store.ReserveDelivery(t.Context(), delivery.ID, time.Now()); err != nil {
			t.Fatal(err)
		}

		if err := f.store.DeliverySent(t.Context(), delivery.ID, 100+int64(len(delivery.ActionToken)), time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	return f, a
}

// TestTelegramControlLaneParity derives callback, explicit answer and reply
// priority from the same peer-owned prompt. Continuation and authoring retain
// ordinary slots because they can run a worker or model.
func TestTelegramControlLaneParity(t *testing.T) {
	tests := []struct {
		name, operation, action string
		lane                    TelegramProcessingLane
		control                 bool
	}{
		{"pause", "work.control", "pause", TelegramProcessingUrgent, true},
		{"cancel", "work.control", "cancel", TelegramProcessingUrgent, true},
		{"reconcile", "work.control", "reconcile", TelegramProcessingUrgent, true},
		{"resume", "work.control", "resume", TelegramProcessingDecision, true},
		{"approve", "work.control", "approve", TelegramProcessingDecision, true},
		{"grant", "work.control", "grant", TelegramProcessingDecision, true},
		{"decision", "work.control", "answer_decision", TelegramProcessingDecision, true},
		{"continue", "work.control", "continue", TelegramProcessingOrdinary, true},
		{"authoring", "authoring.quiz", "confirm", TelegramProcessingOrdinary, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, action := telegramControlLaneFixture(t, tt.operation, tt.action)
			for index, kind := range []string{"callback", "text", "reply"} {
				input := telegramRetainedTestUpdate(t, 50+int64(index), "/answer "+action.Token+" "+tt.action)

				switch kind {
				case "callback":
					input.Kind, input.Text = "callback_query", "tga:"+action.Token+":"+tt.action
				case "reply":
					input.Text, input.ReplyToMessageID = tt.action, 132
				}

				checkTelegramControlInputLane(t, f.store, input, tt.lane, tt.control)
			}
		})
	}
}

func checkTelegramControlInputLane(t *testing.T, store *TelegramReceiverStore, input TelegramReceivedUpdate, lane TelegramProcessingLane, control bool) {
	t.Helper()

	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{input}, time.Now()); err != nil {
		t.Fatal(err)
	}

	var (
		got      TelegramProcessingLane
		reserved bool
	)
	if err := store.db.QueryRowContext(t.Context(), `SELECT processing_lane,control_input FROM telegram_updates WHERE update_id=?`, input.UpdateID).Scan(&got, &reserved); err != nil || got != lane || reserved != control {
		t.Fatal("input priority diverged from retained action", got, reserved, err)
	}
}

// TestTelegramControlDeliveryPriority sends controls before older ordinary
// output without changing the durable per-event part ordering.
func TestTelegramControlDeliveryPriority(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))

	updates := []TelegramReceivedUpdate{telegramRetainedTestUpdate(t, 1, "/help"), telegramRetainedTestUpdate(t, 2, "/status"), telegramRetainedTestUpdate(t, 3, "/approve workflow")}
	if err := store.RetainBatch(t.Context(), 12345, updates, time.Now()); err != nil {
		t.Fatal(err)
	}

	for _, u := range updates {
		if err := store.QueueReadResponse(t.Context(), u, u.Text, telegramReceiverTestConfig().Digest(), time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	due, err := store.DueDeliveries(t.Context(), 12345, time.Now())
	if err != nil || len(due) != 3 {
		t.Fatal(due, err)
	}

	for i, want := range []string{"/status", "/approve workflow", "/help"} {
		if due[i].Text != want {
			t.Fatal("control output followed ordinary backlog", due)
		}
	}
}
