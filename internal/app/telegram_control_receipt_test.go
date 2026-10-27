package app

import (
	"context"
	"errors"
	"maps"
	"strings"
	"sync/atomic"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type telegramControlLostAck struct {
	service
	lost atomic.Bool
}

func (s *telegramControlLostAck) ApplyTelegramControl(ctx context.Context, target telegramcore.OperationTarget, binding workflowstate.MutationBinding, request WorkflowControlRequest, check core.WorkflowAdmissionCheck) (WorkflowWorkStatus, error) {
	status, err := s.service.ApplyTelegramControl(ctx, target, binding, request, check)
	if err == nil && s.lost.CompareAndSwap(false, true) {
		return status, errors.Join(errTelegramLostRecordAck, workflowstate.ErrConflict)
	}

	return status, err
}

func telegramControlInterruptAck(t *testing.T, f *telegramConversationAppFixture, operations service, token, action string) telegramcore.ConversationInput {
	t.Helper()

	f.receiver.operations = &telegramControlLostAck{service: operations}

	input := telegramOperationRetainInput(t, f, "/answer "+token+" "+action)
	if err := f.receiver.respond(t.Context(), input.Update); !errors.Is(err, errTelegramLostRecordAck) {
		t.Fatal("native acknowledgement was not interrupted", err)
	}

	return input
}

func telegramControlRecoverAck(t *testing.T, f *telegramConversationAppFixture, operations service, id string, input telegramcore.ConversationInput) WorkflowWorkStatus {
	t.Helper()

	if _, err := operations.StopWorkflow(t.Context(), id, "pause"); err != nil {
		t.Fatal(err)
	}

	images := workflowCLIImage(t, operations.projectContext.ProjectDir)
	telegramAuthoringReopen(t, f)

	operations.workflowPullRequests = nil

	f.receiver.operations = operations
	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	if !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("receipt replay changed newer native authority")
	}

	checkTelegramControlRecoveredReceipt(t, f, input.Update, id)

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	return status
}

func checkTelegramControlRecoveredReceipt(t *testing.T, f *telegramConversationAppFixture, update telegramcore.ReceivedUpdate, id string) {
	t.Helper()

	request, found, err := f.receiver.store.Operation(t.Context(), update, 0)
	if err != nil || !found || request.State != "confirmed" || !strings.HasPrefix(request.ReceiptRef, "run/operation-receipts/") {
		t.Fatal("original control receipt was not recovered", request, err)
	}

	c := f.sessions(t)[0]
	if c.InvocationID != id || c.Phase != "waiting" {
		t.Fatal("current stopped state was lost", c)
	}
}
