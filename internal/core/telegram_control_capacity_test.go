package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func fillTelegramOrdinaryIntake(t *testing.T, store *TelegramReceiverStore) {
	t.Helper()

	first := telegramRetainedTestUpdate(t, 0, "/help")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{first}, time.Now()); err != nil {
		t.Fatal(err)
	}

	var backlog []TelegramReceivedUpdate
	for id := 1; id < 991; id++ {
		backlog = append(backlog, telegramRetainedTestUpdate(t, int64(id), "/help"))
	}

	data, err := json.Marshal(backlog)
	if err != nil {
		t.Fatal(err)
	}
	// Seed the settled backlog together with its cursor. Actual intake still
	// exercises the final ordinary slot and every overflow/reserved boundary.
	err = store.transaction(t.Context(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), `INSERT INTO telegram_updates(bot_id,update_id,cursor_epoch,payload_digest,update_kind,chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at)
SELECT u.bot_id,json_extract(j.value,'$.UpdateID'),u.cursor_epoch,json_extract(j.value,'$.Digest'),'message',u.chat_id,u.topic_id,u.user_id,json_extract(j.value,'$.MessageID'),u.chat_type,u.received_at,u.payload,'pending','',u.updated_at FROM json_each(?) j JOIN telegram_updates u ON u.bot_id=12345 AND u.update_id=0`, string(data))
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(t.Context(), `UPDATE telegram_receivers SET next_update_id=991 WHERE bot_id=12345`)

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	last := telegramRetainedTestUpdate(t, 991, "/help")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{last}, time.Now()); err != nil {
		t.Fatal(err)
	}
}

// TestTelegramControlReserve preserves 32 control inputs through reopen even
// when ordinary intake is full; the 33rd is explicitly rejected with no effect.
func TestTelegramControlReserve(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)
	fillTelegramOrdinaryIntake(t, store)

	batch := []TelegramReceivedUpdate{telegramRetainedTestUpdate(t, 992, "prepare overflow")}
	for id := 993; id < 1026; id++ {
		batch = append(batch, telegramRetainedTestUpdate(t, int64(id), "/status"))
	}

	if err := store.RetainBatch(t.Context(), 12345, batch, time.Now()); err != nil {
		t.Fatal(err)
	}

	// Simulate the defaults assigned to already-pending input by migration.
	if _, err := store.db.ExecContext(t.Context(), `UPDATE telegram_updates SET processing_lane='ordinary',control_input=0 WHERE payload='/status'`); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)

	checkTelegramReservedInputs(t, store)

	ordinary, err := store.PendingApplicationInputs(t.Context(), 12345, TelegramProcessingOrdinary, time.Now())
	if err != nil || len(ordinary) != 1 || ordinary[0].UpdateID != 0 {
		t.Fatal("ordinary peer order was not preserved", ordinary, err)
	}

	deliveries, err := store.DueDeliveries(t.Context(), 12345, time.Now())
	if err != nil || len(deliveries) != 2 {
		t.Fatal("overflow lacks durable diagnostic", deliveries, err)
	}

	checkTelegramCapacityMessages(t, deliveries)
	checkTelegramControlIntakeCounts(t, store)
}

func checkTelegramReservedInputs(t *testing.T, store *TelegramReceiverStore) {
	t.Helper()

	controls, err := store.PendingApplicationInputs(t.Context(), 12345, TelegramProcessingUrgent, time.Now())
	if err != nil || len(controls) != 32 || controls[0].UpdateID != 993 || controls[31].UpdateID != 1024 {
		t.Fatal("reserved inputs lost or exceeded", controls, err)
	}
}

func checkTelegramControlIntakeCounts(t *testing.T, store *TelegramReceiverStore) {
	t.Helper()

	var pending, rejected int
	if err := store.db.QueryRowContext(t.Context(), `SELECT SUM(state='pending'),SUM(reason_code='intake_capacity') FROM telegram_updates`).Scan(&pending, &rejected); err != nil || pending != 1024 || rejected != 2 {
		t.Fatal("pending hard limit changed", pending, rejected, err)
	}
}

// TestTelegramControlStorageFailure proves a mixed batch cannot advance its
// cursor when persistence of an overflow diagnostic fails.
func TestTelegramControlStorageFailure(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
	fillTelegramOrdinaryIntake(t, store)

	if _, err := store.db.ExecContext(t.Context(), `CREATE TRIGGER fail_notice BEFORE INSERT ON telegram_outbox BEGIN SELECT RAISE(ABORT,'fixture storage failure'); END`); err != nil {
		t.Fatal(err)
	}

	batch := []TelegramReceivedUpdate{telegramRetainedTestUpdate(t, 992, "/status"), telegramRetainedTestUpdate(t, 993, "overflow")}
	if err := store.RetainBatch(t.Context(), 12345, batch, time.Now()); err == nil {
		t.Fatal("failed persistence was accepted")
	}

	state, err := store.Receiver(t.Context(), 12345)
	if err != nil || state.NextUpdateID == nil || *state.NextUpdateID != 992 {
		t.Fatal("failed batch advanced cursor", state, err)
	}

	var rows int
	if err := store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_updates WHERE update_id>=992`).Scan(&rows); err != nil || rows != 0 {
		t.Fatal("partial batch survived", rows, err)
	}
}

// TestTelegramControlModelCapacity shares two provider slots across independent
// interpretation and generation owners without spending a waiting source's retry.
func TestTelegramControlModelCapacity(t *testing.T) {
	f, request := telegramAuthoringModelFixture(t)

	updates := []TelegramReceivedUpdate{telegramRetainedTestUpdate(t, 900, "first intent"), telegramRetainedTestUpdate(t, 901, "next intent")}
	if err := f.store.RetainBatch(t.Context(), 12345, updates, time.Now()); err != nil {
		t.Fatal(err)
	}

	entered, release, finished := make(chan struct{}, 2), make(chan struct{}), make(chan error, 2)
	defer func() {
		close(release)

		for range 2 {
			if err := <-finished; err != nil {
				t.Error(err)
			}
		}
	}()

	invoke := func(ctx context.Context) (json.RawMessage, error) {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 60*time.Second {
			return nil, errors.New("missing model deadline")
		}

		entered <- struct{}{}

		<-release

		return json.RawMessage(`{}`), nil
	}
	go func() {
		_, err := f.store.ResolveTelegramInterpretation(t.Context(), f.env, updates[0], time.Now(), invoke)
		finished <- err
	}()
	go func() {
		_, err := f.store.ResolveTelegramOperationModel(t.Context(), f.env, request, time.Now(), invoke)
		finished <- err
	}()

	for range 2 {
		select {
		case <-entered:
		case <-time.After(3 * time.Second):
			t.Fatal("independent model calls did not enter")
		}
	}

	checkTelegramControlModelWait(t, f, updates[1])
}

func checkTelegramControlModelWait(t *testing.T, f *telegramConversationFixture, update TelegramReceivedUpdate) {
	t.Helper()

	other := openTelegramReceiverTestStore(t, f.env)

	_, err := other.ResolveTelegramInterpretation(t.Context(), f.env, update, time.Now().Add(time.Hour), func(context.Context) (json.RawMessage, error) {
		t.Error("third provider call entered")

		return json.RawMessage(`{}`), nil
	})
	if !errors.Is(err, ErrStructuredModelCapacity) {
		t.Fatal("third model call bypassed shared capacity", err)
	}

	var attempts int
	if err := other.db.QueryRowContext(t.Context(), `SELECT interpretation_attempt FROM telegram_updates WHERE update_id=901`).Scan(&attempts); err != nil || attempts != 0 {
		t.Fatal("waiting for a slot spent retry budget", attempts, err)
	}
}

// TestTelegramControlContextGuard prevents a model result from following a new
// focus or capture revision introduced while the model was outside the DB lock.
func TestTelegramControlContextGuard(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })

	input := telegramOperationActionInput(t, f, "capture original context")

	snapshot, err := f.store.TelegramIntentContext(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}

	input.ContextDigest = snapshot.Digest
	if _, err := f.store.ResolveTelegramConversationTarget(t.Context(), input); err != nil {
		t.Fatal(err)
	}

	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "beta" })

	if _, err := f.store.ResolveTelegramConversationTarget(t.Context(), input); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("model result followed changed context", err)
	}
}

func checkTelegramCapacityMessages(t *testing.T, deliveries []TelegramReceiverDelivery) {
	t.Helper()

	for _, delivery := range deliveries {
		if delivery.Presentation == nil || delivery.Presentation.Family != "42" || !strings.Contains(delivery.Text, "This request was not accepted") {
			t.Fatal("capacity rejection was presented as accepted work", delivery)
		}
	}
}
