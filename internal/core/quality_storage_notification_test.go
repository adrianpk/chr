package core

import (
	"database/sql"
	"os"
	"testing"

	notifycore "forge.adrianpk.com/chr/internal/core/notify"
)

// The old database image is reopened through the public migration-aware boundary.
func qualityNotificationRepository(t *testing.T, db *sql.DB) notifycore.Repository {
	t.Helper()

	repository, err := notifycore.NewRepositoryFromDB(t.Context(), db, os.DirFS("../../cmd/chr"))
	if err != nil {
		t.Fatal(err)
	}

	return repository
}
