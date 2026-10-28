package workflow

import (
	"context"
	"testing"
	"time"
)

// TestTelegramControlStopOwner exercises the worker's actual polling context.
// Native stops and lost credentials cancel it; observing a live owner does not
// release its reservation or create another attempt.
func TestTelegramControlStopOwner(t *testing.T) {
	for _, change := range []string{"none", "pause", "cancel", "credential", "observed"} {
		t.Run(change, func(t *testing.T) { checkTelegramControlStopOwner(t, change) })
	}
}

func checkTelegramControlStopOwner(t *testing.T, change string) {
	t.Helper()

	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request).Invocation
	claim := workflowTestClaim(t, store, run.ID)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := store.CheckpointWorkflow(ctx, claim); err != nil {
			t.Error(err)
		}
	})

	observed := *claim
	if change == "credential" {
		observed.Token += "-invalid"
	}

	if change == "observed" {
		observed = Claim{InvocationID: claim.InvocationID}
	}

	telegramControlOwnerStop(t, store, run.ID, change)

	watched, stop := store.watchWorkflowStop(t.Context(), &observed)
	defer stop()

	waitTelegramControlOwnerStop(t, watched, change)

	retained := workflowTestRead(t, store, run.ID)
	if len(retained.Attempts) != 1 || retained.Cursor != 0 {
		t.Fatal("stop observation changed native work")
	}
}

func telegramControlOwnerStop(t *testing.T, store *Store, id, change string) {
	t.Helper()

	if change == "observed" {
		change = "pause"
	}

	if change == "pause" || change == "cancel" {
		if _, err := store.StopWorkflow(t.Context(), id, change); err != nil {
			t.Fatal(err)
		}
	}
}

func waitTelegramControlOwnerStop(t *testing.T, watched context.Context, change string) {
	t.Helper()

	wait := 200 * time.Millisecond
	if change != "none" {
		wait = 2 * time.Second
	}

	select {
	case <-watched.Done():
		if change == "none" {
			t.Fatal("live claim was cancelled")
		}
	case <-time.After(wait):
		if change != "none" {
			t.Fatal("native stop or lost credential was not observed")
		}
	}
}
