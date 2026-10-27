package telegram

import (
	"os"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/database"
)

// TestTelegramRecoveryOffline proves local retry acquires live receiver
// exclusion before even creating or migrating its database.
func TestTelegramRecoveryOffline(t *testing.T) {
	env := telegramReceiverTestEnv(t)

	if err := BindTelegramReceiverIdentity(env, 12345); err != nil {
		t.Fatal(err)
	}

	lock, err := AcquireTelegramConfigurationLock(env)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.Close(); err != nil {
			t.Error(err)
		}
	}()

	request := TelegramTransportRetry{BotID: 12345, Poll: true, Actor: "local-uid-1000"}
	if err := RetryLocalTelegramTransport(t.Context(), env, os.DirFS("../../../cmd/chr"), request, time.Now()); err == nil {
		t.Fatal("live receiver retry accepted")
	}

	if _, err := os.Stat(database.UserDatabasePath(env)); !os.IsNotExist(err) {
		t.Fatal("retry opened database before receiver exclusion", err)
	}
}
