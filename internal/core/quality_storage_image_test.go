package core

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func qualityStorageImage(t *testing.T, db *sql.DB, table string) string {
	t.Helper()
	// Callers supply only literal test-owned table names, independent of row projections.
	rows, err := db.QueryContext(t.Context(), "SELECT * FROM "+table+" ORDER BY rowid")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Error(err)
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}

	var image [][]any

	for rows.Next() {
		values := make([]any, len(columns))

		targets := make([]any, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}

		if err := rows.Scan(targets...); err != nil {
			t.Fatal(err)
		}

		image = append(image, values)
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(struct {
		Columns []string
		Rows    [][]any
	}{columns, image})
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func qualityOwnerImages(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()

	images := map[string]string{}
	for _, table := range []string{"projects", "execution_records", "remote_command_records", "notification_events", "indexed_tickets", "indexed_ticket_tags", "telegram_receivers", "telegram_peers", "telegram_conversations", "telegram_updates", "telegram_outbox", "telegram_document_snapshots", "telegram_notification_documents"} {
		images[table] = qualityStorageImage(t, db, table)
	}

	return images
}

func qualityStorageExec(t *testing.T, db *sql.DB, query string) {
	t.Helper()

	if _, err := db.ExecContext(t.Context(), query); err != nil {
		t.Fatal(err)
	}
}
