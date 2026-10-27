package telegram

import (
	"strings"
	"testing"
	"time"
)

// TestTelegramConversationCapture checks exact byte/part bounds and duplicate replay after reopening.
func TestTelegramConversationCapture(t *testing.T) {
	tests := []struct {
		name      string
		parts     []string
		wantBytes int
		wantParts int
	}{
		{"byte boundary", []string{strings.Repeat("a", 16384), strings.Repeat("b", 16384), strings.Repeat("c", 16384), strings.Repeat("d", 16381), "overflow"}, 65536, 4},
		{"part boundary", strings.Fields(strings.Repeat("part ", 33)), 159, 32},
		{"message boundary", []string{strings.Repeat("x", 16385)}, 0, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newTelegramConversationFixture(t)
			f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })

			for _, part := range test.parts {
				f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Text = part })
			}

			c := f.sessions(t)[0]

			var count int
			if err := f.store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_capture_parts WHERE conversation_id=?`, c.ID).Scan(&count); err != nil {
				t.Fatal(err)
			}

			if len(c.Capture) != test.wantBytes || count != test.wantParts {
				t.Fatal("capture boundary", len(c.Capture), count)
			}
		})
	}
}

// TestTelegramConversationReplay proves append, source acknowledgement and outbox rollback together.
func TestTelegramConversationReplay(t *testing.T) {
	f := newTelegramConversationFixture(t)
	env := f.env
	f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })

	input := f.input
	input.Update = telegramRetainedTestUpdate(t, 10, "one part")

	input.Action, input.Text = TelegramConversationCapture, "one part"
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := f.store.ApplyTelegramConversation(t.Context(), input, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := f.store.Close(); err != nil {
		t.Fatal(err)
	}

	f.store = openTelegramReceiverTestStore(t, env)
	if err := f.store.ApplyTelegramConversation(t.Context(), input, time.Now()); err != nil {
		t.Fatal(err)
	}

	if got := f.sessions(t)[0].Capture; got != "one part" {
		t.Fatal("duplicate append", got)
	}

	checkTelegramCaptureRollback(t, f)
}

func checkTelegramCaptureRollback(t *testing.T, f *telegramConversationFixture) {
	t.Helper()

	_, err := f.store.db.ExecContext(t.Context(), `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<1022)
INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,state,created_at,updated_at)
SELECT 'filled-'||x,12345,100001,0,100001,'filled-'||x,config_digest,'result','filled','pending',created_at,updated_at FROM n CROSS JOIN (SELECT * FROM telegram_outbox LIMIT 1)`)
	if err != nil {
		t.Fatal(err)
	}

	input := f.input
	input.Update = telegramRetainedTestUpdate(t, 11, "rolled back")

	input.Action, input.Text = TelegramConversationCapture, "rolled back"
	if err := f.store.RetainBatch(t.Context(), 12345, []TelegramReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := f.store.ApplyTelegramConversation(t.Context(), input, time.Now()); err == nil {
		t.Fatal("capacity overflow committed")
	}

	if f.sessions(t)[0].Capture != "one part" {
		t.Fatal("partial state committed")
	}

	var state string
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT state FROM telegram_updates WHERE update_id=11`).Scan(&state); err != nil || state != "pending" {
		t.Fatal("source acknowledged before response", state, err)
	}
}

// TestTelegramConversationCompletion keeps done distinct from generation and discards only canceled capture.
func TestTelegramConversationCompletion(t *testing.T) {
	for _, action := range []TelegramConversationAction{TelegramConversationDone, TelegramConversationCancel} {
		t.Run(string(action), func(t *testing.T) {
			f := newTelegramConversationFixture(t)
			f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })
			f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Text = "requirements" })
			f.apply(t, action, nil)

			c := f.sessions(t)[0]
			f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Conversation = c.ID; i.Text = "late" })

			after := f.sessions(t)[0]
			if c != after {
				t.Fatal("finished capture changed", c, after)
			}

			if action == TelegramConversationCancel && (c.Phase != "canceled" || c.Capture != "") {
				t.Fatal("cancellation retained capture", c)
			}

			if action == TelegramConversationDone && (c.Phase != "preparing" || c.Capture != "requirements") {
				t.Fatal("done did not retain capture", c)
			}
		})
	}
}

// TestTelegramConversationRevocation rechecks queued peer and project authority instead of frozen grants.
func TestTelegramConversationRevocation(t *testing.T) {
	tests := []struct {
		name   string
		revoke func(*TelegramConversationInput)
	}{
		{"removed project", func(i *TelegramConversationInput) { i.Projects = i.Projects[1:] }},
		{"disabled project", func(i *TelegramConversationInput) { i.Projects[0].Enabled = false }},
		{"project allowlist", func(i *TelegramConversationInput) { i.Config.AllowedUsers[0].ProjectIDs = []string{"beta"} }},
		{"user allowlist", func(i *TelegramConversationInput) { i.Config.AllowedUsers = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := newTelegramConversationFixture(t)
			f.apply(t, TelegramConversationStart, func(i *TelegramConversationInput) { i.Project = "alpha" })
			f.apply(t, TelegramConversationCapture, func(i *TelegramConversationInput) { i.Text = "denied"; test.revoke(i) })

			if f.sessions(t)[0].Capture != "" {
				t.Fatal("revoked authority used")
			}
		})
	}
}
