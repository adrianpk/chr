package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramWorkAdmitted(t *testing.T, extents ...string) (*telegramConversationAppFixture, service, telegramcore.TelegramWorkReference) {
	t.Helper()

	extent := "ticket"
	if len(extents) == 1 {
		extent = extents[0]
	}

	f, operations, record := telegramWorkAdmissionFixture(t, extent)

	selector := telegramWorkSelector(t, operations, record, extent)
	action, _ := telegramWorkOffer(t, f, "/work --namespace default "+selector.Extent+" "+selector.Identity)
	f.send(t, "/answer "+action.Token+" start", 0)

	refs, err := f.receiver.store.TelegramWorkReferences(t.Context(), f.receiver.botID, "", 16)
	if err != nil || len(refs) != 1 {
		t.Fatal("confirmed admission was not retained", refs, err)
	}

	return f, operations, refs[0]
}

func telegramWorkAdmissionFixture(t *testing.T, extent string) (*telegramConversationAppFixture, service, core.WorkflowRecord) {
	t.Helper()

	if extent != "ticket" {
		return telegramWorkFixture(t)
	}

	// Ticket control requires one independent target and an unselected grouped
	// task. Multi-extent callers retain the complete six-task fixture plan.
	plan := serviceWorkflowPlanFixture()
	plan.Slices = plan.Slices[:1]
	plan.Slices[0].Tasks = plan.Slices[0].Tasks[:1]

	return telegramWorkFixture(t, plan)
}

func telegramWorkWaitActivation(t *testing.T, h *telegramWorkHost) telegramWorkCompletion {
	t.Helper()

	select {
	case result := <-h.finished:
		if err := h.complete(t.Context(), result); err != nil {
			t.Fatal(err)
		}

		return result
	case <-time.After(45 * time.Second):
		t.Fatal("bounded activation did not finish")

		return telegramWorkCompletion{}
	}
}

// TestTelegramWorkActivation proves an asynchronous call executes exactly one
// common checkpoint and preserves scope, budgets and canonical invocation identity.
func TestTelegramWorkActivation(t *testing.T) {
	for _, extent := range []string{"ticket", "slice", "slice_set", "use_case", "loose"} {
		t.Run(extent, func(t *testing.T) { checkTelegramWorkActivation(t, extent) })
	}
}

func checkTelegramWorkActivation(t *testing.T, extent string) {
	t.Helper()

	f, operations, ref := telegramWorkAdmitted(t, extent)
	t.Setenv(opsnamespace.UserEnv, "elsewhere")

	operations.workflowNamespace = "default"
	h := newTelegramWorkHost(f.receiver)

	if err := h.pump(t.Context()); err != nil {
		t.Fatal(err)
	}

	result := telegramWorkWaitActivation(t, h)
	if result.Failure != nil || !result.Progress.Attempted || result.Progress.Steps != 1 {
		t.Fatal("host did not use one shared checkpoint", result.Progress, result.Failure)
	}

	checkTelegramWorkActivated(t, f, operations, ref, result)
}

func checkTelegramWorkActivated(t *testing.T, f *telegramConversationAppFixture, operations service, ref telegramcore.TelegramWorkReference, result telegramWorkCompletion) {
	t.Helper()

	run := checkTelegramWorkRunCount(t, operations, 1)
	if run.ID != ref.Invocation.ID || run.Cursor != 1 || run.ConfirmedCursor != 0 || len(run.Attempts) != 1 || run.PullRequest != nil {
		t.Fatal("activation replaced or overran admitted work", run)
	}

	c := f.sessions(t)[0]
	if c.LastObservedRevision != result.Progress.Status.Revision || c.InvocationID != ref.Invocation.ID || c.Namespace != "default" {
		t.Fatal("canonical progress lost its local binding", c)
	}
}

// TestTelegramWorkWaits preserves operator stops and exact human-review waits
// without accepting another checkpoint or starting another worker.
func TestTelegramWorkWaits(t *testing.T) {
	for _, wait := range []string{"pause", "cancel", "human_review"} {
		t.Run(wait, func(t *testing.T) {
			f, operations, ref := telegramWorkAdmitted(t)
			if wait == "human_review" {
				if _, err := operations.ContinueWorkExtent(t.Context(), ref.Invocation.ID, 128); err != nil {
					t.Fatal(err)
				}
			} else if _, err := operations.StopWorkflow(t.Context(), ref.Invocation.ID, wait); err != nil {
				t.Fatal(err)
			}

			checkTelegramWorkWait(t, f, operations, ref)
		})
	}
}

func checkTelegramWorkWait(t *testing.T, f *telegramConversationAppFixture, operations service, ref telegramcore.TelegramWorkReference) {
	t.Helper()

	before, err := operations.ReadWorkflow(t.Context(), ref.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	result := f.receiver.activateWork(t.Context(), ref)
	if result.Failure != nil || result.Progress.Attempted || result.Progress.Steps != 0 {
		t.Fatal("host advanced an explicit wait", result.Progress, result.Failure)
	}

	after, err := operations.ReadWorkflow(t.Context(), ref.Invocation.ID)
	if err != nil || after.Revision != before.Revision || after.Policy != before.Policy || len(after.Attempts) != len(before.Attempts) {
		t.Fatal("observing a wait changed native authorization", before, after, err)
	}
}

// TestTelegramWorkOtherOwner retains a live manual owner's process fence and
// descriptor; receiver continuation cannot reconcile it into a new attempt.
func TestTelegramWorkOtherOwner(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)

	claim, err := operations.ClaimWorkflow(t.Context(), ref.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.PrepareWorkflowManualTask(t.Context(), claim, nil); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := operations.CheckpointWorkflow(context.Background(), claim); err != nil {
			t.Error(err)
		}
	})

	result := f.receiver.activateWork(t.Context(), ref)
	if !errors.Is(result.Failure, core.ErrWorkflowOwned) || result.Progress.Steps != 0 {
		t.Fatal("receiver replaced another owner's activation", result.Progress, result.Failure)
	}

	run := checkTelegramWorkRunCount(t, operations, 1)
	if len(run.Attempts) != 1 || run.Cursor != 0 || run.Attempts[0].ID != claim.AttemptID {
		t.Fatal("foreign ownership or progress changed", run)
	}
}

type telegramWorkFinalAuthority struct {
	service
	change func()
}

func (s telegramWorkFinalAuthority) ResumeTelegramWork(ctx context.Context, target telegramcore.TelegramOperationTarget, reference WorkflowRetainedReference, preview WorkflowBoundPreview, check core.WorkflowAdmissionCheck) (RetainedWorkProgress, error) {
	calls := 0
	guard := func(ctx context.Context) error {
		calls++
		if calls == 2 {
			s.change()
		}

		return check(ctx)
	}

	return s.service.ResumeTelegramWork(ctx, target, reference, preview, guard)
}

// TestTelegramWorkActivationAuthority revokes the operator/root at final native
// acceptance, proving the earlier inspection cannot authorize a queued effect.
func TestTelegramWorkActivationAuthority(t *testing.T) {
	for _, change := range []string{"operator", "root"} {
		t.Run(change, func(t *testing.T) {
			f, operations, ref := telegramWorkAdmitted(t)
			before := checkTelegramWorkRunCount(t, operations, 1)
			f.receiver.operations = telegramWorkFinalAuthority{service: operations, change: func() { telegramWorkChangeActor(t, f, change) }}

			result := f.receiver.activateWork(t.Context(), ref)
			if !errors.Is(result.Failure, workflowstate.ErrConflict) || result.Progress.Steps != 0 {
				t.Fatal("changed final authority accepted", result.Progress, result.Failure)
			}

			after := checkTelegramWorkRunCount(t, operations, 1)
			if after.Revision != before.Revision || len(after.Attempts) != 0 {
				t.Fatal("revoked queued activation changed native work", after)
			}
		})
	}
}
