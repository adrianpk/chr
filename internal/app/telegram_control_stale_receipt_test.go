package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/core/database"
)

type telegramStaleControlFixture struct {
	owner      *testing.T
	receiver   *telegramConversationAppFixture
	operations service
	id         string
}

// TestTelegramControlStaleReceipt recovers a committed control after prompt
// expiry or conversation replacement without changing newer native or peer state.
func TestTelegramControlStaleReceipt(t *testing.T) {
	telegramControlParallel(t)

	for _, stale := range []string{"expired", "replaced", "superseded"} {
		t.Run(stale, func(t *testing.T) {
			f, operations, ref := telegramWorkAdmitted(t)
			fixture := telegramStaleControlFixture{owner: t, receiver: f, operations: operations, id: ref.Invocation.ID}

			for _, kind := range []string{"text", "callback", "reply"} {
				for _, accepted := range []bool{false, true} {
					name := kind + "/unaccepted"
					if accepted {
						name = kind + "/accepted"
					}

					t.Run(name, func(t *testing.T) { checkTelegramStaleReceipt(t, fixture, stale, kind, accepted) })
				}
			}
		})
	}
}

func checkTelegramStaleReceipt(t *testing.T, fixture telegramStaleControlFixture, stale, kind string, accepted bool) {
	t.Helper()

	f, operations := fixture.receiver, fixture.operations
	f.receiver.operations = operations
	token := telegramControlOffer(t, f, "/pause "+fixture.id)

	update := telegramControlRetainAnswer(t, f, token, kind)
	if accepted {
		f.receiver.operations = &telegramControlLostAck{service: operations}
		if err := f.receiver.respond(t.Context(), update); !errors.Is(err, errTelegramLostRecordAck) {
			t.Fatal("native acknowledgement was not interrupted", err)
		}
	}

	staleTelegramControlPrompt(t, f, token, stale)

	before := f.sessions(t)
	images := workflowCLIImage(t, operations.projectContext.ProjectDir)

	telegramAuthoringReopen(fixture.owner, f)

	f.receiver.operations = operations
	if err := f.receiver.respond(t.Context(), update); err != nil {
		t.Fatal(err)
	}

	if !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("stale prompt changed the native effect")
	}

	if stale == "replaced" && !slices.Equal(f.sessions(t), before) {
		t.Fatal("receipt recovery changed a replacement conversation")
	}

	checkTelegramStaleReceiptResult(t, f, update, accepted)
}

func checkTelegramStaleReceiptResult(t *testing.T, f *telegramConversationAppFixture, update core.TelegramReceivedUpdate, accepted bool) {
	t.Helper()

	request, found, err := f.receiver.store.TelegramOperation(t.Context(), update, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := f.receiver.drainDeliveries(t.Context()); err != nil {
		t.Fatal(err)
	}

	if accepted {
		if !found || request.State != "confirmed" || !strings.HasPrefix(request.ReceiptRef, "run/operation-receipts/") {
			t.Fatal("stale prompt hid the accepted native receipt", request)
		}

		telegramWaitNotice(t, f, "The original control receipt is retained.", time.Second)

		return
	}

	notice := telegramWaitNotice(t, f, "unavailable, expired or stale", time.Second)
	checkTelegramStaleDiagnostic(t, notice)

	if found && request.State == "confirmed" {
		t.Fatal("unaccepted stale prompt authorized a new control", request)
	}
}

func checkTelegramStaleDiagnostic(t *testing.T, notice telegramConversationNotice) {
	t.Helper()

	if !strings.Contains(notice.Text, "Current state:") || !strings.Contains(notice.Text, "chr work status --json ") {
		t.Fatal("stale control omitted its authorized current native state", notice)
	}
}

func telegramControlRetainAnswer(t *testing.T, f *telegramConversationAppFixture, token, kind string) core.TelegramReceivedUpdate {
	t.Helper()

	message := map[string]any{"message_id": f.next, "from": map[string]any{"id": 100001}, "chat": map[string]any{"id": 100001, "type": "private"}, "text": "/answer " + token + " pause"}

	body := map[string]any{"update_id": f.next, "message": message}
	if kind == "callback" {
		delete(body, "message")

		message["date"] = 1
		message["from"] = map[string]any{"id": 12345, "is_bot": true}
		body["callback_query"] = map[string]any{"from": map[string]any{"id": 100001}, "data": "tga:" + token + ":pause", "message": message}
	}

	if kind == "reply" {
		message["text"] = "pause"
		message["reply_to_message"] = map[string]any{"message_id": telegramControlPromptMessage(t, f, token)}
	}

	payload, err := json.Marshal([]any{body})
	if err != nil {
		t.Fatal(err)
	}

	f.next++
	f.updates <- payload

	updates, err := f.receiver.api.Poll(t.Context(), f.receiver.token, nil, 1)
	if err != nil || len(updates) != 1 {
		t.Fatal("control input unavailable", updates, err)
	}

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	update := core.AuthorizeTelegramReceiverUpdate(config, 12345, updates[0])
	if update.Kind == "rejected" {
		t.Fatal("control fixture rejected", update.Reason)
	}

	if err := f.receiver.store.RetainBatch(t.Context(), 12345, []core.TelegramReceivedUpdate{update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	return update
}

func telegramControlPromptMessage(t *testing.T, f *telegramConversationAppFixture, token string) int64 {
	t.Helper()

	db, err := sql.Open("sqlite", database.UserDatabasePath(f.receiver.env))
	if err != nil {
		t.Fatal(err)
	}

	var id int64

	readErr := db.QueryRowContext(t.Context(), `SELECT message_id FROM telegram_replies WHERE action_token=?`, token).Scan(&id)
	if err := errors.Join(readErr, db.Close()); err != nil {
		t.Fatal(err)
	}

	return id
}

func staleTelegramControlPrompt(t *testing.T, f *telegramConversationAppFixture, token, stale string) {
	t.Helper()

	db, err := sql.Open("sqlite", database.UserDatabasePath(f.receiver.env))
	if err != nil {
		t.Fatal(err)
	}

	query, value := `UPDATE telegram_actions SET expires_at=? WHERE token=?`, time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)
	if stale == "superseded" {
		query, value = `UPDATE telegram_actions SET state=? WHERE token=?`, "superseded"
	}

	if stale == "replaced" {
		query, value = `UPDATE telegram_conversations SET revision=revision+1,capture=? WHERE id=(SELECT conversation_id FROM telegram_requests WHERE id=(SELECT request_id FROM telegram_actions WHERE token=?))`, "Replacement conversation state"
	}

	_, changeErr := db.ExecContext(t.Context(), query, value, token)

	_, invalidateErr := db.ExecContext(t.Context(), `UPDATE telegram_replies SET invalidated_at=? WHERE action_token=?`, time.Now().UTC().Format(time.RFC3339), token)
	if err := errors.Join(changeErr, invalidateErr, db.Close()); err != nil {
		t.Fatal(err)
	}
}
