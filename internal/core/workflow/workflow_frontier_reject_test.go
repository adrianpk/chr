package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowFrontierReject rejects changed canonical history, explicit stops
// and altered archived receipt/proof images before the next task can be claimed.
func TestWorkflowFrontierReject(t *testing.T) {
	for _, stage := range []string{"target moved", "paused", "cancelled", "retained receipt", "retained task proof"} {
		t.Run(stage, func(t *testing.T) {
			fixture := workflowDeliveryExtentFixture(t, "rebase", "slice_set")
			workflowFinalizeFrontier(t, fixture)
			invalidateWorkflowFrontier(t, fixture, stage)

			before := workflowTestRead(t, fixture.store, fixture.run.ID)
			if _, err := fixture.store.AdvanceWorkflowFrontier(t.Context(), before.ID, workflowFrontierPorts(fixture)); err == nil {
				t.Fatal("invalid or stopped frontier was accepted")
			}

			if _, err := fixture.store.ClaimWorkflow(t.Context(), before.ID); err == nil {
				t.Fatal("failed frontier admitted another task")
			}

			after := workflowTestRead(t, fixture.store, before.ID)
			if after.Revision != before.Revision || after.ConfirmedCursor != 3 || len(after.Attempts) != 4 {
				t.Fatal("failed frontier mutated confirmed work")
			}
		})
	}
}

func invalidateWorkflowFrontier(t *testing.T, fixture workflowIntegrationFixture, stage string) {
	t.Helper()

	switch stage {
	case "target moved":
		workflowIntegrationCommand(t, fixture.actor, "push", "--force", "origin", fixture.base+":refs/heads/dev")
	case "paused", "cancelled":
		intent := "pause"
		if stage == "cancelled" {
			intent = "cancel"
		}

		if _, err := fixture.store.StopWorkflow(t.Context(), fixture.run.ID, intent); err != nil {
			t.Fatal(err)
		}
	default:
		run := workflowAdvanceFrontier(t, fixture)

		path := run.Deliveries[0].DeliveryRef
		if stage == "retained task proof" {
			path = run.Attempts[0].ResultRef
		}

		if err := os.WriteFile(filepath.Join(fixture.store.layout.Root(), path), []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// TestWorkflowFrontierRecovery interrupts either replacement in the immutable
// frontier/state journal, then proves restart retains exactly one release and
// cannot lose its receipt or resurrect the previous active PR binding.
func TestWorkflowFrontierRecovery(t *testing.T) {
	for _, index := range []int{0, 1} {
		t.Run(string(rune('0'+index)), func(t *testing.T) {
			fixture := workflowDeliveryExtentFixture(t, "rebase", "slice_set")
			workflowFinalizeFrontier(t, fixture)

			fixture.store.afterReplace = func(position int) error {
				if position == index {
					return errors.New("interrupted frontier journal")
				}

				return nil
			}

			if _, err := fixture.store.AdvanceWorkflowFrontier(t.Context(), fixture.run.ID, workflowFrontierPorts(fixture)); err == nil {
				t.Fatal("frontier replacement was not interrupted")
			}

			run := workflowAdvanceFrontier(t, fixture)
			if run.Status != "pending" || run.PullRequest != nil || len(run.Deliveries) != 1 || run.ConfirmedCursor != 3 {
				t.Fatal("frontier recovery lost or duplicated the released unit")
			}
		})
	}
}
