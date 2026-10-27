package telegram

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestTelegramConversationRootAlias invalidates a changed registry path even when it resolves to the same physical root.
func TestTelegramConversationRootAlias(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationStart, func(i *ConversationInput) { i.Project = "alpha" })

	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(f.input.Projects[0].Path, alias); err != nil {
		t.Fatal(err)
	}

	f.input.Projects[0].Path = alias
	f.apply(t, ConversationCapture, func(i *ConversationInput) { i.Text = "changed binding" })

	if !strings.Contains(f.lastResponse(t), "root changed") || f.sessions(t)[0].Capture != "" {
		t.Fatal("registry alias change accepted")
	}
}

// TestTelegramConversationAcknowledgement proves reply retention failure rolls back definite acknowledgement.
func TestTelegramConversationAcknowledgement(t *testing.T) {
	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationStart, func(i *ConversationInput) { i.Project = "alpha" })

	var id string
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT id FROM telegram_outbox`).Scan(&id); err != nil {
		t.Fatal(err)
	}

	if err := f.store.ReserveDelivery(t.Context(), id, time.Now()); err != nil {
		t.Fatal(err)
	}

	_, err := f.store.db.ExecContext(t.Context(), `CREATE TEMP TRIGGER reply_failure BEFORE INSERT ON telegram_replies BEGIN SELECT RAISE(ABORT,'injected reply retention failure'); END`)
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.DeliverySent(t.Context(), id, 800, time.Now()); err == nil {
		t.Fatal("injected retention failure ignored")
	}

	var state string
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT state FROM telegram_outbox WHERE id=?`, id).Scan(&state); err != nil || state != "sending" {
		t.Fatal("ack committed without association", state, err)
	}

	if _, err := f.store.db.ExecContext(t.Context(), `DROP TRIGGER reply_failure`); err != nil {
		t.Fatal(err)
	}

	f.apply(t, ConversationCapture, func(i *ConversationInput) { i.Text = "new revision" })

	if err := f.store.DeliverySent(t.Context(), id, 800, time.Now()); err != nil {
		t.Fatal(err)
	}

	f.apply(t, ConversationCapture, func(i *ConversationInput) { i.Text = "late answer"; i.Update.ReplyToMessageID = 800 })

	if f.sessions(t)[0].Capture != "new revision" {
		t.Fatal("late acknowledgement revived stale prompt")
	}
}
