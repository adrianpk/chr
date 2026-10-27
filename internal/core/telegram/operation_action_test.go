package telegram

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func telegramOperationActionFixture(t *testing.T, options ...func(OperationRequest, *ActionOffer)) (*telegramConversationFixture, ActionOffer, Action) {
	t.Helper()

	f := newTelegramConversationFixture(t)
	f.apply(t, ConversationStart, func(input *ConversationInput) { input.Project = "alpha" })

	input := telegramOperationActionInput(t, f, "prepare operation")
	input.Conversation = f.sessions(t)[0].ID

	target, err := f.store.ResolveOperationTarget(t.Context(), input, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	request, err := f.store.ReserveOperation(t.Context(), OperationRequest{Source: input.Update, Operator: input.Config.AllowedUsers[0].OperatorID, Target: target, OperationID: "records.ticket.create", Arguments: json.RawMessage(`{"title":"One record"}`)}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if err := f.store.PrepareOperation(t.Context(), request, json.RawMessage(`{"reserved_id":"PRJ-1"}`), time.Now()); err != nil {
		t.Fatal(err)
	}

	offer := ActionOffer{Input: input, RequestID: request.ID, Choices: []string{"confirm", "cancel"}, Text: "Apply the observed record?", ExpiresAt: time.Now().Add(time.Hour)}

	for _, option := range options {
		option(request, &offer)
	}

	action, err := f.store.OfferAction(t.Context(), offer, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	return f, offer, action
}

func telegramOperationActionInput(t *testing.T, f *telegramConversationFixture, text string) ConversationInput {
	t.Helper()

	input := f.input
	input.Update = telegramRetainedTestUpdate(t, f.next, text)

	f.next++
	if err := f.store.RetainBatch(t.Context(), 12345, []ReceivedUpdate{input.Update}, time.Now()); err != nil {
		t.Fatal(err)
	}

	return input
}

// TestTelegramOperationActionReplay reopens a claimed action and retains one original answer across callback/text replay.
func TestTelegramOperationActionReplay(t *testing.T) {
	f, offer, action := telegramOperationActionFixture(t)

	again, err := f.store.OfferAction(t.Context(), offer, time.Now())
	if err != nil || again.Token != action.Token || !again.ExpiresAt.Equal(action.ExpiresAt) {
		t.Fatal("action reoffer changed binding", again, err)
	}

	first := telegramOperationActionInput(t, f, "/answer "+action.Token+" confirm")

	claim, err := f.store.ClaimAction(t.Context(), first, ActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now())
	if err != nil || claim.Action.State != "applying" {
		t.Fatal("action claim failed", claim, err)
	}

	if err := f.store.Close(); err != nil {
		t.Fatal(err)
	}

	f.store = openTelegramReceiverTestStore(t, f.env)
	second := telegramOperationActionInput(t, f, "tga:"+action.Token+":confirm")

	recovered, err := f.store.ClaimAction(t.Context(), second, ActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now())
	if err != nil || recovered.Answer.ID != claim.Answer.ID || recovered.Answer.Source.UpdateID != first.Update.UpdateID {
		t.Fatal("action created a second answer", recovered, err)
	}

	checkTelegramOperationActionFinish(t, f, recovered)
}

func checkTelegramOperationActionFinish(t *testing.T, f *telegramConversationFixture, claim ActionClaim) {
	t.Helper()

	result := ActionResult{State: "confirmed", ReceiptRef: "original/receipt.json", Facts: json.RawMessage(`{"identity":"PRJ-1"}`), Text: "Confirmed PRJ-1", ConfigDigest: f.input.Config.Digest()}
	for range 2 {
		if err := f.store.FinishAction(t.Context(), claim, result, time.Now()); err != nil {
			t.Fatal(err)
		}
	}

	action, err := readAction(t.Context(), f.store.db, claim.Action.Token, claim.Action.Peer)
	if err != nil || action.State != "consumed" {
		t.Fatal("confirmed result did not consume action", action, err)
	}

	for _, id := range []string{claim.Request.ID, claim.Answer.ID} {
		request, err := readOperationID(t.Context(), f.store.db, id)
		if err != nil || request.State != "confirmed" || request.ReceiptRef != result.ReceiptRef {
			t.Fatal("request result was not atomic", request, err)
		}
	}

	input := telegramOperationActionInput(t, f, "changed choice")
	if _, err := f.store.ClaimAction(t.Context(), input, ActionAnswer{Token: action.Token, Choice: "cancel"}, time.Now()); err == nil {
		t.Fatal("consumed choice rebound")
	}
}

// TestTelegramOperationActionGuards expires or supersedes prompts without creating an answer or consuming authority.
func TestTelegramOperationActionGuards(t *testing.T) {
	for _, kind := range []string{"expiry", "actor", "project", "root", "revision", "guard"} {
		t.Run(kind, func(t *testing.T) {
			f, _, action := telegramOperationActionFixture(t)
			input := telegramOperationActionInput(t, f, "answer")
			now := changeTelegramOperationActionGuard(t, f, &input, kind)

			claim, err := f.store.ClaimAction(t.Context(), input, ActionAnswer{Token: action.Token, Choice: "confirm"}, now)
			if err != nil || (claim.Action.State != "expired" && claim.Action.State != "superseded") || claim.Action.AnswerRequestID != "" {
				t.Fatal("stale action created authority", claim, err)
			}

			var answers int
			if err := f.store.db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM telegram_requests WHERE operation_id='actions.answer'`).Scan(&answers); err != nil || answers != 0 {
				t.Fatal("stale action retained an answer", answers, err)
			}
		})
	}
}

func changeTelegramOperationActionGuard(t *testing.T, f *telegramConversationFixture, input *ConversationInput, kind string) time.Time {
	t.Helper()

	now := time.Now()

	switch kind {
	case "expiry":
		return now.Add(2 * time.Hour)
	case "actor":
		input.Config.AllowedUsers[0].OperatorID = "different-operator"
	case "project":
		input.Config.AllowedUsers[0].ProjectIDs = []string{"beta"}
	case "root":
		input.Projects[0].Path = t.TempDir()
	case "revision":
		if _, err := f.store.db.ExecContext(t.Context(), `UPDATE telegram_conversations SET revision=revision+1`); err != nil {
			t.Fatal(err)
		}
	case "guard":
		if _, err := f.store.db.ExecContext(t.Context(), `UPDATE telegram_actions SET guard_digest=?`, fmt.Sprintf("%064d", 1)); err != nil {
			t.Fatal(err)
		}
	}

	return now
}

// TestTelegramOperationActionConcurrency reserves one answer across independent SQLite owners.
func TestTelegramOperationActionConcurrency(t *testing.T) {
	f, _, action := telegramOperationActionFixture(t)
	first := telegramOperationActionInput(t, f, "first answer")
	second := telegramOperationActionInput(t, f, "duplicate callback")

	other, err := NewReceiverStore(t.Context(), f.env, os.DirFS("../../../cmd/chr"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := other.Close(); err != nil {
			t.Error(err)
		}
	})

	type outcome struct {
		claim ActionClaim
		err   error
	}

	results := make(chan outcome, 2)

	for index, store := range []*ReceiverStore{f.store, other} {
		input := []ConversationInput{first, second}[index]
		go func() {
			claim, err := store.ClaimAction(t.Context(), input, ActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now())
			results <- outcome{claim, err}
		}()
	}

	left, right := <-results, <-results
	if left.err != nil || right.err != nil || left.claim.Answer.ID != right.claim.Answer.ID {
		t.Fatal("concurrent action duplicated answer", left, right)
	}
}

// TestTelegramOperationActionRollback preserves both requests and pending action until its result fits the outbox.
func TestTelegramOperationActionRollback(t *testing.T) {
	f, _, action := telegramOperationActionFixture(t)
	input := telegramOperationActionInput(t, f, "answer")

	claim, err := f.store.ClaimAction(t.Context(), input, ActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.store.db.ExecContext(t.Context(), `WITH RECURSIVE n(i) AS (VALUES(1) UNION ALL SELECT i+1 FROM n WHERE i<1022) INSERT INTO telegram_outbox(id,bot_id,chat_id,topic_id,user_id,event_key,config_digest,kind,text,state,created_at,updated_at) SELECT 'fill-'||i,12345,100001,0,100001,'fill-'||i,?,'result','retained','pending','now','now' FROM n`, f.input.Config.Digest())
	if err != nil {
		t.Fatal(err)
	}

	result := ActionResult{State: "confirmed", ReceiptRef: "original-receipt", Facts: json.RawMessage(`{}`), Text: "confirmed", ConfigDigest: f.input.Config.Digest()}
	if err := f.store.FinishAction(t.Context(), claim, result, time.Now()); err == nil {
		t.Fatal("full outbox consumed action")
	}

	if err := f.store.Close(); err != nil {
		t.Fatal(err)
	}

	f.store = openTelegramReceiverTestStore(t, f.env)

	retained, err := readAction(t.Context(), f.store.db, action.Token, action.Peer)
	if err != nil || retained.State != "applying" {
		t.Fatal("failed acknowledgement lost action intent", retained, err)
	}

	checkTelegramOperationActionPending(t, f, claim)
}

func checkTelegramOperationActionPending(t *testing.T, f *telegramConversationFixture, claim ActionClaim) {
	t.Helper()

	for _, id := range []string{claim.Request.ID, claim.Answer.ID} {
		request, err := readOperationID(t.Context(), f.store.db, id)
		if err != nil || request.State != "applying" || request.ReceiptRef != "" {
			t.Fatal("failed acknowledgement changed request", request, err)
		}
	}

	var state string
	if err := f.store.db.QueryRowContext(t.Context(), `SELECT state FROM telegram_updates WHERE bot_id=? AND update_id=?`, claim.Answer.Source.Peer.BotID, claim.Answer.Source.UpdateID).Scan(&state); err != nil || state != "pending" {
		t.Fatal("failed acknowledgement lost answer source", state, err)
	}
}
