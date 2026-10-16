package app

import (
	"maps"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/tui"
)

// TestTelegramWorkEvidence observes one Telegram-created native checkpoint
// through actual CLI, TUI and Web reads, without converting validation to delivery.
func TestTelegramWorkEvidence(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)

	result := f.receiver.activateWork(t.Context(), ref)
	if result.Failure != nil || result.Progress.Status.ValidatedTasks != 1 || result.Progress.Status.ConfirmedTickets != 0 {
		t.Fatal("native validation fixture unavailable", result.Progress, result.Failure)
	}

	status := result.Progress.Status

	text := telegramWorkProgressText(status)
	if !strings.Contains(text, "No confirmed integration receipt") || !strings.Contains(text, status.Evidence[0].ResultHash) || !strings.Contains(text, "workflow:"+status.ID) {
		t.Fatal("projection lost native evidence or claimed delivery", text)
	}

	checkTelegramWorkSurfaces(t, operations, status)
}

func checkTelegramWorkSurfaces(t *testing.T, operations service, status WorkflowWorkStatus) {
	t.Helper()

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)

	observed, err := workflowCLIStatus(t, operations.projectContext.ProjectDir, "status", status.ID)
	if err != nil || !reflect.DeepEqual(status, observed) {
		t.Fatal("CLI changed the Telegram invocation projection", observed, err)
	}

	executor := newTUIWorkActionExecutor(operations)

	view, err := executor.ExecuteWorkAction(tui.WorkActionRequest{Action: tui.WorkActionWorkflowStatus, Workflow: &tui.WorkflowActionRequest{ID: status.ID}})
	if err != nil || view.Workflow == nil || view.Workflow.ID != status.ID || view.Workflow.Revision != status.Revision {
		t.Fatal("TUI changed the Telegram invocation identity", view, err)
	}

	response := webWorkflowHTTP(t, webWorkflowHandler(t, operations), http.MethodGet, "/workflows?id="+status.ID, nil)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), status.Evidence[0].ResultHash) || !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("Web lost shared evidence or a status read changed native work", response.Code)
	}
}

// TestTelegramWorkDeliveryFailure loses a real send acknowledgement, reopens
// SQLite and retries only that notice. Canonical work bytes and attempts stay fixed.
func TestTelegramWorkDeliveryFailure(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)

	result := f.receiver.activateWork(t.Context(), ref)
	if result.Failure != nil {
		t.Fatal(result.Failure)
	}

	if err := f.receiver.completeWorkActivation(t.Context(), result); err != nil {
		t.Fatal(err)
	}

	before := workflowCLIImage(t, operations.projectContext.ProjectDir)
	f.receiver.api.Client = &telegramDropSendAck{client: f.receiver.api.Client}
	first := telegramWorkDrainNotice(t, f)
	reopenTelegramOperationReceiver(t, f)
	setTelegramOperationRetryAt(t, f.receiver.env, time.Now().Add(-time.Minute))

	retry := telegramWorkDrainNotice(t, f)

	if !strings.Contains(first.Text, ref.Invocation.ID) || !strings.Contains(retry.Text, first.Text) || !strings.Contains(retry.Text, "Uncertain delivery retry") {
		t.Fatal("notification retry lost native identity or uncertainty", first, retry)
	}

	if !maps.Equal(before, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("retrying a notification changed native work")
	}
}

// TestTelegramWorkHumanResult projects an actual checked candidate and open PR
// with its exact human-review binding while keeping confirmed delivery absent.
func TestTelegramWorkHumanResult(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)
	if _, err := operations.ContinueWorkExtent(t.Context(), ref.Invocation.ID, 128); err != nil {
		t.Fatal(err)
	}

	result := f.receiver.activateWork(t.Context(), ref)
	if result.Failure != nil || result.Progress.Attempted || result.Progress.Status.ReviewState != "waiting_approval" {
		t.Fatal("human policy did not retain its exact wait", result.Progress, result.Failure)
	}

	if err := f.receiver.completeWorkActivation(t.Context(), result); err != nil {
		t.Fatal(err)
	}

	status := result.Progress.Status

	notice := telegramWorkDrainNotice(t, f)
	for _, value := range []string{status.CandidateHash, status.ApprovalHash, status.PullRequest.URL, "No confirmed integration receipt"} {
		if value == "" || !strings.Contains(notice.Text, value) {
			t.Fatal("human wait lost evidence or claimed delivery", value, notice)
		}
	}

	checkTelegramWorkSurfaces(t, operations, status)
}

func telegramWorkDrainNotice(t *testing.T, f *telegramConversationAppFixture) telegramConversationNotice {
	t.Helper()

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	select {
	case notice := <-f.notices:
		return notice
	default:
		t.Fatal("native progress notice unavailable")

		return telegramConversationNotice{}
	}
}
