package telegram

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func seedTelegramRequestAllowance(t *testing.T, db *sql.DB, template string, total int, state string) {
	t.Helper()

	var present int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_requests`).Scan(&present); err != nil {
		t.Fatal(err)
	}

	_, err := db.ExecContext(t.Context(), `WITH RECURSIVE n(x) AS (SELECT ? UNION ALL SELECT x+1 FROM n WHERE x<?)
INSERT INTO telegram_updates(bot_id,update_id,cursor_epoch,payload_digest,update_kind,chat_id,topic_id,user_id,message_id,chat_type,received_at,payload,state,reason_code,updated_at)
SELECT u.bot_id,100000+n.x,u.cursor_epoch,u.payload_digest,'message',u.chat_id,u.topic_id,u.user_id,n.x,u.chat_type,u.received_at,'','processed','capacity_fixture',u.updated_at FROM n JOIN telegram_requests r ON r.id=? JOIN telegram_updates u ON u.bot_id=r.bot_id AND u.update_id=r.update_id`, present+1, total, template)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.ExecContext(t.Context(), `WITH RECURSIVE n(x) AS (SELECT ? UNION ALL SELECT x+1 FROM n WHERE x<?)
INSERT INTO telegram_requests(id,bot_id,update_id,operation_ordinal,operator_id,project_id,project_binding_digest,namespace,operation_id,arguments,argument_digest,state,created_at,updated_at)
SELECT 'capacity-'||n.x,r.bot_id,100000+n.x,0,r.operator_id,r.project_id,r.project_binding_digest,r.namespace,r.operation_id,r.arguments,r.argument_digest,?,r.created_at,r.updated_at FROM n JOIN telegram_requests r ON r.id=?`, present+1, total, state, template)
	if err != nil {
		t.Fatal(err)
	}
}

// TestTelegramControlRequestReserve admits all 32 prompt/answer pairs after
// ordinary requests reach their allowance. Rejected reservations leave no row.
func TestTelegramControlRequestReserve(t *testing.T) {
	for _, test := range []struct {
		name, state string
		total       int
	}{{"pending", "prepared", 960}, {"retained", "confirmed", 8128}} {
		t.Run(test.name, func(t *testing.T) { checkTelegramRequestReserve(t, test.total, test.state) })
	}
}

func checkTelegramRequestReserve(t *testing.T, total int, state string) {
	t.Helper()

	f, action := telegramPendingControlLaneFixture(t, "work.control", "pause")

	parent, err := f.store.TelegramActionRequest(t.Context(), action.Peer, TelegramActionAnswer{Token: action.Token, Choice: "pause"})
	if err != nil {
		t.Fatal(err)
	}

	template := telegramOperationTestRequest(t, f.store, 90000)
	seedTelegramRequestAllowance(t, f.store.db, template.ID, total, state)
	checkTelegramOrdinaryRequestFull(t, f, template)

	pairs := 32

	if state == "confirmed" {
		// Pending capacity exercises all 32 pairs. At the retained ceiling, prove
		// its first and last pair against a bulk-seeded settled historical backlog.
		reserveTelegramControlPair(t, f, parent.Target)
		seedTelegramRequestAllowance(t, f.store.db, template.ID, total+62, state)

		pairs = 1
	}

	for range pairs {
		reserveTelegramControlPair(t, f, parent.Target)
	}

	input := telegramOperationActionInput(t, f, "pause")

	_, err = f.store.ClaimTelegramAction(t.Context(), input, TelegramActionAnswer{Token: action.Token, Choice: "pause"}, time.Now())
	if !errors.Is(err, ErrTelegramRequestCapacity) {
		t.Fatal("hard request ceiling grew for an extra answer", err)
	}

	checkTelegramRequestCount(t, f.store.db, total+64)

	replay, err := f.store.ReserveTelegramOperation(t.Context(), template, time.Now())
	if err != nil || replay.ID != template.ID {
		t.Fatal("capacity blocked an existing binding", replay, err)
	}

	if err := f.store.QueueReadResponse(t.Context(), input.Update, "Capacity diagnostic", input.Config.Digest(), time.Now()); err != nil {
		t.Fatal("full requests prevented a durable diagnostic", err)
	}
}

func checkTelegramOrdinaryRequestFull(t *testing.T, f *telegramConversationFixture, template TelegramOperationRequest) {
	t.Helper()

	input := telegramOperationActionInput(t, f, "ordinary reservation")
	request := template

	request.Source = input.Update
	if _, err := f.store.ReserveTelegramOperation(t.Context(), request, time.Now()); !errors.Is(err, ErrTelegramRequestCapacity) {
		t.Fatal("ordinary requests consumed the control allowance", err)
	}

	if _, found, err := f.store.TelegramOperation(t.Context(), input.Update, 0); err != nil || found {
		t.Fatal("rejected reservation retained an accepted request", found, err)
	}
}

func reserveTelegramControlPair(t *testing.T, f *telegramConversationFixture, target TelegramOperationTarget) {
	t.Helper()

	input := telegramOperationActionInput(t, f, "control request")

	request, err := f.store.ReservePreparedTelegramOperation(t.Context(), TelegramOperationRequest{Source: input.Update, Operator: "maintainer", Target: target, OperationID: "work.control", Arguments: json.RawMessage(`{"action":"pause"}`)}, json.RawMessage(`{}`), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	input.Projects, input.Config = f.input.Projects, f.input.Config

	action, err := f.store.OfferTelegramAction(t.Context(), TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{"pause"}, Text: "Control allowance", ExpiresAt: time.Now().Add(time.Hour)}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	answer := telegramOperationActionInput(t, f, "pause")
	if _, err := f.store.ClaimTelegramAction(t.Context(), answer, TelegramActionAnswer{Token: action.Token, Choice: "pause"}, time.Now()); err != nil {
		t.Fatal("reserved control answer unavailable", err)
	}
}

func checkTelegramRequestCount(t *testing.T, db *sql.DB, expected int) {
	t.Helper()

	var count int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_requests`).Scan(&count); err != nil || count != expected {
		t.Fatal("request capacity changed", count, expected, err)
	}
}
