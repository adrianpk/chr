package app

import (
	"maps"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// TestTelegramControlLifecycleAck exercises all five lifecycle controls on a
// real native invocation over HTTP/SQLite. Lost acknowledgements recover their
// original receipt after later stop revisions without repeating work or intent.
func TestTelegramControlLifecycleAck(t *testing.T) {
	telegramControlParallel(t)

	f, operations, ref := telegramWorkAdmitted(t)
	for _, action := range []string{"reconcile", "pause", "resume", "continue", "cancel"} {
		prepareTelegramLifecycleAck(t, operations, ref.Invocation.ID, action)

		f.receiver.operations = operations
		token := telegramControlOffer(t, f, "/"+action+" "+ref.Invocation.ID)
		input := telegramControlInterruptAck(t, f, operations, token, action)
		checkTelegramLifecycleAckEffect(t, operations, ref.Invocation.ID, action)

		if action == "pause" {
			if _, err := operations.ResumeWorkflow(t.Context(), ref.Invocation.ID); err != nil {
				t.Fatal(err)
			}
		}

		if _, err := operations.StopWorkflow(t.Context(), ref.Invocation.ID, "pause"); err != nil {
			t.Fatal(err)
		}

		checkTelegramLifecycleAckRecovery(t, f, operations, input)
		checkTelegramLifecycleConsumed(t, f, operations, ref.Invocation.ID, token, action)
	}
}

func prepareTelegramLifecycleAck(t *testing.T, operations service, id, action string) {
	t.Helper()

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if run.Status == "paused" && action != "resume" {
		if _, err := operations.ResumeWorkflow(t.Context(), id); err != nil {
			t.Fatal(err)
		}
	}

	if action == "resume" && run.Status != "paused" {
		if _, err := operations.StopWorkflow(t.Context(), id, "pause"); err != nil {
			t.Fatal(err)
		}
	}
}

func checkTelegramLifecycleAckEffect(t *testing.T, operations service, id, action string) {
	t.Helper()

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if action == "continue" {
		if run.Cursor != 1 || len(run.Attempts) != 1 {
			t.Fatal("continuation did not execute one checkpoint", run)
		}
	} else if action == "pause" && run.StopIntent != "pause" || action == "cancel" && run.StopIntent != "cancel" {
		t.Fatal("stop intent was not retained", action, run)
	}
}

func checkTelegramLifecycleAckRecovery(t *testing.T, f *telegramConversationAppFixture, operations service, input core.TelegramConversationInput) {
	t.Helper()

	images := workflowCLIImage(t, operations.projectContext.ProjectDir)
	telegramAuthoringReopen(t, f)

	operations.workflowPullRequests = nil

	f.receiver.operations = operations
	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	if !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("recovery changed newer native state")
	}

	request, found, err := f.receiver.store.TelegramOperation(t.Context(), input.Update, 0)
	if err != nil || !found || request.State != "confirmed" || !strings.HasPrefix(request.ReceiptRef, "run/operation-receipts/") {
		t.Fatal("original lifecycle receipt unavailable", request, err)
	}

	telegramWaitNotice(t, f, "The original control receipt is retained.", 3*time.Second)
}

func telegramLifecycleAckStatus(t *testing.T, operations service, id string) WorkflowWorkStatus {
	t.Helper()

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	return status
}

// TestTelegramControlReceiptRetry keeps the production receiver alive when an
// accepted control loses its acknowledgement with a domain conflict. Its next
// retained attempt reads the original receipt instead of repeating the effect.
func TestTelegramControlReceiptRetry(t *testing.T) {
	telegramControlParallel(t)

	f, operations, ref := telegramWorkAdmitted(t)

	before, err := operations.StopWorkflow(t.Context(), ref.Invocation.ID, "pause")
	if err != nil {
		t.Fatal(err)
	}

	token := telegramControlOffer(t, f, "/pause "+ref.Invocation.ID)
	lost := &telegramControlLostAck{service: operations}
	f.receiver.operations = lost
	telegramControlRunReceiver(t, f.receiver, func() {})
	telegramWorkQueueMessage(t, f, "/answer "+token+" pause")
	telegramWaitNotice(t, f, "Control confirmed:", 3*time.Second)

	after, err := operations.ReadWorkflow(t.Context(), ref.Invocation.ID)
	if err != nil || !lost.lost.Load() || after.Revision != before.Revision+1 || after.StopIntent != "pause" {
		t.Fatal("acknowledgement retry stopped intake or repeated acceptance", after, err)
	}

	telegramWorkQueueMessage(t, f, "/status")
	telegramWaitNotice(t, f, "Receiver:", 3*time.Second)
}

func checkTelegramLifecycleConsumed(t *testing.T, f *telegramConversationAppFixture, operations service, id, token, action string) {
	t.Helper()

	before := telegramLifecycleAckStatus(t, operations, id)
	if text := f.send(t, "/answer "+token+" "+action, 0).Text; !strings.Contains(text, "already confirmed") {
		t.Fatal("consumed lifecycle action was repeated", action, text)
	}

	if after := telegramLifecycleAckStatus(t, operations, id); after.Revision != before.Revision {
		t.Fatal("duplicate lifecycle action changed work", action)
	}

	if refs, err := f.receiver.store.TelegramWorkReferences(t.Context(), f.receiver.botID, "", 16); err != nil || len(refs) != 1 {
		t.Fatal("controls changed original admission references", refs, err)
	}
}
