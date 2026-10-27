package app

import (
	"context"
	"maps"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramControlReceiptRedelivery recovers the original accepted request
// through a new answer after an older diagnostic acknowledged its first source.
func TestTelegramControlReceiptRedelivery(t *testing.T) {
	telegramControlParallel(t)

	for _, kind := range []string{"text", "callback", "reply"} {
		t.Run(kind, func(t *testing.T) {
			f, operations, ref := telegramWorkAdmitted(t)
			token := telegramControlOffer(t, f, "/pause "+ref.Invocation.ID)
			input := telegramControlInterruptAck(t, f, operations, token, "pause")
			staleTelegramControlPrompt(t, f, token, "superseded")

			if err := f.receiver.store.QueueReadResponse(t.Context(), input.Update, "Earlier stale-prompt diagnostic", input.Config.Digest(), time.Now()); err != nil {
				t.Fatal(err)
			}

			f.receiver.operations = operations
			update := telegramControlRetainAnswer(t, f, token, kind)

			images := workflowCLIImage(t, operations.projectContext.ProjectDir)
			if err := f.receiver.respond(t.Context(), update); err != nil {
				t.Fatal(err)
			}

			checkTelegramStaleReceiptResult(t, f, input.Update, true)

			pending, err := f.receiver.store.PendingUpdates(t.Context(), f.receiver.botID)
			if err != nil || len(pending) != 0 || !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
				t.Fatal("receipt redelivery left input pending or repeated its effect", pending, err)
			}
		})
	}
}

// TestTelegramControlReceiptAuthority prevents a completed receipt from exposing
// project facts after the operator mapping or registered root has changed.
func TestTelegramControlReceiptAuthority(t *testing.T) {
	telegramControlParallel(t)

	for _, change := range []string{"operator", "root"} {
		t.Run(change, func(t *testing.T) {
			f, operations, ref := telegramWorkAdmitted(t)
			token := telegramControlOffer(t, f, "/pause "+ref.Invocation.ID)
			input := telegramControlInterruptAck(t, f, operations, token, "pause")
			staleTelegramControlPrompt(t, f, token, "expired")

			images := workflowCLIImage(t, operations.projectContext.ProjectDir)
			telegramWorkChangeActor(t, f, change)

			f.receiver.operations = operations
			if err := f.receiver.respond(t.Context(), input.Update); err != nil {
				t.Fatal(err)
			}

			request, found, err := f.receiver.store.TelegramOperation(t.Context(), input.Update, 0)
			if err != nil || !found || request.State == "confirmed" || request.ReceiptRef != "" {
				t.Fatal("changed authority recovered private native facts", request, err)
			}

			if !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
				t.Fatal("receipt authority rejection changed native work")
			}
		})
	}
}

type telegramControlReplaceAfterApply struct {
	service
	replace func()
}

func (s telegramControlReplaceAfterApply) ApplyTelegramControl(ctx context.Context, target telegramcore.TelegramOperationTarget, binding workflowstate.MutationBinding, request WorkflowControlRequest, check core.WorkflowAdmissionCheck) (WorkflowWorkStatus, error) {
	status, err := s.service.ApplyTelegramControl(ctx, target, binding, request, check)
	if err == nil {
		s.replace()
	}

	return status, err
}

// TestTelegramControlReceiptProjection accepts a native receipt when another
// conversation revision wins before SQLite acknowledgement, preserving that image.
func TestTelegramControlReceiptProjection(t *testing.T) {
	telegramControlParallel(t)

	f, operations, ref := telegramWorkAdmitted(t)
	token := telegramControlOffer(t, f, "/pause "+ref.Invocation.ID)

	var replacement telegramcore.TelegramConversation

	f.receiver.operations = telegramControlReplaceAfterApply{service: operations, replace: func() {
		staleTelegramControlPrompt(t, f, token, "replaced")

		replacement = f.sessions(t)[0]
	}}

	text := f.send(t, "/answer "+token+" pause", 0).Text
	if !strings.Contains(text, "The original control receipt is retained.") || f.session(t, replacement.ID) != replacement {
		t.Fatal("accepted receipt acknowledgement replaced newer peer state", text)
	}

	checkTelegramControlOutcome(t, "pause", checkTelegramWorkRunCount(t, operations, 1))
}
