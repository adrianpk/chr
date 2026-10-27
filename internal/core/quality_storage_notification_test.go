package core

import (
	"database/sql"
	"os"
	"testing"
)

// The old database image is reopened through the public migration-aware boundary.
func qualityNotificationRepository(t *testing.T, db *sql.DB) NotificationEventRepository {
	t.Helper()

	repository, err := NewNotificationEventRepositoryFromDB(t.Context(), db, os.DirFS("../../cmd/chr"))
	if err != nil {
		t.Fatal(err)
	}

	return repository
}
