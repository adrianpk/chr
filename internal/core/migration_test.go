package core

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"testing/fstest"
	"time"

	_ "modernc.org/sqlite"
)

func TestReadSQLiteMigrationsParsesAndSortsAssets(t *testing.T) {
	migrations, err := ReadSQLiteMigrations(fstest.MapFS{
		"assets/migration/sqlite/20260518000002-second.sql": {
			Data: []byte("-- +migrate Up\nCREATE TABLE second (id text);\n-- +migrate Down\nDROP TABLE second;"),
		},
		"assets/migration/sqlite/20260518000001-first.sql": {
			Data: []byte("-- +migrate Up\nCREATE TABLE first (id text);\n-- +migrate Down\nDROP TABLE first;"),
		},
	})
	if err != nil {
		t.Fatalf("ReadSQLiteMigrations returned error: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("len(migrations) = %d, want 2", len(migrations))
	}

	if migrations[0].Version != "20260518000001" {
		t.Fatalf("first version = %q, want 20260518000001", migrations[0].Version)
	}

	if migrations[1].Version != "20260518000002" {
		t.Fatalf("second version = %q, want 20260518000002", migrations[1].Version)
	}
}

func TestMigratorAppliesPendingMigrations(t *testing.T) {
	db := openTestSQLite(t)
	migrator := NewMigrator(testMigrationFS(), WithMigratorClock(func() time.Time {
		return time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC)
	}))

	err := migrator.Migrate(context.Background(), db)
	if err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	assertTableExists(t, db, "projects")
	assertTableExists(t, db, "loop_runs")
	assertTableExists(t, db, "daemon_state")
	assertTableExists(t, db, "notification_events")
	assertTableExists(t, db, "ticket_key_sequences")
	assertColumnExists(t, db, "indexed_tickets", "synopsis")
	assertColumnExists(t, db, "indexed_tickets", "board_column")
	assertColumnExists(t, db, "indexed_tickets", "source_kind")
	assertColumnExists(t, db, "indexed_tickets", "source_ref")
	assertColumnExists(t, db, "indexed_tickets", "slice")
	assertColumnExists(t, db, "indexed_tickets", "branch")
	assertColumnExists(t, db, "indexed_tickets", "review_refs")
	assertColumnExists(t, db, "indexed_tickets", "review_state")
	assertColumnExists(t, db, "indexed_tickets", "review_round")
	assertColumnExists(t, db, "indexed_tickets", "review_result")
	assertColumnExists(t, db, "indexed_tickets", "review_comments_url")
	assertColumnExists(t, db, "indexed_tickets", "blocking_findings")
	assertColumnExists(t, db, "indexed_tickets", "merge_state")
	assertColumnExists(t, db, "indexed_tickets", "merge_result")
	assertColumnExists(t, db, "indexed_tickets", "merge_actor")
	assertColumnExists(t, db, "indexed_tickets", "merge_commit")
	assertColumnExists(t, db, "indexed_tickets", "merged_at")
	assertColumnExists(t, db, "indexed_tickets", "merge_blocker")
	assertColumnExists(t, db, "execution_records", "workspace_mode")
	assertColumnExists(t, db, "execution_records", "workspace_project_checkout_dir")
	assertColumnExists(t, db, "execution_records", "workspace_execution_dir")
	assertColumnExists(t, db, "execution_records", "workspace_base_ref")
	assertColumnExists(t, db, "execution_records", "workspace_base_sha")
	assertColumnExists(t, db, "execution_records", "workspace_ticket_branch")
	assertMigrationApplied(t, db, "20260518000001", "create_projects", "2026-05-18T12:30:00Z")
}

func TestMigratorSkipsAppliedMigrations(t *testing.T) {
	db := openTestSQLite(t)
	migrator := NewMigrator(testMigrationFS())

	err := migrator.Migrate(context.Background(), db)
	if err != nil {
		t.Fatalf("first Migrate returned error: %v", err)
	}

	err = migrator.Migrate(context.Background(), db)
	if err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}

	var count int

	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", "20260518000001").Scan(&count)
	if err != nil {
		t.Fatalf("count migrations: %v", err)
	}

	if count != 1 {
		t.Fatalf("migration count = %d, want 1", count)
	}
}

func openTestSQLite(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	db.SetMaxOpenConns(1)

	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Fatalf("close sqlite: %v", err)
		}
	})

	return db
}

func testMigrationFS() fstest.MapFS {
	workflowMigration, err := os.ReadFile("../../cmd/chr/assets/migration/sqlite/20261008000014-add_workflow_records.sql")
	if err != nil {
		panic(err)
	}

	return fstest.MapFS{
		"assets/migration/sqlite/20261008000014-add_workflow_records.sql": {Data: workflowMigration},
		"assets/migration/sqlite/20260518000001-create_projects.sql": {
			Data: []byte(testCreateProjectsMigration()),
		},
		"assets/migration/sqlite/20260518000002-create_indexed_tickets.sql": {
			Data: []byte(testCreateIndexedTicketsMigration()),
		},
		"assets/migration/sqlite/20260519000003-create_execution_records.sql": {
			Data: []byte(testCreateExecutionRecordsMigration()),
		},
		"assets/migration/sqlite/20260520000004-create_loop_runs.sql": {
			Data: []byte(testCreateLoopRunsMigration()),
		},
		"assets/migration/sqlite/20260520000005-create_daemon_state.sql": {
			Data: []byte(testCreateDaemonStateMigration()),
		},
		"assets/migration/sqlite/20260521000006-create_notification_events.sql": {
			Data: []byte(testCreateNotificationEventsMigration()),
		},
		"assets/migration/sqlite/20260521000007-create_remote_command_records.sql": {
			Data: []byte(testCreateRemoteCommandRecordsMigration()),
		},
		"assets/migration/sqlite/20260526000008-add_board_ticket_metadata.sql": {
			Data: []byte(testAddBoardTicketMetadataMigration()),
		},
		"assets/migration/sqlite/20260601000009-add_ticket_review_metadata.sql": {
			Data: []byte(testAddTicketReviewMetadataMigration()),
		},
		"assets/migration/sqlite/20260602000010-add_ticket_merge_metadata.sql": {
			Data: []byte(testAddTicketMergeMetadataMigration()),
		},
		"assets/migration/sqlite/20260619000011-add_execution_workspace_metadata.sql": {
			Data: []byte(testAddExecutionWorkspaceMetadataMigration()),
		},
		"assets/migration/sqlite/20260713000012-add_execution_artifact_id.sql": {
			Data: []byte(testAddExecutionArtifactIDMigration()),
		},
		"assets/migration/sqlite/20260713000013-add_loop_run_artifact.sql": {
			Data: []byte(testAddLoopRunArtifactMigration()),
		},
	}
}

func testCreateProjectsMigration() string {
	return `-- +migrate Up
CREATE TABLE projects (
    id text PRIMARY KEY,
    name text NOT NULL UNIQUE,
    path text NOT NULL UNIQUE,
    enabled integer NOT NULL,
    registered_at text NOT NULL,
    last_indexed_at text
);

-- +migrate Down
DROP TABLE IF EXISTS projects;
`
}

func testCreateIndexedTicketsMigration() string {
	return `-- +migrate Up
CREATE TABLE indexed_tickets (
    project_id text NOT NULL,
    path text NOT NULL,
    relative_path text NOT NULL,
    key text NOT NULL,
    title text NOT NULL,
    status text NOT NULL,
    priority text NOT NULL,
    severity text NOT NULL,
    type text NOT NULL,
    use_case text NOT NULL,
    project_area text NOT NULL,
    created_at text,
    updated_at text,
    file_mtime text NOT NULL,
    file_size integer NOT NULL,
    content_hash text NOT NULL,
    indexed_at text NOT NULL,
    PRIMARY KEY (project_id, path),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX indexed_tickets_project_key_idx
ON indexed_tickets (project_id, key);

CREATE INDEX indexed_tickets_project_status_idx
ON indexed_tickets (project_id, status);

CREATE INDEX indexed_tickets_project_type_idx
ON indexed_tickets (project_id, type);

CREATE TABLE indexed_ticket_tags (
    project_id text NOT NULL,
    ticket_path text NOT NULL,
    tag text NOT NULL,
    PRIMARY KEY (project_id, ticket_path, tag),
    FOREIGN KEY (project_id, ticket_path)
        REFERENCES indexed_tickets(project_id, path)
        ON DELETE CASCADE
);

CREATE INDEX indexed_ticket_tags_project_tag_idx
ON indexed_ticket_tags (project_id, tag);

-- +migrate Down
DROP TABLE IF EXISTS indexed_ticket_tags;
DROP TABLE IF EXISTS indexed_tickets;
`
}

func testCreateExecutionRecordsMigration() string {
	return `-- +migrate Up
CREATE TABLE execution_records (
    id text PRIMARY KEY,
    project_id text NOT NULL,
    ticket_key text NOT NULL,
    ticket_path text NOT NULL,
    status text NOT NULL,
    runtime_image text NOT NULL,
    runtime_dockerfile_path text NOT NULL,
    runtime_container_workdir text NOT NULL,
    runtime_validation_command text NOT NULL,
    artifact_run_dir text NOT NULL,
    artifact_result_path text NOT NULL,
    artifact_log_path text NOT NULL,
    artifact_validation_path text NOT NULL,
    failure_reason text NOT NULL,
    created_at text NOT NULL,
    started_at text,
    finished_at text,
    updated_at text NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX execution_records_project_ticket_idx
ON execution_records (project_id, ticket_key);

CREATE INDEX execution_records_project_status_idx
ON execution_records (project_id, status);

-- +migrate Down
DROP TABLE IF EXISTS execution_records;
`
}

func testCreateLoopRunsMigration() string {
	return `-- +migrate Up
CREATE TABLE loop_runs (
    id text PRIMARY KEY,
    project_id text NOT NULL,
    status text NOT NULL,
    eligible_count integer NOT NULL,
    processed_count integer NOT NULL,
    blocked_count integer NOT NULL,
    failed_count integer NOT NULL,
    failure_reason text NOT NULL,
    created_at text NOT NULL,
    started_at text,
    finished_at text,
    updated_at text NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX loop_runs_project_status_idx
ON loop_runs (project_id, status);

CREATE INDEX loop_runs_project_created_idx
ON loop_runs (project_id, created_at);

-- +migrate Down
DROP TABLE IF EXISTS loop_runs;
`
}

func testCreateDaemonStateMigration() string {
	return `-- +migrate Up
CREATE TABLE daemon_state (
    id text PRIMARY KEY,
    status text NOT NULL,
    pid integer NOT NULL,
    started_at text,
    heartbeat_at text,
    stop_requested_at text,
    failure_reason text NOT NULL,
    updated_at text NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS daemon_state;
`
}

func testCreateNotificationEventsMigration() string {
	return `-- +migrate Up
CREATE TABLE notification_events (
    id text PRIMARY KEY,
    project_id text NOT NULL,
    type text NOT NULL,
    severity text NOT NULL,
    status text NOT NULL,
    source text NOT NULL,
    channel text NOT NULL,
    ticket_key text NOT NULL,
    execution_id text NOT NULL,
    loop_run_id text NOT NULL,
    daemon_pid integer NOT NULL,
    summary text NOT NULL,
    detail text NOT NULL,
    failure_reason text NOT NULL,
    created_at text NOT NULL,
    delivered_at text,
    updated_at text NOT NULL
);

CREATE INDEX notification_events_project_created_idx
ON notification_events (project_id, created_at);

CREATE INDEX notification_events_project_status_idx
ON notification_events (project_id, status);

-- +migrate Down
DROP TABLE IF EXISTS notification_events;
`
}

func testCreateRemoteCommandRecordsMigration() string {
	return `-- +migrate Up
CREATE TABLE remote_command_records (
    id text PRIMARY KEY,
    status text NOT NULL,
    channel text NOT NULL,
    sender_external_id text NOT NULL,
    sender_display_name text NOT NULL,
    project_selector text NOT NULL,
    intent text NOT NULL,
    ticket_key text NOT NULL,
    comment text NOT NULL,
    loop_statuses text NOT NULL,
    loop_priorities text NOT NULL,
    loop_severities text NOT NULL,
    loop_project_areas text NOT NULL,
    loop_tags text NOT NULL,
    loop_types text NOT NULL,
    loop_use_cases text NOT NULL,
    loop_max_iterations integer NOT NULL,
    raw_text text NOT NULL,
    handoff_command text NOT NULL,
    rejection_reason text NOT NULL,
    received_at text NOT NULL,
    created_at text NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS remote_command_records;
`
}

func testAddBoardTicketMetadataMigration() string {
	return `-- +migrate Up
ALTER TABLE indexed_tickets
ADD COLUMN synopsis text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN board_column text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN source_kind text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN source_ref text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN slice text NOT NULL DEFAULT '';

CREATE INDEX indexed_tickets_project_board_column_idx
ON indexed_tickets (project_id, board_column, priority, severity);

CREATE INDEX indexed_tickets_project_use_case_idx
ON indexed_tickets (project_id, use_case);

CREATE INDEX indexed_tickets_project_area_idx
ON indexed_tickets (project_id, project_area);

CREATE INDEX indexed_tickets_project_updated_idx
ON indexed_tickets (project_id, updated_at);

CREATE INDEX indexed_tickets_project_source_kind_idx
ON indexed_tickets (project_id, source_kind);

CREATE TABLE ticket_key_sequences (
    project_id text NOT NULL,
    prefix text NOT NULL,
    next_number integer NOT NULL,
    updated_at text NOT NULL,
    PRIMARY KEY (project_id, prefix),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS ticket_key_sequences;
DROP INDEX IF EXISTS indexed_tickets_project_source_kind_idx;
DROP INDEX IF EXISTS indexed_tickets_project_updated_idx;
DROP INDEX IF EXISTS indexed_tickets_project_area_idx;
DROP INDEX IF EXISTS indexed_tickets_project_use_case_idx;
DROP INDEX IF EXISTS indexed_tickets_project_board_column_idx;
`
}

func testAddTicketReviewMetadataMigration() string {
	return `-- +migrate Up
ALTER TABLE indexed_tickets
ADD COLUMN branch text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN review_refs text NOT NULL DEFAULT '[]';

ALTER TABLE indexed_tickets
ADD COLUMN review_state text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN review_round integer NOT NULL DEFAULT 0;

ALTER TABLE indexed_tickets
ADD COLUMN review_result text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN review_comments_url text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN blocking_findings text NOT NULL DEFAULT '[]';

CREATE INDEX indexed_tickets_project_review_state_idx
ON indexed_tickets (project_id, review_state);

CREATE INDEX indexed_tickets_project_branch_idx
ON indexed_tickets (project_id, branch);

-- +migrate Down
DROP INDEX IF EXISTS indexed_tickets_project_branch_idx;
DROP INDEX IF EXISTS indexed_tickets_project_review_state_idx;
`
}

func testAddTicketMergeMetadataMigration() string {
	return `-- +migrate Up
ALTER TABLE indexed_tickets
ADD COLUMN merge_state text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN merge_result text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN merge_actor text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN merge_commit text NOT NULL DEFAULT '';

ALTER TABLE indexed_tickets
ADD COLUMN merged_at text DEFAULT NULL;

ALTER TABLE indexed_tickets
ADD COLUMN merge_blocker text NOT NULL DEFAULT '';

CREATE INDEX indexed_tickets_project_merge_state_idx
ON indexed_tickets (project_id, merge_state);

-- +migrate Down
DROP INDEX IF EXISTS indexed_tickets_project_merge_state_idx;
`
}

func testAddExecutionWorkspaceMetadataMigration() string {
	return `-- +migrate Up
ALTER TABLE execution_records
ADD COLUMN workspace_mode text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_project_checkout_dir text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_execution_dir text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_base_ref text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_base_sha text NOT NULL DEFAULT '';

ALTER TABLE execution_records
ADD COLUMN workspace_ticket_branch text NOT NULL DEFAULT '';

-- +migrate Down
SELECT 1;
`
}

func testAddExecutionArtifactIDMigration() string {
	return `-- +migrate Up
ALTER TABLE execution_records
ADD COLUMN artifact_id text NOT NULL DEFAULT '';

CREATE INDEX execution_records_artifact_id_idx
ON execution_records (artifact_id);

-- +migrate Down
DROP INDEX IF EXISTS execution_records_artifact_id_idx;
`
}

func testAddLoopRunArtifactMigration() string {
	return `-- +migrate Up
ALTER TABLE loop_runs
ADD COLUMN artifact_id text NOT NULL DEFAULT '';

ALTER TABLE loop_runs
ADD COLUMN artifact_dir text NOT NULL DEFAULT '';

CREATE INDEX loop_runs_artifact_id_idx
ON loop_runs (artifact_id);

-- +migrate Down
DROP INDEX IF EXISTS loop_runs_artifact_id_idx;
`
}

func assertTableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()

	var found string

	err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?",
		name,
	).Scan(&found)
	if err != nil {
		t.Fatalf("table %s does not exist: %v", name, err)
	}
}

func assertColumnExists(t *testing.T, db *sql.DB, tableName, columnName string) {
	t.Helper()

	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		t.Fatalf("inspect columns for %s: %v", tableName, err)
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			t.Fatalf("close column rows: %v", err)
		}
	}()

	for rows.Next() {
		var (
			cid          int
			name         string
			columnType   string
			notNull      int
			defaultValue sql.NullString
			primaryKey   int
		)

		err = rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey)
		if err != nil {
			t.Fatalf("scan column: %v", err)
		}

		if name == columnName {
			return
		}
	}

	if err = rows.Err(); err != nil {
		t.Fatalf("iterate columns for %s: %v", tableName, err)
	}

	t.Fatalf("column %s.%s does not exist", tableName, columnName)
}

func assertMigrationApplied(t *testing.T, db *sql.DB, version, name, appliedAt string) {
	t.Helper()

	var gotName string

	var gotAppliedAt string

	err := db.QueryRow(
		"SELECT name, applied_at FROM schema_migrations WHERE version = ?",
		version,
	).Scan(&gotName, &gotAppliedAt)
	if err != nil {
		t.Fatalf("query migration: %v", err)
	}

	if gotName != name {
		t.Fatalf("migration name = %q, want %q", gotName, name)
	}

	if gotAppliedAt != appliedAt {
		t.Fatalf("applied_at = %q, want %q", gotAppliedAt, appliedAt)
	}
}
