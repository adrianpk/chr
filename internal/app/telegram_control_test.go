package app

import (
	"strings"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

func telegramControlOffer(t *testing.T, f *telegramConversationAppFixture, command string) string {
	t.Helper()

	notice := f.send(t, command, 0)
	if !strings.Contains(notice.Text, "for work ") || !strings.Contains(notice.Text, "Current state:") {
		t.Fatal("native control preview unavailable", notice)
	}

	parts := strings.Split(notice.Text, "/answer ")
	if len(parts) != 2 {
		t.Fatal("exact control binding unavailable", notice)
	}

	return strings.Fields(parts[1])[0]
}

// TestTelegramControlLifecycle carries each guarded lifecycle action through
// real HTTP, SQLite confirmation and the existing shared operation service.
func TestTelegramControlLifecycle(t *testing.T) {
	telegramControlParallel(t)

	for _, action := range []string{"pause", "cancel", "resume", "continue", "reconcile"} {
		t.Run(action, func(t *testing.T) { checkTelegramControlLifecycle(t, action) })
	}
}

func checkTelegramControlLifecycle(t *testing.T, action string) {
	t.Helper()

	f, operations, ref := telegramWorkAdmitted(t)
	if action == "resume" {
		if _, err := operations.StopWorkflow(t.Context(), ref.Invocation.ID, "pause"); err != nil {
			t.Fatal(err)
		}
	}

	before := checkTelegramWorkRunCount(t, operations, 1)

	token := telegramControlOffer(t, f, "/"+action+" "+ref.Invocation.ID)
	if previewed := checkTelegramWorkRunCount(t, operations, 1); previewed.Revision != before.Revision {
		t.Fatal("preview changed native work")
	}

	notice := f.send(t, "/answer "+token+" "+action, 0)
	if !strings.Contains(notice.Text, "Control confirmed:") {
		t.Fatal("confirmed control unavailable", notice)
	}

	after := checkTelegramWorkRunCount(t, operations, 1)
	checkTelegramControlOutcome(t, action, after)

	duplicate := f.send(t, "/answer "+token+" "+action, 0)
	if !strings.Contains(duplicate.Text, "already confirmed") {
		t.Fatal("confirmed action was not consumed", duplicate)
	}

	if replayed := checkTelegramWorkRunCount(t, operations, 1); replayed.Revision != after.Revision || len(replayed.Attempts) != len(after.Attempts) {
		t.Fatal("duplicate control changed native work")
	}
}

func checkTelegramControlOutcome(t *testing.T, action string, run workflowcore.Invocation) {
	t.Helper()

	type outcome struct {
		status, stop     string
		cursor, attempts int
	}

	want := map[string]outcome{
		"pause":     {status: "paused", stop: "pause"},
		"cancel":    {status: "cancelled", stop: "cancel"},
		"resume":    {status: "pending"},
		"continue":  {cursor: 1, attempts: 1},
		"reconcile": {status: "pending"},
	}[action]

	if want.status != "" && run.Status != want.status || run.StopIntent != want.stop || run.Cursor != want.cursor || run.ConfirmedCursor != 0 || len(run.Attempts) != want.attempts {
		t.Fatal("control changed stop intent, ownership or the bounded checkpoint", action, run)
	}
}

// TestTelegramControlNatural sends a model-proposed control through the same
// exact-revision confirmation path as slash commands, without eager execution.
func TestTelegramControlNatural(t *testing.T) {
	telegramControlParallel(t)

	f, operations, ref := telegramWorkAdmitted(t)
	calls := telegramAuthoringIntentHTTP(t, f, func(ConversationIntentContext) string {
		return telegramAuthoringIntentJSON(t, ConversationIntent{OperationID: "control", Arguments: ConversationIntentArguments{Action: "pause", Text: ref.Invocation.ID}})
	})
	before := checkTelegramWorkRunCount(t, operations, 1)
	token := telegramControlOffer(t, f, "Pause the retained work.")

	if previewed := checkTelegramWorkRunCount(t, operations, 1); calls.Load() != 1 || previewed.Revision != before.Revision {
		t.Fatal("intent proposal changed native work or was reinterpreted", calls.Load(), previewed)
	}

	f.send(t, "/answer "+token+" pause", 0)
	checkTelegramControlOutcome(t, "pause", checkTelegramWorkRunCount(t, operations, 1))
}

// TestTelegramControlDraft keeps explicit draft cancellation on its capture
// path, without allocating or controlling a native invocation.
func TestTelegramControlDraft(t *testing.T) {
	telegramControlParallel(t)

	f := newTelegramConversationAppFixture(t)
	f.send(t, "/prepare draft", 0)

	c := f.sessions(t)[0]
	f.send(t, "/cancel "+c.ID, 0)

	if current := f.session(t, c.ID); current.Capture != "" || current.InvocationID != "" {
		t.Fatal("draft cancellation crossed into native work", current)
	}
}

// TestTelegramControlStale preserves a newer cancellation and displays current
// native state rather than applying an older pause confirmation.
func TestTelegramControlStale(t *testing.T) {
	telegramControlParallel(t)

	f, operations, ref := telegramWorkAdmitted(t)
	token := telegramControlOffer(t, f, "/pause "+ref.Invocation.ID)

	stopped, err := operations.StopWorkflow(t.Context(), ref.Invocation.ID, "cancel")
	if err != nil {
		t.Fatal(err)
	}

	notice := f.send(t, "/answer "+token+" pause", 0)
	if !strings.Contains(notice.Text, "unavailable or stale") || !strings.Contains(notice.Text, "Work was canceled") {
		t.Fatal("stale control hid current native state", notice)
	}

	after := checkTelegramWorkRunCount(t, operations, 1)
	if after.Revision != stopped.Revision || after.StopIntent != "cancel" {
		t.Fatal("stale pause replaced cancellation", after)
	}
}

// TestTelegramControlConversation resolves only the conversation's bound run;
// draft cancellation and unqualified receiver status retain their old behavior.
func TestTelegramControlConversation(t *testing.T) {
	telegramControlParallel(t)

	f, operations, ref := telegramWorkAdmitted(t)
	token := telegramControlOffer(t, f, "/pause --conversation "+ref.Request.Target.ConversationID)
	f.send(t, "/answer "+token+" pause", 0)
	checkTelegramControlOutcome(t, "pause", checkTelegramWorkRunCount(t, operations, 1))

	refs, err := f.receiver.store.WorkReferences(t.Context(), f.receiver.botID, "", 16)
	if err != nil || len(refs) != 1 || refs[0].Request.ID != ref.Request.ID {
		t.Fatal("control replaced original native admission authority", refs, err)
	}

	notice := f.send(t, "/status "+ref.Invocation.ID, 0)
	if !strings.Contains(notice.Text, ref.Invocation.ID) || !strings.Contains(notice.Text, "Work is paused") {
		t.Fatal("native status lost the exact binding", notice)
	}

	if receiver := f.send(t, "/status", 0); !strings.Contains(receiver.Text, "Receiver:") {
		t.Fatal("receiver status was replaced", receiver)
	}
}
