package app

import (
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramAuthoringPreparedFixture(t *testing.T, start string) (*telegramConversationAppFixture, string, *atomic.Int64) {
	t.Helper()
	t.Setenv(core.OpsUserEnv, "original")

	f := telegramOperationAppFixture(t)
	calls := telegramAuthoringBridgeHTTP(t, f, func(authoringModelInput) string {
		return telegramAuthoringServiceOutput(t, "Retain five retry attempts and preserve existing defaults.", "")
	}, func(input ConversationIntentContext) ConversationIntent {
		if input.Message == "Develop bounded retries." {
			return ConversationIntent{OperationID: "develop", Arguments: ConversationIntentArguments{Text: input.Message}}
		}

		if input.Message == "Prepare the native ticket." {
			return ConversationIntent{OperationID: "prepare-work"}
		}

		return telegramAuthoringBridgeCapture(input)
	})
	f.send(t, start, 0)

	c := f.sessions(t)[0]
	f.send(t, "/done "+c.ID, 0)

	return f, c.ID, calls
}

func telegramAuthoringPreparedLatest(t *testing.T, f *telegramConversationAppFixture, id string) core.TelegramOperationRequest {
	t.Helper()

	c := f.session(t, id)

	config, err := f.receiver.currentConfig()
	if err != nil {
		t.Fatal(err)
	}

	projects, err := f.receiver.operations.ListRegisteredProjects(config.AllowedUsers[0].ProjectIDs)
	if err != nil {
		t.Fatal(err)
	}

	input := core.TelegramConversationInput{Update: core.TelegramReceivedUpdate{Peer: c.Peer, Kind: "message", ChatType: "private"}, Config: config, Projects: projects, Conversation: id}
	target := core.TelegramOperationTarget{ProjectID: c.ProjectID, ProjectBindingDigest: c.ProjectBindingDigest, Namespace: c.Namespace, ConversationID: c.ID, ConversationRevision: c.Revision}

	request, found, err := f.receiver.store.LatestTelegramAuthoringWork(t.Context(), input, target)
	if err != nil || !found {
		t.Fatal(request, err)
	}

	return request
}

func telegramAuthoringPreparedStore(t *testing.T, f *telegramConversationAppFixture) (service, []core.WorkflowRecord) {
	t.Helper()

	operations := telegramOperationConcreteService(t, f.receiver)

	scoped, err := operations.telegramProjectService(t.Context(), core.TelegramOperationTarget{ProjectID: "example-project", Namespace: "original"})
	if err != nil {
		t.Fatal(err)
	}

	store, err := scoped.scopedRecordStore("original")
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	return scoped, snapshot.Records
}

func telegramAuthoringPreparedCheck(t *testing.T, f *telegramConversationAppFixture, id string, authorized bool, count int) AuthoringWorkResult {
	t.Helper()

	request := telegramAuthoringPreparedLatest(t, f, id)

	var result AuthoringWorkResult
	if err := json.Unmarshal(request.Result, &result); err != nil {
		t.Fatal(err)
	}

	if request.State != "confirmed" || result.Result.ReceiptRef == "" || (result.Authorization != nil) != authorized || result.Scope.Namespace != "original" {
		t.Fatal("missing retained scope/authority", request, result)
	}

	operations, records := telegramAuthoringPreparedStore(t, f)
	if len(records) != count {
		t.Fatal("wrong native record count", len(records), count)
	}

	telegramAuthoringPreparedRecord(t, records, result)
	telegramAuthoringWorkNoAdmission(t, operations)

	return result
}

func telegramAuthoringPreparedRecord(t *testing.T, records []core.WorkflowRecord, result AuthoringWorkResult) {
	t.Helper()

	for _, record := range records {
		if nativeRecordView(record, RecordTicket).Identity != result.Result.Record.Identity {
			continue
		}

		var scope core.AuthoringWorkScope
		if err := json.Unmarshal(record.Metadata["authoring_scope"], &scope); err != nil {
			t.Fatal(err)
		}

		if scope != result.Scope || record.Task.Status != core.TicketStatusBacklog || !strings.Contains(string(record.Body), "retry attempts") {
			t.Fatal("native scope or description lost", scope, record)
		}

		return
	}

	t.Fatal("confirmed native record missing")
}

// TestTelegramAuthoringPreparedIntent creates actual native backlog descriptions
// with preparation-only or original explicit development authority, never invocations.
func TestTelegramAuthoringPreparedIntent(t *testing.T) {
	for _, intent := range []string{"prepare", "develop"} {
		t.Run(intent, func(t *testing.T) {
			f, id, calls := telegramAuthoringPreparedFixture(t, "/"+intent+" retry ceiling")

			notice := f.send(t, "/prepare-work "+id, 0)
			if !strings.Contains(notice.Text, "Native backlog ticket prepared") {
				t.Fatal("native preparation failed", notice)
			}

			original := telegramAuthoringPreparedCheck(t, f, id, intent == "develop", 1)

			notice = f.send(t, "/prepare-work "+id, 0)
			if !strings.Contains(notice.Text, "Previously prepared") || calls.Load() != 1 {
				t.Fatal("unchanged scope was prepared again", notice, calls.Load())
			}

			again := telegramAuthoringPreparedCheck(t, f, id, intent == "develop", 1)
			if again.Scope.Hash != original.Scope.Hash {
				t.Fatal("unchanged scope acquired another identity")
			}
		})
	}
}

// TestTelegramAuthoringPreparedParity maps ordinary-language native preparation
// to the same shared writer and retained preparation-only intent as slash controls.
func TestTelegramAuthoringPreparedParity(t *testing.T) {
	f, id, calls := telegramAuthoringPreparedFixture(t, "/prepare retry ceiling")

	notice := f.send(t, "Prepare the native ticket.", 0)
	if !strings.Contains(notice.Text, "Native backlog ticket prepared") || calls.Load() != 1 {
		t.Fatal(notice, calls.Load())
	}

	telegramAuthoringPreparedCheck(t, f, id, false, 1)
}
