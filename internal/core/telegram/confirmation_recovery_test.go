package telegram

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	"forge.adrianpk.com/chr/internal/core/database"
	tgmessage "forge.adrianpk.com/chr/internal/core/telegram/message"
)

// TestTelegramPresentationConfirmationGuards rejects altered scope or retained
// presentation before either transmission or creation of an answer request.
func TestTelegramPresentationConfirmationGuards(t *testing.T) {
	for _, kind := range []string{"caption", "document bytes", "missing document", "source binding", "keyboard choices", "foreign destination", "config binding", "future guard", "future presentation"} {
		t.Run(kind, func(t *testing.T) {
			checkTelegramConfirmationGuard(t, kind)
		})
	}
}

func changeTelegramConfirmation(t *testing.T, f *telegramConversationFixture, action TelegramAction, kind string) {
	t.Helper()

	switch kind {
	case "caption":
		execTelegramRecoverySQL(t, f.store, "UPDATE telegram_outbox SET text='Different scope' WHERE action_token=?", action.Token)
	case "document bytes":
		execTelegramRecoverySQL(t, f.store, "UPDATE telegram_document_snapshots SET payload=zeroblob(byte_count)")
	case "missing document":
		execTelegramRecoverySQL(t, f.store, "UPDATE telegram_document_snapshots SET payload=NULL,purged_at=?", database.FormatTime(time.Now()))
	case "source binding":
		execTelegramRecoverySQL(t, f.store, `UPDATE telegram_requests SET prepared_input='{"reserved_id":"PRJ-2"}' WHERE id=?`, action.RequestID)
	case "keyboard choices":
		execTelegramRecoverySQL(t, f.store, `UPDATE telegram_actions SET allowed_choices='["confirm","extra"]' WHERE token=?`, action.Token)
	case "foreign destination":
		execTelegramRecoverySQL(t, f.store, "UPDATE telegram_document_snapshots SET chat_id=chat_id+1")
	case "config binding":
		execTelegramRecoverySQL(t, f.store, "UPDATE telegram_outbox SET config_digest=? WHERE action_token=?", strings.Repeat("b", 64), action.Token)
	case "future guard":
		execTelegramRecoverySQL(t, f.store, "UPDATE telegram_actions SET guard_version=2 WHERE token=?", action.Token)
	case "future presentation":
		changeTelegramConfirmationVersion(t, f, action)
	}
}

func changeTelegramConfirmationVersion(t *testing.T, f *telegramConversationFixture, action TelegramAction) {
	t.Helper()

	var payload string
	if err := f.store.db.QueryRowContext(t.Context(), "SELECT presentation FROM telegram_outbox WHERE action_token=?", action.Token).Scan(&payload); err != nil {
		t.Fatal(err)
	}

	var presentation tgmessage.Presentation
	if err := json.Unmarshal([]byte(payload), &presentation); err != nil {
		t.Fatal(err)
	}

	presentation.Version = 2

	data, err := json.Marshal(presentation)
	if err != nil {
		t.Fatal(err)
	}

	request, err := readTelegramOperationID(t.Context(), f.store.db, action.RequestID)
	if err != nil {
		t.Fatal(err)
	}
	// Recompute both retained hashes so this checks version refusal rather than
	// accidentally stopping at the ordinary byte-integrity check.
	hash := contenthash.Sum(data)
	execTelegramRecoverySQL(t, f.store, "UPDATE telegram_outbox SET presentation=?,presentation_hash=? WHERE action_token=?", string(data), hash, action.Token)
	execTelegramRecoverySQL(t, f.store, "UPDATE telegram_actions SET presentation_hash=?,guard_digest=? WHERE token=?", hash, telegramActionPresentationDigest(request, hash), action.Token)
}

// TestTelegramPresentationConfirmationConcurrent races independent SQLite
// connections and retains a single answer identity and original result receipt.
func TestTelegramPresentationConfirmationConcurrent(t *testing.T) {
	f, _, action := telegramConfirmationFixture(t)
	second := openTelegramReceiverTestStore(t, f.env)
	inputs := []TelegramConversationInput{telegramOperationActionInput(t, f, "first answer"), telegramOperationActionInput(t, f, "duplicate answer")}

	type result struct {
		claim TelegramActionClaim
		err   error
	}

	results := make(chan result, 2)
	ready := make(chan struct{})

	for index, store := range []*TelegramReceiverStore{f.store, second} {
		go func() {
			<-ready

			claim, err := store.ClaimTelegramAction(t.Context(), inputs[index], TelegramActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now())
			results <- result{claim: claim, err: err}
		}()
	}

	close(ready)

	first, duplicate := <-results, <-results
	if first.err != nil || duplicate.err != nil {
		t.Fatal(first.err, duplicate.err)
	}

	if first.claim.Answer.ID == "" || first.claim.Answer.ID != duplicate.claim.Answer.ID {
		t.Fatal("concurrent callbacks allocated a second answer", first.claim, duplicate.claim)
	}

	var answers int
	if err := f.store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_requests WHERE operation_id='actions.answer'").Scan(&answers); err != nil || answers != 1 {
		t.Fatal("concurrent callbacks changed answer cardinality", answers, err)
	}

	checkTelegramOperationActionFinish(t, f, first.claim)
}

func checkTelegramConfirmationGuard(t *testing.T, kind string) {
	t.Helper()

	f, _, action := telegramConfirmationFixture(t)
	delivery := telegramConfirmationDelivery(t, f, action)
	changeTelegramConfirmation(t, f, action, kind)

	if err := f.store.ValidateTelegramActionDelivery(t.Context(), delivery); err == nil {
		t.Fatal("changed confirmation could be transmitted")
	}

	input := telegramOperationActionInput(t, f, "confirm")

	claim, err := f.store.ClaimTelegramAction(t.Context(), input, TelegramActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now())
	if err != nil && !errors.Is(err, ErrTelegramActionUnavailable) && !errors.Is(err, errTelegramReceiverConflict) {
		t.Fatal(err)
	}

	if err == nil && claim.Action.State != "superseded" {
		t.Fatal("changed presentation authorized an answer", claim)
	}

	var answers int
	if err := f.store.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM telegram_requests WHERE operation_id='actions.answer'").Scan(&answers); err != nil || answers != 0 {
		t.Fatal("changed presentation allocated authority", answers, err)
	}
}
