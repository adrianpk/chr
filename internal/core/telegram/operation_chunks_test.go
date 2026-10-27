package telegram

import (
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// TestTelegramOperationUnicodeParts bounds new views without automatic chunks while retaining an inspection reference.
func TestTelegramOperationUnicodeParts(t *testing.T) {
	for _, text := range []string{strings.Repeat("é", 12000), strings.Repeat("🦋", 8000), strings.Repeat("x", 100000)} {
		parts := tgmessage.ResponseParts(text, "ops/default/ticket/PRJ-1")
		if len(parts) != 1 {
			t.Fatal("invalid part count", len(parts))
		}

		for _, part := range parts {
			if !validTelegramReceiverText(part) || len(utf16.Encode([]rune(part))) > 1800 {
				t.Fatal("Unicode delivery bound exceeded")
			}
		}

		if !strings.Contains(parts[0], "ops/default/ticket/PRJ-1") {
			t.Fatal("shortened view reference lost")
		}
	}
}

// TestTelegramOperationPartOrder keeps later parts behind an unacknowledged or blocked predecessor.
func TestTelegramOperationPartOrder(t *testing.T) {
	store := openTelegramReceiverTestStore(t, telegramReceiverTestEnv(t))
	update := telegramRetainedTestUpdate(t, 1, "/tickets show")

	now := time.Now()
	if err := store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{update}, now); err != nil {
		t.Fatal(err)
	}

	if err := store.QueueReadResponse(t.Context(), update, strings.Repeat("🦋", 4000), telegramReceiverTestConfig().Digest(), now); err != nil {
		t.Fatal(err)
	}

	// Physical legacy rows still obey predecessor acknowledgement and blocking.
	if _, err := store.db.ExecContext(t.Context(), "UPDATE telegram_outbox SET text='legacy first part',part_count=3 WHERE event_key='update:1'"); err != nil {
		t.Fatal(err)
	}

	if _, err := store.db.ExecContext(t.Context(), `INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,part_index,part_count,state,created_at,updated_at)
 SELECT o.id||'-legacy-'||n.i,o.bot_id,o.chat_id,o.topic_id,o.user_id,o.event_key,o.config_digest,o.kind,'legacy later part',n.i,3,'pending',o.created_at,o.updated_at FROM telegram_outbox o JOIN (SELECT 2 i UNION ALL SELECT 3) n WHERE o.event_key='update:1' AND o.part_index=1`); err != nil {
		t.Fatal(err)
	}

	deliveries, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(deliveries) != 1 {
		t.Fatal("later parts bypassed first part", deliveries, err)
	}

	if err := store.ReserveDelivery(t.Context(), deliveries[0].ID, now); err != nil {
		t.Fatal(err)
	}

	if err := store.DeliverySent(t.Context(), deliveries[0].ID, 99, now); err != nil {
		t.Fatal(err)
	}

	checkTelegramOperationPartSuccessor(t, store, deliveries[0].ID, now)
}

func checkTelegramOperationPartSuccessor(t *testing.T, store *TelegramReceiverStore, firstID string, now time.Time) {
	t.Helper()

	next, err := store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(next) != 1 || next[0].ID == firstID {
		t.Fatal("part acknowledgement did not advance", next, err)
	}

	if err := store.BlockDelivery(t.Context(), next[0].ID, now); err != nil {
		t.Fatal(err)
	}

	next, err = store.DueDeliveries(t.Context(), 12345, now)
	if err != nil || len(next) != 0 {
		t.Fatal("blocked part released successors", next, err)
	}
}
