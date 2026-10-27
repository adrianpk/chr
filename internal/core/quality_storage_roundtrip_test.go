package core

import (
	"database/sql"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

func qualityLegacyExpected(t *testing.T) qualityLegacyRecords {
	t.Helper()

	data, err := os.ReadFile("testdata/quality-storage/legacy-records.json")
	if err != nil {
		t.Fatal(err)
	}

	var records qualityLegacyRecords
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}

	return records
}

// TestQualityStorageBindings compares every stored column against the old
// writers, so matching mistakes in a new binder and scanner cannot cancel out.
func TestQualityStorageBindings(t *testing.T) {
	for _, test := range []struct {
		name, table string
		write       func(*testing.T, *sql.DB, qualityLegacyRecords)
	}{
		{"execution insert", "execution_records", qualityExecutionInsert},
		{"execution update", "execution_records", qualityExecutionUpdate},
		{"remote insert", "remote_command_records", qualityRemoteInsert},
		{"notification update", "notification_events", qualityNotificationUpdate},
		{"ticket upsert", "indexed_tickets", qualityTicketUpsert},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := openQualityLegacyDB(t)
			want := qualityStorageImage(t, db, test.table)
			test.write(t, db, qualityLegacyExpected(t))

			if got := qualityStorageImage(t, db, test.table); got != want {
				t.Fatalf("stored columns changed:\ngot %s\nwant %s", got, want)
			}
		})
	}
}

func qualityExecutionInsert(t *testing.T, db *sql.DB, want qualityLegacyRecords) {
	t.Helper()
	qualityStorageExec(t, db, "DELETE FROM execution_records")

	if err := qualityExecutionRepository(t, db).CreateExecution(t.Context(), want.Execution); err != nil {
		t.Fatal(err)
	}
}

func qualityExecutionUpdate(t *testing.T, db *sql.DB, want qualityLegacyRecords) {
	t.Helper()
	// Poison same-type values first; an omitted assignment cannot retain its expected value.
	qualityStorageExec(t, db, "UPDATE execution_records SET project_id='value-002',ticket_key='poison-key',ticket_path='poison-path',runtime_image='poison-image',workspace_base_ref='poison-ref',artifact_log_path='poison-log',failure_reason='poison-failure'")

	if err := qualityExecutionRepository(t, db).UpdateExecution(t.Context(), want.Execution); err != nil {
		t.Fatal(err)
	}
}

func qualityRemoteInsert(t *testing.T, db *sql.DB, want qualityLegacyRecords) {
	t.Helper()
	qualityStorageExec(t, db, "DELETE FROM remote_command_records")

	if err := (&sqliteRemoteCommandRecordRepository{db: db}).CreateRemoteCommandRecord(t.Context(), want.Remote); err != nil {
		t.Fatal(err)
	}
}

func qualityNotificationUpdate(t *testing.T, db *sql.DB, want qualityLegacyRecords) {
	t.Helper()
	qualityStorageExec(t, db, "UPDATE notification_events SET summary='poison-summary',detail='poison-detail',failure_reason='poison-failure',channel='poison-channel',loop_run_id='poison-loop'")

	if err := qualityNotificationRepository(t, db).Update(t.Context(), want.Notification); err != nil {
		t.Fatal(err)
	}
}

func qualityTicketUpsert(t *testing.T, db *sql.DB, want qualityLegacyRecords) {
	t.Helper()
	qualityStorageExec(t, db, "UPDATE indexed_tickets SET title='poison-title',synopsis='poison-synopsis',source_ref='poison-source',branch='poison-branch',review_comments_url='poison-comments',merge_commit='poison-commit'")

	if err := qualityTicketIndex(t, db).Upsert(t.Context(), want.Ticket); err != nil {
		t.Fatal(err)
	}
}

// TestQualityStorageNulls exercises both NULL and populated time transitions,
// while the conversation projection retains its original COALESCE defaults.
func TestQualityStorageNulls(t *testing.T) {
	db := openQualityLegacyDB(t)
	want := qualityLegacyExpected(t)
	execution := qualityExecutionRepository(t, db)
	notification := qualityNotificationRepository(t, db)

	for _, populated := range []bool{true, false} {
		t.Run(map[bool]string{true: "populated", false: "null"}[populated], func(t *testing.T) {
			changeQualityOptionalTimes(&want, populated)

			if err := execution.UpdateExecution(t.Context(), want.Execution); err != nil {
				t.Fatal(err)
			}

			if err := notification.Update(t.Context(), want.Notification); err != nil {
				t.Fatal(err)
			}

			got := readQualityLegacyRecords(t, db, want)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("nullable values changed: got %#v want %#v", got, want)
			}
		})
	}

	qualityConversationDefaults(t, db, want.Conversation)
}

func changeQualityOptionalTimes(records *qualityLegacyRecords, populated bool) {
	records.Execution.StartedAt = nil
	records.Execution.FinishedAt = nil
	records.Notification.Event.DeliveredAt = nil
	records.Notification.Event.Status = notifycore.StatusPending

	records.Execution.Status = ExecutionStatusPending
	if populated {
		started := records.Execution.CreatedAt.Add(2 * time.Minute)
		finished := started.Add(3 * time.Minute)
		delivered := records.Notification.Event.CreatedAt.Add(4 * time.Minute)
		records.Execution.Status = ExecutionStatusSucceeded
		records.Notification.Event.Status = notifycore.StatusDelivered
		records.Execution.StartedAt = &started
		records.Execution.FinishedAt = &finished
		records.Notification.Event.DeliveredAt = &delivered
	}
}

func qualityConversationDefaults(t *testing.T, db *sql.DB, want telegramcore.Conversation) {
	t.Helper()
	qualityStorageExec(t, db, "UPDATE telegram_conversations SET authoring_session_ref=NULL,authorization_ref=NULL,invocation_id=NULL,last_observed_revision=NULL,guided_run_id=NULL")

	want.AuthoringSessionRef = ""
	want.AuthorizationRef = ""
	want.InvocationID = ""
	want.LastObservedRevision = 0
	want.GuidedRunID = ""

	got := qualityReadConversation(t, db, want)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("conversation defaults changed: %#v", got)
	}
}
