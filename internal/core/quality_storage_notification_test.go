package core

import notifycore "forge.adrianpk.com/chr/internal/core/notify"

import (
	"database/sql"
	"os"
	"testing"
)

// The old database image is reopened through the public migration-aware boundary.
func qualityNotificationRepository(t *testing.T, db *sql.DB) notifycore.NotificationEventRepository {
	t.Helper()

	repository, err := notifycore.NewNotificationEventRepositoryFromDB(t.Context(), db, os.DirFS("../../cmd/chr"))
	if err != nil {
		t.Fatal(err)
	}

	return repository
}
