package core

import (
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// TestTelegramRecoverySchema upgrades retained receiver state, proves a failed
// additive migration rolls back, and refuses a future schema without takeover.
func TestTelegramRecoverySchema(t *testing.T) {
	env := telegramReceiverTestEnv(t)
	assets := telegramRecoveryMigrationAssets(t)
	path := "assets/migration/sqlite/20261017000023-add_telegram_recovery.sql"
	prior := telegramRecoveryMigrationAssets(t)
	delete(prior, path)

	old, err := NewTelegramReceiverStore(t.Context(), env, prior)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := old.Close(); err != nil {
			t.Error(err)
		}
	})

	now := database.FormatTime(time.Now())
	execTelegramRecoverySQL(t, old, `INSERT INTO telegram_receivers(bot_id,config_digest,next_update_id,retry_attempt,retry_cycle,total_retry_attempts,updated_at) VALUES(12345,?,501,5,7,11,?)`, telegramReceiverTestConfig().Digest(), now)
	execTelegramRecoverySQL(t, old, `INSERT INTO telegram_updates(bot_id,update_id,cursor_epoch,payload_digest,update_kind,chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at) VALUES(12345,499,1,?,'message',100001,0,100001,1,'private',?,'/status','pending','',?)`, strings.Repeat("a", 64), now, now)

	broken := telegramRecoveryMigrationAssets(t)

	broken[path] = &fstest.MapFile{Data: []byte("-- +migrate Up\nALTER TABLE telegram_receivers ADD COLUMN retry_actor TEXT; invalid SQL;\n-- +migrate Down\nALTER TABLE telegram_receivers DROP COLUMN retry_actor;\n")}
	if migrated, err := NewTelegramReceiverStore(t.Context(), env, broken); err == nil {
		t.Fatal("failed recovery migration was accepted", migrated.Close())
	}

	var columns int
	if err := old.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM pragma_table_info('telegram_receivers') WHERE name='retry_actor'`).Scan(&columns); err != nil || columns != 0 {
		t.Fatal("partial recovery schema remained", columns, err)
	}

	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	store := openTelegramReceiverTestStore(t, env)
	checkTelegramRecoveryUpgraded(t, store)
	execTelegramRecoverySQL(t, store, `INSERT INTO schema_migrations VALUES('99999999999999','future','2026-01-01T00:00:00Z')`)

	if newer, err := NewTelegramReceiverStore(t.Context(), env, assets); err == nil {
		t.Fatal("future schema accepted", newer.Close())
	}
}

func checkTelegramRecoveryUpgraded(t *testing.T, store *TelegramReceiverStore) {
	t.Helper()

	state, err := store.Receiver(t.Context(), 12345)
	if err != nil || state.RetryAttempt != 5 || state.RetryCycle != 7 || state.TotalRetryAttempts != 11 || state.NextUpdateID == nil || *state.NextUpdateID != 501 {
		t.Fatal("upgrade reset receiver history", state, err)
	}

	pending, err := store.PendingUpdates(t.Context(), 12345)
	if err != nil || len(pending) != 1 || pending[0].UpdateID != 499 {
		t.Fatal("upgrade lost pending input", pending, err)
	}
}

func telegramRecoveryMigrationAssets(t *testing.T) fstest.MapFS {
	t.Helper()

	source := os.DirFS("../../cmd/chr")

	entries, err := fs.ReadDir(source, "assets/migration/sqlite")
	if err != nil {
		t.Fatal(err)
	}

	assets := fstest.MapFS{}

	for _, entry := range entries {
		path := "assets/migration/sqlite/" + entry.Name()

		data, err := fs.ReadFile(source, path)
		if err != nil {
			t.Fatal(err)
		}

		assets[path] = &fstest.MapFile{Data: data}
	}

	return assets
}

// TestTelegramRecoveryMalformed fails closed on corrupt retired-ID metadata;
// no source or cursor can be changed by a partially successful maintenance pass.
func TestTelegramRecoveryMalformed(t *testing.T) {
	tests := []struct{ name, encoded string }{
		{"null", "null"}, {"shape", "[[]]"}, {"negative", "[[-1,0]]"}, {"overlap", "[[1,2],[2,3]]"}, {"adjacent", "[[1,2],[3,4]]"}, {"overflow", "[[9223372036854775807,9223372036854775807]]"}, {"size", strings.Repeat(" ", 65537)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
			execTelegramRecoverySQL(t, store, `UPDATE telegram_receivers SET retired_update_ranges=?,next_update_id=41`, test.encoded)

			if _, err := store.PrepareTelegramPoll(t.Context(), 12345, time.Now()); err == nil {
				t.Fatal("malformed retired identity metadata was accepted")
			}

			state, err := store.Receiver(t.Context(), 12345)
			if err != nil || state.NextUpdateID == nil || *state.NextUpdateID != 41 {
				t.Fatal("failed maintenance moved the cursor", state, err)
			}
		})
	}
}

// TestTelegramRecoveryPending preserves an aged full pending inbox. New control
// overflow becomes a durable rejection instead of evicting unresolved input.
func TestTelegramRecoveryPending(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
	old := database.FormatTime(time.Now().Add(-8 * 24 * time.Hour))
	execTelegramRecoverySQL(t, store, `WITH RECURSIVE n(id) AS (SELECT 1 UNION ALL SELECT id+1 FROM n WHERE id<1024) INSERT INTO telegram_updates(bot_id,update_id,cursor_epoch,payload_digest,update_kind,chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at,processing_lane,control_input) SELECT 12345,id,1,?,'message',100001,0,100001,id,'private',?,'/status','pending','',?,CASE WHEN id>992 THEN 'urgent' ELSE 'ordinary' END,id>992 FROM n`, strings.Repeat("a", 64), old, old)

	if _, err := store.PrepareTelegramPoll(t.Context(), 12345, time.Now()); err != nil {
		t.Fatal(err)
	}

	update := telegramRetainedTestUpdate(t, 1025, "/pause run")
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	var pending int
	if err := store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_updates WHERE state='pending'`).Scan(&pending); err != nil || pending != 1024 {
		t.Fatal("retention evicted pending input", pending, err)
	}

	var state, reason string
	if err := store.db.QueryRowContext(t.Context(), `SELECT state,reason_code FROM telegram_updates WHERE update_id=1025`).Scan(&state, &reason); err != nil || state != "rejected" || reason != "intake_capacity" {
		t.Fatal("full inbox accepted a control effect", state, reason, err)
	}
}
