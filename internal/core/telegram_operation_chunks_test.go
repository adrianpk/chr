package core

import (
	"strings"
	"testing"
	"time"
)

// TestTelegramOperationUnicodeParts bounds astral runes, chunk counts and stable output references.
func TestTelegramOperationUnicodeParts(t *testing.T) {
	for _, text := range []string{strings.Repeat("é", 12000), strings.Repeat("🦋", 8000), strings.Repeat("x", 100000)} {
		parts := TelegramResponseParts(text, "ops/default/ticket/PRJ-1")
		if len(parts) == 0 || len(parts) > 8 {
			t.Fatal("invalid part count", len(parts))
		}

		for _, part := range parts {
			if !validTelegramReceiverText(part) {
				t.Fatal("Unicode delivery bound exceeded")
			}
		}

		if !strings.Contains(parts[0], "ops/default/ticket/PRJ-1") {
			t.Fatal("multi-part reference lost")
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
