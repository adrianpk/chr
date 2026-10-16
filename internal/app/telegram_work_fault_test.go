package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramWorkLostAck struct {
	service
	lost atomic.Bool
}

func (s *telegramWorkLostAck) StartTelegramWork(ctx context.Context, target core.TelegramOperationTarget, preview WorkflowBoundPreview, binding WorkflowBoundAdmission) (AgentWorkResult, error) {
	result, err := s.service.StartTelegramWork(ctx, target, preview, binding)
	if err == nil && s.lost.CompareAndSwap(false, true) {
		return result, errTelegramLostRecordAck
	}

	return result, err
}

// TestTelegramWorkLostAck reopens SQLite after native admission commits but its
// acknowledgement is lost. Later source/policy edits survive original receipt replay.
func TestTelegramWorkLostAck(t *testing.T) {
	f, operations, record := telegramWorkFixture(t)
	action, _ := telegramWorkOffer(t, f, "/work ticket "+record.Task.Key)
	f.receiver.operations = &telegramWorkLostAck{service: operations}

	input := telegramOperationRetainInput(t, f, "/answer "+action.Token+" start")
	if err := f.receiver.respond(t.Context(), input.Update); !errors.Is(err, errTelegramLostRecordAck) {
		t.Fatal("native acknowledgement was not interrupted", err)
	}

	before := checkTelegramWorkRunCount(t, operations, 1)
	telegramWorkEditSource(t, operations, record)
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\nreview:\n  mode: trusted_auto_merge\n")
	telegramAuthoringReopen(t, f)

	f.receiver.operations = operations
	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	checkTelegramWorkRecovery(t, f, operations, input.Update, before)
}

func checkTelegramWorkRecovery(t *testing.T, f *telegramConversationAppFixture, operations service, update core.TelegramReceivedUpdate, before core.WorkflowInvocation) {
	t.Helper()

	after := checkTelegramWorkRunCount(t, operations, 1)
	if after.ID != before.ID || after.Policy != before.Policy || after.RequestHash != before.RequestHash || len(after.Attempts) != 0 || len(f.notices) != 1 {
		t.Fatal("replay changed native admission or repeated work", before, after)
	}

	checkTelegramWorkRecoveryBinding(t, f, update, before)
}

func checkTelegramWorkRecoveryBinding(t *testing.T, f *telegramConversationAppFixture, update core.TelegramReceivedUpdate, before core.WorkflowInvocation) {
	t.Helper()

	request, found, err := f.receiver.store.TelegramOperation(t.Context(), update, 0)
	if err != nil || !found || request.State != "confirmed" || request.ReceiptRef != "workflow:"+before.ID {
		t.Fatal("original answer was not confirmed", request, err)
	}

	if result := telegramWorkDecodeResult(t, request); result.Invocation.ID != before.ID {
		t.Fatal("receipt names a different invocation", result)
	}

	c := f.sessions(t)[0]
	if c.InvocationID != before.ID || c.LastObservedRevision != before.Revision || c.Phase != "executing" {
		t.Fatal("conversation did not bind original native admission", c)
	}
}

func telegramWorkEditSource(t *testing.T, operations service, record core.WorkflowRecord) {
	t.Helper()

	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: operations.projectContext.ProjectDir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(layout.Root(), record.Path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, append(data, ' '), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestTelegramWorkStalePreview rejects current record, policy or target movement
// without refreshing the displayed extent or allocating any native invocation.
func TestTelegramWorkStalePreview(t *testing.T) {
	for _, change := range []string{"source", "policy", "target"} {
		t.Run(change, func(t *testing.T) {
			f, operations, record := telegramWorkFixture(t)
			action, _ := telegramWorkOffer(t, f, "/work ticket "+record.Task.Key)
			telegramWorkChangePreview(t, operations, record, change)

			text := f.send(t, "/answer "+action.Token+" start", 0).Text
			if !strings.Contains(text, "conflict") {
				t.Fatal("stale preview was refreshed or accepted", text)
			}

			checkTelegramWorkRunCount(t, operations, 0)
		})
	}
}

func telegramWorkChangePreview(t *testing.T, operations service, record core.WorkflowRecord, change string) {
	t.Helper()

	project := operations.projectContext.ProjectDir

	switch change {
	case "source":
		telegramWorkEditSource(t, operations, record)
	case "policy":
		writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\nreview:\n  mode: trusted_auto_merge\n")
	case "target":
		provider, ok := operations.workflowPullRequests.(*serviceWorkEntryProvider)
		if !ok {
			t.Fatal("native fixture provider unavailable")
		}

		serviceWorkflowDeliveryGit(t, project, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-qm", "Later target")
		serviceWorkflowDeliveryGit(t, project, "push", provider.remote, "HEAD:refs/heads/dev")
	}
}

// TestTelegramWorkNamespace retains the preview namespace after environment
// changes and refuses to replace a conversation's already bound invocation.
func TestTelegramWorkNamespace(t *testing.T) {
	f, operations, record := telegramWorkFixture(t)
	action, _ := telegramWorkOffer(t, f, "/work --namespace default ticket "+record.Task.Key)
	t.Setenv(core.OpsUserEnv, "elsewhere")

	operations.workflowNamespace = "default"

	if text := f.send(t, "/answer "+action.Token+" start", 0).Text; !strings.Contains(text, "Native work admitted.") {
		t.Fatal("environment retargeted frozen work", text)
	}

	run := checkTelegramWorkRunCount(t, operations, 1)

	c := f.sessions(t)[0]
	if c.Namespace != "default" || run.Namespace != "default" || c.InvocationID != run.ID {
		t.Fatal("namespace changed between preview and admission", c, run)
	}

	text := f.send(t, "/work --conversation "+c.ID+" loose", 0).Text
	if !strings.Contains(text, "conflict") {
		t.Fatal("new extent replaced the conversation invocation", text)
	}

	checkTelegramWorkRunCount(t, operations, 1)
}
