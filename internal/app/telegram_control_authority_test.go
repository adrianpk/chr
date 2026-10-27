package app

import (
	"context"
	"errors"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramControlAuthorization checks authority at the native effect boundary
// for every lifecycle action, including paths that release or recover ownership.
func TestTelegramControlAuthorization(t *testing.T) {
	telegramControlParallel(t)

	for _, action := range []string{"pause", "cancel", "resume", "continue", "reconcile"} {
		t.Run(action, func(t *testing.T) { checkTelegramControlAuthorization(t, action) })
	}
}

func checkTelegramControlAuthorization(t *testing.T, action string) {
	t.Helper()

	_, operations, ref := telegramWorkAdmitted(t)
	if action == "resume" {
		if _, err := operations.StopWorkflow(t.Context(), ref.Invocation.ID, "pause"); err != nil {
			t.Fatal(err)
		}
	}

	before := checkTelegramWorkRunCount(t, operations, 1)
	checks := 0
	request := WorkflowControlRequest{Action: action, ID: before.ID, ExpectedRevision: before.Revision, StepLimit: 1, AuthorizationCheck: func(context.Context) error {
		checks++

		return workflowstate.ErrConflict
	}}

	status, err := operations.ControlWorkExtent(t.Context(), request)
	if !errors.Is(err, workflowstate.ErrConflict) || checks != 1 || status.Revision != before.Revision {
		t.Fatal("native control did not enforce the final authority check", status, checks, err)
	}

	after := checkTelegramWorkRunCount(t, operations, 1)
	if after.Revision != before.Revision || after.Status != before.Status || len(after.Attempts) != len(before.Attempts) {
		t.Fatal("rejected authority changed native state", after)
	}
}

type telegramControlChangeAuthority struct {
	service
	change func()
}

func (s telegramControlChangeAuthority) ApplyTelegramControl(ctx context.Context, target telegramcore.TelegramOperationTarget, binding workflowstate.MutationBinding, request WorkflowControlRequest, check core.WorkflowAdmissionCheck) (WorkflowWorkStatus, error) {
	return s.service.ApplyTelegramControl(ctx, target, binding, request, func(ctx context.Context) error {
		s.change()

		return check(ctx)
	})
}

// TestTelegramControlAuthority rejects operator and registry-root changes after
// the Telegram claim, while native exclusion is already held.
func TestTelegramControlAuthority(t *testing.T) {
	telegramControlParallel(t)

	for _, change := range []string{"operator", "root"} {
		t.Run(change, func(t *testing.T) {
			f, operations, ref := telegramWorkAdmitted(t)
			token := telegramControlOffer(t, f, "/pause "+ref.Invocation.ID)
			before := checkTelegramWorkRunCount(t, operations, 1)
			checks := 0
			f.receiver.operations = telegramControlChangeAuthority{service: operations, change: func() {
				checks++

				telegramWorkChangeActor(t, f, change)
			}}
			f.send(t, "/answer "+token+" pause", 0)

			if after := checkTelegramWorkRunCount(t, operations, 1); checks != 1 || after.Revision != before.Revision || after.StopIntent != "" {
				t.Fatal("changed authority reached a native effect", checks, after)
			}
		})
	}
}
