package core

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type qualityReceiverDirectories struct{ data, state string }

func (e qualityReceiverDirectories) DataDir() string  { return e.data }
func (e qualityReceiverDirectories) StateDir() string { return e.state }

// The receiver opens its own connection to the unmodified frozen database image.
func qualityReceiverStore(t *testing.T, db *sql.DB) *telegramcore.ReceiverStore {
	t.Helper()

	var (
		sequence   int
		name, path string
	)
	if err := db.QueryRowContext(t.Context(), "PRAGMA database_list").Scan(&sequence, &name, &path); err != nil {
		t.Fatal(err)
	}

	store, err := telegramcore.NewReceiverStore(t.Context(), qualityReceiverDirectories{data: filepath.Dir(path), state: t.TempDir()}, os.DirFS("../../cmd/chr"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})

	return store
}

func qualityReadConversation(t *testing.T, db *sql.DB, want telegramcore.Conversation) telegramcore.Conversation {
	t.Helper()

	conversationLookup, err := qualityReceiverStore(t, db).ConversationReference(t.Context(), want.Peer, want.ID)

	got, found := conversationLookup.Conversation, conversationLookup.Found
	if err != nil || !found {
		t.Fatal("frozen conversation unavailable", found, err)
	}

	return got
}
