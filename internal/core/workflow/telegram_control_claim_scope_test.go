package workflow

import (
	"context"
	"errors"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestTelegramControlClaimScope checks actual credentials through their exact
// owner, a project alias and a foreign repository with a colliding invocation ID.
// The foreign caller cannot validate or mutate the retained owner.
func TestTelegramControlClaimScope(t *testing.T) {
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

	foreign, _, other := workflowClaimFixture(t)

	otherRun := workflowTestStart(t, foreign, other).Invocation
	if otherRun.ID != run.ID {
		t.Fatal("fixture must collide invocation identities")
	}

	for _, test := range []struct {
		name  string
		store *Store
		valid bool
	}{
		{"owner", store, true},
		{"project alias", NewWorkflowStore(store.layout, "project-alias"), true},
		{"foreign root", foreign, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkTelegramControlClaimScope(t, test.store, claim, test.valid)
		})
	}

	current := workflowTestRead(t, store, run.ID)

	unclaimed := workflowTestRead(t, foreign, otherRun.ID)
	if len(current.Attempts) != 1 || current.Cursor != 0 || len(unclaimed.Attempts) != 0 || unclaimed.Revision != otherRun.Revision {
		t.Fatal("scope checks changed native ownership")
	}
}

func checkTelegramControlClaimScope(t *testing.T, store *Store, claim *Claim, valid bool) {
	t.Helper()

	err := store.CheckWorkflowClaim(t.Context(), claim)
	if valid && err != nil || !valid && !errors.Is(err, workflowstate.ErrOwned) {
		t.Fatal("claim scope changed", err)
	}
}
