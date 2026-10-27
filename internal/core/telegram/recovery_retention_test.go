package telegram

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// TestTelegramRecoverySparse retires only exact acknowledged IDs, retains local
// pending input, and resets a stale remote offset once without a low watermark.
func TestTelegramRecoverySparse(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	store := openTelegramReceiverTestStore(t, env)

	old := time.Now().UTC().Add(-8 * 24 * time.Hour)
	for _, id := range []int64{3, 4, 9000} {
		seedTelegramRecoveryDelivered(t, store, id, old)
	}

	pending := telegramRetainedTestUpdate(t, 30, "pending local work")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{pending}, old); err != nil {
		t.Fatal(err)
	}

	for range 2 {
		checkTelegramRecoveryGap(t, store)
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store = openTelegramReceiverTestStore(t, env)

	replay := []TelegramReceivedUpdate{telegramRetainedTestUpdate(t, 3, "changed retired payload"), telegramRetainedTestUpdate(t, 5, "new sparse input"), telegramRetainedTestUpdate(t, 9000, "/status")}
	if err := store.RetainBatch(t.Context(), 12345, replay, time.Now()); err != nil {
		t.Fatal(err)
	}

	checkTelegramRecoverySparseResult(t, store)
	// The successful recovery poll can share the epoch marker's second. Another
	// outage must still create a fresh epoch exactly once.
	for range 2 {
		checkTelegramRecoveryNextGap(t, store)
	}
}

func checkTelegramRecoverySparseResult(t *testing.T, store *TelegramReceiverStore) {
	t.Helper()

	updates, err := store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(updates) != 2 || updates[0].UpdateID != 5 || updates[1].UpdateID != 30 {
		t.Fatal("retirement recreated old work or lost a sparse source", updates, err)
	}

	var intervals string
	if err := store.db.QueryRowContext(t.Context(), `SELECT retired_update_ranges FROM telegram_receivers WHERE bot_id=12345`).Scan(&intervals); err != nil {
		t.Fatal(err)
	}

	if intervals != "[[3,4],[9000,9000]]" {
		t.Fatal("retirement bridged an unobserved gap", intervals)
	}
}

func checkTelegramRecoveryGap(t *testing.T, store *TelegramReceiverStore) {
	t.Helper()

	state, err := store.PrepareTelegramPoll(t.Context(), 12345, time.Now())
	if err != nil || state.Epoch != 2 || state.NextUpdateID != nil || state.RetentionGapAt.IsZero() {
		t.Fatal("retention gap was missed or repeatedly reset", state, err)
	}
}

func seedTelegramRecoveryDelivered(t *testing.T, store *TelegramReceiverStore, id int64, now time.Time) {
	t.Helper()

	update := telegramRetainedTestUpdate(t, id, "/status")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	if err := store.QueueReadResponse(t.Context(), update, "recorded result", telegramReceiverTestConfig().Digest(), now); err != nil {
		t.Fatal(err)
	}

	due, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(due) != 1 {
		t.Fatal(due, err)
	}

	if err := store.ReserveDelivery(t.Context(), due[0].ID, now); err != nil {
		t.Fatal(err)
	}

	if err := store.DeliverySent(t.Context(), due[0].ID, id+1, now); err != nil {
		t.Fatal(err)
	}
}

// TestTelegramRecoveryIntervals leaves a source intact when the exact interval
// ceiling prevents retirement, while an adjacent known ID can still coalesce.
func TestTelegramRecoveryIntervals(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))

	pairs := make([][2]int64, 1024)
	for i := range pairs {
		pairs[i] = [2]int64{int64(i * 2), int64(i * 2)}
	}

	encoded, err := json.Marshal(pairs)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.db.ExecContext(t.Context(), `UPDATE telegram_receivers SET retired_update_ranges=?`, string(encoded)); err != nil {
		t.Fatal(err)
	}

	old := time.Now().Add(-8 * 24 * time.Hour)
	seedTelegramRecoveryDelivered(t, store, 2047, old)
	seedTelegramRecoveryDelivered(t, store, 99999, old)

	if _, err := store.PrepareTelegramPoll(t.Context(), 12345, time.Now()); err != nil {
		t.Fatal(err)
	}

	var kept int
	if err := store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_updates WHERE update_id IN (2047,99999)`).Scan(&kept); err != nil || kept != 1 {
		t.Fatal("interval pressure lost an unretired identity", kept, err)
	}

	if err := store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_updates WHERE update_id=99999`).Scan(&kept); err != nil || kept != 1 {
		t.Fatal("unrepresentable retirement was acknowledged", kept, err)
	}
}

// TestTelegramRecoveryCompaction preserves current focus, unresolved effects,
// live prompts and unsent output even when their payloads are older than seven days.
func TestTelegramRecoveryCompaction(t *testing.T) {
	for _, guard := range []string{"terminal", "focused", "unresolved", "prompt", "outbox"} {
		t.Run(guard, func(t *testing.T) { checkTelegramRecoveryCompaction(t, guard) })
	}
}

func checkTelegramRecoveryCompaction(t *testing.T, guard string) {
	t.Helper()

	f, request := telegramAuthoringModelFixture(t)
	old := database.FormatTime(time.Now().Add(-8 * 24 * time.Hour))

	execTelegramRecoverySQL(t, f.store, `UPDATE telegram_requests SET state='confirmed',domain_receipt_ref='canonical:retained'`)
	execTelegramRecoverySQL(t, f.store, `UPDATE telegram_updates SET state='processed',updated_at=?`, old)
	execTelegramRecoverySQL(t, f.store, `UPDATE telegram_conversations SET phase='completed',capture='retained capture',updated_at=?`, old)
	execTelegramRecoverySQL(t, f.store, `UPDATE telegram_peers SET focused_conversation_id=NULL`)
	execTelegramRecoverySQL(t, f.store, `UPDATE telegram_outbox SET state='delivered',updated_at=?`, old)
	protectTelegramRecoveryCompaction(t, f, request, guard)

	if _, err := f.store.PrepareTelegramPoll(t.Context(), 12345, time.Now()); err != nil {
		t.Fatal(err)
	}

	var prepared, capture, receipt string

	err := f.store.db.QueryRowContext(t.Context(), `SELECT COALESCE(r.prepared_input,''),c.capture,COALESCE(r.domain_receipt_ref,'') FROM telegram_requests r JOIN telegram_conversations c ON c.id=r.conversation_id WHERE r.id=?`, request.ID).Scan(&prepared, &capture, &receipt)
	if err != nil || receipt != "canonical:retained" {
		t.Fatal("compaction removed a canonical receipt", receipt, err)
	}

	if guard == "terminal" {
		if prepared != "" || capture != "" {
			t.Fatal("eligible terminal payload was not compacted")
		}
	} else if prepared == "" || capture != "retained capture" {
		t.Fatal("compaction discarded unresolved authority", guard)
	}
}

func protectTelegramRecoveryCompaction(t *testing.T, f *telegramConversationFixture, request TelegramOperationRequest, guard string) {
	t.Helper()

	switch guard {
	case "focused":
		execTelegramRecoverySQL(t, f.store, `UPDATE telegram_peers SET focused_conversation_id=?`, request.Target.ConversationID)
	case "unresolved":
		execTelegramRecoverySQL(t, f.store, `UPDATE telegram_requests SET state='applying'`)
	case "outbox":
		execTelegramRecoverySQL(t, f.store, `UPDATE telegram_outbox SET state='blocked'`)
	case "prompt":
		execTelegramRecoverySQL(t, f.store, `INSERT INTO telegram_actions(token,request_id,bot_id,chat_id,topic_id,user_id,conversation_revision,guard_digest,allowed_choices,state,expires_at,created_at,updated_at) VALUES(?,?,12345,100001,0,100001,1,?,'["confirm"]','pending',?,?,?)`, strings.Repeat("a", 32), request.ID, strings.Repeat("b", 64), database.FormatTime(time.Now().Add(time.Hour)), database.FormatTime(time.Now()), database.FormatTime(time.Now()))
	}
}

func execTelegramRecoverySQL(t *testing.T, store *TelegramReceiverStore, query string, args ...any) {
	t.Helper()

	if _, err := store.db.ExecContext(t.Context(), query, args...); err != nil {
		t.Fatal(err)
	}
}

func checkTelegramRecoveryNextGap(t *testing.T, store *TelegramReceiverStore) {
	t.Helper()

	state, err := store.PrepareTelegramPoll(t.Context(), 12345, time.Now().Add(25*time.Hour))
	if err != nil || state.Epoch != 3 || state.NextUpdateID != nil {
		t.Fatal("a later outage could not establish one fresh epoch", state, err)
	}
}
