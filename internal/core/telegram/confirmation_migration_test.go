package telegram

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// TestTelegramPresentationConfirmationMigration restores the actual old action
// schema and upgrades its pending prompt without changing any legacy authority.
func TestTelegramPresentationConfirmationMigration(t *testing.T) {
	f, _, action := telegramOperationActionFixture(t)

	request, err := readTelegramOperationID(t.Context(), f.store.db, action.RequestID)
	if err != nil {
		t.Fatal(err)
	}

	execTelegramRecoverySQL(t, f.store, "UPDATE telegram_outbox SET state='uncertain',attempts=2,total_attempts=2 WHERE action_token=?", action.Token)

	migrations, err := database.ReadSQLiteMigrations(telegramRecoveryMigrationAssets(t))
	if err != nil {
		t.Fatal(err)
	}

	for index := len(migrations) - 1; index >= 0; index-- {
		migration := migrations[index]
		if migration.Version == "20261021000028" || migration.Version == "20261021000027" {
			execTelegramRecoverySQL(t, f.store, migration.DownSQL)
		}
	}

	execTelegramRecoverySQL(t, f.store, "DELETE FROM schema_migrations WHERE version IN ('20261021000027','20261021000028')")

	if err := f.store.Close(); err != nil {
		t.Fatal(err)
	}

	f.store = openTelegramReceiverTestStore(t, f.env)
	checkTelegramLegacyConfirmation(t, f, action, request)

	input := telegramOperationActionInput(t, f, "confirm legacy action")

	claim, err := f.store.ClaimTelegramAction(t.Context(), input, TelegramActionAnswer{Token: action.Token, Choice: "confirm"}, time.Now())
	if err != nil || claim.Action.State != "applying" {
		t.Fatal("legacy guard no longer authorizes its original operation", claim, err)
	}

	checkTelegramOperationActionFinish(t, f, claim)
}

func checkTelegramLegacyConfirmation(t *testing.T, f *telegramConversationFixture, original TelegramAction, request TelegramOperationRequest) {
	t.Helper()

	action, err := readTelegramAction(t.Context(), f.store.db, original.Token, original.Peer)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(action, original) {
		t.Fatal("migration rewrote legacy authority", action, original)
	}

	retained, err := readTelegramOperationID(t.Context(), f.store.db, original.RequestID)
	if err != nil || !bytes.Equal(retained.PreparedInput, request.PreparedInput) || retained.ArgumentDigest != request.ArgumentDigest {
		t.Fatal("migration rewrote prepared input", retained, err)
	}

	delivery := telegramConfirmationLegacyDelivery(t, f, original)
	if delivery.State != "uncertain" || delivery.Attempts != 2 || delivery.Presentation != nil || delivery.DocumentID != "" {
		t.Fatal("migration changed pending delivery meaning or retry allowance", delivery)
	}
}

func telegramConfirmationLegacyDelivery(t *testing.T, f *telegramConversationFixture, action TelegramAction) TelegramReceiverDelivery {
	t.Helper()

	rows, err := f.store.DueDeliveries(t.Context(), 12345, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	for _, row := range rows {
		if row.ActionToken == action.Token {
			return row
		}
	}

	t.Fatal("migration lost the original queued prompt")

	return TelegramReceiverDelivery{}
}
