package telegram

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

// TestTelegramRecoveryCanceledTx forces database/sql's asynchronous rollback
// before the owner commits. Shutdown must report cancellation, while preserving
// an independent checkpoint failure even when it is joined with ErrTxDone.
func TestTelegramRecoveryCanceledTx(t *testing.T) {
	env := telegramReceiverTestEnv(t)

	store := openTelegramReceiverTestStore(t, env)
	for _, test := range []struct {
		phase string
		want  error
	}{
		{"action", context.Canceled}, {"commit", context.Canceled}, {"independent", errors.New("checkpoint failed")}, {"uncanceled", sql.ErrTxDone},
	} {
		t.Run(test.phase, func(t *testing.T) { checkTelegramRecoveryCanceledTx(t, store, test.phase, test.want) })
	}
}

func checkTelegramRecoveryCanceledTx(t *testing.T, store *TelegramReceiverStore, phase string, want error) {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	err := store.transaction(ctx, func(tx *sql.Tx) error {
		if phase == "uncanceled" {
			return sql.ErrTxDone
		}

		cancel()

		rolledBack := waitTelegramRecoveryRollback(t, tx)

		switch phase {
		case "action":
			return rolledBack
		case "independent":
			return errors.Join(want, rolledBack)
		default:
			return nil
		}
	})
	if !errors.Is(err, want) {
		t.Fatal("transaction cancellation lost its failure classification", err)
	}
}

func waitTelegramRecoveryRollback(t *testing.T, tx *sql.Tx) error {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		_, err := tx.ExecContext(t.Context(), "SELECT 1")
		if errors.Is(err, sql.ErrTxDone) {
			return err
		}

		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("database/sql did not roll back its canceled transaction")

	return nil
}
