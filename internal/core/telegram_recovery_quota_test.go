package core

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// TestTelegramRecoveryQuota checks actual receipt and outbound boundaries. A
// rejected batch/response leaves both the source and sparse cursor unchanged.
func TestTelegramRecoveryQuota(t *testing.T) {
	for _, kind := range []string{"receipts", "outbox"} {
		t.Run(kind, func(t *testing.T) { checkTelegramRecoveryQuota(t, kind) })
	}
}

func checkTelegramRecoveryQuota(t *testing.T, kind string) {
	t.Helper()

	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))

	now := time.Now()
	if kind == "receipts" {
		fillTelegramReceiptQuota(t, store, now)
	} else {
		fillTelegramOutboxQuota(t, store, now)
	}

	pending, err := store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 1 || pending[0].UpdateID != 10000 {
		t.Fatal("quota rejection acknowledged unresolved input", pending, err)
	}

	state, err := store.Receiver(t.Context(), 12345)
	if err != nil || state.NextUpdateID == nil || *state.NextUpdateID != 10001 {
		t.Fatal("quota rejection moved the cursor", state, err)
	}
}

// TestTelegramRecoveryPayload reaches the real 64 MiB logical ceiling, then
// proves a second intake cannot exceed it or acknowledge its unretained source.
func TestTelegramRecoveryPayload(t *testing.T) {
	if testing.Short() {
		t.Skip("64 MiB persistence boundary is selected in focused recovery checks")
	}

	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
	now := time.Now()

	for id := range int64(8) {
		update := telegramRetainedTestUpdate(t, id, "/help")
		if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, now); err != nil {
			t.Fatal(err)
		}
	}

	seedTelegramRecoveryPayload(t, store, now)

	boundary := telegramRetainedTestUpdate(t, 8, "/status")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{boundary}, now); err != nil {
		t.Fatal("exact logical boundary was refused", err)
	}

	overflow := telegramRetainedTestUpdate(t, 9, "/help")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{overflow}, now); err == nil {
		t.Fatal("logical transport payload exceeded 64 MiB")
	}

	state, err := store.Receiver(t.Context(), 12345)
	if err != nil || state.NextUpdateID == nil || *state.NextUpdateID != 9 {
		t.Fatal("payload overflow acknowledged unretained input", state, err)
	}
}

func seedTelegramRecoveryPayload(t *testing.T, store *TelegramReceiverStore, now time.Time) {
	t.Helper()

	tx, err := store.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			t.Error(err)
		}
	}()

	for index := range 64 {
		size := 1 << 20
		if index == 63 {
			size -= 8*len("/help") + 64*len("{}") + len("/status")
		}

		encoded := `"` + strings.Repeat("x", size-2) + `"`

		_, err := tx.ExecContext(t.Context(), `INSERT INTO telegram_requests(id,bot_id,update_id,operation_ordinal,operator_id,project_id,project_binding_digest,namespace,operation_id,arguments,argument_digest,prepared_input,domain_guard,state,created_at,updated_at) VALUES(?,12345,?,?,'maintainer','alpha',?,'default','records.show','{}',?,?,?,'confirmed',?,?)`, fmt.Sprintf("quota-request-%d", index), index/8, index%8, strings.Repeat("a", 64), strings.Repeat("b", 64), encoded, strings.Repeat("c", 64), database.FormatTime(now), database.FormatTime(now))
		if err != nil {
			t.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func fillTelegramReceiptQuota(t *testing.T, store *TelegramReceiverStore, now time.Time) {
	t.Helper()

	execTelegramRecoverySQL(t, store, `WITH RECURSIVE n(id) AS (SELECT 0 UNION ALL SELECT id+1 FROM n WHERE id<9999) INSERT INTO telegram_updates(bot_id,update_id,cursor_epoch,payload_digest,update_kind,chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at) SELECT 12345,id,1,?,'ignored',100001,0,100001,1,'private',?,'','ignored','unsupported',? FROM n`, strings.Repeat("a", 64), database.FormatTime(now), database.FormatTime(now))

	pending := telegramRetainedTestUpdate(t, 10000, "/status")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{pending}, now); err != nil {
		t.Fatal("receipt ceiling consumed the pending reserve", err)
	}

	overflow := telegramRetainedTestUpdate(t, 10001, "/status")

	overflow.Kind, overflow.Text = "ignored", ""
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{overflow}, now); err == nil {
		t.Fatal("receipt ceiling was exceeded")
	}
}

func fillTelegramOutboxQuota(t *testing.T, store *TelegramReceiverStore, now time.Time) {
	t.Helper()

	update := telegramRetainedTestUpdate(t, 10000, "/status")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	execTelegramRecoverySQL(t, store, `WITH RECURSIVE n(id) AS (SELECT 1 UNION ALL SELECT id+1 FROM n WHERE id<1024) INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,state,created_at,updated_at) SELECT 'quota-'||id,12345,100001,0,100001,'quota:'||id,?,'result','unresolved','blocked',?,? FROM n`, telegramReceiverTestConfig().Digest(), database.FormatTime(now.Add(-8*24*time.Hour)), database.FormatTime(now.Add(-8*24*time.Hour)))

	if _, err := store.PrepareTelegramPoll(t.Context(), 12345, now); err != nil {
		t.Fatal(err)
	}

	if err := store.QueueReadResponse(t.Context(), update, "new result", telegramReceiverTestConfig().Digest(), now); err == nil {
		t.Fatal("retention evicted unresolved output to admit another response")
	}
}
